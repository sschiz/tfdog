package bot

import (
	"git.sr.ht/~mcldresner/tfdog/beta"
	"go.uber.org/zap"
	tb "gopkg.in/tucnak/telebot.v2"
)

// NewBetaPayload returns beta payload.
func NewBetaPayload(b *tb.Bot, chat tb.ChatID) func(*beta.Beta) {
	return func(beta *beta.Beta) {
		logger := zap.L().
			Named("beta_payload").
			With(zap.String("link", beta.GetLink())).
			With(zap.String("app_name", beta.GetAppName()))
		defer logger.Debug("done")

		logger.Debug("payload is started")
		isFull, err := beta.IsFull()
		if err != nil {
			logger.
				With(zap.Error(err)).
				Error("failed to check whether beta is full")
			return
		}

		if isFull {
			logger.Debug("beta is full")
			return
		}

		text := "✅ [" + beta.GetAppName() + "](" + beta.GetLink() + ") beta has free slots!"
		_, err = b.Send(
			chat,
			text,
			tb.NoPreview,
			tb.ModeMarkdown,
		)
		if err != nil {
			logger.With(zap.Error(err)).Error("failed to send message")
			return
		}
	}
}
