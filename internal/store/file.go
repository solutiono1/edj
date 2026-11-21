package store

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"edj.cash/edj/internal/core"
)

type snapshot struct {
	Links         map[string]*core.Link          `json:"links"`
	Runs          map[string]*core.PayrollRun    `json:"runs"`
	Proofs        map[string]*core.DisclosureProof `json:"proofs"`
	Users         map[string]*User               `json:"users"`
	Sessions      map[string]string              `json:"sessions"`
	Notifications []Notification                 `json:"notifications"`
	SignerKey     string                         `json:"signerKey"`
}

func NewPersistent(path string) (*Memory, *core.Signer, error) {
	m := New()
	m.path = path

	var signer *core.Signer
	if data, err := os.ReadFile(path); err == nil {
		var snap snapshot
		if err := json.Unmarshal(data, &snap); err != nil {
			return nil, nil, fmt.Errorf("read store: %w", err)
		}
		if snap.Links != nil {
			m.links = snap.Links
		}
		if snap.Runs != nil {
			m.runs = snap.Runs
		}
		if snap.Proofs != nil {
			m.proofs = snap.Proofs
		}
		if snap.Users != nil {
			m.users = snap.Users
		}
		if snap.Sessions != nil {
			m.sessions = snap.Sessions
		}
		if snap.Notifications != nil {
			m.notifications = snap.Notifications
		}
		if snap.SignerKey != "" {
			raw, err := base64.RawURLEncoding.DecodeString(snap.SignerKey)
			if err != nil {
				return nil, nil, fmt.Errorf("read signing key: %w", err)
			}
			signer, err = core.NewSignerFromKey(ed25519.PrivateKey(raw))
			if err != nil {
				return nil, nil, fmt.Errorf("load signing key: %w", err)
			}
		}
	} else if !os.IsNotExist(err) {
		return nil, nil, fmt.Errorf("open store: %w", err)
	}

	if signer == nil {
		signer = core.NewSigner()
	}
	m.signerKey = base64.RawURLEncoding.EncodeToString(signer.PrivateKey())
	if err := m.saveLocked(); err != nil {
		return nil, nil, err
	}
	return m, signer, nil
}

func (m *Memory) saveLocked() error {
	if m.path == "" {
		return nil
	}
	snap := snapshot{
		Links:         m.links,
		Runs:          m.runs,
		Proofs:        m.proofs,
		Users:         m.users,
		Sessions:      m.sessions,
		Notifications: m.notifications,
		SignerKey:     m.signerKey,
	}
	data, err := json.MarshalIndent(snap, "", "  ")
	if err != nil {
		return fmt.Errorf("encode store: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(m.path), 0o700); err != nil {
		return fmt.Errorf("create store dir: %w", err)
	}
	tmp := m.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return fmt.Errorf("write store: %w", err)
	}
	if err := os.Rename(tmp, m.path); err != nil {
		return fmt.Errorf("commit store: %w", err)
	}
	return nil
}
