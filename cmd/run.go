package cmd

import (
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"github.com/yukiteruamano/koma/provider/custom"
	"github.com/yukiteruamano/koma/source"
)

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().BoolP("lenient", "l", false, "do not warn about missing functions")
}

var runCmd = &cobra.Command{
	Use:   "run [file]",
	Short: "Run lua file",
	Long: `Runs Lua5.1 VM. Useful for debugging.
Or you can use koma as a standalone lua interpreter.`,
	Args:    cobra.ExactArgs(1),
	Example: "  koma run ./test.lua",
	Run: func(cmd *cobra.Command, args []string) {
		sourcePath := args[0]

		// LoadSource runs file when it's loaded
		src, err := custom.LoadSource(sourcePath, !lo.Must(cmd.Flags().GetBool("lenient")))
		handleErr(err)

		if src != nil {
			source.CloseSource(src)
		}
	},
}
