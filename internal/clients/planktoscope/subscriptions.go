package planktoscope

import (
	"sync"
)

type Broadcaster struct {
	channel  chan struct{}
	channelL *sync.RWMutex
}

func NewBroadcaster() *Broadcaster {
	return &Broadcaster{
		channel:  make(chan struct{}),
		channelL: &sync.RWMutex{},
	}
}

func (b *Broadcaster) BroadcastNext() {
	b.channelL.Lock()
	defer b.channelL.Unlock()

	close(b.channel)
	b.channel = make(chan struct{})
}

func (b *Broadcaster) Broadcasted() <-chan struct{} {
	b.channelL.RLock()
	defer b.channelL.RUnlock()
	return b.channel
}
