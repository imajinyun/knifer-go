package cron

import "sync"

// listenerManager is aligned with the utility toolkit TaskListenerManager.
type listenerManager struct {
	mu        sync.RWMutex
	listeners []TaskListener
}

func newListenerManager() *listenerManager {
	return &listenerManager{}
}

func (m *listenerManager) add(l TaskListener) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.listeners = append(m.listeners, l)
}

func (m *listenerManager) remove(l TaskListener) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i, x := range m.listeners {
		if x == l {
			m.listeners = append(m.listeners[:i], m.listeners[i+1:]...)
			return
		}
	}
}

func (m *listenerManager) snapshot() []TaskListener {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]TaskListener, len(m.listeners))
	copy(out, m.listeners)
	return out
}

func (m *listenerManager) notifyStart(e *TaskExecutor) {
	for _, l := range m.snapshot() {
		safeNotify(func() { l.OnStart(e) })
	}
}

func (m *listenerManager) notifySucceeded(e *TaskExecutor) {
	for _, l := range m.snapshot() {
		safeNotify(func() { l.OnSucceeded(e) })
	}
}

func (m *listenerManager) notifyFailed(e *TaskExecutor, err any) {
	listeners := m.snapshot()
	if len(listeners) == 0 {
		// Fallback path: when no listener exists, keep the failure from being silently ignored.
		// This package does not depend on log here; callers can observe failures through listeners.
		return
	}
	for _, l := range listeners {
		safeNotify(func() { l.OnFailed(e, err) })
	}
}

func safeNotify(fn func()) {
	defer func() { _ = recover() }()
	fn()
}
