package dfslib

import (
	"fmt"
	"testing"
)

// func TestTwoClients(t *testing.T) {
// 	serverAddr := "127.0.0.1:8080"
// 	localIP := "127.0.0.1"
// 	localPath := "/Users/luorock/Documents/UBCfolder/year4/cpsc416/a2_file_dir_0"
// 	_, error := MountDFS(serverAddr, localIP, localPath)

// 	if error != nil {
// 		fmt.Println(error)
// 	}
// 	// localPath = "/Users/luorock/Documents/UBCfolder/year4/cpsc416/a2_file_dir_1"
// 	// _, error = MountDFS(serverAddr, localIP, localPath)
// 	// if error != nil {
// 	// 	fmt.Println(error)
// 	// }
// }

func TestLocalPath(t *testing.T) {
	serverAddr := "127.0.0.1:8080"
	localIP := "127.0.0.1"
	localPath := "/Users/luorock/Documents/UBCfolder/year4/cpsc416/a2_file_dir_1"
	_, error := MountDFS(serverAddr, localIP, localPath)
	fmt.Println(error)
}

func TestLocalFileExistst(*testing.T) {
	serverAddr := "127.0.0.1:8080"
	localIP := "127.0.0.1"
	localPath := "/Users/luorock/Documents/UBCfolder/year4/cpsc416/a2_file_dir_0/"
	dfs, error := MountDFS(serverAddr, localIP, localPath)
	_, error = dfs.LocalFileExists("helloworld")
	fmt.Println(error)
}

func TestOpenReadNewFile(*testing.T) {
	serverAddr := "127.0.0.1:8080"
	localIP := "127.0.0.1"
	localPath := "/Users/luorock/Documents/UBCfolder/year4/cpsc416/a2_file_dir_1/"
	dfs, _ := MountDFS(serverAddr, localIP, localPath)
	localResult, _ := dfs.LocalFileExists("helloworld")
	fmt.Println("does hello wolrd exist locally?")
	fmt.Println(localResult)
	dfs.Open("helloworld", READ)
	result, _ := dfs.GlobalFileExists("helloworld")
	fmt.Println("does hello wolrd exist globally?")
	fmt.Println(result)
	localResult, _ = dfs.LocalFileExists("helloworld")
	fmt.Println("does hello wolrd exist locally?")
	fmt.Println(localResult)
}
