package database

import (
	"context"
	"fmt"
	apipb "github.com/graphikDB/graphik/gen/grpc/go"
	"github.com/graphikDB/raft/fsm"
	"github.com/hashicorp/raft"
	"github.com/pkg/errors"
	"go.etcd.io/bbolt"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"io"
	"time"
)

func (g *Graph) fsm() *fsm.FSM {
	return &fsm.FSM{
		ApplyFunc: func(log *raft.Log) interface{} {
			start := time.Now()
			var cmd = &apipb.RaftCommand{}

			if err := proto.Unmarshal(log.Data, cmd); err != nil {
				return err
			}
			defer func() {
				g.logger.Debug("applied raft log",
					zap.Duration("dur", time.Since(start)),
					zap.String("method", cmd.Method),
					zap.String("user", cmd.User.GetRef().GetGid()),
				)
			}()
			ctx := g.methodToContext(context.Background(), cmd.Method)
			ctx, usr, err := g.userToContext(ctx, cmd.User.GetAttributes().AsMap())
			if err != nil {
				return errors.Wrap(err, "raft: userToContext")
			}
			cmd.User = usr
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()
			if err := g.db.Update(func(tx *bbolt.Tx) error {
				if cmd.GetSetAuthorizers() != nil {
					for _, a := range cmd.GetSetAuthorizers().GetAuthorizers() {
						_, err := g.setAuthorizer(ctx, tx, a)
						if err != nil {
							return errors.Wrap(err, "raft: setAuthorizer")
						}
					}
				}
				if cmd.GetSetTypeValidators() != nil {
					for _, a := range cmd.GetSetTypeValidators().GetValidators() {
						_, err := g.setTypedValidator(ctx, tx, a)
						if err != nil {
							return errors.Wrap(err, "raft: setTypedValidator")
						}
					}
				}
				if cmd.SetIndexes != nil {
					for _, a := range cmd.GetSetIndexes().GetIndexes() {
						_, err := g.setIndex(ctx, tx, a)
						if err != nil {
							return errors.Wrap(err, "raft: setIndex")
						}
					}
				}
				if cmd.SetTriggers != nil {
					for _, a := range cmd.GetSetTriggers().GetTriggers() {
						_, err := g.setTrigger(ctx, tx, a)
						if err != nil {
							return errors.Wrap(err, "raft: setTrigger")
						}
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
					connections, err := g.setConnections(ctx, tx, cmd.SetConnections...)
					if err != nil {
						return errors.Wrap(err, "raft: setConnections")
					}
					cmd.SetConnections = connections.Connections
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
			if cmd.SetAuthorizers != nil {
				if err := g.cacheAuthorizers(); err != nil {
					return err
				}
			}
			if cmd.SetTypeValidators != nil {
				if err := g.cacheTypeValidators(); err != nil {
					return err
				}
			}
			if cmd.SetIndexes != nil {
				if err := g.cacheIndexes(); err != nil {
					return err
				}
			}
			if cmd.SetTriggers != nil {
				if err := g.cacheTriggers(); err != nil {
					return err
				}
			}
			if cmd.GetSendMessage() != nil {
				if err := g.machine.PubSub().Publish(cmd.SendMessage.Channel, cmd.SendMessage); err != nil {
					return status.Error(codes.Internal, err.Error())
				}
			}
			return cmd
		},
		SnapshotFunc: func() (*fsm.Snapshot, error) {
			g.logger.Info("raft: snapshot unimplemented")
			return nil, fmt.Errorf("raft: snapshot unimplemented")
		},
		RestoreFunc: func(closer io.ReadCloser) error {
			g.logger.Info("raft: restore unimplemented")
			return fmt.Errorf("raft: restore unimplemented")
		},
	}
}

func (g *Graph) applyCommand(rft *apipb.RaftCommand) (*apipb.RaftCommand, error) {
	if rft == nil {
		return nil, errors.New("empty raft command")
	}
	bits, err := proto.Marshal(rft)
	if err != nil {
		return nil, err
	}
	i, err := g.raft.Apply(bits)
	if err != nil {
		return nil, err
	}
	return i.(*apipb.RaftCommand), nil
}
