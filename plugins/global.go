package plugins
type AuthParm struct {
	Action string		//操作
	ClientID string
	Username string
	Password string
	RemoteIP string
	Topic string
	Certificate []byte	//tls秘钥
}