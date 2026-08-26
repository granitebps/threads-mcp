package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/granitebps/threads-mcp/internal/config"
	provideradapter "github.com/granitebps/threads-mcp/internal/provider/threadscli"
	"github.com/granitebps/threads-mcp/internal/server"
	"github.com/granitebps/threads-mcp/internal/version"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load(version.Version)
	if err != nil {
		fmt.Fprintln(os.Stderr, "threads-mcp:", err)
		os.Exit(1)
	}

	providerClient, err := provideradapter.New(cfg, version.Version)
	if err != nil {
		fmt.Fprintln(os.Stderr, "threads-mcp:", err)
		os.Exit(1)
	}

	mcpServer := server.New(providerClient, cfg, server.BuildInfo{
		Version:    version.Version,
		Commit:     version.Commit,
		Date:       version.Date,
		SDKVersion: "v1.7.0",
		Protocol:   "2026-07-28",
	})
	if err := mcpServer.Run(ctx, &mcp.StdioTransport{}); err != nil && !errors.Is(err, context.Canceled) {
		fmt.Fprintln(os.Stderr, "threads-mcp:", err)
		os.Exit(1)
	}
}
