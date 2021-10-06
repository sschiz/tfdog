package bot

import (
	"errors"

	"git.sr.ht/~mcldresner/tfdog/service"
	"go.uber.org/zap"
	tb "gopkg.in/tucnak/telebot.v2"
)

type handler struct {
	bot *tb.Bot
	srv service.Service
}

func newHandler(bot *tb.Bot, srv service.Service) *handler {
	return &handler{bot: bot, srv: srv}
}

func (h *handler) Subscribe(m *tb.Message) {
	logger := zap.L().
		Named("handler").
		With(zap.String("command", "subscribe"))

	payload := NewBetaPayload(h.bot, m.Sender)
	err := h.srv.Subscribe(m.Sender.ID, m.Payload, payload)
	if err != nil {
		if errors.Is(err, service.ErrAlreadySubscribed) {
			_, err = h.bot.Send(m.Sender, "You have already subscribed this beta.")
			if err != nil {
				logger.With(zap.Error(err)).Error("failed to send message")
				return
			}
		}
		return
	}

	text := "⚡️ You have been subscribed the beta"
	_, err = h.bot.Send(
		m.Sender,
		text,
		tb.NoPreview,
		tb.ModeMarkdown,
	)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to send message")
		return
	}
}

func (h *handler) Unsubscribe(m *tb.Message) {
	logger := zap.L().
		Named("handler").
		With(zap.String("command", "unsubscribe"))

	keyboard, err := h.generateDeletionKeyboard(m.Sender.ID)
	if err != nil {
		return
	}

	_, err = h.bot.Send(m.Sender, "List of subscriptions:", keyboard)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to send message")
		return
	}
}

func (h *handler) UnsubscribeInline(c *tb.Callback) {
	logger := zap.L().
		Named("handler").
		With(zap.String("command", "unsubscribe_inline"))

	resp := &tb.CallbackResponse{
		CallbackID: c.ID,
		ShowAlert:  true,
		Text:       "Something went wrong",
	}
	defer func(bot *tb.Bot, c *tb.Callback, resp *tb.CallbackResponse) {
		err := bot.Respond(c, resp)
		if err != nil {
			logger.With(zap.Error(err)).Error("failed to respond")
		}
	}(h.bot, c, resp)

	err := h.srv.Unsubscribe(c.Sender.ID, c.Data)
	if err != nil {
		return
	}
	resp.Text = "Successfully unsubscribed"
	resp.ShowAlert = false

	keyboard, err := h.generateDeletionKeyboard(c.Sender.ID)
	if err != nil {
		return
	}

	_, err = h.bot.EditReplyMarkup(c.Message, keyboard)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to edit reply markup")
		return
	}
}

func (h *handler) generateDeletionKeyboard(userID int) (*tb.ReplyMarkup, error) {
	subs, err := h.srv.GetUserSubscriptions(userID)
	if err != nil {
		return nil, err
	}

	keyboard := generateReplyMarkup(subs)
	return keyboard, nil
}

func generateReplyMarkup(subs []service.Subscription) *tb.ReplyMarkup {
	selector := new(tb.ReplyMarkup)
	rows := make([]tb.Row, len(subs))
	for i, val := range subs {
		rows[i] = selector.Row(
			selector.Data(
				val.AppName,
				"",
				val.Link,
			),
		)
	}

	selector.Inline(rows...)
	return selector
}
