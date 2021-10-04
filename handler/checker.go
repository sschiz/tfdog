package handler

import (
	tb "gopkg.in/tucnak/telebot.v2"
)

func CheckerMiddleware(f func(m *tb.Message) error) func(m *tb.Message) error {
	return func(m *tb.Message) error {
		if !m.Private() {
			return nil
		}

		if m.Sender.IsBot {
			return nil
		}

		err := f(m)
		if err != nil {
			return err
		}

		return nil
	}
}

func ErrorMiddleware(f func(m *tb.Message) error, b *tb.Bot) func(m *tb.Message) {
	return func(m *tb.Message) {
		err := f(m)
		if err != nil {
			_, _ = b.Send(m.Sender, "error occurred: ", err.Error())
		}
	}
}
