module github.com/videocoin/cloud-billing

go 1.13

replace github.com/videocoin/cloud-api => ../cloud-api

replace github.com/videocoin/cloud-pkg => ../cloud-pkg

require (
	github.com/go-playground/locales v0.13.0
	github.com/go-playground/universal-translator v0.17.0
	github.com/go-sql-driver/mysql v1.4.1
	github.com/gogo/protobuf v1.3.1
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/leodido/go-urn v1.2.0 // indirect
	github.com/mailru/dbr v3.0.0+incompatible
	github.com/sirupsen/logrus v1.4.2
	github.com/stripe/stripe-go v70.9.0+incompatible
	github.com/videocoin/cloud-api v0.0.17
	github.com/videocoin/cloud-pkg v0.0.5
	google.golang.org/grpc v1.21.1
	gopkg.in/go-playground/validator.v9 v9.31.0
)
