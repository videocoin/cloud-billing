package service

import (
	"github.com/sirupsen/logrus"
)

type Config struct {
	Name    string        `envconfig:"-"`
	Version string        `envconfig:"-"`
	Logger  *logrus.Entry `envconfig:"-"`

	RPCAddr         string `default:"0.0.0.0:5008" envconfig:"RPC_ADDR"`
	AccountsRPCAddr string `default:"0.0.0.0:5001" envconfig:"ACCOUNTS_RPC_ADDR"`
	DBURI           string `default:"root:@/videocoin?charset=utf8&parseTime=True&loc=Local" envconfig:"DBURI"`
	MQURI           string `default:"amqp://guest:guest@127.0.0.1:5672" envconfig:"MQURI"`
	AuthTokenSecret string `default:"" envconfig:"AUTH_TOKEN_SECRET"`
}
