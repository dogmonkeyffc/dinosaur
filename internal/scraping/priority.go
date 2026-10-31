package scraping

import (
	"dinosaur/internal/media"
	"dinosaur/internal/metadata"
)

type scheduledJob struct {
	job      metadata.Job
	request  metadata.ScrapeRequest
	settings metadata.Settings
	items    []media.Item
}

// One worker owns all metadata writes. Manual confirmations run between bulk
// items, after an in-flight lookup finishes and before the next automatic one.
func (m *Manager) work(task scheduledJob) {
	defer m.workers.Done()
	for {
		m.run(task.job, task.request, task.settings, task.items)
		m.mu.Lock()
		if len(m.manual) > 0 {
			task = m.takeManualLocked()
			m.mu.Unlock()
			continue
		}
		m.active = false
		pending := m.pending && !m.closed
		m.pending = false
		m.mu.Unlock()
		if pending {
			m.Schedule()
		}
		return
	}
}

func (m *Manager) takeManualLocked() scheduledJob {
	task := m.manual[0]
	m.manual[0] = scheduledJob{}
	m.manual = m.manual[1:]
	return task
}

func (m *Manager) runManualJobs() {
	for {
		m.mu.Lock()
		if m.closed || len(m.manual) == 0 {
			m.mu.Unlock()
			return
		}
		task := m.takeManualLocked()
		m.mu.Unlock()
		m.run(task.job, task.request, task.settings, task.items)
	}
}
