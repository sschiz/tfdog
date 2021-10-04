package handler

import (
	"go.uber.org/zap"
	tb "gopkg.in/tucnak/telebot.v2"
)

// Stringer is simple handler that just sends text.
func Stringer(bot *tb.Bot, text string) func(*tb.Message) {
	return func(msg *tb.Message) {
		_, err := bot.Send(msg.Sender, text)
		if err != nil {
			zap.L().
				With(zap.Error(err)).
				With(zap.String("text", text)).
				Error("failed to send text")
		}
	}
}
