package handler

import (
	"sync"

	"github.com/google/uuid"
	tb "gopkg.in/tucnak/telebot.v2"
)

type sub struct {
	id      uuid.UUID
	appName string
}

type subscriptions struct {
	mu sync.Mutex
	m  map[int][]sub
}

func newSubscriptions() *subscriptions {
	return &subscriptions{m: make(map[int][]sub)}
}

func (s *subscriptions) Add(senderID int, id uuid.UUID, appName string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m[senderID] = append(s.m[senderID], sub{id: id, appName: appName})
}

func (s *subscriptions) Remove(senderID int, id uuid.UUID) {
	s.mu.Lock()
	defer s.mu.Unlock()
	subs, ok := s.m[senderID]
	if !ok {
		return
	}

	i := len(subs)
	for j, val := range subs {
		if val.id == id {
			i = j
			break
		}
	}
	if i == len(subs) {
		return
	}

	subs[i] = subs[len(subs)-1]
	s.m[senderID] = subs[:len(subs)-1]
}

func (s *subscriptions) RemoveAll(senderID int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.m[senderID]; ok {
		delete(s.m, senderID)
	}
}

func (s *subscriptions) GenerateReplyMarkup(senderID int) *tb.ReplyMarkup {
	s.mu.Lock()
	defer s.mu.Unlock()
	subs, ok := s.m[senderID]
	if !ok {
		return nil
	}

	if len(subs) == 0 {
		return nil
	}

	selector := new(tb.ReplyMarkup)
	rows := make([]tb.Row, len(subs))
	for i, val := range subs {
		rows[i] = selector.Row(
			selector.Data(
				val.appName,
				val.id.String(),
			),
		)
	}

	selector.Inline(rows...)

	return selector
}
