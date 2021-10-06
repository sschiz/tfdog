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
	if !m.Private() || m.Sender.IsBot {
		return false
	}

	if !strings.HasPrefix(m.Text, "/subscribe") {
		return true
	}

	if len(m.Entities) != 2 {
		return false
	}

	entity := m.Entities[1]
	if entity.Type != tb.EntityURL {
		return false
	}

	return true
}
