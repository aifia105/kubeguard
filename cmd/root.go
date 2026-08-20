package cmd

import (
	"context"
	"os"
	"os/signal"

	"github.com/aifia105/kubeguard/pkg/k8sclient"
	"github.com/aifia105/kubeguard/pkg/logger"
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
	metricsclientset "k8s.io/metrics/pkg/client/clientset/versioned"
)

var (
	namespaceFlag string
	outputFlag    string
	k8sClient     *kubernetes.Clientset
	mclientset    *metricsclientset.Clientset
	ctx           context.Context
)

var rootCmd = &cobra.Command{
	Use:   "kubeguard",
	Short: "kubeguard is a Kubernetes security scanner that helps you identify potential security risks and misconfigurations in your Kubernetes cluster.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		k8sClient, mclientset, err = k8sclient.NewK8sClient()
		if err != nil {
			logger.LogFatal("failed to create Kubernetes client: %v", err)
			return err
		}
		return nil
	},
}

func Execute() {
	var cancel context.CancelFunc
	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	go func() {
		<-sigCh
		cancel()
	}()

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&namespaceFlag, "namespace", "n", "", "limit scan to a single namespace (default: whole cluster)")
	rootCmd.PersistentFlags().StringVarP(&outputFlag, "output", "o", "text", "Output format: text or json")
	rootCmd.AddCommand(scanCmd)
	rootCmd.AddCommand(logsCmd)
	rootCmd.AddCommand(auditCmd)
}

func resolveNamespace(args []string) string {
	if namespaceFlag != "" {
		return namespaceFlag
	}
	if len(args) > 0 {
		return args[0]
	}
	return ""
}
