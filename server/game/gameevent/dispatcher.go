package gameevent

type Handler interface {
	HandleGameEvent(event Event)
}

type HandlerFunc func(event Event)

func (handler HandlerFunc) HandleGameEvent(event Event) {
	handler(event)
}

type Dispatcher struct {
	subscriptions []subscription
}

type subscription struct {
	eventId string
	handler Handler
}

func NewDispatcher() *Dispatcher {
	return &Dispatcher{subscriptions: []subscription{}}
}

func (dispatcher *Dispatcher) Register(handler Handler) {
	if handler == nil {
		return
	}
	dispatcher.subscriptions = append(dispatcher.subscriptions, subscription{handler: handler})
}

func (dispatcher *Dispatcher) Subscribe(eventId string, handler Handler) {
	if eventId == "" || handler == nil {
		return
	}
	dispatcher.subscriptions = append(dispatcher.subscriptions, subscription{
		eventId: eventId,
		handler: handler,
	})
}

func (dispatcher *Dispatcher) Emit(event Event) {
	subscriptions := append([]subscription(nil), dispatcher.subscriptions...)
	for _, subscription := range subscriptions {
		if subscription.eventId != "" && subscription.eventId != event.Id {
			continue
		}
		subscription.handler.HandleGameEvent(event)
	}
}
