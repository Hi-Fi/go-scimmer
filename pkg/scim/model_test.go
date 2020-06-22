package scim

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestToJson(t *testing.T) {
	pathchOp := newPatchOp()
	pathchOp.Operations = []PatchOperation{
		{
			Op: "add",
			Value: Group{
				Members: []Member{
					{
						Value: "123",
					},
					{
						Value: "234",
					},
				},
			},
		},
	}

	marshalled, _ := json.Marshal(pathchOp)
	fmt.Println(string(marshalled))
}
