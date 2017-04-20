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
		t.Errorf("[ERROR] Write while WriteLock!")
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
		t.Errorf("[ERROR] Write while ReadLock!")
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
	var i, j int
	for i = 0; i < readers/2; i++ {
		go reader(rwm, &c, done, t)
	}
	for j = 0; j < writers; j++ {
		go writer(rwm, &c, done, t)
	}
	for ; i < readers; i++ {
		go reader(rwm, &c, done, t)
	}

	for i := 0; i < readers+writers; i++ {
		<-done
	}
}

var MutexesToTest = []struct {
	Alias string
	Mutex RWMutex
}{
	{
		Alias: "BuildInMutex",
		Mutex: NewBuildInMutex(),
	},
	// {
	// 	Alias: "CustomMutex",
	// 	Mutex: NewMighlighHighMutex(),
	// },
	// {
	// 	Alias: "ChannelMutex",
	// 	Mutex: NewChannelMutex(),
	// },
	// {
	// 	Alias: "ClosingMutex",
	// 	Mutex: NewClosingMutex(),
	// },
	// {
	// 	Alias: "FakeMutex-TestsShouldFail",
	// 	Mutex: NewFakeMutex(),
	// },
	// {
	// 	Alias: "MaxReaderMutex1000",
	// 	Mutex: NewMaxReaderMutex(1000),
	// },
	{
		Alias: "MaxReaderMutex10",
		Mutex: NewMaxReaderMutex(10),
	},
	// {
	// 	Alias: "MaxReaderMutex1",
	// 	Mutex: NewMaxReaderMutex(1),
	// },
}

func Test_Lock(t *testing.T) {
	var testCases = []struct {
		Alias         string
		NumberWriters int
	}{
		{
			Alias:         "10Writers",
			NumberWriters: 10,
		},
		{
			Alias:         "100Writers",
			NumberWriters: 100,
		},
		{
			Alias:         "1Writers",
			NumberWriters: 1,
		},
	}

	for _, mutex := range MutexesToTest {
		mutexalias := mutex.Alias
		rwm := mutex.Mutex

		for _, testCase := range testCases {
			alias := mutexalias + "-" + testCase.Alias
			writers := testCase.NumberWriters

			test := func(tt *testing.T) {
				writetester(rwm, writers, tt)
			}

			t.Run(alias, test)
		}
	}
}

func Test_RLock(t *testing.T) {

	testCases := []struct {
		Alias   string
		Writers int
		Readers int
	}{
		{
			Alias:   "10w-10r",
			Writers: 10,
			Readers: 10,
		},
		{
			Alias:   "100w-0r",
			Writers: 100,
			Readers: 0,
		},
		{
			Alias:   "1000w-1000r",
			Writers: 1000,
			Readers: 1000,
		},
		{
			Alias:   "0w-1r",
			Writers: 0,
			Readers: 1,
		},
	}

	for _, mutex := range MutexesToTest {
		mutexalias := mutex.Alias
		rwm := mutex.Mutex

		for _, testCase := range testCases {
			alias := mutexalias + "-" + testCase.Alias
			w := testCase.Writers
			r := testCase.Readers

			test := func(tt *testing.T) {
				readtester(rwm, r, w, tt)
			}

			t.Run(alias, test)
		}

	}

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

func BenchmarkRWMutex(b *testing.B) {
	testCases := []struct {
		Alias      string
		LocalWork  int
		WriteRatio int
		Mutex      RWMutex
	}{
		{
			Alias:      "LW0-WR100",
			LocalWork:  0,
			WriteRatio: 100,
		},
		{
			Alias:      "LW20-WR50",
			LocalWork:  20,
			WriteRatio: 50,
		},
	}

	for _, mutex := range MutexesToTest {
		mutexalias := mutex.Alias
		rwm := mutex.Mutex

		for _, testCase := range testCases {
			alias := mutexalias + "-" + testCase.Alias
			lw := testCase.LocalWork
			wr := testCase.WriteRatio

			benchmark := func(pb *testing.B) {
				benchmarkRWMutex(pb, lw, wr, rwm)
			}

			b.Run(alias, benchmark)
		}
	}
}
