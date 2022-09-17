package model

type EventHandleFunc func(e string, args ...interface{})

type IEventDrivenModel interface {
	Hook(e string, handleFunc EventHandleFunc)
	Raise(e string, args ...interface{})
}

type EventDrivenModel struct {
	items map[string][]EventHandleFunc
}

func (edm *EventDrivenModel) Hook(s string, handler EventHandleFunc) {
	evf, ok := edm.items[s]
	if ok {
		edm.items[s] = append(evf, handler)
	} else {
		edm.items[s] = []EventHandleFunc{
			handler,
		}
	}
}

func (edm *EventDrivenModel) Raise(s string, args ...interface{}) {
	if handlers, ok := edm.items[s]; ok {
		for _, item := range handlers {
			item(s, args)
		}
	}
}
