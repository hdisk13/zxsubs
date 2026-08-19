package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/hdisk13/zxsubs/internal/azure"
	"github.com/hdisk13/zxsubs/internal/picker"
)

type azClient interface {
	List() ([]azure.Subscription, error)
	Set(id string) error
}

type pickFunc func([]azure.Subscription) (azure.Subscription, bool, error)

func main() {
	os.Exit(run(os.Args[1:], azure.New(), os.Stdout, os.Stderr, picker.Run))
}

func run(args []string, cli azClient, stdout, stderr io.Writer, pick pickFunc) int {
	fs := flag.NewFlagSet("zxsubs", flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.Usage = func() {
		fmt.Fprint(stderr, `zxsubs — pick the active Azure CLI subscription.

Usage:
  zxsubs

Requires Azure CLI on PATH and an existing login (az login).
Opens a full-terminal picker: arrow keys move, type to filter,
Enter selects, q or Esc cancels.

Exit codes:
  0  subscription was set
  1  canceled, error, or az is missing / not logged in
`)
	}
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 1
	}
	if fs.NArg() > 0 {
		fmt.Fprintln(stderr, "zxsubs takes no arguments")
		fs.Usage()
		return 1
	}

	subs, err := cli.List()
	if err != nil {
		fmt.Fprintf(stderr, "zxsubs: %v\n", err)
		return 1
	}

	chosen, ok, err := pick(subs)
	if err != nil {
		fmt.Fprintf(stderr, "zxsubs: %v\n", err)
		return 1
	}
	if !ok {
		fmt.Fprintln(stderr, "Canceled.")
		return 1
	}

	if err := cli.Set(chosen.ID); err != nil {
		fmt.Fprintf(stderr, "zxsubs: %v\n", err)
		return 1
	}

	fmt.Fprintf(stdout, "Active subscription: %s (%s)\n", chosen.Name, chosen.ID)
	return 0
}
