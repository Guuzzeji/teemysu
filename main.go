package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Guuzzeji/teemysu/internal/ai"
	"github.com/Guuzzeji/teemysu/internal/db"
	"github.com/Guuzzeji/teemysu/internal/discord"
	"github.com/joho/godotenv"
)

var logger = log.New(os.Stderr, "", log.LstdFlags|log.Lmicroseconds)

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func main() {
	_ = godotenv.Load()

	token := os.Getenv("DISCORD_BOT_TOKEN")
	if token == "" {
		fmt.Fprintln(os.Stderr, "DISCORD_BOT_TOKEN not set")
		os.Exit(1)
	}

	aiClient, err := ai.New()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	store, err := db.New(envOr("DATABASE_PATH", "data.db"))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer store.Close()
	logger.Printf("ai ready: chat=%s embed=%s", aiClient.ChatModel(), aiClient.EmbedModel())
	logger.Printf("db ready: %s", envOr("DATABASE_PATH", "data.db"))

	bot, err := discord.New(token, store, aiClient)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer bot.Close()

	if err := bot.Open(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	logger.Printf("bot online. press Ctrl-C to stop.")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	logger.Printf("shutting down.")
}
