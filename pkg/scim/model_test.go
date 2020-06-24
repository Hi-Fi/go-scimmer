package scim

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestToJson(t *testing.T) {
	pathchOp := newPatchOp("add", Group{
		Members: []Member{
			{
				Value: "123",
			},
			{
				Value: "234",
			},
		},
	})

	marshalled, _ := json.Marshal(pathchOp)
	fmt.Println(string(marshalled))
}

func TestEmptyValueToJson(t *testing.T) {
	pathchOp := newPatchOp("replace", Group{
		Members: []Member{},
	})

	marshalled, _ := json.Marshal(pathchOp)
	fmt.Println(string(marshalled))
}
