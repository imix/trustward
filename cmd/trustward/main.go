package main

import (
	"errors"
	"fmt"
	"os"
	"text/template"

	"github.com/spf13/cobra"

	"github.com/imix/trustward/internal/mermaid"
	"github.com/imix/trustward/internal/project"
	"github.com/imix/trustward/internal/quarto"
	"github.com/imix/trustward/internal/validate"
)

const reportTmplPath = "report.tmpl"

func main() {
	root := &cobra.Command{
		Use:   "trustward",
		Short: "Security model tooling for YAML-based threat models",
	}

	diagramCmd := &cobra.Command{
		Use:   "diagram",
		Short: "Generate diagrams from the security model",
	}
	diagramCmd.AddCommand(&cobra.Command{
		Use:   "dataflow",
		Short: "Render a Mermaid data flow diagram",
		Args:  cobra.NoArgs,
		RunE:  runDataflow,
	})

	reportCmd := &cobra.Command{
		Use:   "report",
		Short: "Render the Quarto risk-management report",
		Args:  cobra.NoArgs,
		RunE:  runReport,
	}
	reportCmd.Flags().Bool("pdf", false, "include PDF format in the Quarto front matter (requires Chrome headless)")

	templateCmd := &cobra.Command{
		Use:   "template",
		Short: "Manage report templates",
	}
	templateCmd.AddCommand(&cobra.Command{
		Use:   "export",
		Short: "Write the built-in report template to " + reportTmplPath,
		Args:  cobra.NoArgs,
		RunE:  runTemplateExport,
	})

	validateCmd := &cobra.Command{
		Use:          "validate",
		Short:        "Check referential integrity of the security model",
		Args:         cobra.NoArgs,
		RunE:         runValidate,
		SilenceUsage: true,
	}

	root.AddCommand(diagramCmd, reportCmd, templateCmd, validateCmd)

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

func runDataflow(_ *cobra.Command, _ []string) error {
	proj, err := project.Load(".")
	if err != nil {
		return fmt.Errorf("loading project: %w", err)
	}
	fmt.Print(mermaid.DataFlow(proj))
	return nil
}

func runReport(cmd *cobra.Command, _ []string) error {
	proj, err := project.Load(".")
	if err != nil {
		return fmt.Errorf("loading project: %w", err)
	}
	tmpl, err := loadReportTemplate()
	if err != nil {
		return fmt.Errorf("loading template: %w", err)
	}
	pdf, _ := cmd.Flags().GetBool("pdf")
	diagram := mermaid.DataFlow(proj)
	out, err := quarto.Report(proj, tmpl, diagram, pdf)
	if err != nil {
		return fmt.Errorf("rendering report: %w", err)
	}
	fmt.Print(out)
	return nil
}

func runValidate(_ *cobra.Command, _ []string) error {
	proj, err := project.Load(".")
	if err != nil {
		return fmt.Errorf("loading project: %w", err)
	}
	issues := validate.Check(proj)
	for _, issue := range issues {
		fmt.Fprintln(os.Stderr, issue)
	}
	if len(issues) > 0 {
		return fmt.Errorf("%d validation issue(s)", len(issues))
	}
	fmt.Println("model is consistent")
	return nil
}

func runTemplateExport(_ *cobra.Command, _ []string) error {
	if _, err := os.Stat(reportTmplPath); err == nil {
		return fmt.Errorf("%s already exists — delete it first if you want to reset it", reportTmplPath)
	}
	if err := os.WriteFile(reportTmplPath, quarto.DefaultTemplateContent(), 0644); err != nil {
		return err
	}
	fmt.Printf("wrote %s\n", reportTmplPath)
	return nil
}

// loadReportTemplate returns a project-local template if one exists,
// otherwise falls back to the built-in default.
func loadReportTemplate() (*template.Template, error) {
	data, err := os.ReadFile(reportTmplPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			fmt.Fprintln(os.Stderr, "note: using the built-in report template — run 'trustward template export' to customize branding and link out to your system-design docs")
			return quarto.DefaultTemplate(), nil
		}
		return nil, err
	}
	return quarto.ParseTemplate(data)
}
