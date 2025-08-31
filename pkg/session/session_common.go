package session

import "net/http"

type SessionManagerInterface interface {
	Check(*http.Request) (*Session, error)
	Create(http.ResponseWriter, UserInterface) (*Session, error)
	DestroyCurrent(http.ResponseWriter, *http.Request) error
	DestroyAll(http.ResponseWriter, UserInterface) error
}
