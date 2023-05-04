package requests

import (
	"github.com/anaxita/wvmc/internal/entity"
)

type ControlServer struct {
	ServerID int64          `json:"server_id"`
	Command  entity.Command `json:"command"`
}
