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

func NewRWMutex() RWMutex {
	// m := sync.RWMutex{}

	c := make(chan bool, 1)
	m := wMutex{c}

	return &m
}

type wMutex struct {
	locked chan bool
}

func (this *wMutex) Lock() {
	this.locked <- true
}

func (this *wMutex) Unlock() {
	<-this.locked
}

func (this *wMutex) RLock() {
}

func (this *wMutex) RUnlock() {
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

}

func (this *channelMutex) RUnlock() {

}

func (this *channelMutex) cleanChannels() {
	for {
		select {
		case <-this.acquireWriteLock:
			this.grantWriteLock <- true
			<-this.releaseWriteLock

		}
	}
}
