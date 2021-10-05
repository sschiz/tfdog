package bot

import (
	tb "gopkg.in/tucnak/telebot.v2"
)

func ErrorMiddleware(f func(m *tb.Message) error, b *tb.Bot) func(m *tb.Message) {
	return func(m *tb.Message) {
		err := f(m)
		if err != nil {
			_, _ = b.Send(m.Sender, "error occurred: ", err.Error())
		}
	}
}
