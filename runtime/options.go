package runtime

import (
	"github.com/autom8ter/graphik/jwks"
	"github.com/autom8ter/machine"
)

type Opts struct {
	localID  string
	raftDir  string
	bindAddr string
	leader   bool
	machine  *machine.Machine
	jwks     *jwks.Auth
}

type Opt func(s *Opts)

func WithID(id string) Opt {
	return func(s *Opts) {
		s.localID = id
	}
}

func WithRaftDir(dir string) Opt {
	return func(s *Opts) {
		s.raftDir = dir
	}
}

func WithBindAddr(addr string) Opt {
	return func(s *Opts) {
		s.bindAddr = addr
	}
}

func WithLeader(leader bool) Opt {
	return func(s *Opts) {
		s.leader = leader
	}
}

func WithMachine(machine *machine.Machine) Opt {
	return func(s *Opts) {
		s.machine = machine
	}
}


func WithJWKS(jwks *jwks.Auth) Opt {
	return func(s *Opts) {
		s.jwks = jwks
	}
}
