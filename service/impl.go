package service

import (
	"errors"
	"sync"
	"time"

	"git.sr.ht/~mcldresner/tfdog/beta"
	"git.sr.ht/~mcldresner/tfdog/repository"
	"github.com/go-co-op/gocron"
	"github.com/google/uuid"
	"go.uber.org/atomic"
	"go.uber.org/zap"
)

var (
	// ErrLinkNotFound can be returned when link wasn't subscribed
	ErrLinkNotFound = errors.New("link not found")

	// ErrJobNotFound can be returned when job is not found
	ErrJobNotFound = errors.New("job not found")
)

type srv struct {
	sc        *gocron.Scheduler // scheduler will be started after first subscription
	isStarted *atomic.Bool
	interval  time.Duration

	repo   repository.Repository
	logger *zap.Logger
	links  sync.Map
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

func (s *srv) Subscribe(userID int, link string, payload func(*beta.Beta)) (string, error) {
	logger := s.logger.
		With(zap.String("method", "subscribe")).
		With(zap.Int("user_id", userID)).
		With(zap.String("link", link))

	logger.Info("got request")
	defer logger.Debug("done")

	b, err := beta.NewTFBeta(link)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to create beta")
		return "", err
	}

	job, err := s.sc.Do(payload, b)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to schedule payload")
	}
	id := uuid.NewString()
	job.Tag(id)
	s.links.Store(id, link)

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
		return "", err
	}

	if !s.isStarted.Load() {
		s.sc.StartAsync()
		s.isStarted.Store(true)
		logger.Debug("scheduler is started")
	}

	return id, nil
}

func (s *srv) Unsubscribe(userID int, link string) error {
	logger := s.logger.
		With(zap.String("method", "unsubscribe")).
		With(zap.Int("user_id", userID)).
		With(zap.String("link", link))

	logger.Info("got request")
	defer logger.Debug("done")

	var id string
	s.links.Range(func(key, value interface{}) bool {
		k, v := key.(string), value.(string)
		if v == link {
			id = k
			return false
		}
		return true
	})
	if len(id) == 0 {
		logger.Error("link not found")
		return ErrLinkNotFound
	}

	err := s.sc.RemoveByTag(id)
	if err != nil {
		logger.With(zap.Error(err)).Error("job not found")
		return ErrJobNotFound
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

	s.links.Delete(id)

	return nil
}

func (s *srv) GetUserSubscriptions(userID int) ([]Subscription, error) {
	logger := s.logger.
		With(zap.String("method", "get_user_subscriptions")).
		With(zap.Int("user_id", userID))

	logger.Info("got request")
	defer logger.Debug("done")

	repoSubs, err := s.repo.GetUserSubscriptions(userID)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to get user subscriptions")
		return nil, err
	}

	subs := make([]Subscription, len(repoSubs))
	for i, sub := range repoSubs {
		subs[i].Subscription = sub
	}

	return subs, nil
}

func (s *srv) Close() error {
	if s.isStarted.Load() {
		s.sc.Stop()
	}
	return nil
}
