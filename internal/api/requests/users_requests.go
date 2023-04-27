package requests

type AddServers struct {
	ServerIDs []int64 `json:"servers"`
}
