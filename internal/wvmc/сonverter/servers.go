package converter

import (
	"errors"

	"github.com/anaxita/wvmc/internal/wvmc/control"
	"github.com/anaxita/wvmc/internal/wvmc/model"
)

// ServerVM получает слайсы серверов и виртуалок, возвращает сервера, в которых значения свойств state и network заполнены полями из слайса ВМ.
// Слайсы должны быть одной длины
func ServerVM(servers *[]model.Server, vms []control.VM) error {
	if len(*servers) != len(vms) {
		return errors.New("слайсы должны быть одинаковой длины")
	}

	for k, v := range *servers {
		v.State = vms[k].Status
		v.Network = vms[k].Network
	}

	return nil
}
