package scim

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/hi-fi/go-scimmer/pkg/model"
)

func TestPatchOpCreation(t *testing.T) {
	modelGroup := model.Group{
		CommonName: "TestGroup1",
		Members:    []string{"test@test1.com", "test@test2.com"},
	}

	scimGroup := newGroup(&modelGroup)

	patchRequest := newPatchOp("replace", Group{
		Members: scimGroup.Members,
	})
	marshalled, _ := json.MarshalIndent(patchRequest, "", "    ")
	fmt.Println(string(marshalled))

}

// func TestBulkRequestCreation(t *testing.T) {
// 	modelUsers := []model.User{
// 		{
// 			FirstName: "Test",
// 			LastName:  "Test1",
// 			Username:  "test@test1.com",
// 			Email:     "test@test1.com",
// 		},
// 		{
// 			FirstName: "Test",
// 			LastName:  "Test2",
// 			Username:  "test@test2.com",
// 			Email:     "test@test2.com",
// 		},
// 	}

// 	modelGroups := []model.Group{
// 		{
// 			CommonName: "TestGroup1",
// 			Members:    []string{"test@test1.com", "test@test2.com"},
// 		},
// 	}
// 	idMap := fillIdMap(modelUsers, modelGroups)
// 	bulkRequest := handleNewIdentities(idMap, modelUsers, modelGroups)
// 	marshalled, _ := json.MarshalIndent(bulkRequest, "", "    ")
// 	fmt.Println(string(marshalled))
// }
