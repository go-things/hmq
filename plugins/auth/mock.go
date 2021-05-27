package auth

import "gitee.com/godLei6/hmq/plugins"

type mockAuth struct{}

func (m *mockAuth) CheckACL(plugins.AuthParm) bool {
	return true
}

func (m *mockAuth) CheckConnect(plugins.AuthParm) bool {
	return true
}
