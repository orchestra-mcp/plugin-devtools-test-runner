package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/orchestra-mcp/plugin-devtools-test-runner/internal"
	"github.com/orchestra-mcp/sdk-go/plugin"
)

func main() {
	builder := plugin.New("devtools.test-runner").
		Version("0.1.0").
		Description("Test runner tools — run, list, and check coverage for Go, Rust, Python, and Node projects").
		Author("Orchestra").
		Binary("devtools-test-runner")

	tp := &internal.ToolsPlugin{}
	tp.RegisterTools(builder)

	p := builder.BuildWithTools()
	p.ParseFlags()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		cancel()
	}()

	if err := p.Run(ctx); err != nil {
		log.Fatalf("devtools.test-runner: %v", err)
	}
}
