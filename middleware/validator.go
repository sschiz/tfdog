package middleware

import (
	"github.com/google/uuid"
	tb "gopkg.in/tucnak/telebot.v2"
)

// WithValidator validates incoming updates.
func WithValidator() Middleware {
	return func(upd *tb.Update) bool {
		if upd.Message != nil && !validateMessage(upd.Message) {
			return false
		}

		if upd.Callback != nil && !validateCallback(upd.Callback) {
			return false
		}

		return true
	}
}

func validateMessage(m *tb.Message) bool {
	return m.Private() && !m.Sender.IsBot
}

func validateCallback(c *tb.Callback) bool {
	_, err := uuid.Parse(c.Data[1:])
	if err != nil {
		return false
	}

	c.Data = c.Data[1:]

	return true
}
