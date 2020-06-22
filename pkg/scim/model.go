package scim

import (
	"fmt"

	"github.com/hi-fi/go-scimmer/pkg/model"
)

type User struct {
	ID          string   `json:"id,omitempty"`
	UserName    string   `json:"userName"`
	DisplayName string   `json:"displayName"`
	ExternalID  string   `json:"externalID"`
	Schemas     []string `json:"schemas"`
	Active      bool     `json:"active"`
	Name        Name
	Emails      []Email
}

type Name struct {
	GivenName  string `json:"givenName"`
	FamilyName string `json:"familyName"`
}

func newUser(modelUser model.User) User {
	user := User{
		Schemas: []string{"urn:ietf:params:scim:schemas:core:2.0:User"},
		Active:  true,
	}
	user.Name = Name{
		FamilyName: modelUser.LastName,
		GivenName:  modelUser.FirstName,
	}
	user.DisplayName = fmt.Sprintf("%s %s", modelUser.FirstName, modelUser.LastName)
	user.ID = modelUser.ScimID
	user.ExternalID = modelUser.Email
	user.UserName = modelUser.Username
	user.Emails = []Email{
		{
			Type:    "work",
			Primary: true,
			Value:   modelUser.Email,
		},
	}
	return user
}

type Email struct {
	Value   string `json:"value"`
	Type    string `json:"type"`
	Primary bool
}

type Group struct {
	ID          string   `json:"id,omitempty"`
	DisplayName string   `json:"displayName,omitempty"`
	Members     []Member `json:"members,omitempty"`
	Schemas     []string `json:"schema,omitempty"`
}

func newGroup(modelGroup model.Group) Group {
	group := Group{
		Schemas:     []string{"urn:ietf:params:scim:schemas:core:2.0:Group"},
		DisplayName: modelGroup.CommonName,
		ID:          modelGroup.ScimID,
	}

	for _, member := range modelGroup.Members {
		group.Members = append(group.Members, Member{
			Value: member,
		})
	}
	return group
}

type Member struct {
	Value       string `json:"value,omitempty"`
	Ref         string `json:"$ref,omitempty"`
	DisplayName string `json:"displayName,omitempty"`
}

type PatchOp struct {
	Schemas    []string         `json:"schema"`
	Operations []PatchOperation `json:"Operations"`
}

func newPatchOp() PatchOp {
	patchOp := PatchOp{
		Schemas: []string{"urn:ietf:params:scim:api:messages:2.0:PatchOp"},
	}
	return patchOp
}

type PatchOperation struct {
	Op    string      `json:"op"`
	Value interface{} `json:"value"`
}

type BulkOperation struct {
	Method string      `json:"method"`
	Path   string      `json:"path"`
	BulkID string      `json:"bulkId"`
	Data   interface{} `json:"data"`
}

type BulkRequest struct {
	Schemas    []string        `json:"schema"`
	Operations []BulkOperation `json:"Operations"`
}

func newBulkRequest() BulkRequest {
	bulkRequest := BulkRequest{
		Schemas: []string{"urn:ietf:params:scim:api:messages:2.0:BulkRequest"},
	}
	return bulkRequest
}
