package dfslib

import (
	"fmt"
	"os"
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
	localPath := "/Users/luorock/Documents/UBCfolder/year4/cpsc416/a2_file_dir_0/"
	dfs, _ := MountDFS(serverAddr, localIP, localPath)
	localResult, _ := dfs.LocalFileExists("helloworld")
	fmt.Println("does hello world exist locally?")
	fmt.Println(localResult)
	_, err := dfs.Open("helloworld", READ)
	if err != nil {
		fmt.Println(err)
		return
	}
	result, err := dfs.GlobalFileExists("helloworld")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("does hello world exist globally?")
	fmt.Println(result)
	localResult, _ = dfs.LocalFileExists("helloworld")
	fmt.Println("does hello world exist locally?")
	fmt.Println(localResult)
}

func TestOpenWriteAndReadFile1(*testing.T) {
	serverAddr := "127.0.0.1:8080"
	localIP := "127.0.0.1"
	localPath := "/Users/luorock/Documents/UBCfolder/year4/cpsc416/a2_file_dir_0/"
	dfs, _ := MountDFS(serverAddr, localIP, localPath)
	localResult, _ := dfs.LocalFileExists("helloworld")
	if localResult {
		fmt.Println("delete a2_file_dir_0/helloworld before running this test")
		return
	}
	file, _ := dfs.Open("helloworld", WRITE)
	byteArray := [32]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	var chunk Chunk
	copy(chunk[:], byteArray[:])
	err := file.Write(0, &chunk)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = file.Read(0, &chunk)
	fmt.Println(chunk)
	if err != nil {
		fmt.Println(err)
		return
	}
	file.Close()
	dfs.UMountDFS()
	os.Remove("/Users/luorock/Documents/UBCfolder/year4/cpsc416/a2_file_dir_0/helloworld.dfs")
}

func TestOpenDread(*testing.T) {
	serverAddr := "127.0.0.1:8080"
	localIP := "127.0.0.1"
	localPath := "/Users/luorock/Documents/UBCfolder/year4/cpsc416/a2_file_dir_1/"
	dfs, _ := MountDFS(serverAddr, localIP, localPath)
	localResult, _ := dfs.LocalFileExists("helloworld")
	fmt.Println("does hello world exist locally?")
	fmt.Println(localResult)
	file, err := dfs.Open("helloworld", DREAD)
	if err != nil {
		fmt.Println(err)
		return
	}
	file.Close()
}
