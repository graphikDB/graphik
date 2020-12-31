package raft

import (
	"net"
	"os"
	"time"
)

type Options struct {
	raftDir                  string
	peerID                   string
	isLeader                 bool
	maxPool                  int
	timeout                  time.Duration
	retainSnapshots          int
	restoreSnapshotOnRestart bool
	advertise                net.Addr
	electionTimeout          time.Duration
	heartbeatTimeout         time.Duration
	commitTimeout            time.Duration
	leaseTimeout             time.Duration
	debug                    bool
}

func (o *Options) setDefaults() {
	if o.maxPool <= 0 {
		o.maxPool = 5
	}
	if o.timeout <= 0 {
		o.timeout = 5 * time.Second
	}
	if o.retainSnapshots == 0 {
		o.retainSnapshots = 1
	}
	if o.peerID == "" {
		h, _ := os.Hostname()
		o.peerID = hash([]byte(h))
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

func WithTimeout(timeout time.Duration) Opt {
	return func(o *Options) {
		o.timeout = timeout
	}
}

func WithElectionTimeout(timeout time.Duration) Opt {
	return func(o *Options) {
		o.electionTimeout = timeout
	}
}

func WithHeartbeatTimeout(timeout time.Duration) Opt {
	return func(o *Options) {
		o.heartbeatTimeout = timeout
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

func WithRestoreSnapshotOnRestart(restore bool) Opt {
	return func(o *Options) {
		o.restoreSnapshotOnRestart = restore
	}
}

func WithAdvertiseAddr(addr net.Addr) Opt {
	return func(o *Options) {
		o.advertise = addr
	}
}

func WithLeaderLeaseTimeout(timeout time.Duration) Opt {
	return func(o *Options) {
		o.leaseTimeout = timeout
	}
}

func WithCommitTimeout(timeout time.Duration) Opt {
	return func(o *Options) {
		o.commitTimeout = timeout
	}
}

func WithDebug(debug bool) Opt {
	return func(o *Options) {
		o.debug = debug
	}
}
