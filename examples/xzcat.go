package main

import (
	"fmt"
	"io"
	"os"
	xz "github.com/remyoudompheng/go-liblzma"
)

func main() {
	dec, er := xz.NewReader(os.Stdin)
	if er != nil {
		fmt.Println(er)
		os.Exit(1)
	}

	io.Copy(os.Stdout, dec)
}
