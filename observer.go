package main

type publisher interface {
	publish(value any)
}

type observer interface {
	notify(status map[string][]string)
}

type observerFunc func(status map[string][]string)

func (fn observerFunc) notify(status map[string][]string) {
	fn(status)
}

type observable []observer

func (observers *observable) attachObserver(a observer) {
	*observers = append(*observers, a)
}

func (observers observable) publish(status map[string][]string) {
	for _, obs := range observers {
		obs.notify(status)
	}
}
