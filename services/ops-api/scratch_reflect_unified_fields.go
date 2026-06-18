package main

import (
	"fmt"
	"reflect"

	unified "github.com/unified-to/unified-go-sdk"
)

func main() {
	sdk := unified.New()
	t := reflect.TypeOf(sdk)
	fmt.Printf("SDK fields:\n")
	for i := 0; i < t.Elem().NumField(); i++ {
		fmt.Println(t.Elem().Field(i).Name)
	}
}
