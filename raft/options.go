package raft

import (
	"fmt"
	"github.com/graphikDB/graphik/helpers"
	"os"
	"time"
)

type Options struct {
	raftDir         string
	peerID          string
	isLeader        bool
	port            int
	maxPool         int
	timeout         time.Duration
	retainSnapshots int
}

func (o *Options) setDefaults() {
	if o.port == 0 {
		o.port = 7810
	}
	if o.maxPool <= 0 {
		o.maxPool = 5
	}
	if o.timeout == 0 {
		o.timeout = 5 * time.Second
	}
	if o.retainSnapshots == 0 {
		o.retainSnapshots = 1
	}
	if o.peerID == "" {
		h, _ := os.Hostname()
		o.peerID = helpers.Hash([]byte(fmt.Sprintf("%s%v", h, o.port)))
	}
	if o.raftDir == "" {
		o.raftDir = "/tmp/graphik/raft"
	}
	os.MkdirAll(o.raftDir, 0700)
}

type Opt func(o *Options)

func WithPeerID(peerID string) Opt {
	return func(o *Options) {
		o.peerID = peerID
	}
}

func WithIsLeader(isLeader bool) Opt {
	return func(o *Options) {
		o.isLeader = isLeader
	}
}

func WithListenPort(port int) Opt {
	return func(o *Options) {
		o.port = port
	}
}

func WithTimeout(timeout time.Duration) Opt {
	return func(o *Options) {
		o.timeout = timeout
	}
}

func WithMaxPool(max int) Opt {
	return func(o *Options) {
		o.maxPool = max
	}
}

func WithSnapshotRetention(retention int) Opt {
	return func(o *Options) {
		o.retainSnapshots = retention
	}
}

func WithRaftDir(dir string) Opt {
	return func(o *Options) {
		o.raftDir = dir
	}
}
