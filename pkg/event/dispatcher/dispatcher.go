package dispatcher

import (
	"context"
	"sync"

	"github.com/krixlion/dev_forum-user/pkg/event"
)

type Dispatcher struct {
	maxWorkers int
	handlers   map[event.EventType][]event.Handler
	events     <-chan event.Event
	mu         sync.Mutex
	broker     event.Broker
}

func NewDispatcher(broker event.Broker, maxWorkers int) *Dispatcher {
	return &Dispatcher{
		maxWorkers: maxWorkers,
		broker:     broker,
		handlers:   make(map[event.EventType][]event.Handler),
	}
}

// AddEventSources registers provided channels as an event source.
// This method is not thread safe and should be called before Run().
func (d *Dispatcher) AddEventSources(sources ...<-chan event.Event) {
	d.events = mergeChans(sources...)
}

func (d *Dispatcher) Run(ctx context.Context) {
	for {
		select {
		case event := <-d.events:
			d.Dispatch(event)
		case <-ctx.Done():
			return
		}
	}
}

func (d *Dispatcher) Subscribe(handler event.Handler, eTypes ...event.EventType) {
	d.mu.Lock()
	defer d.mu.Unlock()
	for _, eType := range eTypes {
		d.handlers[eType] = append(d.handlers[eType], handler)
	}
}

func (d *Dispatcher) Dispatch(e event.Event) {
	limit := make(chan struct{}, d.maxWorkers)

	if err := d.broker.ResilientPublish(e); err != nil {
		panic(err)
	}

	for _, handler := range d.handlers[e.Type] {
		limit <- struct{}{}
		go func(handler event.Handler) {
			handler.Handle(e)
			<-limit
		}(handler)
	}
}

func mergeChans(cs ...<-chan event.Event) <-chan event.Event {
	out := make(chan event.Event)

	wg := sync.WaitGroup{}
	wg.Add(len(cs))

	for _, c := range cs {
		go func(c <-chan event.Event) {
			for v := range c {
				out <- v
			}
			wg.Done()
		}(c)
	}

	return out
}
