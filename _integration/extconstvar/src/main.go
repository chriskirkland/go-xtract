package main

import (
	"fmt"
	"github.com/chriskirkland/go-xtract/_integration/extconstvar/src/pkg"
)

const (
	str1 = "this is a constant"
)

func main() {
	fmt.Println(pkg.Constant)
	other()
}

func other() {
	fmt.Println(str1)
	fmt.Println(pkg.Variable)
}
