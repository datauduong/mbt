package cmd

import (
	"errors"

	"github.com/mbtproject/mbt/lib"
	"github.com/spf13/cobra"
)

var (
	out string
)

func init() {
	applyCmd.PersistentFlags().StringVar(&to, "to", "", "template to apply")
	applyCmd.PersistentFlags().StringVar(&out, "out", "", "output path")
	applyCmd.AddCommand(applyBranchCmd)
	applyCmd.AddCommand(applyCommitCmd)
	RootCmd.AddCommand(applyCmd)
}

var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Main command for applying the repository manifest over a template",
	Long: `Main command for applying the repository manifest over a template 

Repository manifest is a data structure created by inspecting .mbt.yml files.
It contains the information about the modules stored within the repository therefore,
can be used for generating artifacts such as deployment scripts.

Apply command transforms the specified go template with the manifest. 

Template must be committed to the repository.
	`,
}

var applyBranchCmd = &cobra.Command{
	Use:   "branch <branch>",
	Short: "Applies the manifest of specified branch over a template",
	Long: `Applies the manifest of specified branch over a template 

Calculated manifest and the template is based on the tip of the specified branch.
	`,
	RunE: func(cmd *cobra.Command, args []string) error {
		branch := "master"
		if len(args) > 0 {
			branch = args[0]
		}

		if to == "" {
			return errors.New("requires the path to template")
		}

		return handle(lib.ApplyBranch(in, to, branch, out))
	},
}

var applyCommitCmd = &cobra.Command{
	Use:   "commit <sha>",
	Short: "Applies the manifest of specified commit over a template",
	Long: `Applies the manifest of specified commit over a template

Calculated manifest and the template is based on the specified commit.

Commit SHA must be the complete 40 character SHA1 string.
	`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("requires the commit sha")
		}

		commit := args[0]

		return handle(lib.ApplyCommit(in, commit, to, out))
	},
}
