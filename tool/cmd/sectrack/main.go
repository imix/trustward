package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/spf13/cobra"

	"sectrack/internal/mermaid"
	"sectrack/internal/project"
	"sectrack/internal/quarto"
	"sectrack/internal/validate"
)

const threatModelTmplPath = "templates/threat-model.tmpl"

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
		Short: "Render a Mermaid data flow diagram",
		RunE:  runDataflow,
	})

	reportCmd := &cobra.Command{
		Use:   "report",
		Short: "Generate documents from the security model",
	}
	threatModelCmd := &cobra.Command{
		Use:   "threat-model",
		Short: "Render a Quarto threat model report",
		RunE:  runThreatModelReport,
	}
	threatModelCmd.Flags().Bool("pdf", false, "include PDF format in the Quarto front matter (requires Chrome headless)")
	reportCmd.AddCommand(threatModelCmd)

	templateCmd := &cobra.Command{
		Use:   "template",
		Short: "Manage report templates",
	}
	templateCmd.AddCommand(&cobra.Command{
		Use:   "export threat-model",
		Short: "Write the built-in threat model template to " + threatModelTmplPath,
		RunE:  runTemplateExport,
	})

	validateCmd := &cobra.Command{
		Use:          "validate",
		Short:        "Check referential integrity of the security model",
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

func runThreatModelReport(cmd *cobra.Command, _ []string) error {
	proj, err := project.Load(".")
	if err != nil {
		return fmt.Errorf("loading project: %w", err)
	}
	tmpl, err := loadThreatModelTemplate()
	if err != nil {
		return fmt.Errorf("loading template: %w", err)
	}
	pdf, _ := cmd.Flags().GetBool("pdf")
	diagram := mermaid.DataFlow(proj)
	out, err := quarto.ThreatModel(proj, tmpl, diagram, pdf)
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
	if _, err := os.Stat(threatModelTmplPath); err == nil {
		return fmt.Errorf("%s already exists — delete it first if you want to reset it", threatModelTmplPath)
	}
	if err := os.MkdirAll(filepath.Dir(threatModelTmplPath), 0755); err != nil {
		return err
	}
	if err := os.WriteFile(threatModelTmplPath, quarto.DefaultTemplateContent(), 0644); err != nil {
		return err
	}
	fmt.Printf("wrote %s\n", threatModelTmplPath)
	return nil
}

// loadThreatModelTemplate returns a project-local template if one exists,
// otherwise falls back to the built-in default.
func loadThreatModelTemplate() (*template.Template, error) {
	data, err := os.ReadFile(threatModelTmplPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			fmt.Fprintln(os.Stderr, "note: using the built-in report template — run 'sectrack template export threat-model' to customize branding and link out to your system-design docs")
			return quarto.DefaultTemplate(), nil
		}
		return nil, err
	}
	return quarto.ParseTemplate(data)
}
