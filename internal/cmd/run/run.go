package run

import (
	"log"
	"tekton-exporter/internal/globals"
	"tekton-exporter/internal/kubernetes"

	"github.com/spf13/cobra"
)

const (
	descriptionShort = `Execute metrics analyzer`

	descriptionLong = `
	Run execute metrics analyzer`

	LogLevelFlagErrorMessage     = "impossible to get flag --log-level: %s"
	DisableTraceFlagErrorMessage = "impossible to get flag --disable-trace: %s"
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
	// Automatically watched by 'kubernetes' package:
	//cmd.Flags().String("kubeconfig", "", "Path to kubeconfig")

	//
	//cmd.Flags().String("grafana-url", "", "Grafana URL to send requests (https://example.com:8080)")
	//cmd.Flags().String("grafana-auth-token", "", "Grafana API token")

	//cmd.Flags().String("prometheus-url", "", "Prometheus URL to send requests (https://example.com:9090)")
	//cmd.Flags().Bool("use-mimir-endpoint", false, "Use Grafana Mimir endpoint instead of pure Prometheus")
	//cmd.Flags().String("mimir-tenant", "", "Select a tenant on Grafana Mimir")

	//
	//cmd.Flags().Bool("show-not-ingested-metrics", false, "Enable showing used-but-not-ingested metrics")
	//cmd.Flags().Bool("show-not-used-metrics", false, "Enable showing ingested unused metrics")

	//cmd.Flags().Bool("not-ingested-exclude-dashboards", false, "Exclude Grafana dashboards from used-but-not-ingested analysis")
	//cmd.Flags().Bool("not-ingested-exclude-rules", false, "Exclude Prometheus rules from used-but-not-ingested analysis")

	// Set flags conditions
	//cmd.MarkFlagRequired("prometheus-url")
	//cmd.MarkFlagRequired("grafana-url")
	//cmd.MarkFlagRequired("grafana-auth-token")
	//cmd.MarkFlagsRequiredTogether("use-mimir-endpoint", "mimir-tenant")
	//cmd.MarkFlagsOneRequired("show-not-ingested-metrics", "show-not-used-metrics")
	//cmd.MarkFlagsMutuallyExclusive("show-not-ingested-metrics", "show-not-used-metrics")
	//cmd.MarkFlagsMutuallyExclusive("not-ingested-exclude-dashboards", "not-ingested-exclude-rules")

	return cmd
}

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

	client, err := kubernetes.NewClient()
	_, err = kubernetes.GetAllPipelineRuns(globals.ExecContext.Context, client)

	//log.Print(recursos)

	// TODO
	//prometheusUrlFlag, err := cmd.Flags().GetString("prometheus-url")
	//if err != nil {
	//	globals.ExecContext.Logger.Fatalf(PrometheusUrlFlagErrorMessage, err)
	//}

	//useMimirEndpointFlag, err := cmd.Flags().GetBool("use-mimir-endpoint")
	//if err != nil {
	//	globals.ExecContext.Logger.Fatalf(UseMimirEndpointFlagErrorMessage, err)
	//}

	//globals.ExecContext.Logger.Infow("results", zap.Any("data", resultingMetrics))
}
