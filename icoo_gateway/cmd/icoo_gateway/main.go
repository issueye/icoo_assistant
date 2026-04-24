package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"icoo_gateway/internal/api"
	"icoo_gateway/internal/config"
	"icoo_gateway/internal/server"
)

const Version = "0.0.1"

func main() {
	if isHelpRequest(os.Args[1:]) {
		printUsage()
		return
	}
	if isVersionRequest(os.Args[1:]) {
		_, _ = os.Stdout.WriteString(Version + "\n")
		return
	}

	root, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	cfg, err := config.Load(root)
	if err != nil {
		log.Fatal(err)
	}

	handler := api.NewMux(api.NewApp())
	srv := server.New(cfg, handler)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(stop)

	go func() {
		log.Printf("icoo_gateway listening on http://%s", cfg.Addr())
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	<-stop
	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}
}

func isVersionRequest(args []string) bool {
	if len(args) != 1 {
		return false
	}
	switch strings.TrimSpace(args[0]) {
	case "--version", "-version", "version":
		return true
	default:
		return false
	}
}

func isHelpRequest(args []string) bool {
	if len(args) != 1 {
		return false
	}
	switch strings.TrimSpace(args[0]) {
	case "--help", "-h", "help":
		return true
	default:
		return false
	}
}

func printUsage() {
	_, _ = fmt.Fprintf(os.Stdout, "icoo_gateway %s\n\n", Version)
	_, _ = fmt.Fprintln(os.Stdout, "Usage:")
	_, _ = fmt.Fprintln(os.Stdout, "  go run ./cmd/icoo_gateway")
	_, _ = fmt.Fprintln(os.Stdout, "  go run ./cmd/icoo_gateway --help")
	_, _ = fmt.Fprintln(os.Stdout, "  go run ./cmd/icoo_gateway --version")
	_, _ = fmt.Fprintln(os.Stdout)
	_, _ = fmt.Fprintln(os.Stdout, "Configuration:")
	_, _ = fmt.Fprintln(os.Stdout, "  Load environment variables from .env in the current working directory.")
	_, _ = fmt.Fprintln(os.Stdout, "  See .env.example for supported settings.")
}
