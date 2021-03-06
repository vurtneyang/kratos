package main

import (
	"flag"
	"fmt"
	"os"

	"kratos/tool/protobuf/pkg/gen"
	"kratos/tool/protobuf/pkg/generator"
)

func main() {
	versionFlag := flag.Bool("version", false, "print version and exit")
	flag.Parse()
	if *versionFlag {
		fmt.Println(generator.Version)
		os.Exit(0)
	}

	g := NewSwaggerGenerator()
	gen.Main(g)
}
