package store

import (
	"sync"
	"time"

	"edj.cash/edj/internal/core"
)

type User struct {
	Sub   string `json:"sub"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

type Notification struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"createdAt"`
	Read      bool      `json:"read"`
}

type Memory struct {
	mu            sync.RWMutex
	path          string
	signerKey     string
	links         map[string]*core.Link
	runs          map[string]*core.PayrollRun
	proofs        map[string]*core.DisclosureProof
	users         map[string]*User
	sessions      map[string]string
	notifications []Notification
}

func New() *Memory {
	return &Memory{
		links:    map[string]*core.Link{},
		runs:     map[string]*core.PayrollRun{},
		proofs:   map[string]*core.DisclosureProof{},
		users:    map[string]*User{},
		sessions: map[string]string{},
		notifications: []Notification{
			{ID: "N_1", Title: "Welcome to edj", Body: "Your private receiving address is ready in Wallet settings.", CreatedAt: time.Now().UTC(), Read: false},
			{ID: "N_2", Title: "BLAZE 2026", Body: "Submissions close 22 November, 14:00 UTC. Ship early and post your demo.", CreatedAt: time.Now().UTC(), Read: false},
		},
	}
}

func (m *Memory) PutLink(l *core.Link) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.links[l.ID] = l
	m.saveLocked()
}

func (m *Memory) GetLink(id string) (*core.Link, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	l, ok := m.links[id]
	return l, ok
}

func (m *Memory) Links() []*core.Link {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]*core.Link, 0, len(m.links))
	for _, l := range m.links {
		out = append(out, l)
	}
	return out
}

func (m *Memory) PutRun(r *core.PayrollRun) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.runs[r.ID] = r
	m.saveLocked()
}

func (m *Memory) GetRun(id string) (*core.PayrollRun, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	r, ok := m.runs[id]
	return r, ok
}

func (m *Memory) PutProof(p *core.DisclosureProof) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.proofs[p.ID] = p
	m.saveLocked()
}

func (m *Memory) GetProof(id string) (*core.DisclosureProof, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	p, ok := m.proofs[id]
	return p, ok
}

func (m *Memory) UpsertUser(u *User) (*User, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, exists := m.users[u.Sub]
	m.users[u.Sub] = u
	m.saveLocked()
	return u, !exists
}

func (m *Memory) User(sub string) (*User, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	u, ok := m.users[sub]
	return u, ok
}

func (m *Memory) PutSession(token, sub string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sessions[token] = sub
	m.saveLocked()
}

func (m *Memory) Session(token string) (string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, ok := m.sessions[token]
	return s, ok
}

func (m *Memory) DeleteSession(token string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.sessions, token)
	m.saveLocked()
}

func (m *Memory) Notifications() []Notification {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return append([]Notification(nil), m.notifications...)
}
