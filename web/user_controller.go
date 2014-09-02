package web

import (
	"encoding/json"
	"fmt"
	// sj "github.com/bitly/go-simplejson"
	"github.com/dgryski/go-md5crypt"
	"github.com/egggo/inbucket/database"
	"github.com/egggo/inbucket/log"
	// "errors"
	"io/ioutil"
	"net/http"
	"strconv"
)

const (
	REPLY_CODE_OK = "0"

	REPLY_CODE_FAIL          = "1"
	REPLY_CODE_NO_SUCH_USER  = "10"
	REPLY_CODE_BAD_PASSWD    = "11"
	REPLY_CODE_ALREADY_EXIST = "12"
)

const (
	QUERY_USER_TYPE_NORMAL = 1
	QUERY_USER_TYPE_BATCH  = 2
)

type Reply map[string]interface{}

type PasswdPair struct {
	Old string `json:"oldPass"`
	New string `json:"newPass"`
}

type BatchUser struct {
	Type        int       `json:"type"`
	OrderColumn string    `json:"orderColumn"`
	Order       string    `json:"order"`
	Match       string    `json:"match"`
	UserList    []db.User `json:"users"`
}

func UserAdd(w http.ResponseWriter, req *http.Request, ctx *Context) error {
	reply := make(Reply)
	reply["code"] = REPLY_CODE_OK
	reply["msg"] = "OK"

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.LogError("read req %v", err)
		reply["code"] = REPLY_CODE_FAIL
		reply["msg"] = err.Error()
		RenderJson(w, reply)
		return nil
	}

	log.LogTrace("body: %v", body)
	user := new(db.User)
	err = json.Unmarshal(body, user)
	if err != nil {
		log.LogError("unmarshal user %v", err)
		reply["code"] = REPLY_CODE_FAIL
		reply["msg"] = err.Error()
		RenderJson(w, reply)
		return nil
	}

	log.LogTrace("add user %v", user)

	salt := "$1$" + salt()

	cryptPass, err := md5crypt.Crypt([]byte(user.Password), []byte(salt))
	if err != nil {
		log.LogError("crypt pass %v", err)
		reply["code"] = REPLY_CODE_FAIL
		reply["msg"] = err.Error()
		RenderJson(w, reply)
		return nil
	}

	user.Password = string(cryptPass)
	err = ctx.Database.UserAdd(user)
	if err != nil {
		log.LogError("add user %v", err)
		reply["code"] = REPLY_CODE_FAIL
		reply["msg"] = err.Error()
		RenderJson(w, reply)
		return nil
	}

	log.LogTrace("add user suc %v", user)

	reply["id"] = user.Id
	RenderJson(w, reply)
	return nil
}

func UserUpdate(w http.ResponseWriter, req *http.Request, ctx *Context) error {
	reply := make(Reply)
	reply["code"] = REPLY_CODE_OK
	reply["msg"] = "OK"

	id := ctx.Vars["id"]

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.LogError("read req %v", err)
		reply["code"] = REPLY_CODE_FAIL
		reply["msg"] = err.Error()
		RenderJson(w, reply)
		return nil
	}

	log.LogTrace("body: %v", body)
	user := new(db.User)
	user.Id, err = strconv.ParseUint(id, 10, 0)
	if err != nil {
		log.LogError("Bad user id %v", err)
		reply["code"] = REPLY_CODE_FAIL
		reply["msg"] = err.Error()
		RenderJson(w, reply)
		return nil
	}

	err = json.Unmarshal(body, user)
	if err != nil {
		log.LogError("unmarshal user %v", err)
		reply["code"] = REPLY_CODE_FAIL
		reply["msg"] = err.Error()
		RenderJson(w, reply)
		return nil
	}

	log.LogTrace("update user %v", user)

	err = ctx.Database.UserUpdate(user)
	if err != nil {
		log.LogError("update user %v", err)
		reply["code"] = REPLY_CODE_FAIL
		reply["msg"] = err.Error()
		RenderJson(w, reply)
		return nil
	}

	log.LogTrace("update user suc %v", user)

	RenderJson(w, reply)

	return nil
}

func UserDel(w http.ResponseWriter, req *http.Request, ctx *Context) error {
	reply := make(Reply)
	reply["code"] = REPLY_CODE_OK
	reply["msg"] = "OK"

	id := ctx.Vars["id"]
	userId, err := strconv.ParseUint(id, 10, 0)
	if err != nil {
		log.LogError("Bad user id %v", err)
		reply["code"] = REPLY_CODE_FAIL
		reply["msg"] = err.Error()
		RenderJson(w, reply)
		return nil
	}

	log.LogTrace("del user id %d", userId)

	err = ctx.Database.UserDel(userId)
	if err != nil {
		log.LogError("del user %v", err)
		reply["code"] = REPLY_CODE_FAIL
		reply["msg"] = err.Error()
		RenderJson(w, reply)
		return nil
	}

	log.LogTrace("del user suc %d", id)

	RenderJson(w, reply)

	return nil
}

func UserGet(w http.ResponseWriter, req *http.Request, ctx *Context) error {
	reply := make(Reply)
	reply["code"] = REPLY_CODE_OK
	reply["msg"] = "OK"

	id := ctx.Vars["id"]
	userId, err := strconv.ParseUint(id, 10, 0)
	if err != nil {
		log.LogError("Bad user id %v", err)
		reply["code"] = REPLY_CODE_FAIL
		reply["msg"] = err.Error()
		RenderJson(w, reply)
		return nil
	}

	log.LogTrace("get user id %d", userId)

	user, err := ctx.Database.UserGet(userId)
	if err != nil {
		log.LogError("get user %v", err)
		reply["code"] = REPLY_CODE_FAIL
		reply["msg"] = err.Error()
		RenderJson(w, reply)
		return nil
	}

	if user == nil {
		reply["code"] = REPLY_CODE_NO_SUCH_USER
		reply["msg"] = fmt.Errorf("no such user").Error()
		RenderJson(w, reply)
		return nil
	}
	log.LogTrace("get user suc %d", id)

	reply["user"] = user
	RenderJson(w, reply)

	return nil
}

func UserChangePasswd(w http.ResponseWriter, req *http.Request, ctx *Context) error {
	reply := make(Reply)
	reply["code"] = REPLY_CODE_OK
	reply["msg"] = "OK"

	id := ctx.Vars["id"]

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.LogError("read req %v", err)
		reply["code"] = REPLY_CODE_FAIL
		reply["msg"] = err.Error()
		RenderJson(w, reply)
		return nil
	}

	log.LogTrace("body: %v", body)
	user := new(db.User)
	passwdPair := new(PasswdPair)

	user.Id, err = strconv.ParseUint(id, 10, 0)
	if err != nil {
		log.LogError("Bad user id %v", err)
		reply["code"] = REPLY_CODE_FAIL
		reply["msg"] = err.Error()
		RenderJson(w, reply)
		return nil
	}

	err = json.Unmarshal(body, passwdPair)
	if err != nil {
		log.LogError("unmarshal passwdPair %v", err)
		reply["code"] = REPLY_CODE_FAIL
		reply["msg"] = err.Error()
		RenderJson(w, reply)
		return nil
	}

	// pass, err := ctx.Database.Auth(user.Id, passwdPair.Old)
	// if err != nil {
	// 	log.LogError("unmarshal passwdPair %v", err)
	// 	reply["code"] = REPLY_CODE_FAIL
	// 	reply["msg"] = err.Error()
	// 	RenderJson(w, reply)
	// 	return nil
	// }

	// if !pass {
	// 	log.LogError("bad passwd %v", passwdPair)
	// 	reply["code"] = REPLY_CODE_BAD_PASSWD
	// 	reply["msg"] = fmt.Errorf("bad passwd").Error()
	// 	RenderJson(w, reply)
	// 	return nil
	// }

	log.LogTrace("change user passwd %v", id)

	salt := "$1$" + salt()

	cryptPass, err := md5crypt.Crypt([]byte(passwdPair.New), []byte(salt))
	if err != nil {
		log.LogError("crypt pass %v", err)
		reply["code"] = REPLY_CODE_FAIL
		reply["msg"] = err.Error()
		RenderJson(w, reply)
		return nil
	}

	user.Password = string(cryptPass)

	err = ctx.Database.UserUpdate(user)
	if err != nil {
		log.LogError("change user passwd %v", err)
		reply["code"] = REPLY_CODE_FAIL
		reply["msg"] = err.Error()
		RenderJson(w, reply)
		return nil
	}

	log.LogTrace("change user passwd suc %v", user)

	RenderJson(w, reply)

	return nil
}

func UserList(w http.ResponseWriter, req *http.Request, ctx *Context) error {
	reply := make(Reply)
	reply["code"] = REPLY_CODE_OK
	reply["msg"] = "OK"

	pageno := ctx.Vars["pageno"]
	count := ctx.Vars["count"]

	pagenoNum, err := strconv.Atoi(pageno)

	if err != nil {
		log.LogError("Bad pageno %v", err)
		reply["code"] = REPLY_CODE_FAIL
		reply["msg"] = err.Error()
		RenderJson(w, reply)
		return nil
	}

	countNum, err := strconv.Atoi(count)

	if err != nil {
		log.LogError("Bad count %v", err)
		reply["code"] = REPLY_CODE_FAIL
		reply["msg"] = err.Error()
		RenderJson(w, reply)
		return nil
	}

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.LogError("read req %v", err)
		reply["code"] = REPLY_CODE_FAIL
		reply["msg"] = err.Error()
		RenderJson(w, reply)
		return nil
	}

	batchUser := new(BatchUser)

	log.LogTrace("body: %v ", body)
	err = json.Unmarshal(body, batchUser)
	if err != nil {
		log.LogError("unmarshal user %v", err)
		reply["code"] = REPLY_CODE_FAIL
		reply["msg"] = err.Error()
		RenderJson(w, reply)
		return nil
	}

	log.LogTrace("get user list %d, %d, %d, %v", pagenoNum, countNum, batchUser.Type, batchUser)

	ids := make([]uint64, 0)

	if batchUser.Type == QUERY_USER_TYPE_BATCH {
		for _, v := range batchUser.UserList {
			ids = append(ids, v.Id)
		}
	}

	total, users, err := ctx.Database.UserList(pagenoNum, countNum, ids, batchUser.OrderColumn, batchUser.Order, batchUser.Match)
	if err != nil {
		log.LogError("get user list %v", err)
		reply["code"] = REPLY_CODE_FAIL
		reply["msg"] = err.Error()
		RenderJson(w, reply)
		return nil
	}

	log.LogTrace("get user list suc %d %d", pagenoNum, countNum)

	reply["total"] = total
	reply["users"] = users
	RenderJson(w, reply)

	return nil
}
