package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"icoo_assistant/internal/config"
)

func main() {
	if isHelpRequest(os.Args[1:]) {
		printUsage(os.Stdout)
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
	application, err := newApp(cfg)
	if err != nil {
		log.Fatal(err)
	}
	if query := strings.TrimSpace(strings.Join(os.Args[1:], " ")); query != "" {
		if err := application.runOnce(os.Stdout, query); err != nil {
			log.Fatal(err)
		}
		return
	}
	if err := application.runREPL(os.Stdin, os.Stdout); err != nil {
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

func printUsage(out io.Writer) {
	_, _ = fmt.Fprintf(out, "icoo_assistant %s\n\n", Version)
	_, _ = fmt.Fprintln(out, "Usage:")
	_, _ = fmt.Fprintln(out, "  assistant [query]")
	_, _ = fmt.Fprintln(out, "  assistant --version")
	_, _ = fmt.Fprintln(out, "  assistant --help")
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, "Examples:")
	_, _ = fmt.Fprintln(out, "  assistant")
	_, _ = fmt.Fprintln(out, "  assistant \"read README and summarize the project\"")
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, "Configuration:")
	_, _ = fmt.Fprintln(out, "  Load environment variables from .env in the current working directory.")
	_, _ = fmt.Fprintln(out, "  See .env.example for supported settings.")
}
