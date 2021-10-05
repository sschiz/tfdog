package bot

import (
	"time"

	"git.sr.ht/~mcldresner/tfdog/middleware"
	"git.sr.ht/~mcldresner/tfdog/scheduler"
	tb "gopkg.in/tucnak/telebot.v2"
)

// NewBot constructs new bot.
func NewBot(pollerTimeout time.Duration,
	token string,
	sc *scheduler.Scheduler,
	helpText, startText string,
) (*tb.Bot, error) {
	poller := &tb.LongPoller{
		Timeout:        pollerTimeout,
		AllowedUpdates: []string{"message", "callback_query"},
	}
	mid := tb.NewMiddlewarePoller(poller, middleware.BuildMiddlewares(
		middleware.WithValidator(),
	))

	b, err := tb.NewBot(tb.Settings{
		Token:       token,
		Poller:      mid,
		Synchronous: false,
	})
	if err != nil {
		return nil, err
	}

	h := newHandler(sc, b)

	b.Handle("/subscribe", ErrorMiddleware(h.Subscribe, b))
	b.Handle("/unsubscribe", ErrorMiddleware(h.Unsubscribe, b))
	b.Handle(tb.OnCallback, h.UnsubscribeInline)

	b.Handle("/ping", Stringer(b, "pong!"))
	b.Handle("/help", Stringer(b, helpText))
	b.Handle("/start", Stringer(b, startText))

	return b, nil
}
