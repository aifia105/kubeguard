package collectors

import (
	"bufio"
	"context"
	"strings"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

func FetchPodLogs(ctx context.Context, clientset *kubernetes.Clientset, namespace, podName, containerName string, tailLines int64) (string, error) {
	opts := &v1.PodLogOptions{
		Container: containerName,
		TailLines: &tailLines,
	}
	req := clientset.CoreV1().Pods(namespace).GetLogs(podName, opts)
	stream, err := req.Stream(ctx)
	if err != nil {

		return "", nil
	}
	defer stream.Close()

	var lines []string
	scanner := bufio.NewScanner(stream)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return strings.Join(lines, "\n"), nil
}
