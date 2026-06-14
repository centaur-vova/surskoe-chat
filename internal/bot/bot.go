// Package bot implements the Telegram bot core functionality.
//
// It provides bot setup, long polling subscription, and command dispatching
// for /start and /news handlers.
package bot

import (
	"context"
	"log"

	"github.com/centaur-vova/telegram-go-chat-skeleton/internal/config"
	"github.com/centaur-vova/telegram-go-chat-skeleton/internal/gemini"
	"github.com/mymmrac/telego"
)

// Bot represents a Telegram bot instance with its dependencies.
type Bot struct {
	bot    *telego.Bot
	gemini *gemini.Client
	cfg    config.Config
}

// New creates a new Bot instance. It initializes the Telegram bot and
// Gemini client. If bot creation fails, it logs a fatal error.
func New(cfg config.Config) *Bot {
	// Create bot instance
	bot, err := telego.NewBot(cfg.TelegramToken, telego.WithDefaultDebugLogger())
	if err != nil {
		log.Fatal(err)
	}

	// Gemini client
	client := gemini.New(cfg.GeminiModel, cfg.GeminiKey)

	return &Bot{
		bot:    bot,
		gemini: client,
		cfg:    cfg,
	}
}

// Subscribe starts long polling for Telegram updates and returns a channel
// of updates. The channel is closed when the context is cancelled.
func (b *Bot) Subscribe(ctx context.Context) (<-chan telego.Update, error) {
	// Get updates channel via long polling with context
	params := &telego.GetUpdatesParams{
		Timeout: b.cfg.PollTimeout,
	}
	return b.bot.UpdatesViaLongPolling(ctx, params)
}

// Handle processes incoming updates and dispatches commands to handlers.
// It filters messages from chats other than the configured chat ID.
func (b *Bot) Handle(ctx context.Context, updates <-chan telego.Update) {
	chatID := b.cfg.ChatID
	prompts := b.cfg.Prompts

	for update := range updates {
		if update.Message == nil {
			continue
		}

		// Ignore messages from other chats
		if update.Message.Chat.ID != chatID.ID {
			continue
		}

		text := update.Message.Text

		// /news command — generate a real news article, well, "almost" real haha
		if text == "/news" {
			b.news(ctx, chatID, prompts)
		}

		// /start command — welcome message
		if text == "/start" {
			b.start(ctx, chatID, prompts)
		}
	}
}
