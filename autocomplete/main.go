package main

import (
	"autocomplete/ollamastream"
	"context"
	"fmt"
	"log"
	"time"
)

func generate(prompt string) {
	const ollamaEndpoint = "http://localhost:11434/api/generate"
	var fullText string

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := ollamastream.GenerateStream(
		ctx,
		prompt,
		ollamaEndpoint,
		"deepseek-r1",
		0.7,
		80,
		func(token string) {
			fullText += token
			fmt.Print(token)
		},
	)
	if err != nil {
		log.Fatalf("GenerateStream error: %v", err)
	}

	fmt.Printf("\n\n---\nFull text:\n%s\n", fullText)
}

func main() {
	generate("Write a short poem about autumn leaves.")
}
