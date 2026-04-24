package main

import (
	"log"
	"os"
	"strings"

	"icoo_assistant/internal/config"
)

func main() {
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
