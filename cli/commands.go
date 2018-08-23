package cli

import (
	"fmt"
	"os"
	"sort"
	"text/tabwriter"

	"github.com/jinbanglin/cli"
)

func quit(c *cli.Context, args []string) {
	os.Exit(0)
}

func help(c *cli.Context, args []string) {
	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)

	fmt.Fprintln(os.Stdout)

	var keys []string
	for k, _ := range commands {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		cmd := commands[k]
		fmt.Fprintln(w, "\t", cmd.name, "\t\t", cmd.usage)
	}

	w.Flush()
}
