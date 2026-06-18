package main

import (
	"fmt"
	"reflect"

	unified "github.com/unified-to/unified-go-sdk"
)

func main() {
	sdk := unified.New()
	t := reflect.TypeOf(sdk.Commerce)
	fmt.Printf("Commerce methods:\n")
	for i := 0; i < t.NumMethod(); i++ {
		fmt.Println(t.Method(i).Name)
	}

	t2 := reflect.TypeOf(sdk.Accounting)
	fmt.Printf("\nAccounting methods:\n")
	for i := 0; i < t2.NumMethod(); i++ {
		fmt.Println(t2.Method(i).Name)
	}
}
