package scim

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/hi-fi/go-scimmer/pkg/model"
)

func TestBulkRequestCreation(t *testing.T) {
	modelUsers := []*model.User{
		{
			FirstName: "Test",
			LastName:  "Test1",
			Username:  "test@test1.com",
			Email:     "test@test1.com",
		},
		{
			FirstName: "Test",
			LastName:  "Test2",
			Username:  "test@test2.com",
			Email:     "test@test2.com",
		},
	}

	modelGroups := []*model.Group{
		{
			CommonName: "TestGroup1",
			Members:    []string{"test@test1.com", "test@test2.com"},
		},
	}

	config := &Config{
		DryRun: false,
	}
	idMap := &model.IDMap{}
	bulkRequest := config.generateBulkRequest(idMap, modelUsers, modelGroups)
	marshalled, _ := json.MarshalIndent(bulkRequest, "", "    ")
	fmt.Println(string(marshalled))
}
