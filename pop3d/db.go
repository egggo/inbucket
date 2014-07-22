package pop3d

// import (
// 	"github.com/dgryski/go-md5crypt"
// 	"github.com/egggo/inbucket/log"
// 	"strings"
// )

// type Mailbox struct {
// 	Name string `xorm:"varchar(255) not null unique 'username'"`
// 	Pass string `xorm:"varchar(255) not null 'password'"`
// }

// func (ses *Session) auth(name string, pass string) (bool, error) {
// 	mailbox := new(Mailbox)

// 	log.LogInfo("Auth - name: %v, pass: <%v>", name, pass)

// 	has, err := ses.server.dbEngine.Where("username=?", name+"@"+ses.server.domain).Get(mailbox)
// 	if err != nil {
// 		return false, err
// 	}
// 	log.LogInfo("has: %v, user pass: %v", has, mailbox.Pass)
// 	if has {

// 		substrs := strings.Split(mailbox.Pass, "$")
// 		if len(substrs) < 4 {
// 			return false, nil
// 		}

// 		salt := "$" + substrs[1] + "$" + substrs[2]

// 		cryptPass, err := md5crypt.Crypt([]byte(pass), []byte(salt))
// 		if err != nil {
// 			return false, nil
// 		}

// 		log.LogInfo("salt : %v, cryptPass: %v", salt, string(cryptPass))

// 		if string(cryptPass) != mailbox.Pass {
// 			return false, nil
// 		} else {
// 			return true, nil
// 		}
// 	}
// 	return has, err
// }
