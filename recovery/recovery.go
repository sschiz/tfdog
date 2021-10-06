package recovery

import (
	"git.sr.ht/~mcldresner/tfdog/repository"
	"git.sr.ht/~mcldresner/tfdog/service"
	"git.sr.ht/~mcldresner/tfdog/transport/bot"
	tb "gopkg.in/tucnak/telebot.v2"
)

// ServiceFromRepository restores the service using a repository.
func ServiceFromRepository(srv service.Service, repo repository.Repository, b *tb.Bot) error {
	subs, err := repo.GetAllSubscriptions()
	if err != nil {
		return err
	}
	if len(subs) == 0 {
		return nil
	}

	err = repo.DeleteAllSubscriptions()
	if err != nil {
		return err
	}

	for _, sub := range subs {
		payload := bot.NewBetaPayload(b, &tb.User{ID: sub.UserID})
		err = srv.Subscribe(sub.UserID, sub.Link, payload)
		if err != nil {
			return err
		}
	}

	return nil
}
