package randemoter

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/storskegg/randemoter/emojilib"
)

var num int

var cmdRoot = &cobra.Command{
	Use:   "randemoter",
	Short: "Produces a given number of cryptographically random emojis",
	Long:  "Annoy or communicate with your kids with whatever emojis this thing spits out!",
	//Example:                    "",
	ArgAliases: nil,
	Version:    "0.0.1",
	Run:        execRoot,
}

func execRoot(cmd *cobra.Command, args []string) {
	sb := strings.Builder{}
	for i := 0; i < num; i++ {
		sb.WriteRune(emojilib.RandEmoji())
	}
	fmt.Println(sb.String())
}

func registerFlags() {
	cmdRoot.Flags().IntVarP(&num, "num", "n", 10, "number of emojis to generate")
}

func Execute() error {
	registerFlags()
	return cmdRoot.Execute()
}
