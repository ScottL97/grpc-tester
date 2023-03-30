package tester

import (
	"crypto/tls"
	"encoding/json"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"os"
	"path"
	"time"
)

type Config struct {
	CockroachDBConfig *CockroachDBConfig
	Host              string        `json:"host"`          // gRPC 服务端地址
	Concurrency       int           `json:"concurrency"`   // 并发量
	RequestsTimes     int           `json:"RequestsTimes"` // 请求次数
	Timeout           time.Duration `json:"timeout"`       // 超时时间
	SkipTLS           bool          `json:"skipTLS"`       // 是否跳过 TLS 验证
	InSecure          bool          `json:"inSecure"`      // gRPC 是否使用 InSecure 模式
	creds             credentials.TransportCredentials
	ConnsNum          int
	TestDir           string
}

type CockroachDBConfig struct {
	Host            string `json:"host"`            // 数据库 URL，不带端口
	Port            int    `json:"port"`            // 数据库端口
	User            string `json:"user"`            // 数据库用户名
	Password        string `json:"password"`        // 数据库密码
	Database        string `json:"database"`        // 数据库名称
	ApplicationName string `json:"applicationName"` // 数据库应用名
	SSLMode         string `json:"sslMode"`         // SSL 模式，可选值为 disable, allow, prefer, require, verify-ca, verify-full
}

func (c *Config) createClientTransportCredentials() credentials.TransportCredentials {
	if c.InSecure {
		return insecure.NewCredentials()
	}
	var tlsConf tls.Config
	tlsConf.InsecureSkipVerify = c.SkipTLS
	return credentials.NewTLS(&tlsConf)
}

func NewConfig(testDir string) *Config {
	fp, err := os.Open(path.Join(testDir, "config.json"))
	if err != nil {
		panic(err)
	}
	defer fp.Close()

	config := &Config{}
	decoder := json.NewDecoder(fp)
	err = decoder.Decode(config)
	if err != nil {
		panic(err)
	}

	config.TestDir = testDir
	config.creds = config.createClientTransportCredentials()
	return config
}

func (c *Config) ParseCockroachDBConfig() {
	fp, err := os.Open(path.Join(c.TestDir, "cockroachdb.json"))
	if err != nil {
		panic(err)
	}
	defer fp.Close()

	config := &CockroachDBConfig{}
	decoder := json.NewDecoder(fp)
	err = decoder.Decode(config)
	if err != nil {
		panic(err)
	}

	c.CockroachDBConfig = config
}
