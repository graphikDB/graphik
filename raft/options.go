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
	listenAddr      string
	maxPool         int
	timeout         time.Duration
	retainSnapshots int
}

func (o *Options) setDefaults() {
	if o.listenAddr == "" {
		o.listenAddr = "localhost:7822"
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
		o.peerID = helpers.Hash([]byte(fmt.Sprintf("%s%s", h, o.listenAddr)))
	}
	if o.raftDir == "" {
		o.raftDir = fmt.Sprintf("/tmp/graphik/raft/%s", o.peerID)
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

func WithListenAddr(addr string) Opt {
	return func(o *Options) {
		o.listenAddr = addr
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
