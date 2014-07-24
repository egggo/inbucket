package db

import (
	"fmt"
	"github.com/dgryski/go-md5crypt"
	"github.com/egggo/inbucket/config"
	"github.com/egggo/inbucket/log"
	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
	"strings"
	"time"
)

type User struct {
	Id       uint64    `xorm:"pk autoincr" json:"id"`
	Username string    `xorm:"varchar(255) not null unique 'username'" json:"username"`
	Password string    `xorm:"varchar(255) not null 'password'" json:"password"`
	Domain   string    `xorm:"varchar(255) not null 'domain'" json:"domain"`
	Created  time.Time `xorm:"created" json:"created"`
	Updated  time.Time `xorm:"updated" json:"updated"`
}

type Group struct {
	Id      uint64    `xorm:"pk autoincr" json:"id"`
	Name    string    `xorm:"varchar(255) not null unique 'name'" json:"name"`
	Domain  string    `xorm:"varchar(255) not null 'domain'" json:"domain"`
	Created time.Time `xorm:"created" json:"created"`
	Updated time.Time `xorm:"updated" json:"updated"`
}

type GroupMember struct {
	Id      uint64    `xorm:"pk autoincr" json:"id"`
	UserId  uint64    `xorm:"BigInt not null" json:"userId"`
	GroupId uint64    `xorm:"BigInt not null" json:"groupId"`
	Created time.Time `xorm:"created" json:"created"`
	Updated time.Time `xorm:"updated" json:"updated"`
}

type Database struct {
	engine *xorm.Engine
}

func New() *Database {
	cfg := config.GetDatabaseConfig()
	params := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8", cfg.DBUser, cfg.DBPass, cfg.DBHost, cfg.DBName)

	engine, err := xorm.NewEngine(cfg.DBDriver, params)

	if err != nil {
		log.LogError(" create Database engine  fail: %v", err)
		// TODO More graceful early-shutdown procedure
		panic(err)
	}

	err = engine.Sync(
		new(User),
		new(Group),
		new(GroupMember),
	)

	if err != nil {
		log.LogError(" engine sync  fail: %v", err)
		panic(err)
	}

	engine.ShowSQL = true

	return &Database{engine: engine}
}

func (db *Database) Close() {
	db.engine.Close()
}
func (db *Database) UserAdd(user *User) error {

	_, err := db.engine.Insert(user)
	return err
}

func (db *Database) UserDel(id uint64) error {
	user := new(User)
	_, err := db.engine.Id(id).Delete(user)
	return err
}

func (db *Database) UserUpdate(user *User) error {

	_, err := db.engine.Id(user.Id).Update(user)
	return err
}

func (db *Database) UserGet(id uint64) (*User, error) {
	user := new(User)
	has, err := db.engine.Id(id).Get(user)
	if !has || err != nil {
		return nil, err
	}
	return user, nil
}

func (db *Database) UserList(pageno int, count int) (int64, []*User, error) {
	users := make([]*User, 0)

	user := new(User)
	total, err := db.engine.Count(user)
	if err != nil {
		return 0, nil, err
	}
	err = db.engine.Limit(count, pageno*count).Find(&users)

	return total, users, nil
}

func (db *Database) Auth(id uint64, pass string) (bool, error) {
	user := new(User)

	log.LogInfo("Auth - id: %v, pass: <%v>", id, pass)

	has, err := db.engine.Where("id=?", id).Get(user)
	if err != nil {
		return false, err
	}
	log.LogInfo("has: %v, user pass: %v", has, user.Password)
	if has {

		substrs := strings.Split(user.Password, "$")
		if len(substrs) < 4 {
			return false, nil
		}

		salt := "$" + substrs[1] + "$" + substrs[2]

		cryptPass, err := md5crypt.Crypt([]byte(pass), []byte(salt))
		if err != nil {
			return false, nil
		}

		log.LogInfo("salt : %v, cryptPass: %v", salt, string(cryptPass))

		if string(cryptPass) != user.Password {
			return false, nil
		} else {
			return true, nil
		}
	}
	return has, err
}

func (db *Database) GroupAdd(group *Group) error {

	_, err := db.engine.Insert(group)
	return err
}

func (db *Database) GroupDel(id uint64) error {
	group := new(Group)
	_, err := db.engine.Id(id).Delete(group)
	return err
}

func (db *Database) GroupUpdate(group *Group) error {

	_, err := db.engine.Id(group.Id).Update(group)
	return err
}

func (db *Database) GroupGet(id uint64) (*Group, error) {
	group := new(Group)
	has, err := db.engine.Id(id).Get(group)
	if !has || err != nil {
		return nil, err
	}
	return group, nil
}

func (db *Database) GroupList(pageno int, count int) (int64, []*Group, error) {
	groups := make([]*Group, 0)

	group := new(Group)
	total, err := db.engine.Count(group)
	if err != nil {
		return 0, nil, err
	}
	err = db.engine.Limit(count, pageno*count).Find(&groups)

	return total, groups, nil
}

func (db *Database) GroupMemberAdd(groupMember *GroupMember) error {

	_, err := db.engine.Insert(groupMember)
	return err
}

func (db *Database) GroupMemberDel(id uint64) error {
	groupMember := new(GroupMember)
	_, err := db.engine.Id(id).Delete(groupMember)
	return err
}

func (db *Database) GroupMemberGet(id uint64) (*GroupMember, error) {
	groupMember := new(GroupMember)
	has, err := db.engine.Id(id).Get(groupMember)
	if !has || err != nil {
		return nil, err
	}
	return groupMember, nil
}

func (db *Database) GroupMemberList(groupId uint64, pageno int, count int) (int64, []*GroupMember, error) {
	groupMembers := make([]*GroupMember, 0)

	groupMember := new(GroupMember)
	total, err := db.engine.Count(groupMember)
	if err != nil {
		return 0, nil, err
	}

	if groupId > 0 {
		err = db.engine.Where("groupId=?", groupId).Limit(count, pageno*count).Find(&groupMembers)
	} else {
		err = db.engine.Limit(count, pageno*count).Find(&groupMembers)
	}
	return total, groupMembers, nil
}
