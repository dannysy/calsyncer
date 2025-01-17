package cmd

import (
	"calsyncer/internal/config"
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

// Non-blocking async func
type command func(ctx context.Context)

type commandRegistry map[string]command

var commands = commandRegistry{
	"noop": noopCmd,
	"sync": syncCmd,
}

func Run() {
	cmd := config.Gist().String(config.CMD)
	cmdFn, ok := commands[cmd]
	if !ok {
		help()
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	doneCh := make(chan os.Signal, 1)
	signal.Notify(doneCh, os.Interrupt, syscall.SIGTERM)
	cmdFn(ctx)
	<-doneCh
	cancel()
}

func help() {
	fmt.Println("Usage: calsyncer [command]")
	fmt.Println("Commands: noop, sync")
	fmt.Println("Example: calsyncer noop")
	fmt.Println("Config params (name|required|default):\v")
	fmt.Println(config.Sprint())
}
