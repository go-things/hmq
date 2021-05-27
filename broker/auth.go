package broker

import (
	"gitee.com/godLei6/hmq/plugins"
	"strings"
)

const (
	SUB = "1"
	PUB = "2"
)

func (b *Broker) CheckTopicAuth(param plugins.AuthParm) bool {
	if b.auth != nil {
		if strings.HasPrefix(param.Topic, "$SYS/broker/connection/clients/") {
			return true
		}

		if strings.HasPrefix(param.Topic, "$share/") && param.Action == SUB {
			substr := groupCompile.FindStringSubmatch(param.Topic)
			if len(substr) != 3 {
				return false
			}
			param.Topic = substr[2]
		}

		return b.auth.CheckACL(param)
	}

	return true

}

func (b *Broker) CheckConnectAuth(param plugins.AuthParm) bool {
	if b.auth != nil {
		return b.auth.CheckConnect(param)
	}

	return true

}
