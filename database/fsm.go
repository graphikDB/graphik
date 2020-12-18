package database

import (
	"context"
	"fmt"
	apipb "github.com/graphikDB/graphik/gen/grpc/go"
	"github.com/graphikDB/graphik/logger"
	"github.com/graphikDB/graphik/raft/fsm"
	"github.com/hashicorp/raft"
	"github.com/pkg/errors"
	"go.etcd.io/bbolt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
	"io"
	"time"
)

func (g *Graph) fsm() *fsm.FSM {
	return &fsm.FSM{
		ApplyFunc: func(log *raft.Log) interface{} {
			var cmd = &apipb.RaftCommand{}
			if err := proto.Unmarshal(log.Data, cmd); err != nil {
				return err
			}
			ctx := g.methodToContext(context.Background(), cmd.Method)
			ctx, usr, err := g.userToContext(ctx, cmd.User.GetAttributes().AsMap())
			if err != nil {
				return errors.Wrap(err, "raft: userToContext")
			}
			cmd.User = usr
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()
			if msg := cmd.GetSendMessage(); msg != nil {
				if err := g.machine.PubSub().Publish(msg.Channel, &apipb.Message{
					Channel:   msg.Channel,
					Data:      msg.Data,
					User:      cmd.User.GetRef(),
					Timestamp: timestamppb.Now(),
					Method:    g.getMethod(ctx),
				}); err != nil {
					return status.Error(codes.Internal, err.Error())
				}
			}
			if err := g.db.Update(func(tx *bbolt.Tx) error {
				if cmd.GetSetAuthorizers() != nil {
					for _, a := range cmd.GetSetAuthorizers().GetAuthorizers() {
						_, err := g.setAuthorizer(ctx, tx, a)
						if err != nil {
							return errors.Wrap(err, "raft: setAuthorizer")
						}
					}
					if err := g.cacheAuthorizers(); err != nil {
						return errors.Wrap(err, "raft: cacheAuthorizers")
					}
				}
				if cmd.GetSetTypeValidators() != nil {
					for _, a := range cmd.GetSetTypeValidators().GetValidators() {
						_, err := g.setTypedValidator(ctx, tx, a)
						if err != nil {
							return errors.Wrap(err, "raft: setTypedValidator")
						}
					}
					if err := g.cacheTypeValidators(); err != nil {
						return errors.Wrap(err, "raft: cacheTypeValidators")
					}
				}
				if cmd.SetIndexes != nil {
					for _, a := range cmd.GetSetIndexes().GetIndexes() {
						_, err := g.setIndex(ctx, tx, a)
						if err != nil {
							return errors.Wrap(err, "raft: setIndex")
						}
					}
					if err := g.cacheIndexes(); err != nil {
						return errors.Wrap(err, "raft: cacheIndexes")
					}
				}
				if len(cmd.GetSetDocs()) > 0 {
					docs, err := g.setDocs(ctx, tx, cmd.SetDocs...)
					if err != nil {
						return errors.Wrap(err, "raft: setDocs")
					}
					cmd.SetDocs = docs.GetDocs()
				}
				if len(cmd.GetSetConnections()) > 0 {
					connections, err := g.setConnections(ctx, tx, cmd.GetSetConnections()...)
					if err != nil {
						return errors.Wrap(err, "raft: setConnections")
					}
					cmd.SetConnections = connections.GetConnections()
				}
				if len(cmd.DelConnections) > 0 {
					for _, r := range cmd.DelConnections {
						err := g.delConnection(ctx, tx, r)
						if err != nil {
							return errors.Wrap(err, "raft: delConnection")
						}
					}
				}
				if len(cmd.DelDocs) > 0 {
					for _, r := range cmd.DelDocs {
						err := g.delDoc(ctx, tx, r)
						if err != nil {
							return errors.Wrap(err, "raft: delDoc")
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
			logger.Info("raft: snapshot unimplemented")
			return nil, fmt.Errorf("raft: snapshot unimplemented")
		},
		RestoreFunc: func(closer io.ReadCloser) error {
			logger.Info("raft: restore unimplemented")
			return fmt.Errorf("raft: restore unimplemented")
		},
	}
}

func (g *Graph) applyCommand(rft *apipb.RaftCommand) (*apipb.RaftCommand, error) {
	bits, err := proto.Marshal(rft)
	i, err := g.raft.Apply(bits)
	if err != nil {
		return nil, err
	}
	return i.(*apipb.RaftCommand), nil
}
