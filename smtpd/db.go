package smtpd

import (
	"github.com/egggo/inbucket/log"
	// "github.com/go-xorm/core"
	"strings"
)

type Alias struct {
	Address string `xorm:"varchar(255) not null unique 'address'"`
	Goto    ConvAlias
}

type ConvAlias []string

func (c *ConvAlias) FromDB(data []byte) error {
	*c = strings.Split(string(data), ",")
	return nil
}

func (c *ConvAlias) ToDB() ([]byte, error) {
	return []byte(strings.Join(*c, ",")), nil
}

func (ses *Session) verifyAlias(address string) ([]string, error) {
	alias := new(Alias)
	// alias.Goto = new(ConvAlias)

	log.LogInfo("Auth - address: %v", address)

	has, err := ses.server.dbEngine.Where("address=?", address).Get(alias)
	if err != nil {
		return nil, err
	}

	log.LogInfo("has: %v , alias:%v", has, alias)

	if has {
		return alias.Goto, nil
	}
	return nil, nil
}
