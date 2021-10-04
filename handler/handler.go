package handler

import (
	"fmt"
	"log"

	"git.sr.ht/~mcldresner/tfdog/scheduler"
	"git.sr.ht/~mcldresner/tfdog/tfbeta"

	"github.com/google/uuid"
	tb "gopkg.in/tucnak/telebot.v2"
)

type Handler struct {
	sch  *scheduler.Scheduler
	bot  *tb.Bot
	subs *subscriptions
}

func NewHandler(s *scheduler.Scheduler, b *tb.Bot) *Handler {
	return &Handler{sch: s, bot: b, subs: newSubscriptions()}
}

func (h *Handler) Subscribe(m *tb.Message) error {
	if len(m.Payload) == 0 {
		_, _ = h.bot.Send(
			m.Sender,
			"Please, specify TestFlight link.\n"+
				"Example:\n`/subscribe https://testflight.apple.com/join/u6iogfd0`",
			tb.ModeMarkdown,
		)
	}

	beta, err := tfbeta.NewTFBeta(m.Payload)
	if err != nil {
		log.Printf("failed to create beta: %s", err)
		return err
	}
	name := beta.GetAppName()

	bot := h.bot
	sender := tb.ChatID(m.Chat.ID)
	id, err := h.sch.Schedule(func() {
		isFull, err := beta.IsFull()
		if err != nil {
			_, _ = bot.Send(sender, "error occurred: "+err.Error())
			log.Printf("failed to check whether beta is full or not: %s", err)
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
				log.Printf("failed to send message: %s", err)
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
		log.Printf("failed to send message: %s", err)
		return err
	}

	return nil
}

func (h *Handler) Unsubscribe(m *tb.Message) error {
	if err := h.sendSubscriptionList(m.Sender); err != nil {
		return err
	}

	return nil
}

func (h *Handler) sendSubscriptionList(sender *tb.User) error {
	markup := h.subs.GenerateReplyMarkup(sender.ID)
	text := "List of subscriptions:"
	if markup == nil {
		text = "❌ List of subscriptions is empty"
		markup = &tb.ReplyMarkup{}
	}

	_, err := h.bot.Send(sender, text, markup)
	if err != nil {
		log.Printf("failed to send message: %s", err)
		return err
	}
	return nil
}

func (h *Handler) UnsubscribeInline(c *tb.Callback) {
	resp := &tb.CallbackResponse{
		CallbackID: c.ID,
		ShowAlert:  true,
	}

	id, err := uuid.Parse(c.Data[1:])
	if err != nil {
		resp.Text = "id is invalid"
		log.Printf("failed to parse id: data = %s, err = %s", c.Data, err)
		_ = h.bot.Respond(c, resp)
		return
	}

	h.subs.Remove(c.Sender.ID, id)
	h.sch.RemoveJob(id)

	resp.Text = "Successfully unsubscribed"
	resp.ShowAlert = false
	_ = h.bot.Respond(c, resp)

	err = h.bot.Delete(c.Message)
	if err != nil {
		log.Printf("failed to delete message: %s", err)
		return
	}

	_ = h.sendSubscriptionList(c.Sender)
}
