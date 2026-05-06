package main

import "strings"

const (
	sourceCommandPrefix = "go run ./cmd/assistant"
	binaryCommandPrefix = "icoo"
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
