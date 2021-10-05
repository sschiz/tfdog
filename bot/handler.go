package bot

import (
	"fmt"

	"git.sr.ht/~mcldresner/tfdog/scheduler"
	"git.sr.ht/~mcldresner/tfdog/tfbeta"
	"github.com/google/uuid"
	"go.uber.org/zap"
	tb "gopkg.in/tucnak/telebot.v2"
)

type handler struct {
	sch  *scheduler.Scheduler
	bot  *tb.Bot
	subs *subscriptions
}

func newHandler(s *scheduler.Scheduler, b *tb.Bot) *handler {
	return &handler{sch: s, bot: b, subs: newSubscriptions()}
}

func (h *handler) Subscribe(m *tb.Message) error {
	if len(m.Payload) == 0 {
		_, err := h.bot.Send(
			m.Sender,
			"Please, specify TestFlight link.\n"+
				"Example:\n`/subscribe https://testflight.apple.com/join/u6iogfd0`",
			tb.ModeMarkdown,
		)

		if err != nil {
			zap.L().With(zap.Error(err)).Error("failed to send message")
			return err
		}

		return nil
	}

	beta, err := tfbeta.NewTFBeta(m.Payload)
	if err != nil {
		zap.L().With(zap.Error(err)).Error("failed to create beta")
		return err
	}
	name := beta.GetAppName()

	bot := h.bot
	sender := tb.ChatID(m.Chat.ID)
	id, err := h.sch.Schedule(func() {
		isFull, err := beta.IsFull()
		if err != nil {
			_, _ = bot.Send(sender, "error occurred: "+err.Error())
			zap.L().With(zap.Error(err)).Error("failed to check beta")
			return
		}

		if !isFull {
			_, err = bot.Send(
				sender,
				fmt.Sprintf(
					"✅ [%s](%s) beta has free slots!",
					beta.GetAppName(),
					beta.GetLink(),
				),
				tb.NoPreview,
				tb.ModeMarkdown,
			)
			if err != nil {
				zap.L().With(zap.Error(err)).Error("failed to send message")
				return
			}
		}
	})
	if err != nil {
		return err
	}

	h.subs.Add(m.Sender.ID, id, name)
	_, err = h.bot.Send(
		sender,
		fmt.Sprintf(
			"⚡️ You have been subscribed [%s](%s) beta",
			beta.GetAppName(),
			beta.GetLink(),
		),
		tb.NoPreview,
		tb.ModeMarkdown,
	)
	if err != nil {
		zap.L().With(zap.Error(err)).Error("failed to send message")
		return err
	}

	return nil
}

func (h *handler) Unsubscribe(m *tb.Message) error {
	if err := h.sendSubscriptionList(m.Sender); err != nil {
		return err
	}

	return nil
}

func (h *handler) sendSubscriptionList(sender *tb.User) error {
	markup := h.subs.GenerateReplyMarkup(sender.ID)
	text := "List of subscriptions:"
	if markup == nil {
		text = "❌ List of subscriptions is empty"
		markup = &tb.ReplyMarkup{}
	}

	_, err := h.bot.Send(sender, text, markup)
	if err != nil {
		zap.L().With(zap.Error(err)).Error("failed to send message")
		return err
	}
	return nil
}

func (h *handler) UnsubscribeInline(c *tb.Callback) {
	resp := &tb.CallbackResponse{
		CallbackID: c.ID,
		ShowAlert:  true,
		Text:       "id is invalid",
	}
	defer func(bot *tb.Bot, c *tb.Callback, resp *tb.CallbackResponse) {
		_ = bot.Respond(c, resp)
	}(h.bot, c, resp)

	id, err := uuid.Parse(c.Data[1:])
	if err != nil {
		zap.L().
			With(zap.Error(err)).
			With(zap.String("data", c.Data)).
			Error("failed to parse id")
		return
	}

	h.subs.Remove(c.Sender.ID, id)
	h.sch.RemoveJob(id)

	resp.Text = "Successfully unsubscribed"
	resp.ShowAlert = false

	err = h.bot.Delete(c.Message)
	if err != nil {
		zap.L().With(zap.Error(err)).Error("failed to delete message")
		return
	}

	err = h.sendSubscriptionList(c.Sender)
	if err != nil {
		zap.L().With(zap.Error(err)).Error("failed to send subscription list")
		return
	}
}
