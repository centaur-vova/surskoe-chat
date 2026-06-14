// Command bot is a Telegram bot for local community chat.
//
// It listens for /start and /news commands, generates localized news
// using Gemini AI, and handles graceful shutdown on interrupt.
package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/centaur-vova/telegram-go-chat-skeleton/internal/bot"
	"github.com/centaur-vova/telegram-go-chat-skeleton/internal/config"
	"github.com/centaur-vova/telegram-go-chat-skeleton/internal/logger"
)

func main() {
	cfg := config.Load()

	logger.Init(cfg.LogLevel)

	// Context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Bot
	bot := bot.New(cfg)

	// Subscribe for updates
	updates, err := bot.Subscribe(ctx)
	if err != nil {
		log.Fatal(err)
	}

	logger.Info("StarPom of the Spacefleet is online! Listening for commands in chat…")

	// Run bot
	go bot.Handle(ctx, updates)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	// Damn, I love Go, ahaha.

	logger.Info("Shutting down…")
	cancel()
}
