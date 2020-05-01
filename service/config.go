package service

import (
	"github.com/sirupsen/logrus"
)

type Config struct {
	Name    string        `envconfig:"-"`
	Version string        `envconfig:"-"`
	Logger  *logrus.Entry `envconfig:"-"`

	RPCAddr              string `envconfig:"RPC_ADDR" default:"0.0.0.0:5020"`
	PRPCAddr             string `envconfig:"PRPC_ADDR" default:"0.0.0.0:5120"`
	StripeHookServerAddr string `envconfig:"STRIPE_HOOK_SERVER_ADDR" default:"0.0.0.0:5028"`
	UsersRPCAddr         string `envconfig:"USERS_RPC_ADDR" default:"0.0.0.0:5000"`
	AccountsRPCAddr      string `envconfig:"ACCOUNTS_RPC_ADDR" default:"0.0.0.0:5001"`
	EmitterRPCAddr       string `envconfig:"EMITTER_RPC_ADDR" default:"0.0.0.0:5003"`
	DBURI                string `envconfig:"DBURI" default:"root:@/videocoin?charset=utf8&parseTime=True&loc=Local"`
	MQURI                string `envconfig:"MQURI" default:"amqp://guest:guest@127.0.0.1:5672"`
	AuthTokenSecret      string `envconfig:"AUTH_TOKEN_SECRET" default:"secret"`
	StripeKey            string `envconfig:"STRIPE_KEY" required:"true"`
	StripeWHSecret       string `envconfig:"STRIPE_WH_SECRET" required:"true"`
}
