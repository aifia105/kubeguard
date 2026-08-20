package cmd

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/aifia105/kubeguard/pkg/logger"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
)

var (
	containerFlag string
	tailFlag      int64
	followFlag    bool
)

var logsCmd = &cobra.Command{
	Use:   "logs <pod-name> [namespace]",
	Short: "Fetch logs from a pod",
	Args:  cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		podName := args[0]
		ns := namespaceFlag

		if strings.Contains(podName, "/") {
			parts := strings.SplitN(podName, "/", 2)
			ns, podName = parts[0], parts[1]
		} else if ns == "" && len(args) > 1 {
			ns = args[1]
		}

		nsExplicit := ns != ""

		if !nsExplicit {
			foundNs, err := findPodNamespace(podName)
			if err != nil {
				logger.LogFatal("failed to search for pod %s across namespaces: %v", podName, err)
			}
			if foundNs == "" {
				logger.LogFatal("pod %q not found in any namespace", podName)
			}
			if foundNs != "default" {
				logger.LogInfo("no namespace specified, found pod %q in namespace %q", podName, foundNs)
			}
			ns = foundNs
		}

		opts := &v1.PodLogOptions{
			Follow:     followFlag,
			Timestamps: true,
		}
		if containerFlag != "" {
			opts.Container = containerFlag
		}
		if tailFlag > 0 {
			opts.TailLines = &tailFlag
		}

		req := k8sClient.CoreV1().Pods(ns).GetLogs(podName, opts)
		stream, err := req.Stream(ctx)
		if err != nil {
			logger.LogFatal("failed to get logs for pod %s/%s: %v", ns, podName, err)
		}
		defer stream.Close()

		scanner := bufio.NewScanner(stream)
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}

	},
}

func findPodNamespace(podName string) (string, error) {
	pods, err := k8sClient.CoreV1().Pods("").List(ctx, metav1.ListOptions{
		FieldSelector: fields.OneTermEqualSelector("metadata.name", podName).String(),
	})
	if err != nil {
		return "", err
	}
	if len(pods.Items) == 0 {
		return "", nil
	}
	if len(pods.Items) > 1 {
		var namespaces []string
		for _, pod := range pods.Items {
			namespaces = append(namespaces, pod.Namespace)
		}
		return "", fmt.Errorf("pod %q found in multiple namespaces: %s", podName, strings.Join(namespaces, ", "))
	}
	return pods.Items[0].Namespace, nil
}

func init() {
	logsCmd.Flags().StringVarP(&containerFlag, "container", "c", "", "Container name (if multiple containers in pod)")
	logsCmd.Flags().Int64VarP(&tailFlag, "tail", "t", 10, "Number of lines to show from the end of the logs")
	logsCmd.Flags().BoolVarP(&followFlag, "follow", "f", false, "Follow the logs output")
}
