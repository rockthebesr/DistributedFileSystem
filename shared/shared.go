package shared

type ClientInfo struct {
	ClientLocalPath string
	ClientIP        string
}

type Reply struct {
	Connected bool
}

type FileName struct {
	FileName string
}

type FileExists struct {
	FileExists bool
}

type NotifyNewFile struct {
	FileName string
	ClientID int
}

func Contains(s []int, e int) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
