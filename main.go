package main

import (
	"fmt"
	"os"

	"github.com/Guuzzeji/teemysu/internal/ai"
)

func main() {
	c, err := ai.New()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("chat model: %s\nembed model: %s\n", c.ChatModel(), c.EmbedModel())
}