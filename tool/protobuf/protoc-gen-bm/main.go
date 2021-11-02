package main

import (
	"flag"
	"fmt"
	"os"

	"kratos/tool/protobuf/pkg/gen"
	"kratos/tool/protobuf/pkg/generator"
	bmgen "kratos/tool/protobuf/protoc-gen-bm/generator"
)

func main() {
	versionFlag := flag.Bool("version", false, "print version and exit")
	flag.Parse()
	if *versionFlag {
		fmt.Println(generator.Version)
		os.Exit(0)
	}

	g := bmgen.BmGenerator()
	gen.Main(g)
}
