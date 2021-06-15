package authgrpc

import (
	"context"
	"encoding/json"
	"gitee.com/godLei6/hmq/logger"
	"gitee.com/godLei6/hmq/plugins"
	"gitee.com/godLei6/things/src/dmsvr/dmclient"
	"github.com/tal-tech/go-zero/zrpc"
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
func (a *authGRPC) CheckConnect(param plugins.AuthParm) bool {
	ctx,cancel := context.WithTimeout(context.Background(),time.Minute)
	defer cancel()
	_, err := a.LoginAuth(ctx,&dmclient.LoginAuthReq{
		Username :param.Username, //用户名
		Password :param.Password, //密码
		ClientID :param.ClientID, //clientID
		Ip       :param.RemoteIP,	//访问的ip地址
		Certificate :param.Certificate,
	})
	if err != nil {
		return false
	}
	return true
}


//CheckACL check mqtt connect
func (a *authGRPC) CheckACL(param plugins.AuthParm) bool {
	ctx,cancel := context.WithTimeout(context.Background(),time.Minute)
	defer cancel()
	_, err := a.AccessAuth(ctx,&dmclient.AccessAuthReq{
		Username :param.Username, 	//用户名
		Topic    :param.Topic,  	//主题
		ClientID :param.ClientID,  	//clientID
		Access   :param.Action,		//操作
		Ip       :param.RemoteIP,	//访问的ip地址
	})
	if err != nil {
		return false
	}
	return true
}


