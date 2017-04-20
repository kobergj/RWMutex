package CustomRWMutex

import (
	"sync"
)

type RWMutex interface {
	Lock()
	Unlock()

	RLock()
	RUnlock()
}

func NewFakeMutex() RWMutex {
	m := fakeMutex{}

	return &m
}

type fakeMutex struct {
}

func (this *fakeMutex) Lock() {
}

func (this *fakeMutex) Unlock() {
}

func (this *fakeMutex) RLock() {
}

func (this *fakeMutex) RUnlock() {
}

func NewBuildInMutex() RWMutex {
	return &sync.RWMutex{}
}

func NewMighlighHighMutex() RWMutex {
	awl := make(chan bool)
	gwl := make(chan bool)
	rwl := make(chan bool)

	rl := make(chan int)

	mhm := &milighHighMutex{
		acquireWriteLock: awl,
		grantWriteLock:   gwl,
		releaseWriteLock: rwl,

		readLock: rl,
	}

	go mhm.cleanChannels()

	return mhm
}

type milighHighMutex struct {
	acquireWriteLock chan bool
	grantWriteLock   chan bool
	releaseWriteLock chan bool

	readLock        chan int
	activeReadLocks int
}

func (this *milighHighMutex) Lock() {
	this.acquireWriteLock <- true
	<-this.grantWriteLock
}

func (this *milighHighMutex) Unlock() {
	this.releaseWriteLock <- true
}

func (this *milighHighMutex) RLock() {
	this.readLock <- 1
}

func (this *milighHighMutex) RUnlock() {
	this.readLock <- -1
}

func (this *milighHighMutex) cleanChannels() {
	for {
		select {
		case n := <-this.readLock:
			this.activeReadLocks += n
		case <-this.acquireWriteLock:
			for this.activeReadLocks != 0 {
				n := <-this.readLock
				this.activeReadLocks += n
			}
			this.grantWriteLock <- true
			<-this.releaseWriteLock
		}
	}
}

func NewChannelMutex() RWMutex {
	awl := make(chan bool)
	gwl := make(chan bool)
	rwl := make(chan bool)

	arl := make(chan bool)
	rrl := make(chan bool)

	rwm := &channelMutex{
		acquireWriteLock: awl,
		grantWriteLock:   gwl,
		releaseWriteLock: rwl,

		acquireReadLock: arl,
		releaseReadLock: rrl,
	}

	go rwm.cleanChannels()

	return rwm
}

type channelMutex struct {
	acquireWriteLock chan bool
	grantWriteLock   chan bool
	releaseWriteLock chan bool

	acquireReadLock chan bool
	releaseReadLock chan bool
	activeReadLocks int
}

func (this *channelMutex) Lock() {
	this.acquireWriteLock <- true
	<-this.grantWriteLock
}

func (this *channelMutex) Unlock() {
	go func() {
		this.releaseWriteLock <- true
	}()
}

func (this *channelMutex) RLock() {
	this.acquireReadLock <- true
}

func (this *channelMutex) RUnlock() {
	go func() {
		this.releaseReadLock <- true
	}()
}

func (this *channelMutex) cleanChannels() {
	for {
		select {
		case <-this.acquireWriteLock:
			for this.activeReadLocks != 0 {
				<-this.releaseReadLock
				this.activeReadLocks -= 1
			}
			this.grantWriteLock <- true
			<-this.releaseWriteLock
		case <-this.acquireReadLock:
			this.activeReadLocks += 1
		}
	}
}

func NewClosingMutex() RWMutex {
	awl := make(chan bool)
	rwl := make(chan bool)

	arl := make(chan bool)
	rrl := make(chan bool)

	rwm := &closingMutex{
		acquireWriteLock: awl,
		releaseWriteLock: rwl,

		acquireReadLock: arl,
		releaseReadLock: rrl,
	}

	go rwm.cleanChannels()

	return rwm
}

type closingMutex struct {
	acquireWriteLock chan bool
	writelocked      chan bool
	releaseWriteLock chan bool

	acquireReadLock chan bool
	readlocked      chan bool
	releaseReadLock chan bool

	activeReadLocks int
}

func (this *closingMutex) Lock() {
	this.acquireWriteLock <- true
}

func (this *closingMutex) Unlock() {
	go func() {
		this.releaseWriteLock <- true
	}()
}

func (this *closingMutex) RLock() {
	this.acquireReadLock <- true
}

func (this *closingMutex) RUnlock() {
	go func() {
		this.releaseReadLock <- true
	}()
}

func (this *closingMutex) cleanChannels() {
	for {
		select {

		case <-this.acquireReadLock:
			if this.readlocked == nil {
				this.readlocked = make(chan bool)
			}
			this.activeReadLocks += 1
		case <-this.releaseReadLock:
			this.activeReadLocks -= 1
			if this.activeReadLocks == 0 {
				close(this.readlocked)
				this.readlocked = nil
			}
		default:
			if this.readlocked == nil {
				select {
				case <-this.acquireWriteLock:
					if this.writelocked == nil {
						this.writelocked = make(chan bool)
					}
					<-this.releaseWriteLock
					close(this.writelocked)
					this.writelocked = nil
				default:
					continue
				}
			}
		}
	}
}

func NewMaxReaderMutex(maxReaders int) RWMutex {
	return &maxReaderMutex{
		maxReaders: maxReaders,
	}
}

type maxReaderMutex struct {
	activeWriteLocks chan bool
	activeReadLocks  chan bool

	maxReaders int
}

func (this *maxReaderMutex) Lock() {
	if this.activeWriteLocks == nil {
		this.activeWriteLocks = make(chan bool, 1)
	}

	this.activeWriteLocks <- true

	if this.activeReadLocks == nil {
		this.activeReadLocks = make(chan bool, this.maxReaders)
	}

	for i := 0; i < this.maxReaders; i++ {
		this.activeReadLocks <- true
	}

}

func (this *maxReaderMutex) Unlock() {
	<-this.activeWriteLocks

	for i := 0; i < this.maxReaders; i++ {
		<-this.activeReadLocks
	}
}

func (this *maxReaderMutex) RLock() {
	if this.activeReadLocks == nil {
		this.activeReadLocks = make(chan bool, this.maxReaders)
	}

	this.activeReadLocks <- true
}

func (this *maxReaderMutex) RUnlock() {
	<-this.activeReadLocks
}
