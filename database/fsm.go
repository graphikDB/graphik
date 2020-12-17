package database

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	apipb "github.com/graphikDB/graphik/gen/grpc/go"
	"github.com/graphikDB/graphik/raft/fsm"
	"github.com/hashicorp/raft"
	"go.etcd.io/bbolt"
	"io"
	"time"
)

func init() {
	gob.Register(&raftCommand{})
}

type raftCommand struct {
	setDocs        []*apipb.Doc
	setConnections []*apipb.Connection
	delDocs        []*apipb.Ref
	delConnections []*apipb.Ref
}

func (g *Graph) fsm() *fsm.FSM {
	return &fsm.FSM{
		ApplyFunc: func(log *raft.Log) interface{} {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			var cmd raftCommand
			if err := gob.NewDecoder(bytes.NewBuffer(log.Data)).Decode(&cmd); err != nil {
				return err
			}
			if err := g.db.Batch(func(tx *bbolt.Tx) error {
				if len(cmd.setDocs) > 0 {
					docs, err := g.setDocs(ctx, tx, cmd.setDocs...)
					if err != nil {
						return err
					}
					cmd.setDocs = docs.GetDocs()
				}
				if len(cmd.setConnections) > 0 {
					connections, err := g.setConnections(ctx, tx, cmd.setConnections...)
					if err != nil {
						return err
					}
					cmd.setConnections = connections.GetConnections()
				}
				if len(cmd.delConnections) > 0 {
					for _, r := range cmd.delConnections {
						err := g.delConnection(ctx, tx, r)
						if err != nil {
							return err
						}
					}
				}
				if len(cmd.delDocs) > 0 {
					for _, r := range cmd.delDocs {
						err := g.delDoc(ctx, tx, r)
						if err != nil {
							return err
						}
					}
				}
				return nil
			}); err != nil {
				return err
			}
			return cmd
		},
		SnapshotFunc: func() (*fsm.Snapshot, error) {
			return nil, fmt.Errorf("raft: snapshot unimplemented")
		},
		RestoreFunc: func(closer io.ReadCloser) error {
			return fmt.Errorf("raft: restore unimplemented")
		},
	}
}
