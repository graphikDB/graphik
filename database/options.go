package database

import (
	"context"
	"fmt"
	"github.com/autom8ter/machine"
	"github.com/graphikDB/graphik/helpers"
	"github.com/graphikDB/graphik/logger"
	"github.com/graphikDB/graphik/version"
)

type Options struct {
	storagePath                string
	raftSecret                 string
	machine                    *machine.Machine
	logger                     *logger.Logger
	requireRequestAuthorizers  bool
	requireResponseAuthorizers bool
	rootUsers                  []string
}

func (o *Options) SetDefaults() error {
	if o.storagePath == "" {
		o.storagePath = "/tmp/graphik"
	}
	if o.machine == nil {
		o.machine = machine.New(context.Background())
	}
	if o.logger == nil {
		o.logger = logger.New(false)
	}
	if o.raftSecret == "" {
		o.raftSecret = helpers.Hash([]byte(fmt.Sprintf("defaultRaftSecret_%s", version.Version)))
	}
	return nil
}

type Opt func(o *Options)

func WithStoragePath(storagePath string) Opt {
	return func(o *Options) {
		o.storagePath = storagePath
	}
}

func WithRaftSecret(raftSecret string) Opt {
	return func(o *Options) {
		o.raftSecret = raftSecret
	}
}

func WithMachine(mach *machine.Machine) Opt {
	return func(o *Options) {
		o.machine = mach
	}
}

func WithLogger(lgger *logger.Logger) Opt {
	return func(o *Options) {
		o.logger = lgger
	}
}

func WithRequireRequestAuthorizers(require bool) Opt {
	return func(o *Options) {
		o.requireRequestAuthorizers = require
	}
}

func WithRequireResponseAuthorizers(require bool) Opt {
	return func(o *Options) {
		o.requireResponseAuthorizers = require
	}
}

func WithRootUsers(rootUsers []string) Opt {
	return func(o *Options) {
		o.rootUsers = append(rootUsers, rootUsers...)
	}
}
