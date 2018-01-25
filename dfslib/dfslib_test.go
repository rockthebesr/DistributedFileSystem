package dfslib

import (
	"fmt"
	"testing"
)

func TestSomething(t *testing.T) {
	serverAddr := "127.0.0.1:8080"
	localIP := "127.0.0.1"
	localPath := "dfs"
	dfsInstance, error := MountDFS(serverAddr, localIP, localPath)
	fmt.Println(dfsInstance)
	fmt.Println(error)
	// println(serverAddr)
	// println(localIP)
	// println(localPath)
}
