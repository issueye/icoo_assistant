package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
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
	if isCheckRequest(os.Args[1:]) {
		if err := runSelfCheck(os.Stdout, cfg); err != nil {
			log.Fatal(err)
		}
		return
	}
	if isInitRequest(os.Args[1:]) {
		if err := config.GenerateDefaultTOML(root); err != nil {
			log.Fatal(err)
		}
		_, _ = fmt.Fprintf(os.Stdout, "config.toml created at %s\nEdit it and set anthropic.api_key to get started.\n", filepath.Join(root, "config.toml"))
		return
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

func isCheckRequest(args []string) bool {
	if len(args) != 1 {
		return false
	}
	switch strings.TrimSpace(args[0]) {
	case "--check", "check", "doctor":
		return true
	default:
		return false
	}
}

func isInitRequest(args []string) bool {
	if len(args) != 1 {
		return false
	}
	switch strings.TrimSpace(args[0]) {
	case "--init", "init":
		return true
	default:
		return false
	}
}

func printUsage(out io.Writer) {
	_, _ = fmt.Fprintf(out, "icoo_assistant %s\n\n", Version)
	_, _ = fmt.Fprintln(out, "Usage:")
	_, _ = fmt.Fprintf(out, "  %s [query]\n", sourceCommandPrefix)
	_, _ = fmt.Fprintf(out, "  %s\n", sourceCommand("check"))
	_, _ = fmt.Fprintf(out, "  %s\n", sourceCommand("doctor"))
	_, _ = fmt.Fprintf(out, "  %s\n", sourceCommand("init"))
	_, _ = fmt.Fprintf(out, "  %s\n", sourceCommand("--version"))
	_, _ = fmt.Fprintf(out, "  %s\n", sourceCommand("--help"))
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, "Examples:")
	_, _ = fmt.Fprintf(out, "  %s\n", sourceCommand(""))
	_, _ = fmt.Fprintf(out, "  %s\n", sourceCommand("check"))
	_, _ = fmt.Fprintf(out, "  %s\n", sourceCommand(`"read README and summarize the project"`))
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, "Configuration:")
	_, _ = fmt.Fprintln(out, "  Load from config.toml (primary) or .env (fallback) in the current working directory.")
	_, _ = fmt.Fprintf(out, "  Generate defaults with %s.\n", sourceCommand("init"))
	_, _ = fmt.Fprintln(out, "  See .env.example for supported settings.")
	_, _ = fmt.Fprintln(out, "  Environment variables override both file sources.")
	_, _ = fmt.Fprintf(out, "  Replace `%s` with `%s` if the binary is already installed.\n", sourceCommandPrefix, binaryCommandPrefix)
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, "First Use:")
	_, _ = fmt.Fprintf(out, "  Run `%s` before the first real task to verify the workspace and view the minimal happy path.\n", sourceCommand("check"))
	_, _ = fmt.Fprintln(out, "  Recommended first-run path:")
	_, _ = fmt.Fprintf(out, "    1. %s\n", sourceCommand("check"))
	_, _ = fmt.Fprintf(out, "    2. %s\n", sourceCommand(`"先用 tool_catalog 总结当前可用工具，再说明 project_task、task_audit 和 agent_hook_audit 的边界"`))
	_, _ = fmt.Fprintf(out, "    3. %s\n", sourceCommand(`"创建一个项目任务，用于验证后台测试"`))
	_, _ = fmt.Fprintf(out, "    4. %s\n", sourceCommand(`"使用 tool_catalog action=audit_paths 说明审计入口，再给出 task_audit 和 agent_hook_audit 的查询示例"`))
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, "Mode Notes:")
	_, _ = fmt.Fprintln(out, "  Without ANTHROPIC_API_KEY, assistant runs in fake mode for local dry runs and setup validation.")
	_, _ = fmt.Fprintln(out, "  In fake mode, steps 2-4 still work as a dry run, but model-generated answers remain unavailable by design.")
	_, _ = fmt.Fprintln(out, "  With ANTHROPIC_API_KEY, assistant uses the real Anthropic client.")
}
