package session

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"hw11_shopql/pkg/utils/randutils"
)

type SessionsDB struct {
	DB *sql.DB
}

type Session struct {
	UserID uint32
	ID     string
}

/*
	https://medium.com/@dotronglong/interface-naming-convention-in-golang-f53d9f471593
	user
	User
	Userer
	IUser
	UserI
	UserInterface
*/

type UserInterface interface {
	GetID() uint32
	// GetVer() int32
}

type SessionManager interface {
	Check(*http.Request) (*Session, error)
	Create(http.ResponseWriter, UserInterface) (*Session, error)
	DestroyCurrent(http.ResponseWriter, *http.Request) error
	DestroyAll(http.ResponseWriter, UserInterface) error
}

// линтер ругается если используем базовые типы в Value контекста
// типа так безопаснее разграничивать
type ctxKey string

const sessionKey ctxKey = "Token1"

var (
	ErrNoAuth = errors.New("No session found")
)

func SessionFromContext(ctx context.Context) (*Session, error) {
	sess, ok := ctx.Value(sessionKey).(*Session)
	if !ok {
		return nil, ErrNoAuth
	}
	return sess, nil
}

var (
	noAuthUrls = map[string]struct{}{
		"/user/login_oauth": struct{}{},
		"/user/login":       struct{}{},
		"/user/reg":         struct{}{},
		"/":                 struct{}{},
	}
)

func AuthMiddleware(sm SessionManager, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := noAuthUrls[r.URL.Path]; ok {
			next.ServeHTTP(w, r)
			return
		}
		sess, err := sm.Check(r)
		if err != nil {
			http.Error(w, "No auth", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), sessionKey, sess)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func NewSessionsDB(db *sql.DB) *SessionsDB {
	return &SessionsDB{
		DB: db,
	}
}

func (sm *SessionsDB) Check(r *http.Request) (*Session, error) {
	sessionId := r.Header.Get("Authorization")
	if len(sessionId) < 6 {
		return nil, ErrNoAuth
	}
	sess := &Session{}
	row := sm.DB.QueryRow(`SELECT user_id FROM sessions WHERE id = $1`, strings.Split(sessionId, " ")[1])
	err := row.Scan(&sess.UserID)
	if err == sql.ErrNoRows {
		log.Println("CheckSession no rows")
		return nil, ErrNoAuth
	} else if err != nil {
		log.Println("CheckSession err:", err)
		return nil, err
	}
	sess.ID = sessionId
	return sess, nil
}

func (sm *SessionsDB) Create(w http.ResponseWriter, user UserInterface) (*Session, error) {
	sessID := randutils.RandStringRunes(32)
	_, err := sm.DB.Exec("INSERT INTO sessions(id, user_id) VALUES($1, $2)", sessID, user.GetID())
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	w.Header().Set("Authorization", sessID)
	return &Session{UserID: user.GetID(), ID: sessID}, nil
}

func (sm *SessionsDB) DestroyCurrent(w http.ResponseWriter, r *http.Request) error {
	sess, err := SessionFromContext(r.Context())
	if err == nil {
		_, err = sm.DB.Exec("DELETE FROM sessions WHERE id = ?", sess.ID)
		if err != nil {
			return err
		}
	}
	cookie := http.Cookie{
		Name:    "session_id",
		Expires: time.Now().AddDate(0, 0, -1),
		Path:    "/",
	}
	http.SetCookie(w, &cookie)
	return nil
}

func (sm *SessionsDB) DestroyAll(w http.ResponseWriter, user UserInterface) error {
	result, err := sm.DB.Exec("DELETE FROM sessions WHERE user_id = ?",
		user.GetID())
	if err != nil {
		return err
	}

	affected, _ := result.RowsAffected()
	log.Println("destroyed sessions", affected, "for user", user.GetID())

	return nil
}
