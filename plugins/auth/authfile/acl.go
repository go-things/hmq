package acl

import (
	"gitee.com/godLei6/hmq/plugins"
)

type aclAuth struct {
	config *ACLConfig
}

func Init() *aclAuth {
	aclConfig, err := AclConfigLoad("./plugins/auth/authfile/acl.conf")
	if err != nil {
		panic(err)
	}
	return &aclAuth{
		config: aclConfig,
	}
}

func (a *aclAuth) CheckConnect(param plugins.AuthParm) bool {
	return true
}

func (a *aclAuth) CheckACL(param plugins.AuthParm) bool {
	return checkTopicAuth(a.config, param.Action, param.RemoteIP, param.Username, param.ClientID, param.Topic)
}
