package pkg

import (
	"fmt"
	"os"
)

func Fn(s string) {
	fmt.Fprintf(os.Stderr, "Fn(%q)\n", s)
}
