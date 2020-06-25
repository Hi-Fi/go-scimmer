package scim

import (
	"encoding/json"
	"testing"

	"github.com/hi-fi/go-scimmer/pkg/model"
	"github.com/stretchr/testify/assert"
)

func TestPatchOpCreation(t *testing.T) {
	expected := `{
	"schemas": [
		"urn:ietf:params:scim:api:messages:2.0:PatchOp"
	],
	"Operations": [
		{
			"op": "replace",
			"value": {
				"members": [
					{
						"value": "test@test1.com"
					},
					{
						"value": "test@test2.com"
					}
				]
			}
		}
	]
}`
	modelGroup := model.Group{
		CommonName: "TestGroup1",
		Members:    []string{"test@test1.com", "test@test2.com"},
	}

	scimGroup := newGroup(&modelGroup)

	patchRequest := newPatchOp("replace", Group{
		Members: scimGroup.Members,
	})
	marshalled, _ := json.MarshalIndent(patchRequest, "", "\t")
	assert.Equal(t, expected, string(marshalled), "Requests should be same.")
}
