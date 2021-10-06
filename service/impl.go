package service

import (
	"errors"
	"strconv"
	"time"

	"git.sr.ht/~mcldresner/tfdog/beta"
	"git.sr.ht/~mcldresner/tfdog/repository"
	"github.com/go-co-op/gocron"
	"go.uber.org/atomic"
	"go.uber.org/zap"
)

var (
	// ErrSubscriptionNotFound may be returned when job is not found.
	ErrSubscriptionNotFound = errors.New("subscription not found")

	// ErrAlreadySubscribed may be returned if link is already subscribed.
	ErrAlreadySubscribed = errors.New("link already subscribed")
)

type srv struct {
	sc        *gocron.Scheduler // scheduler will be started after first subscription
	isStarted *atomic.Bool
	interval  time.Duration

	repo   repository.Repository
	logger *zap.Logger
}

// NewService new Service instance.
func NewService(repo repository.Repository, interval time.Duration) Service {
	return &srv{
		sc:        gocron.NewScheduler(time.UTC),
		isStarted: atomic.NewBool(false),
		repo:      repo,
		interval:  interval,
		logger:    zap.L().Named("service"),
	}
}

func (s *srv) Subscribe(userID int, link string, payload func(*beta.Beta)) error {
	logger := s.logger.
		With(zap.String("method", "subscribe")).
		With(zap.Int("user_id", userID)).
		With(zap.String("link", link))

	logger.Info("got request")
	defer logger.Debug("done")

	isSubscribed, err := s.isSubscribed(userID, link)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to check whether link is subscribed")
		return err
	}
	if isSubscribed {
		return ErrAlreadySubscribed
	}

	b, err := beta.NewTFBeta(link)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to create beta")
		return err
	}

	job, err := s.sc.Do(payload, b)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to schedule payload")
	}
	job.Tag(strconv.Itoa(userID) + link)

	err = s.repo.SaveSubscription(
		repository.Subscription{
			UserID:  userID,
			Link:    link,
			AppName: b.GetAppName(),
		},
	)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to save subscription")
		s.sc.RemoveByReference(job)
		return err
	}

	if !s.isStarted.Load() {
		s.sc.StartAsync()
		s.isStarted.Store(true)
		logger.Debug("scheduler is started")
	}

	return nil
}

func (s *srv) Unsubscribe(userID int, link string) error {
	logger := s.logger.
		With(zap.String("method", "unsubscribe")).
		With(zap.Int("user_id", userID)).
		With(zap.String("link", link))

	logger.Info("got request")
	defer logger.Debug("done")

	err := s.sc.RemoveByTag(strconv.Itoa(userID) + link)
	if err != nil {
		logger.With(zap.Error(err)).Error("job not found")
		return ErrSubscriptionNotFound
	}

	err = s.repo.RemoveSubscription(repository.Subscription{
		UserID: userID,
		Link:   link,
	})
	if err != nil {
		s.logger.
			With(zap.Error(err)).
			Error("failed to remove subscription")
		return err
	}

	return nil
}

func (s *srv) GetUserSubscriptions(userID int) ([]Subscription, error) {
	logger := s.logger.
		With(zap.String("method", "get_user_subscriptions")).
		With(zap.Int("user_id", userID))

	logger.Info("got request")
	defer logger.Debug("done")

	subs, err := s.repo.GetUserSubscriptions(userID)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to get user subscriptions")
		return nil, err
	}

	return castSubscriptions(subs), nil
}

func (s *srv) Close() error {
	if s.isStarted.Load() {
		s.sc.Stop()
	}
	return nil
}

func (s *srv) isSubscribed(userID int, link string) (bool, error) {
	subs, err := s.repo.GetUserSubscriptions(userID)
	if err != nil {
		return false, err
	}

	var isSubscribed bool
	for _, sub := range subs {
		if sub.Link == link {
			isSubscribed = true
			break
		}
	}

	return isSubscribed, nil
}

func castSubscriptions(repoSubs []repository.Subscription) []Subscription {
	subs := make([]Subscription, len(repoSubs))
	for i, sub := range repoSubs {
		subs[i].Subscription = sub
	}

	return subs
}
