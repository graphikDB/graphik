package vm

import "github.com/pkg/errors"

type VM struct {
	connectionVM *ConnectionVM
	docVM        *DocVM
	messageVM    *MessageVM
	changeVM     *ChangeVM
	authVm       *AuthVM
}

func NewVM() (*VM, error) {
	connection, err := NewConnectionVM()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create connection vm")
	}
	doc, err := NewDocVM()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create doc vm")
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
	return &VM{connectionVM: connection, docVM: doc, messageVM: message, changeVM: change, authVm: auth}, nil
}

func (v *VM) Connection() *ConnectionVM {
	return v.connectionVM
}

func (v *VM) Doc() *DocVM {
	return v.docVM
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
