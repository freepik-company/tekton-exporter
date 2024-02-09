package cmd

import (
	"tekton-exporter/internal/cmd/run"
	"tekton-exporter/internal/cmd/version"

	"github.com/spf13/cobra"
)

const (
	descriptionShort = `Tekton Prometheus exporter`

	// descriptionLong TODO
	descriptionLong = `
	Tekton Exporter is a simple Prometheus exporter.
	It exposes non standard (but useful) metrics`
)

func NewMetricosoCommand(name string) *cobra.Command {
	c := &cobra.Command{
		Use:   name,
		Short: descriptionShort,
		Long:  descriptionLong,
	}

	c.AddCommand(
		version.NewCommand(),
		run.NewCommand(),
	)

	return c
}
