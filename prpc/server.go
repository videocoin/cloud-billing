package prpc

import (
	"net"

	"github.com/sirupsen/logrus"
	pv1 "github.com/videocoin/cloud-api/billing/private/v1"
	"github.com/videocoin/cloud-billing/manager"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	logger *logrus.Entry
	addr   string
	grpc   *grpc.Server
	listen net.Listener
	dm     *manager.Manager
}

func NewServer(opts ...Option) (*Server, error) {
	s := &Server{
		grpc: grpc.NewServer(),
	}
	for _, o := range opts {
		if err := o(s); err != nil {
			return nil, err
		}
	}

	pv1.RegisterBillingServiceServer(s.grpc, s)
	reflection.Register(s.grpc)

	return s, nil
}

func (s *Server) Start() error {
	s.logger.Infof("starting private rpc server on %s", s.addr)
	return s.grpc.Serve(s.listen)
}
