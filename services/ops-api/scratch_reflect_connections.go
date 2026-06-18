package main

import (
	"fmt"
	"reflect"

	"github.com/unified-to/unified-go-sdk/pkg/models/operations"
)

func main() {
	t := reflect.TypeOf(operations.ListUnifiedConnectionsRequest{})
	fmt.Printf("Fields of ListUnifiedConnectionsRequest:\n")
	for i := 0; i < t.NumField(); i++ {
		fmt.Println(t.Field(i).Name, t.Field(i).Type.String())
	}
}
