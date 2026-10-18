package main

import (
	"context"
	"fmt"
	"github.com/cskwork/scopeweaver-cli/internal/cli"
	"github.com/cskwork/scopeweaver-cli/internal/portable"
	"github.com/cskwork/scopeweaver-cli/server"
	"os"
	"os/signal"
	"syscall"
)

var version = "0.1.0"

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	cli.Version = version
	server.BuildVersion = version
	if len(os.Args) > 1 && os.Args[1] == "__runtime" {
		if len(os.Args) != 4 || os.Args[2] != "--home" {
			fmt.Fprintln(os.Stderr, "invalid private runtime invocation")
			os.Exit(2)
		}
		if err := portable.Run(ctx, os.Args[3]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(3)
		}
		return
	}
	os.Exit(cli.New().Run(ctx, os.Args[1:]))
}
