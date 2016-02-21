package main

type Mutex struct {
	mc chan int
}

func NewMutex() *Mutex {
	mu := new(Mutex)
	mu.mc = make(chan int, 1)
	// equivalently
	// mu := &Mutex{make(chan int, 1)}
	mu.Unlock()
	return mu
}

func (mu *Mutex) Lock() {
	<-mu.mc
	// don't care what is removed
}

func (mu *Mutex) Unlock() {
	mu.mc <- 1
}
