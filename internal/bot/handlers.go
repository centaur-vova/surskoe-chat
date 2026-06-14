// Package bot implements Telegram bot handlers for local community chat.
//
// It provides functionality for news generation using Gemini AI and
// basic command handling (/start, /news).
package bot

import (
	"context"

	"github.com/centaur-vova/telegram-go-chat-skeleton/internal/config"
	"github.com/centaur-vova/telegram-go-chat-skeleton/internal/logger"
	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoutil"
)

// news handles the /news command. It sends a "recording" message, generates
// a localized news article via Gemini AI, deletes the temporary message,
// and sends the generated news to the chat.
//
// The prompt asks Gemini to recall a real positive event from rural Russia
// and adapt it to the local context of Surskoye.
func (b *Bot) news(ctx context.Context, chatID telego.ChatID, prompts *config.PromptsConfig) {
	msg, _ := b.bot.SendMessage(ctx, telegoutil.Message(chatID, prompts.Messages.Recording))

	userPrompt := `Вспомни одно реальное позитивное событие из жизни любой российской сельской глубинки за последние два года (например: благоустройство сквера, ремонт дороги, открытие почты, победа на районных соревнованиях, закупка тракторов). Перенеси это событие в р.п. Сурское и перепиши строго по инструкции`
	news, err := b.gemini.Ask(prompts.News.System, userPrompt, false)
	if err != nil {
		logger.Error("Gemini error", "error", err)
		_, err = b.bot.SendMessage(ctx, telegoutil.Message(chatID, prompts.Messages.Error))
		if err != nil {
			logger.Error("Error sending message", "error", err)
		}
		return
	}

	// Delete "recording..." message and send the news
	if err = b.bot.DeleteMessage(ctx, telegoutil.Delete(chatID, msg.MessageID)); err != nil {
		logger.Error("Error deleting bot message", "error", err)
	}
	_, err = b.bot.SendMessage(ctx, telegoutil.Message(chatID, "📡 "+news))
	if err != nil {
		logger.Error("Error sending message", "error", err)
	}
}

// start handles the /start command. It sends a welcome message to the chat.
func (b *Bot) start(ctx context.Context, chatID telego.ChatID, prompts *config.PromptsConfig) {
	_, err := b.bot.SendMessage(ctx, telegoutil.Message(chatID, prompts.Messages.Welcome))
	if err != nil {
		logger.Error("Error sending message", "error", err)
	}
}
