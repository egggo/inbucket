package web

import (
	"encoding/json"
	"fmt"

	"github.com/egggo/inbucket/database"
	"github.com/egggo/inbucket/log"
	// "errors"
	"io/ioutil"
	"net/http"
	"strconv"
)

func GroupAdd(w http.ResponseWriter, req *http.Request, ctx *Context) error {
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
	group := new(db.Group)
	err = json.Unmarshal(body, group)
	if err != nil {
		log.LogError("unmarshal group %v", err)
		reply["code"] = REPLY_CODE_FAIL
		reply["msg"] = err.Error()
		RenderJson(w, reply)
		return nil
	}

	user, err := ctx.Database.UserGetByName(group.Name)
	if err != nil {
		log.LogError("check group and user %v", err)
		reply["code"] = REPLY_CODE_FAIL
		reply["msg"] = err.Error()
		RenderJson(w, reply)
		return nil
	}

	if user != nil {
		log.LogError("already exist %v", user.Username)
		reply["code"] = REPLY_CODE_ALREADY_EXIST
		reply["msg"] = "already exist user"
		RenderJson(w, reply)
		return nil
	}

	log.LogTrace("add group %v", group)

	err = ctx.Database.GroupAdd(group)
	if err != nil {
		log.LogError("add group %v", err)
		reply["code"] = REPLY_CODE_FAIL
		reply["msg"] = err.Error()
		RenderJson(w, reply)
		return nil
	}

	log.LogTrace("add group suc %v", group)

	reply["id"] = group.Id
	RenderJson(w, reply)
	return nil
}

func GroupUpdate(w http.ResponseWriter, req *http.Request, ctx *Context) error {
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
	group := new(db.Group)
	group.Id, err = strconv.ParseUint(id, 10, 0)
	if err != nil {
		log.LogError("Bad group id %v", err)
		reply["code"] = REPLY_CODE_FAIL
		reply["msg"] = err.Error()
		RenderJson(w, reply)
		return nil
	}

	err = json.Unmarshal(body, group)
	if err != nil {
		log.LogError("unmarshal group %v", err)
		reply["code"] = REPLY_CODE_FAIL
		reply["msg"] = err.Error()
		RenderJson(w, reply)
		return nil
	}

	log.LogTrace("update group %v", group)

	err = ctx.Database.GroupUpdate(group)
	if err != nil {
		log.LogError("update group %v", err)
		reply["code"] = REPLY_CODE_FAIL
		reply["msg"] = err.Error()
		RenderJson(w, reply)
		return nil
	}

	log.LogTrace("update group suc %v", group)

	RenderJson(w, reply)

	return nil
}

func GroupDel(w http.ResponseWriter, req *http.Request, ctx *Context) error {
	reply := make(Reply)
	reply["code"] = REPLY_CODE_OK
	reply["msg"] = "OK"

	id := ctx.Vars["id"]
	userId, err := strconv.ParseUint(id, 10, 0)
	if err != nil {
		log.LogError("Bad group id %v", err)
		reply["code"] = REPLY_CODE_FAIL
		reply["msg"] = err.Error()
		RenderJson(w, reply)
		return nil
	}

	log.LogTrace("del group id %d", userId)

	err = ctx.Database.GroupDel(userId)
	if err != nil {
		log.LogError("del group %v", err)
		reply["code"] = REPLY_CODE_FAIL
		reply["msg"] = err.Error()
		RenderJson(w, reply)
		return nil
	}

	log.LogTrace("del group suc %d", id)

	RenderJson(w, reply)

	return nil
}

func GroupGet(w http.ResponseWriter, req *http.Request, ctx *Context) error {
	reply := make(Reply)
	reply["code"] = REPLY_CODE_OK
	reply["msg"] = "OK"

	id := ctx.Vars["id"]
	userId, err := strconv.ParseUint(id, 10, 0)
	if err != nil {
		log.LogError("Bad group id %v", err)
		reply["code"] = REPLY_CODE_FAIL
		reply["msg"] = err.Error()
		RenderJson(w, reply)
		return nil
	}

	log.LogTrace("get group id %d", userId)

	group, err := ctx.Database.GroupGet(userId)
	if err != nil {
		log.LogError("get group %v", err)
		reply["code"] = REPLY_CODE_FAIL
		reply["msg"] = err.Error()
		RenderJson(w, reply)
		return nil
	}

	if group == nil {
		reply["code"] = REPLY_CODE_NO_SUCH_USER
		reply["msg"] = fmt.Errorf("no such group").Error()
		RenderJson(w, reply)
		return nil
	}
	log.LogTrace("get group suc %d", id)

	reply["group"] = group
	RenderJson(w, reply)

	return nil
}

func GroupList(w http.ResponseWriter, req *http.Request, ctx *Context) error {
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

	log.LogTrace("get group list %d, %d", pagenoNum, countNum)

	total, groups, err := ctx.Database.GroupList(pagenoNum, countNum)
	if err != nil {
		log.LogError("get group list %v", err)
		reply["code"] = REPLY_CODE_FAIL
		reply["msg"] = err.Error()
		RenderJson(w, reply)
		return nil
	}

	log.LogTrace("get group list suc %d %d", pagenoNum, countNum)

	reply["total"] = total
	reply["groups"] = groups
	RenderJson(w, reply)

	return nil
}

func GroupMemberAdd(w http.ResponseWriter, req *http.Request, ctx *Context) error {
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
	groupMember := new(db.GroupMember)
	err = json.Unmarshal(body, groupMember)
	if err != nil {
		log.LogError("unmarshal groupMember %v", err)
		reply["code"] = REPLY_CODE_FAIL
		reply["msg"] = err.Error()
		RenderJson(w, reply)
		return nil
	}

	log.LogTrace("add groupMember %v", groupMember)

	err = ctx.Database.GroupMemberAdd(groupMember)
	if err != nil {
		log.LogError("add groupMember %v", err)
		reply["code"] = REPLY_CODE_FAIL
		reply["msg"] = err.Error()
		RenderJson(w, reply)
		return nil
	}

	log.LogTrace("add groupMember suc %v", groupMember)

	reply["id"] = groupMember.Id
	RenderJson(w, reply)
	return nil
}

func GroupMemberDel(w http.ResponseWriter, req *http.Request, ctx *Context) error {
	reply := make(Reply)
	reply["code"] = REPLY_CODE_OK
	reply["msg"] = "OK"

	id := ctx.Vars["id"]
	groupMemberId, err := strconv.ParseUint(id, 10, 0)
	if err != nil {
		log.LogError("Bad groupMember id %v", err)
		reply["code"] = REPLY_CODE_FAIL
		reply["msg"] = err.Error()
		RenderJson(w, reply)
		return nil
	}

	log.LogTrace("del groupMember id %d", groupMemberId)

	err = ctx.Database.GroupMemberDel(groupMemberId)
	if err != nil {
		log.LogError("del group %v", err)
		reply["code"] = REPLY_CODE_FAIL
		reply["msg"] = err.Error()
		RenderJson(w, reply)
		return nil
	}

	log.LogTrace("del groupMember suc %d", id)

	RenderJson(w, reply)

	return nil
}

func GroupMemberGet(w http.ResponseWriter, req *http.Request, ctx *Context) error {
	reply := make(Reply)
	reply["code"] = REPLY_CODE_OK
	reply["msg"] = "OK"

	id := ctx.Vars["id"]
	groupMemberId, err := strconv.ParseUint(id, 10, 0)
	if err != nil {
		log.LogError("Bad groupMember id %v", err)
		reply["code"] = REPLY_CODE_FAIL
		reply["msg"] = err.Error()
		RenderJson(w, reply)
		return nil
	}

	log.LogTrace("get groupMember id %d", groupMemberId)

	groupMember, err := ctx.Database.GroupMemberGet(groupMemberId)
	if err != nil {
		log.LogError("get groupMember %v", err)
		reply["code"] = REPLY_CODE_FAIL
		reply["msg"] = err.Error()
		RenderJson(w, reply)
		return nil
	}

	if groupMember == nil {
		reply["code"] = REPLY_CODE_NO_SUCH_USER
		reply["msg"] = fmt.Errorf("no such groupMember").Error()
		RenderJson(w, reply)
		return nil
	}
	log.LogTrace("get groupMember suc %d", id)

	reply["groupMember"] = groupMember
	RenderJson(w, reply)

	return nil
}

func GroupMemberList(w http.ResponseWriter, req *http.Request, ctx *Context) error {
	reply := make(Reply)
	reply["code"] = REPLY_CODE_OK
	reply["msg"] = "OK"

	groupIdStr := ctx.Vars["groupId"]
	pageno := ctx.Vars["pageno"]
	count := ctx.Vars["count"]

	groupId, err := strconv.ParseUint(groupIdStr, 10, 0)
	if err != nil {
		log.LogError("Bad group id %v", err)
		reply["code"] = REPLY_CODE_FAIL
		reply["msg"] = err.Error()
		RenderJson(w, reply)
		return nil
	}

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

	log.LogTrace("get groupMember list %d, %d, %d", groupId, pagenoNum, countNum)

	total, groupMembers, err := ctx.Database.GroupMemberList(groupId, pagenoNum, countNum)
	if err != nil {
		log.LogError("get groupMember list %v", err)
		reply["code"] = REPLY_CODE_FAIL
		reply["msg"] = err.Error()
		RenderJson(w, reply)
		return nil
	}

	reply["total"] = total
	reply["groupMembers"] = groupMembers
	RenderJson(w, reply)

	return nil
}
