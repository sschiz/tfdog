package service

import (
	"io"

	"git.sr.ht/~mcldresner/tfdog/beta"
	"git.sr.ht/~mcldresner/tfdog/repository"
)

// Service describes subscription service.
// It will be periodically do payload.
type Service interface {
	Subscribe(userID int, link string, payload func(*beta.Beta)) error
	Unsubscribe(userID int, link string) error
	GetUserSubscriptions(userID int) ([]Subscription, error)

	io.Closer
}

// Subscription describes user subscription.
type Subscription struct {
	repository.Subscription
}
