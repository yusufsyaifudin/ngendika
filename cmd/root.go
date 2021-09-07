package cmd

import (
	"fmt"
	"os"

	"github.com/yusufsyaifudin/ngendika/cmd/api"
	"github.com/yusufsyaifudin/ngendika/cmd/migrate"
	"github.com/yusufsyaifudin/ngendika/cmd/worker"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

var _rootCmd = &cobra.Command{
	Use:   "ngendika",
	Short: "ngendika application",
}

// init register all sub-command. This way, we can generate docs for all sub-command.
func init() {
	_rootCmd.AddCommand(api.Execute())
	_rootCmd.AddCommand(migrate.Execute())
	_rootCmd.AddCommand(worker.Execute())
	_rootCmd.AddCommand(GenerateDocCmd())
}

// Execute will execute the command
func Execute() {
	_rootCmd.PersistentFlags().StringP("config", "c", "", "Config file")

	if err := _rootCmd.Execute(); err != nil {
		err = fmt.Errorf("error running program: %w", err)
		_, _ = fmt.Fprintf(os.Stderr, err.Error())
	}
}

// GenerateDoc generate CLI docs to specific directory.
func GenerateDoc(dir string) {
	if err := doc.GenMarkdownTree(_rootCmd, dir); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%+v", err)
		os.Exit(1)
	}
	fmt.Println("All files have been generated and updated.")
}

func GenerateDocCmd() *cobra.Command {
	var docCmd = &cobra.Command{
		Use:   "docs",
		Short: "Generate documentation",
		Long:  "",
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				err = fmt.Errorf("error get current working directory: %w", err)
				return err
			}

			dir := cwd + "/docs"
			err = os.MkdirAll(dir, os.ModePerm)
			if err != nil {
				err = fmt.Errorf("error create directory: %w", err)
				return err
			}

			GenerateDoc(dir)
			return nil
		},
	}

	return docCmd
}
