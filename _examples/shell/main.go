package main

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"
	"github.com/posener/complete"
	"github.com/willabides/kongplete"
)

var cli struct {
	Debug bool `help:"Debug mode."`

	Rm struct {
		User      string `help:"Run as user." short:"u" default:"default"`
		Force     bool   `help:"Force removal." short:"f"`
		Recursive bool   `help:"Recursively remove files." short:"r"`

		Paths []string `arg:"" help:"Paths to remove." type:"path" name:"path" predictor:"file"`
	} `cmd:"" help:"Remove files."`

	Ls struct {
		Paths []string `arg:"" optional:"" help:"Paths to list." type:"path" predictor:"file"`
	} `cmd:"" help:"List paths."`

	InstallCompletions kongplete.InstallCompletions `cmd:"" help:"install shell completions"`
}

func main() {
	parser := kong.Must(&cli,
		kong.Name("shell"),
		kong.Description("A shell-like example app."),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
			Summary: true,
		}))
	kongplete.Complete(parser,
		kongplete.WithPredictor("file", complete.PredictFiles("*")),
	)
	ctx, err := parser.Parse(os.Args[1:])
	parser.FatalIfErrorf(err)
	switch ctx.Command() {
	case "rm <path>":
		fmt.Println(cli.Rm.Paths, cli.Rm.Force, cli.Rm.Recursive)

	case "ls":
	}
}
