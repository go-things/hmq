package authhttp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gitee.com/godLei6/hmq/logger"
	"gitee.com/godLei6/hmq/plugins"
	"go.uber.org/zap"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

//Config device kafka config
type Config struct {
	AuthURL  string `json:"auth"`
	ACLURL   string `json:"acl"`
	SuperURL string `json:"super"`
}

type authHTTP struct {
	client *http.Client
}

var (
	config     Config
	log        = logger.Get().Named("authhttp")
	httpClient *http.Client
)

//Init init kafak client
func Init() *authHTTP {
	content, err := ioutil.ReadFile("./plugins/auth/authhttp/http.json")
	if err != nil {
		log.Fatal("Read config file error: ", zap.Error(err))
	}
	// log.Info(string(content))

	err = json.Unmarshal(content, &config)
	if err != nil {
		log.Fatal("Unmarshal config file error: ", zap.Error(err))
	}
	// fmt.Println("http: config: ", config)

	httpClient = &http.Client{
		Transport: &http.Transport{
			MaxConnsPerHost:     100,
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
		},
		Timeout: time.Second * 100,
	}
	return &authHTTP{client: httpClient}
}

//CheckAuth check mqtt connect
func (a *authHTTP) CheckConnect(param plugins.AuthParm) bool {
	action := "connect"
	{
		aCache := checkCache(action, param.ClientID, param.Username, param.Password, "")
		if aCache != nil {
			if aCache.password == param.Password && aCache.username == param.Username && aCache.action == action {
				return true
			}
		}
	}

	data := make(map[string] string,3)
	data["username"] = param.Username
	data["clientid"] = param.ClientID
	data["password"] = param.Password
	data["ip"] = param.RemoteIP
	jData,_ := json.Marshal(data)
	resp,err := HttpPost(jData,config.AuthURL)
	if err != nil {
		log.Error("new request super: ", zap.Error(err))
		return false
	}
	defer resp.Body.Close()
	io.Copy(ioutil.Discard, resp.Body)
	if resp.StatusCode == http.StatusOK {
		addCache(action, param.ClientID, param.Username, param.Password, "")
		return true
	}

	return false
}

// //CheckSuper check mqtt connect
// func CheckSuper(clientID, username, password string) bool {
// 	action := "connect"
// 	{
// 		aCache := checkCache(action, clientID, username, password, "")
// 		if aCache != nil {
// 			if aCache.password == password && aCache.username == username && aCache.action == action {
// 				return true
// 			}
// 		}
// 	}

// 	data := url.Values{}
// 	data.Add("username", username)
// 	data.Add("clientid", clientID)
// 	data.Add("password", password)

// 	req, err := http.NewRequest("POST", config.SuperURL, strings.NewReader(data.Encode()))
// 	if err != nil {
// 		log.Error("new request super: ", zap.Error(err))
// 		return false
// 	}
// 	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
// 	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

// 	resp, err := httpClient.Do(req)
// 	if err != nil {
// 		log.Error("request super: ", zap.Error(err))
// 		return false
// 	}

// 	defer resp.Body.Close()
// 	io.Copy(ioutil.Discard, resp.Body)

// 	if resp.StatusCode == http.StatusOK {
// 		return true
// 	}
// 	return false
// }

//CheckACL check mqtt connect
func (a *authHTTP) CheckACL(param plugins.AuthParm) bool {

	{
		aCache := checkCache(param.Action, "", param.Username, "", param.Topic)
		if aCache != nil {
			if aCache.topic == param.Topic && aCache.action == param.Action {
				return true
			}
		}
	}


	data := make(map[string] string,3)
	data["username"] = param.Username
	data["clientid"] = param.ClientID
	data["topic"] = param.Topic
	data["access"] = param.Action
	data["ip"] = param.RemoteIP
	jData,_ := json.Marshal(data)
	resp,err := HttpPost(jData,config.ACLURL)
	if err != nil {
		log.Error("new request super: ", zap.Error(err))
		return false
	}
	defer resp.Body.Close()
	io.Copy(ioutil.Discard, resp.Body)

	if resp.StatusCode == http.StatusOK {
		addCache(param.Action, "", param.Username, "", param.Topic)
		return true
	}
	return false
}


func HttpPost(data []byte, url string) (*http.Response,error){

	request, err := http.NewRequest("POST", url,bytes.NewReader(data) )
	if err != nil {
		return nil,err
	}
	request.Header.Set("Content-Type", "application/json;charset=UTF-8")
	client := http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		fmt.Println(err.Error())
		return nil,err
	}
	return resp,nil
}