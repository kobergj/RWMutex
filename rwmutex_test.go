package CustomRWMutex

import (
	"testing"
	"time"
)

type Counter struct {
	counter int
}

func (this *Counter) Increment(i int) {
	this.counter += i
}

func (this *Counter) Decrememt(i int) {
	this.counter -= i
}

func (this *Counter) Read() int {
	return this.counter
}

func writer(rwm RWMutex, counter *Counter, done chan<- bool, t *testing.T) {
	rwm.Lock()

	counter.Increment(6291)

	time.Sleep(1)

	if c := counter.Read(); c != 6291 {
		t.Error("Multiple Writes!")
	}

	counter.Decrememt(6291)
	rwm.Unlock()
	done <- true
}

func reader(rwm RWMutex, counter *Counter, done chan<- bool, t *testing.T) {
	rwm.RLock()

	first := counter.Read()

	time.Sleep(1)

	second := counter.Read()

	if first != second {
		t.Error("Write while ReadLock!")
	}
	rwm.RUnlock()
	done <- true
}

func writetester(rwm RWMutex, writers int, t *testing.T) {
	c := Counter{}
	done := make(chan bool)
	defer close(done)
	for i := 0; i < writers; i++ {
		go writer(rwm, &c, done, t)
	}
	for i := 0; i < writers; i++ {
		<-done
	}
}

func readtester(rwm RWMutex, readers, writers int, t *testing.T) {
	c := Counter{}
	done := make(chan bool)
	defer close(done)
	for i := 0; i < readers/2; i++ {
		go reader(rwm, &c, done, t)
	}
	for i := 0; i < writers; i++ {
		go writer(rwm, &c, done, t)
	}
	for i := 0; i < readers/2; i++ {
		go reader(rwm, &c, done, t)
	}

	for i := 0; i < readers+writers; i++ {
		<-done
	}
}

func Test_Lock(t *testing.T) {
	rwm := NewMighlighHighMutex()
	writetester(rwm, 10, t)
}

func Test_RLock(t *testing.T) {
	rwm := NewMighlighHighMutex()

	readtester(rwm, 10, 10, t)
}
