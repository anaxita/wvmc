package requests

type AddServers struct {
	ServerIDs []string `json:"servers"`
}
