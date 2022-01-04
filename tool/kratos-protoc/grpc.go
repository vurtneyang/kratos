package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
)

const (
	_getGRPCGen = "go get -u github.com/gogo/protobuf/protoc-gen-gofast"
	_grpcProtoc = "protoc --proto_path=%s --proto_path=%s --proto_path=%s --gofast_out=plugins=grpc," +
		"Mgoogle/protobuf/any.proto=github.com/gogo/protobuf/types," +
		"Mgoogle/protobuf/duration.proto=github.com/gogo/protobuf/types," +
		"Mgoogle/protobuf/struct.proto=github.com/gogo/protobuf/types," +
		"Mgoogle/protobuf/timestamp.proto=github.com/gogo/protobuf/types," +
		"Mgoogle/protobuf/wrappers.proto=github.com/gogo/protobuf/types:."
)

func installGRPCGen() error {
	if _, err := exec.LookPath("protoc-gen-gofast"); err != nil {
		if err := goget(_getGRPCGen); err != nil {
			return err
		}
	}
	return nil
}

func genGRPC(files []string) error {
	pwd, _ := os.Getwd()
	gosrc := path.Join(gopath(), "src")
	ext, err := latestKratos()
	if err != nil {
		return err
	}
	line := fmt.Sprintf(_grpcProtoc, gosrc, ext, pwd)
	log.Println(line, strings.Join(files, " "))
	args := strings.Split(line, " ")
	args = append(args, files...)
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = pwd
	cmd.Env = os.Environ()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
