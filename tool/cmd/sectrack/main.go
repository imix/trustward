package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"sectrack/internal/mermaid"
	"sectrack/internal/project"
	"sectrack/internal/quarto"
)

func main() {
	root := &cobra.Command{
		Use:   "sectrack",
		Short: "Security model tooling for YAML-based threat models",
	}

	diagramCmd := &cobra.Command{
		Use:   "diagram",
		Short: "Generate diagrams from the security model",
	}
	diagramCmd.AddCommand(&cobra.Command{
		Use:   "dataflow",
		Short: "Generate a Mermaid data flow diagram from system.yaml",
		RunE:  runDataflow,
	})

	reportCmd := &cobra.Command{
		Use:   "report",
		Short: "Generate documents from the security model",
	}
	threatModelCmd := &cobra.Command{
		Use:   "threat-model",
		Short: "Generate a Quarto threat model report from system.yaml and threat-model.yaml",
		RunE:  runThreatModelReport,
	}
	threatModelCmd.Flags().Bool("pdf", false, "include PDF format in the Quarto front matter (requires Chrome headless)")
	reportCmd.AddCommand(threatModelCmd)

	root.AddCommand(diagramCmd, reportCmd)

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

func runDataflow(_ *cobra.Command, _ []string) error {
	proj, err := project.Load(".")
	if err != nil {
		return fmt.Errorf("loading project: %w", err)
	}
	fmt.Print(mermaid.DataFlow(&proj.System))
	return nil
}

func runThreatModelReport(cmd *cobra.Command, _ []string) error {
	proj, err := project.Load(".")
	if err != nil {
		return fmt.Errorf("loading project: %w", err)
	}
	pdf, _ := cmd.Flags().GetBool("pdf")

	sys := proj.System
	meta := quarto.ReportMeta{
		Date:    sys.Version.ReleaseDate,
		Version: sys.Version.Semver,
	}
	if sys.SystemMeta != nil {
		meta.Title = sys.SystemMeta.Title
		meta.Description = sys.SystemMeta.Description
	}
	diagram := mermaid.DataFlow(&sys)

	out, err := quarto.ThreatModel(meta, &proj.ThreatModel, &proj.Company, diagram, pdf)
	if err != nil {
		return fmt.Errorf("rendering report: %w", err)
	}
	fmt.Print(out)
	return nil
}
