package handler

import tb "gopkg.in/tucnak/telebot.v2"

// Stringer is simple handler that just sends text.
func Stringer(bot *tb.Bot, text string) func(*tb.Message) {
	return func(msg *tb.Message) {
		_, _ = bot.Send(msg.Sender, text)
	}
}
