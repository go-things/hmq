package authgrpc

import (
	"context"
	"encoding/json"
	"gitee.com/godLei6/hmq/logger"
	"github.com/tal-tech/go-zero/zrpc"
	"gitee.com/godLei6/things/src/dmsvr/dmclient"
	"go.uber.org/zap"
	"io/ioutil"
	"time"
)



type authGRPC struct {
	dmclient.Dm
}

var (
	config     zrpc.RpcClientConf
	log        = logger.Get().Named("authgrpc")
)

//Init init kafak client
func Init() *authGRPC {
	content, err := ioutil.ReadFile("./plugins/auth/authgrpc/grpc.json")
	if err != nil {
		log.Fatal("Read config file error: ", zap.Error(err))
	}
	// log.Info(string(content))

	err = json.Unmarshal(content, &config)
	if err != nil {
		log.Fatal("Unmarshal config file error: ", zap.Error(err))
	}
	// fmt.Println("http: config: ", config)

	client := dmclient.NewDm(zrpc.MustNewClient(config))
	return &authGRPC{Dm: client}
}

//CheckAuth check mqtt connect
func (a *authGRPC) CheckConnect(clientID, username, password, ip string) bool {
	ctx,cancel := context.WithTimeout(context.Background(),time.Minute)
	defer cancel()
	_, err := a.LoginAuth(ctx,&dmclient.LoginAuthReq{
		Username :username, //用户名
		Password :password, //密码
		Clientid :clientID, //clientID
		Ip       :ip,	//访问的ip地址
	})
	if err != nil {
		return false
	}
	return true
}


//CheckACL check mqtt connect
func (a *authGRPC) CheckACL(action, clientID, username, ip, topic string) bool {
	ctx,cancel := context.WithTimeout(context.Background(),time.Minute)
	defer cancel()
	_, err := a.AccessAuth(ctx,&dmclient.AccessAuthReq{
		Username :username, //用户名
		Topic    :topic,  	//主题
		ClientID :clientID,  	//clientID
		Access   :action,	//操作
		Ip       :ip,	//访问的ip地址
	})
	if err != nil {
		return false
	}
	return true
}


