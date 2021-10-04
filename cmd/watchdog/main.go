package main

import (
	"git.sr.ht/~mcldresner/tfdog/config"
	"git.sr.ht/~mcldresner/tfdog/handler"
	"git.sr.ht/~mcldresner/tfdog/middleware"
	"git.sr.ht/~mcldresner/tfdog/scheduler"
	tb "gopkg.in/tucnak/telebot.v2"
)

func main() {
	cfg, err := config.ParseConfig()
	if err != nil {
		panic(err)
	}

	poller := &tb.LongPoller{Timeout: cfg.PollerTimeout}
	mid := tb.NewMiddlewarePoller(poller, middleware.BuildMiddlewares(
		middleware.WithValidator(),
	))

	b, err := tb.NewBot(tb.Settings{
		Token:       cfg.Token,
		Poller:      mid,
		Synchronous: false,
	})
	if err != nil {
		panic(err)
	}

	sc := scheduler.NewScheduler(cfg.SchedulerInterval)
	h := handler.NewHandler(sc, b)

	b.Handle("/subscribe", handler.ErrorMiddleware(h.Subscribe, b))
	b.Handle("/unsubscribe", handler.ErrorMiddleware(h.Unsubscribe, b))
	b.Handle(tb.OnCallback, h.UnsubscribeInline)

	b.Handle("/ping", handler.Stringer(b, "pong!"))
	b.Handle("/help", handler.Stringer(b, cfg.HelpText))
	b.Handle("/start", handler.Stringer(b, cfg.StartText))

	sc.Start()
	b.Start()
}
