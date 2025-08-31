package user

import (
	"encoding/json"
	"fmt"
	"hw11_shopql/pkg/session"
	"net/http"
)

type Resp map[string]map[string]string

type UserHandler struct {
	St UserRepoInterface
	SM session.SessionManagerInterface
}

func (uh *UserHandler) Log(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		user := make(map[string]*User)
		err := json.NewDecoder(r.Body).Decode(&user)
		if err != nil {
			fmt.Println(err)
		}
		if err != nil {
			http.Error(w, "failed to read json", 401)
			return
		}
		User := user["user"]
		id, err := uh.St.CheckUser(User.EMAIL, User.Password)
		if err != nil {
			http.Error(w, "Failed", http.StatusBadRequest)
			return
		}
		User.ID = id
		sess, _ := uh.SM.Create(w, User)
		res, _ := json.Marshal(Resp{"body": map[string]string{"token": sess.ID}})
		w.Write(res)
		w.WriteHeader(200)
	} else {
		w.WriteHeader(401)
	}
}

func (uh *UserHandler) Reg(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		user := make(map[string]*User)
		err := json.NewDecoder(r.Body).Decode(&user)
		User := user["user"]
		//err := json.NewDecoder(r.Body).Decode(&user)
		if err != nil {
			http.Error(w, "failed to read json", 401)
			return
		}
		id, err := uh.St.AddUser(User)
		if err != nil {
			http.Error(w, "Failed", http.StatusBadRequest)
			return
		}
		User.ID = id
		sess, _ := uh.SM.Create(w, User)
		res, _ := json.Marshal(Resp{"body": map[string]string{"token": sess.ID}})
		w.Write(res)
		w.WriteHeader(200)
	} else {
		w.WriteHeader(401)
	}
}

func CreateUserHandler(userRepo UserRepoInterface, SM session.SessionManagerInterface) *UserHandler {
	return &UserHandler{
		St: userRepo,
		SM: SM,
	}
}
