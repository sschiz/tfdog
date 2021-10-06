package middleware

import (
	"strings"

	tb "gopkg.in/tucnak/telebot.v2"
)

// WithValidator validates incoming updates.
func WithValidator() Middleware {
	return func(upd *tb.Update) bool {
		if upd.Message != nil && !validateMessage(upd.Message) {
			return false
		}

		return true
	}
}

func validateMessage(m *tb.Message) bool {
	condition := m.Private() && !m.Sender.IsBot
	if strings.HasPrefix(m.Text, "/subscribe") {
		condition = condition && len(m.Payload) != 0
	}
	return condition
}
