package cmd

import (
	"context"
	"errors"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"sort"
	"strings"
)

var (
	ilogsExample = ""

	ilogsLong = ""

	ilogsUse = "ilogs [filter] [-c container] [-F tail lines] [flags]"
)

type IlogsOptions struct {
	configFlags *genericclioptions.ConfigFlags
	clientSet   *kubernetes.Clientset

	filter    string
	container string
	tail      int64

	currentNamespace string
	allNamespaces    bool
	genericclioptions.IOStreams

	selectedPods []corev1.Pod
}

// NewIlogsOptions provides an instance of IlogsOptions with default values
func NewIlogsOptions(streams genericclioptions.IOStreams) *IlogsOptions {
	return &IlogsOptions{
		configFlags: genericclioptions.NewConfigFlags(true),
		IOStreams:   streams,
	}
}

func NewCmdIlogs(streams genericclioptions.IOStreams) *cobra.Command {
	o := NewIlogsOptions(streams)

	cmd := &cobra.Command{
		Use:          ilogsUse,
		Short:        "",
		Long:         ilogsLong,
		Example:      ilogsExample,
		SilenceUsage: true,
		RunE: func(c *cobra.Command, args []string) error {
			if err := o.Complete(c, args); err != nil {
				return err
			}
			if err := o.Validate(); err != nil {
				return err
			}
			if err := o.Run(); err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&o.container, "container", "c", "", "Container name. If omitted, logs from all containers are shown.")
	cmd.Flags().Int64VarP(&o.tail, "tail", "f", 100, "Lines of recent log file to display. Defaults to 50.")
	cmd.Flags().BoolVarP(&o.allNamespaces, "all-namespaces", "A", false, "If present, list the requested object(s) across all namespaces. Namespace in current\ncontext is ignored even if specified with --namespace.")
	o.configFlags.AddFlags(cmd.Flags())
	return cmd
}

func (o *IlogsOptions) Complete(c *cobra.Command, args []string) error {
	if len(args) == 0 {
		return errors.New("pod filter is required")
	}
	restConfig, err := o.configFlags.ToRESTConfig()
	if err != nil {
		return err
	}
	o.clientSet, err = kubernetes.NewForConfig(restConfig)
	if err != nil {
		return err
	}
	if len(args) == 0 || args[0] == "" {
		return errors.New("pod filter is required")
	}
	o.filter = args[0]
	o.currentNamespace, _, err = o.configFlags.ToRawKubeConfigLoader().Namespace()
	if err != nil {
		return err
	}
	if o.allNamespaces {
		o.currentNamespace = ""
	}
	pods, err := o.filterPods()
	if err != nil {
		return err
	}
	o.selectedPods, err = o.selectPod(pods)
	if err != nil {
		return err
	}
	return nil
}

func (o *IlogsOptions) filterPods() ([]corev1.Pod, error) {
	podItems, err := o.clientSet.CoreV1().Pods(o.currentNamespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	if len(podItems.Items) == 0 {
		return nil, errors.New("no pods was found in current namespace")
	}

	pods := make([]corev1.Pod, 0, len(podItems.Items))
	for _, pod := range podItems.Items {
		if strings.Contains(pod.Name, o.filter) {
			pods = append(pods, pod)
		}
	}
	if len(pods) == 0 {
		return nil, errors.New(fmt.Sprintf("no pods was found with filter:  %s", o.filter))
	}

	sort.Slice(pods, func(i, j int) bool {
		return pods[i].CreationTimestamp.After(pods[j].CreationTimestamp.Time)
	})

	return pods, nil
}

func (o *IlogsOptions) selectPod(pods []corev1.Pod) ([]corev1.Pod, error) {
	selectedPods := make([]corev1.Pod, 0, len(pods))

	podsNames := make([]string, 0, len(pods))

	podsNames = append(podsNames, "All")
	for _, pod := range pods {
		podsNames = append(podsNames, pod.GetName())
	}
	// the questions to ask
	var qs = &survey.Select{
		Message: "Select pods:",
		Options: podsNames,
	}

	var selectedPodName string
	if err := survey.AskOne(qs, &selectedPodName); err != nil {
		return nil, err
	}

	if selectedPodName == "All" {
		return pods, nil
	}
	// TODO: multi select
	for _, pod := range pods {
		if pod.GetName() == selectedPodName {
			selectedPods = append(selectedPods, pod)
			return selectedPods, nil
		}
	}
	return selectedPods, nil
}

func (o *IlogsOptions) Validate() error {
	return nil
}

func (o *IlogsOptions) Run() error {
	return nil
}
