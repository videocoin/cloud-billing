package service

import (
	"github.com/sirupsen/logrus"
)

type Config struct {
	Name    string        `envconfig:"-"`
	Version string        `envconfig:"-"`
	Logger  *logrus.Entry `envconfig:"-"`

	RPCAddr               string `default:"0.0.0.0:5020" envconfig:"RPC_ADDR"`
	AccountsRPCAddr       string `default:"0.0.0.0:5001" envconfig:"ACCOUNTS_RPC_ADDR"`
	DBURI                 string `default:"root:@/cloud?charset=utf8&parseTime=True&loc=Local" envconfig:"DBURI"`
	MQURI                 string `default:"amqp://guest:guest@127.0.0.1:5672" envconfig:"MQURI"`
	AuthTokenSecret       string `default:"" envconfig:"AUTH_TOKEN_SECRET"`
	StripeKey             string `envconfig:"STRIPE_KEY" required:"true"`
	StripeBaseCallbackURL string `envconfig:"STRIPE_BASE_CALLBACK_URL" required:"true"`
}
