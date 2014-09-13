package core

import "time"

const (
	TIME_FORMAT_YYYYMMDD string  = "2006-01-02"
	TIME_FORMAT_YYYYMMDDHHMMSS string  = "2006-01-02 15:04:05"
)

type Timer struct {
	duration int64
	updateTime int64
	ch chan bool
}

func (this *Timer) GetChannel() chan bool {
	return this.ch
}

func (this *Timer) isActive() bool{
	return this.duration != 0
}

func (this *Timer) Stop() {
	this.duration = 0
	this.updateTime = 0
}

func (this *Timer) Start(duration int64) {
	this.duration = duration
	ch := make(chan bool, 1)
	this.ch = ch
	go func(timer *Timer) {
		for {
			if !timer.isActive() {
				break
			}
			time.Sleep(100 * time.Millisecond)
			now := GetMilliSeconds()
			if  now - this.updateTime >= duration {
				timer.updateTime = now
				ch <- true
			}
		}
	}(this)
}

func GetMilliSeconds() int64 {
	return time.Now().UnixNano() / 1e6
}
