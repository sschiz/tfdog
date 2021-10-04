package main

import (
	"time"

	"git.sr.ht/~mcldresner/tfdog/handler"
	"git.sr.ht/~mcldresner/tfdog/scheduler"

	tb "gopkg.in/tucnak/telebot.v2"
)

func main() {
	b, err := tb.NewBot(tb.Settings{
		Token:       "token here",
		Poller:      &tb.LongPoller{Timeout: 10 * time.Second},
		Synchronous: false,
	})
	if err != nil {
		panic(err)
	}

	sc := scheduler.NewScheduler(3 * time.Minute)
	h := handler.NewHandler(sc, b)

	b.Handle("/subscribe", handler.ErrorMiddleware(handler.CheckerMiddleware(h.Subscribe), b))
	b.Handle("/unsubscribe", handler.ErrorMiddleware(handler.CheckerMiddleware(h.Unsubscribe), b))
	b.Handle(tb.OnCallback, h.UnsubscribeInline)

	b.Handle("/ping", func(m *tb.Message) {
		_, _ = b.Send(m.Chat, "Pong!")
	})
	b.Handle("/help", help(b))
	b.Handle("/start", help(b))

	sc.Start()
	b.Start()
}

const helpText = `
TestFlight Watcher is a bot that let you to periodically check whether TestFlight beta is full.
If beta is not full, the bot will send notifications about it.

/subscribe [beta link] - subscribe to notifications about beta fullness
/unsubscribe - unsubscribe from notifications
/ping - pong!
/help - get help
`

func help(bot *tb.Bot) func(*tb.Message) {
	return func(msg *tb.Message) {
		_, _ = bot.Send(msg.Sender, helpText)
	}
}
