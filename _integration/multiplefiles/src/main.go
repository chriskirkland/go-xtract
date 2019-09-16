package pkg

import "fmt"

const (
	str1 = "this is a constant"

	str2 = `this is another constant`
)

var str3 = "variables work too"

func thing() {
	fmt.Println(str1)
}
