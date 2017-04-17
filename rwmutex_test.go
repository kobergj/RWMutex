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

func writer(rwm RWMutex, counter *Counter, done chan<- bool, t *testing.T, testAlias string) {
	rwm.Lock()

	counter.Increment(6291)

	time.Sleep(1)

	if c := counter.Read(); c != 6291 {
		t.Errorf("[%s] Write while WriteLock!", testAlias)
	}

	counter.Decrememt(6291)
	rwm.Unlock()
	done <- true
}

func reader(rwm RWMutex, counter *Counter, done chan<- bool, t *testing.T, testAlias string) {
	rwm.RLock()

	first := counter.Read()

	time.Sleep(1)

	second := counter.Read()

	if first != second {
		t.Errorf("[%s] Write while ReadLock!", testAlias)
	}
	rwm.RUnlock()
	done <- true
}

func writetester(rwm RWMutex, writers int, t *testing.T, testAlias string) {
	c := Counter{}
	done := make(chan bool)
	defer close(done)
	for i := 0; i < writers; i++ {
		go writer(rwm, &c, done, t, testAlias)
	}
	for i := 0; i < writers; i++ {
		<-done
	}
}

func readtester(rwm RWMutex, readers, writers int, t *testing.T, testAlias string) {
	c := Counter{}
	done := make(chan bool)
	defer close(done)
	for i := 0; i < readers/2; i++ {
		go reader(rwm, &c, done, t, testAlias)
	}
	for i := 0; i < writers; i++ {
		go writer(rwm, &c, done, t, testAlias)
	}
	for i := 0; i < readers/2; i++ {
		go reader(rwm, &c, done, t, testAlias)
	}

	for i := 0; i < readers+writers; i++ {
		<-done
	}
}

func Test_Lock(t *testing.T) {
	var testCases = []struct {
		Alias         string
		Mutex         RWMutex
		NumberWriters int
	}{
		{
			Alias:         "CustomMutex10Writers",
			Mutex:         NewMighlighHighMutex(),
			NumberWriters: 10,
		},
		{
			Alias:         "ChannelMutex10Writers",
			Mutex:         NewChannelMutex(),
			NumberWriters: 10,
		},
	}

	for _, testCase := range testCases {
		rwm := testCase.Mutex
		alias := testCase.Alias
		writers := testCase.NumberWriters
		writetester(rwm, writers, t, alias)
	}
}

func Test_RLock(t *testing.T) {
	rwm := NewMighlighHighMutex()

	readtester(rwm, 10, 10, t, "TestRLock")
}

func benchmarkRWMutex(b *testing.B, localWork, writeRatio int, rwm RWMutex) {
	b.RunParallel(func(pb *testing.PB) {
		foo := 0
		for pb.Next() {
			foo++
			if foo%writeRatio == 0 {
				rwm.Lock()
				rwm.Unlock()
			} else {
				rwm.RLock()
				for i := 0; i != localWork; i += 1 {
					foo *= 2
					foo /= 2
				}
				rwm.RUnlock()
			}
		}
		_ = foo
	})
}

func BenchmarkRWMutexWrite100(b *testing.B) {
	rwm := NewMighlighHighMutex()
	benchmarkRWMutex(b, 0, 100, rwm)
}

func BenchmarkBuildinRWMutexWrite100(b *testing.B) {
	rwm := NewBuildInMutex()
	benchmarkRWMutex(b, 0, 100, rwm)
}
