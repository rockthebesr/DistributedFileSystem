package shared

type ClientInfo struct {
	ClientLocalPath string
	ClientAddr      string
}

type ClientID struct {
	ClientID int
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

type FileNameAndClientID struct {
	FileName string
	ClientID int
}

type FileNameAndChunkNumberAndClientID struct {
	FileName    string
	ChunkNumber int
	ClientID    int
}

type FileData struct {
	FileData      [32 * 256]byte
	ChunkVersions [256]int
}

type ChunkData struct {
	ChunkData [32]byte
}

func Contains(s []int, e int) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
