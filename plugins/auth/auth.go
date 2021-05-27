package auth

import (
	"fmt"
	"gitee.com/godLei6/hmq/plugins"
	authfile "gitee.com/godLei6/hmq/plugins/auth/authfile"
	"gitee.com/godLei6/hmq/plugins/auth/authgrpc"
	"gitee.com/godLei6/hmq/plugins/auth/authhttp"
)

const (
	AuthHTTP = "authhttp"
	AuthFile = "authfile"
	AuthGrpc = "authgrpc"
)



type Auth interface {
	CheckACL(plugins.AuthParm) bool
	CheckConnect(plugins.AuthParm) bool
}

func NewAuth(name string) Auth {
	fmt.Printf("NewAuth|name=%s\n",name)
	switch name {
	case AuthHTTP:
		return authhttp.Init()
	case AuthFile:
		return authfile.Init()
	case AuthGrpc:
		return authgrpc.Init()
	default:
		return &mockAuth{}
	}
}
