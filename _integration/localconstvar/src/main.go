package main

import "fmt"

const (
	str1 = "this is a constant"

	str2 = `this is another constant`
)

var str3 = "variables work too"

func main() {
	fmt.Println(str1)
	other()
}

func other() {
	fmt.Println(str2)
	fmt.Println(str3)
}
