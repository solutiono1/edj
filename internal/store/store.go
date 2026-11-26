package store

import "edj.cash/edj/internal/core"

type Store interface {
	PutLink(*core.Link)
	GetLink(string) (*core.Link, bool)
	Links() []*core.Link
	PutRun(*core.PayrollRun)
	GetRun(string) (*core.PayrollRun, bool)
	Runs() []*core.PayrollRun
	PutProof(*core.DisclosureProof)
	GetProof(string) (*core.DisclosureProof, bool)
	Proofs() []*core.DisclosureProof
	UpsertUser(*User) (*User, bool)
	User(string) (*User, bool)
	UserByEmail(string) (*User, bool)
	CreateEmailUser(*User, string) (*User, bool)
	PasswordHash(string) (string, bool)
	PutSession(string, string)
	Session(string) (string, bool)
	DeleteSession(string)
	Notifications() []Notification
}
