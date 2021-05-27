package broker

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"

	"gitee.com/godLei6/hmq/logger"
	"gitee.com/godLei6/hmq/plugins/auth"
	"gitee.com/godLei6/hmq/plugins/bridge"
	"go.uber.org/zap"
)

type Config struct {
	Worker   int       `json:"workerNum"`
	HTTPPort string    `json:"httpPort"`
	Host     string    `json:"host"`
	Port     string    `json:"port"`
	Cluster  RouteInfo `json:"cluster"`
	Router   string    `json:"router"`
	TlsHost  string    `json:"tlsHost"`
	TlsPort  string    `json:"tlsPort"`
	WsPath   string    `json:"wsPath"`
	WsPort   string    `json:"wsPort"`
	WsTLS    bool      `json:"wsTLS"`
	TlsInfo  TLSInfo   `json:"tlsInfo"`
	Debug    bool      `json:"debug"`
	Plugin   Plugins   `json:"plugins"`
}

type Plugins struct {
	Auth   auth.Auth
	Bridge bridge.BridgeMQ
}

type NamedPlugins struct {
	Auth   string
	Bridge string
}

type RouteInfo struct {
	Host string `json:"host"`
	Port string `json:"port"`
}

type TLSInfo struct {
	Verify   bool   `json:"verify"`
	CaFile   string `json:"caFile"`
	CertFile string `json:"certFile"`
	KeyFile  string `json:"keyFile"`
}

var DefaultConfig *Config = &Config{
	Worker: 4096,
	Host:   "0.0.0.0",
	Port:   "1883",
}

var (
	log = logger.Prod().Named("broker")
)

func showHelp() {
	fmt.Printf("%s\n", usageStr)
	os.Exit(0)
}

func ConfigureConfig(args []string) (*Config, error) {
	config := &Config{}
	var (
		help       bool
		configFile string
	)
	fs := flag.NewFlagSet("hmq-broker", flag.ExitOnError)
	fs.Usage = showHelp

	fs.BoolVar(&help, "h", false, "Show this message.")
	fs.BoolVar(&help, "help", false, "Show this message.")
	fs.IntVar(&config.Worker, "w", 1024, "worker num to process message, perfer (client num)/10.")
	fs.IntVar(&config.Worker, "worker", 1024, "worker num to process message, perfer (client num)/10.")
	fs.StringVar(&config.HTTPPort, "httpport", "8080", "Port to listen on.")
	fs.StringVar(&config.HTTPPort, "hp", "8080", "Port to listen on.")
	fs.StringVar(&config.Port, "port", "1883", "Port to listen on.")
	fs.StringVar(&config.Port, "p", "1883", "Port to listen on.")
	fs.StringVar(&config.Host, "host", "0.0.0.0", "Network host to listen on")
	fs.StringVar(&config.Cluster.Port, "cp", "", "Cluster port from which members can connect.")
	fs.StringVar(&config.Cluster.Port, "clusterport", "", "Cluster port from which members can connect.")
	fs.StringVar(&config.Router, "r", "", "Router who maintenance cluster info")
	fs.StringVar(&config.Router, "router", "", "Router who maintenance cluster info")
	fs.StringVar(&config.WsPort, "ws", "", "port for ws to listen on")
	fs.StringVar(&config.WsPort, "wsport", "", "port for ws to listen on")
	fs.StringVar(&config.WsPath, "wsp", "", "path for ws to listen on")
	fs.StringVar(&config.WsPath, "wspath", "", "path for ws to listen on")
	fs.StringVar(&configFile, "config", "", "config file for hmq")
	fs.StringVar(&configFile, "c", "", "config file for hmq")
	fs.BoolVar(&config.Debug, "debug", false, "enable Debug logging.")
	fs.BoolVar(&config.Debug, "d", false, "enable Debug logging.")

	fs.Bool("D", true, "enable Debug logging.")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	if help {
		showHelp()
		return nil, nil
	}

	fs.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "D":
			config.Debug = true
		}
	})

	if configFile != "" {
		tmpConfig, e := LoadConfig(configFile)
		if e != nil {
			return nil, e
		} else {
			config = tmpConfig
		}
	}

	if config.Debug {
		log = logger.Debug().Named("broker")
	}

	if err := config.check(); err != nil {
		return nil, err
	}

	return config, nil

}

func LoadConfig(filename string) (*Config, error) {

	content, err := ioutil.ReadFile(filename)
	if err != nil {
		// log.Error("Read config file error: ", zap.Error(err))
		return nil, err
	}
	// log.Info(string(content))

	var config Config
	err = json.Unmarshal(content, &config)
	if err != nil {
		// log.Error("Unmarshal config file error: ", zap.Error(err))
		return nil, err
	}

	return &config, nil
}


func (p *Plugins) UnmarshalJSON(b []byte) error {
	var named NamedPlugins
	err := json.Unmarshal(b, &named)
	if err != nil {
		return err
	}
	p.Auth = auth.NewAuth(named.Auth)
	p.Bridge = bridge.NewBridgeMQ(named.Bridge)
	return nil
}

func (config *Config) check() error {

	if config.Worker == 0 {
		config.Worker = 1024
	}

	if config.Port != "" {
		if config.Host == "" {
			config.Host = "0.0.0.0"
		}
	}

	if config.Cluster.Port != "" {
		if config.Cluster.Host == "" {
			config.Cluster.Host = "0.0.0.0"
		}
	}
	if config.Router != "" {
		if config.Cluster.Port == "" {
			return errors.New("cluster port is null")
		}
	}

	if config.TlsPort != "" {
		if config.TlsInfo.CertFile == "" || config.TlsInfo.KeyFile == "" {
			log.Error("tls config error, no cert or key file.")
			return errors.New("tls config error, no cert or key file.")
		}
		if config.TlsHost == "" {
			config.TlsHost = "0.0.0.0"
		}
	}
	return nil
}

var clientCert string = `-----BEGIN CERTIFICATE-----
MIIC3zCCAcegAwIBAgIBAjANBgkqhkiG9w0BAQsFADATMREwDwYDVQQDEwhNeVRl
c3RDQTAeFw0xNjEyMjYwMzA4MjNaFw0xNzEyMjYwMzA4MjNaMCIxDzANBgNVBAMM
BmNsb3VkMzEPMA0GA1UECgwGY2xpZW50MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8A
MIIBCgKCAQEA40Q+bMUjxCOdDcdC2jZaX8HuNCdm6Mu1rgj8ZfyTJIzsKtv00LYd
xfdhlNFj1uq8wi/zK/cB95wBpG1Loo/WicqSP2G/A7aPnzIBPj3zzP7HdyM5EaHW
zDWLzK+f0+MmAsrp7UW/zBR5O+ScnmIWm2H7KJY36dJPKllzzw+R6a4eJ6vthBcm
nueIYrhdXnunaWzkWQqAWlSZCzD8/MfTkgAPYW7OoS6aAQugTBzhHRo1meOVIT7u
y+hmZE4kE8V98Iy1rGPV5Uz/1vSEJziJGvQkyVr3gcAv5DwLWnyX1vOXKIKOgRAz
VYB4tbfEP5tQ0jimG5ErftF/sGYFNTRslQIDAQABoy8wLTAJBgNVHRMEAjAAMAsG
A1UdDwQEAwIHgDATBgNVHSUEDDAKBggrBgEFBQcDAjANBgkqhkiG9w0BAQsFAAOC
AQEAMtVzZdj1y/TLxP7KZcqkd/Z/vdW6moo12tahDHR4vPq0NdGaHADRfZHbCBmb
JEI9Qz3CKSdKZRZ4/A3ui/ZltbvkCao9ilmhQXXDT3Yz5hxk5ZBC9+Zs1IZmrsis
Qg/cdLUx4+ei/eR0OgWyC5D9AKNzshQExKBGGojedb98VcuS5ccJKrq0kVzZ/BZQ
k1EswNC9ifKBcqPIO1rTD9T3PB7dv9ZRpxwslmgYWWsqQu9x/dnOPEHJ1yXr7KJh
47NP7OrtG21la8EcAtjA3GXDjiZ+tXIR1RMbx1iuJAQPBeddWJCPbyofdNRL59BQ
caLVQ77r0hpvnNpkafa5QptyfQ==
-----END CERTIFICATE-----`

var clientKey string = `-----BEGIN RSA PRIVATE KEY-----
MIIEpQIBAAKCAQEA40Q+bMUjxCOdDcdC2jZaX8HuNCdm6Mu1rgj8ZfyTJIzsKtv0
0LYdxfdhlNFj1uq8wi/zK/cB95wBpG1Loo/WicqSP2G/A7aPnzIBPj3zzP7HdyM5
EaHWzDWLzK+f0+MmAsrp7UW/zBR5O+ScnmIWm2H7KJY36dJPKllzzw+R6a4eJ6vt
hBcmnueIYrhdXnunaWzkWQqAWlSZCzD8/MfTkgAPYW7OoS6aAQugTBzhHRo1meOV
IT7uy+hmZE4kE8V98Iy1rGPV5Uz/1vSEJziJGvQkyVr3gcAv5DwLWnyX1vOXKIKO
gRAzVYB4tbfEP5tQ0jimG5ErftF/sGYFNTRslQIDAQABAoIBAQCT4CzKM4AxOIcR
lw0t1V36nsIy10yDv0EI67nnVnAbwUJOJO7n+wfmby/kWFahWf3WUMLmYYO7LJx4
89DaBsOuxstgSGa0sM5E5JGggUkoosMBBz8z9N1B5LmBRuk1QsDR4lxR0ieZT90O
lpM+D07sbdWxtATPtNNkF+5d1aC4riPaNenwPXdb88bamcqCcARExwNxVhUogu88
frBeIfBdvNTZTmsqiqWrmAm4l1QnoQ1kCd3br4vbOlI4aQZCAPhECMaBSM7soNax
6XHUAA35vB3njNgvQYb6X2HvfktenwKXxDKDm7T8E6Ckof0kySncu2tpcIU/aHi4
QxS2TenhAoGBAPeU48RIbKYt158xmBbiY6EzMRHI1mq+iItiYGwjt43td4l1nEX+
UVGNnRJDffPPWIwNabPnOw9ZClwyEWgkJNJ/OS542B5QtFA5don5uAiX5OZCtQ6/
jyedC2HLq+e4No00pBkko3sVKbUHD98qRd45PFwhC34HJGjzxj3C16ZdAoGBAOr+
hN/2JSDOW+0dpbwVJUT1u3Ir9nOFZ3N5LDgvkE7dosKaHY9+AtUjMvhJ8Vea3jbJ
3VZrmacVtOPrrbsVWeacibqdDvRkPbQeg8vJjymLAFuvpv9WI6rih7PoiiG0HfSR
8aS14QTId31+6d9vgH/oWQuqNcTnqG3xWkK8HAuZAoGBAKE0INm9DoFld+//qre7
0IM1gc3Cp1n5lY6sD3xaBTo0VJD8MzSf0vL28j7iEzCc4VrPoPOyq5HiuAwvzYWx
gwhMLj9ED/QtODrEL5rHLjzqKfCDnsBrmhqA9thGdTf7igmHLRHx+UA7F1z3rC3y
qGt5eQPDwGfe3qY3k+zC4QdBAoGBANvaF3J5FS9mITbr39zhY6bqx93/J2nYy3qL
SUWfqkE+tkGecj2HRRsm/U6xzyuI5pEXtw5dSLm7YytBmZ5IUX2hwnFm81DOX7Qe
QGvuPRQ+yaz93x1P97quiQtWabUykDv6NrtEtisFalVs4V17Mht4w6ZYLknz+e4y
OaHp38sxAoGAF2ZBRadUjrfYN+BKJxekdvLEzGlRICvBRdB6vDfJPNULp1cVIzbF
rNhpjJJb53OvSJwI6OwRt5ehfIg1sRjoSYXhE6yJyEBQRIRdPLbxSAaQB20P9ZlL
blA0kLm6HiGNSu1CTAst23i2WueGQgOHHdBQoLUU5xEBNFYB2S7OB74=
-----END RSA PRIVATE KEY-----`
func GenerateTLSConfig() *tls.Config {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		panic(err)
	}
	template := x509.Certificate{SerialNumber: big.NewInt(1)}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		panic(err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		panic(err)
	}
	return &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		ClientAuth: tls.RequestClientCert,
		//VerifyPeerCertificate: func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
		//	clientCertificate, err := tls.X509KeyPair([]byte(clientCert), []byte(clientKey))
		//	if err != nil {
		//		fmt.Printf("X509KeyPair err=%+v\n",err)
		//	}
		//	cert, err:= x509.ParseCertificate(rawCerts[0])
		//	//cert2,err:= x509.ParseCertificate(clientCertificate.Certificate[0])
		//	//if err != nil {
		//	//	fmt.Printf("ParseCertificate err=%+v\n",err)
		//	//}
		//	//fmt.Printf("cert true=%+v||getcert=%+v||verifiedChains=%+v||len=%d\n",
		//	//	cert,cert2,verifiedChains,len(verifiedChains))
		//	fmt.Printf("cert true=%+v||getcert=%+v\n",
		//		len(clientCertificate.Certificate),cert)
		//	return nil
		//}}
	}
}

func NewTLSConfig(tlsInfo TLSInfo) (*tls.Config, error) {

	cert, err := tls.LoadX509KeyPair(tlsInfo.CertFile, tlsInfo.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("error parsing X509 certificate/key pair: %v", zap.Error(err))
	}
	//cert.Leaf, err = x509.ParseCertificate(cert.Certificate[0])
	//if err != nil {
	//	return nil, fmt.Errorf("error parsing certificate: %v", zap.Error(err))
	//}

	// Create TLSConfig
	// We will determine the cipher suites that we prefer.
	config := tls.Config{
		Certificates: []tls.Certificate{cert},
		//MinVersion:   tls.VersionTLS12,
	}

	// Require client certificates as needed
	//if tlsInfo.Verify {
	//	config.ClientAuth = tls.RequireAndVerifyClientCert
	//} else {
	//	config.ClientAuth = tls.NoClientCert
	//}
	//config.InsecureSkipVerify = true
	// Add in CAs if applicable.
	//if tlsInfo.CaFile != "" && false {
	//	rootPEM, err := ioutil.ReadFile(tlsInfo.CaFile)
	//	if err != nil || rootPEM == nil {
	//		return nil, err
	//	}
	//	pool := x509.NewCertPool()
	//	ok := pool.AppendCertsFromPEM([]byte(rootPEM))
	//	if !ok {
	//		return nil, fmt.Errorf("failed to parse root ca certificate")
	//	}
	//	config.ClientCAs = pool
	//}

	return &config, nil
}
