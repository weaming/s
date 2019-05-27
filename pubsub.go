package main

import (
	"log"
	"sync"
)

type MsgHandlerFunc func(interface{}) error
type PubSub map[string]*Topic

// return success
func (p *PubSub) Pub(topic string, msg interface{}, init bool) bool {
	t, ok := (*p)[topic]
	if !ok {
		if !init {
			return false
		}
		t = NewTopic(topic, p)
	}
	t.Pub <- msg
	return true
}

func (p *PubSub) Get(topic string) *Topic {
	return (*p)[topic]
}

func (p *PubSub) GetValid(topic string) *Topic {
	t := p.Get(topic)
	if t == nil {
		t = NewTopic(topic, p)
	}
	return t
}

func (p *PubSub) Subscribe(topic string, subKey string, fn MsgHandlerFunc) *Topic {
	t := p.GetValid(topic)
	// TODO: handle old fn
	// if _, ok := t.Subs[subKey]; ok {
	// }
	t.Subs[subKey] = fn
	return t
}

func (p *PubSub) StopTopic(topic string) {
	if t, ok := (*p)[topic]; ok {
		t.Stop <- true
	}
}

func (p *PubSub) StopIdleTopic(topic string) {
	if t, ok := (*p)[topic]; ok {
		if len(t.Subs) == 0 {
			t.Stop <- true
		}
	}
}

type Topic struct {
	Topic  string
	Stop   chan bool
	Pub    chan interface{}
	Subs   map[string]MsgHandlerFunc
	pubsub *PubSub
	sync.RWMutex
}

func NewTopic(topic string, pubsub *PubSub) *Topic {
	t := &Topic{
		Topic:  topic,
		Stop:   make(chan bool, 1),
		Pub:    make(chan interface{}, 10000),
		Subs:   map[string]MsgHandlerFunc{},
		pubsub: pubsub,
	}
	(*pubsub)[topic] = t
	go t.Start()
	return t
}

func (t *Topic) Start() {
	log.Println("[pubsub] starting topic", t.Topic)
	for {
		select {
		case x := <-t.Pub:
			t.RLock()
			for _, fn := range t.Subs {
				// avoid block
				go func(fn func(interface{}) error) {
					err := fn(x)
					if err != nil {
						// put back
						t.Pub <- x
						log.Println(red(err.Error()))
					} else {
						log.Println(green("sent message on topic \"%v\", message: %v", t.Topic, x))
					}
				}(fn)
			}
			t.RUnlock()
		case <-t.Stop:
			delete(*t.pubsub, t.Topic)
			log.Printf("stopped topic %v", t.Topic)
			return
		}
	}

}

func (t *Topic) Unsubscribe(subKey string) {
	t.Lock()
	defer t.Unlock()
	if _, ok := t.Subs[subKey]; ok {
		delete(t.Subs, subKey)
		log.Printf("unsubscribed on topic %s for key %s", t.Topic, subKey)
	}
}
