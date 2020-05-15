package main

import (
	"github.com/chriskirkland/go-xtract/_integration/nestedpkgfn/src/pkg"
)

func main() {
	pkg.Fn("string passed to function in nested pkg")
}
