package shared

type ClientInfo struct {
	ClientLocalPath string
	ClientIP        string
}

type Reply struct {
	Connected bool
}
