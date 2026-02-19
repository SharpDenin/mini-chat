package service

import "sync"

type PresenceSubscriber chan PresenceEvent

type PresenceEventBus struct {
	mu          sync.RWMutex
	subscribers map[PresenceSubscriber]struct{}
}

func NewPresenceEventBus() *PresenceEventBus {
	return &PresenceEventBus{
		subscribers: make(map[PresenceSubscriber]struct{}),
	}
}

func (b *PresenceEventBus) Subscribe() PresenceSubscriber {
	ch := make(PresenceSubscriber, 16)

	b.mu.Lock()
	b.subscribers[ch] = struct{}{}
	b.mu.Unlock()

	return ch
}

func (b *PresenceEventBus) Unsubscribe(ch PresenceSubscriber) {
	b.mu.Lock()
	delete(b.subscribers, ch)
	close(ch)
	b.mu.Unlock()
}

func (b *PresenceEventBus) Publish(event PresenceEvent) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for sub := range b.subscribers {
		select {
		case sub <- event:
		default:
		}
	}
}
