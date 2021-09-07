package main

import (
	"fmt"
	"os"

	"github.com/yusufsyaifudin/ngendika/cmd"
)

func main() {
	dir := os.Args[1]
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%+v", err)
		os.Exit(1)
	}

	cmd.GenerateDoc(dir)
}
