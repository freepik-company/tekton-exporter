package run

import (
	"context"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"tekton-exporter/internal/globals"
	"tekton-exporter/internal/kubernetes"
	"tekton-exporter/internal/metrics"

	"github.com/spf13/cobra"
)

const (
	descriptionShort = `Execute metrics exporter`

	descriptionLong = `
	Run execute metrics exporter`

	LogLevelFlagErrorMessage        = "impossible to get flag --log-level: %s"
	DisableTraceFlagErrorMessage    = "impossible to get flag --disable-trace: %s"
	MetricsPortFlagErrorMessage     = "impossible to get flag --metrics-port: %s"
	MetricsHostFlagErrorMessage     = "impossible to get flag --metrics-host: %s"
	MetricsWebserverErrorMessage    = "imposible to launch metrics webserver: %s"
	PopulatedLabelsFlagErrorMessage = "impossible to get flag --populated-labels: %s"
	//WatchAllNamespacesFlagErrorMessage = "impossible to get flag --watch-all-namespaces: %s"
	//WatchNamespaceFlagErrorMessage     = "impossible to get flag --watch-namespace: %s"
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "run",
		DisableFlagsInUseLine: true,
		Short:                 descriptionShort,
		Long:                  descriptionLong,

		Run: RunCommand,
	}

	//
	cmd.Flags().String("log-level", "info", "Verbosity level for logs")
	cmd.Flags().Bool("disable-trace", false, "Disable showing traces in logs")

	cmd.Flags().String("metrics-port", "2112", "Port where metrics web-server will run")
	cmd.Flags().String("metrics-host", "0.0.0.0", "Host where metrics web-server will run")

	// This flag is directly defined and used by client-go.
	// It's declared here for documentation purposes only
	cmd.Flags().String("kubeconfig", "~/.kube/config", "Path to the kubeconfig file")

	cmd.Flags().StringSlice("populated-labels", []string{}, "(Repeatable or comma-separated list) Object labels populated on metrics")

	return cmd
}

// RunCommand TODO
// Ref: https://pkg.go.dev/github.com/spf13/pflag#StringSlice
func RunCommand(cmd *cobra.Command, args []string) {

	// Init the logger
	logLevelFlag, err := cmd.Flags().GetString("log-level")
	if err != nil {
		log.Fatalf(LogLevelFlagErrorMessage, err)
	}

	disableTraceFlag, err := cmd.Flags().GetBool("disable-trace")
	if err != nil {
		log.Fatalf(DisableTraceFlagErrorMessage, err)
	}

	err = globals.SetLogger(logLevelFlag, disableTraceFlag)
	if err != nil {
		log.Fatal(err)
	}

	// TODO
	metricsPortFlag, err := cmd.Flags().GetString("metrics-port")
	if err != nil {
		log.Fatalf(MetricsPortFlagErrorMessage, err)
	}

	metricsHostFlag, err := cmd.Flags().GetString("metrics-host")
	if err != nil {
		log.Fatalf(MetricsHostFlagErrorMessage, err)
	}

	populatedLabelsFlag, err := cmd.Flags().GetStringSlice("populated-labels")
	if err != nil {
		log.Fatalf(PopulatedLabelsFlagErrorMessage, err)
	}

	// Handle a potentially confusing situation:
	// Cobra flags' library does not properly parse
	// comma-separated lists depending on the environment
	// the CLI is running (i.e. Kubernetes),
	populatedLabelsFlag = globals.SplitCommaSeparatedValues(populatedLabelsFlag)

	// Store populated labels in context to use them later
	globals.ExecContext.Context = context.WithValue(globals.ExecContext.Context,
		"flag-populated-labels", populatedLabelsFlag)

	// Register metrics into Prometheus Registry
	metrics.RegisterMetrics(populatedLabelsFlag)

	// Create a Kubernetes client for Unstructured resources (CRs)
	client, err := kubernetes.NewClient()

	// Process PipelineRun resources in the background
	// Hey!, errors for watcher must be shown inside the watcher as this is a goroutine
	go func() {
		// Following loop grants re-launching the goroutine when it fails
		for {
			err := kubernetes.WatchPipelineRuns(&globals.ExecContext.Context, client)
			if err != nil {
				globals.ExecContext.Logger.Errorf("error on PipelineRun objects watcher: %s", err)
			}
		}
	}()

	// Process TaskRun resources in the background
	// Hey, errors for watcher must be shown inside the watcher as this is a goroutine
	go func() {
		// Following loop grants re-launching the goroutine when it fails
		for {
			err := kubernetes.WatchTaskRuns(&globals.ExecContext.Context, client)
			if err != nil {
				globals.ExecContext.Logger.Errorf("error on TaskRun objects watcher: %s", err)
			}
		}
	}()

	// Start a webserver for exposing metrics endpoint
	metricsHost := metricsHostFlag + ":" + metricsPortFlag
	http.Handle("/metrics", promhttp.Handler())
	err = http.ListenAndServe(metricsHost, nil)
	if err != nil {
		globals.ExecContext.Logger.Fatalf(MetricsWebserverErrorMessage, err)
	}
}
