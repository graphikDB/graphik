package vm

import "github.com/pkg/errors"

type VM struct {
	edgeVM    *EdgeVM
	nodeVM    *NodeVM
	messageVM *MessageVM
	changeVM  *ChangeVM
	authVm    *AuthVM
}

func NewVM() (*VM, error) {
	edge, err := NewEdgeVM()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create edge vm")
	}
	node, err := NewNodeVM()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create node vm")
	}
	message, err := NewMessageVM()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create message vm")
	}
	change, err := NewChangeVM()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create change vm")
	}
	auth, err := NewAuthVM()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create auth vm")
	}
	return &VM{edgeVM: edge, nodeVM: node, messageVM: message, changeVM: change, authVm: auth}, nil
}

func (v *VM) Edge() *EdgeVM {
	return v.edgeVM
}

func (v *VM) Node() *NodeVM {
	return v.nodeVM
}

func (v *VM) Message() *MessageVM {
	return v.messageVM
}

func (v *VM) Change() *ChangeVM {
	return v.changeVM
}

func (v *VM) Auth() *AuthVM {
	return v.authVm
}
