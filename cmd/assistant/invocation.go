package main

import "strings"

const (
	sourceCommandPrefix = "go run ./icoo_runtime/cmd/assistant"
	binaryCommandPrefix = "assistant"
)

func sourceCommand(args string) string {
	args = strings.TrimSpace(args)
	if args == "" {
		return sourceCommandPrefix
	}
	return sourceCommandPrefix + " " + args
}

func binaryCommand(args string) string {
	args = strings.TrimSpace(args)
	if args == "" {
		return binaryCommandPrefix
	}
	return binaryCommandPrefix + " " + args
}
