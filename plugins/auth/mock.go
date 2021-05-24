package auth

type mockAuth struct{}

func (m *mockAuth) CheckACL(action, clientID, username, ip, topic string) bool {
	return true
}

func (m *mockAuth) CheckConnect(clientID, username, password, ip string) bool {
	return true
}
