package cron

import (
	"sync/atomic"
	"time"
)

const (
	timerUnitSecond = int64(time.Second / time.Millisecond)
	timerUnitMinute = int64(time.Minute / time.Millisecond)
)

// cronTimer is aligned with the utility toolkit CronTimer and triggers launchers by time unit.
type cronTimer struct {
	scheduler *Scheduler
	stop      atomic.Bool
	stopCh    chan struct{}
}

func newCronTimer(s *Scheduler) *cronTimer {
	return &cronTimer{scheduler: s, stopCh: make(chan struct{})}
}

// run starts the loop, triggers tasks by unit, and automatically corrects clock drift.
func (t *cronTimer) run() {
	unit := timerUnitMinute
	if t.scheduler.config.MatchSecond {
		unit = timerUnitSecond
	}
	thisTime := t.scheduler.nowMillis()
	for !t.stop.Load() {
		thisTime += unit
		now := t.scheduler.nowMillis()
		sleep := thisTime - now
		if sleep > 0 {
			if !t.sleep(time.Duration(sleep) * time.Millisecond) {
				return
			}
		}
		// Reset the baseline when time moves backward or jumps significantly.
		now = t.scheduler.nowMillis()
		if now-thisTime > 2*unit || thisTime-now > 2*unit {
			thisTime = now
		}
		if t.stop.Load() {
			return
		}
		t.scheduler.launcherMgr.spawn(thisTime)
	}
}

// sleep returns false immediately when stopped.
func (t *cronTimer) sleep(d time.Duration) bool {
	if d <= 0 {
		return true
	}
	return t.scheduler.sleep(d, t.stopCh)
}

func defaultTimerSleep(d time.Duration, stopCh <-chan struct{}) bool {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-stopCh:
		return false
	case <-timer.C:
		return true
	}
}

func (t *cronTimer) stopTimer() {
	if t.stop.CompareAndSwap(false, true) {
		close(t.stopCh)
	}
}

// timeFromMillisInLocation converts milliseconds to time.Time in the specified time zone.
func timeFromMillisInLocation(ms int64, loc *time.Location) time.Time {
	if loc == nil {
		loc = time.Local
	}
	return time.UnixMilli(ms).In(loc)
}
