package scim

import (
	"encoding/base64"
	"fmt"

	"github.com/hi-fi/go-scimmer/pkg/model"
)

func SyncIdentities(users []model.User, groups []model.Group) (updatedUsers []model.User, updatedGroups []model.Group, err error) {
	idMap := getIDMap(users, groups)
	handleNewIdentities(idMap, users, groups)
	return
}

func getIDMap(users []model.User, groups []model.Group) map[string]string {
	idMap := make(map[string]string)
	for _, user := range users {
		if len(user.ScimID) > 0 {
			idMap[getHash(user.DistinguishedName)] = user.ScimID
		} else {
			fmt.Println("Adding bulkid to user")
			idMap[getHash(user.DistinguishedName)] = getHash(user.DistinguishedName)
		}
	}
	for _, group := range groups {
		if len(group.ScimID) > 0 {
			idMap[getHash(group.DistinguishedName)] = group.ScimID
		} else {
			fmt.Println("Adding bulkid to group")

			idMap[getHash(group.DistinguishedName)] = getHash(group.DistinguishedName)
		}
	}
	return idMap
}

func handleNewIdentities(idMap map[string]string, users []model.User, groups []model.Group) BulkRequest {
	bulkRequest := newBulkRequest()
	for _, user := range users {
		if user.ScimID == "" {
			newUser := newUser(user)
			bulkOperation := BulkOperation{
				BulkID: getHash(user.DistinguishedName),
				Method: "POST",
				Path:   "/Users",
				Data:   newUser,
			}
			bulkRequest.Operations = append(bulkRequest.Operations, bulkOperation)
		}
	}

	fmt.Printf("%v", idMap)
	for _, group := range groups {
		if group.ScimID == "" {
			newGroup := newGroup(group)
			for index := range newGroup.Members {
				newGroup.Members[index].Value = idMap[getHash(newGroup.Members[index].Value)]
				// If not user, try group
				if len(newGroup.Members[index].Value) == 0 {
					newGroup.Members[index].Value = idMap[getHash(newGroup.Members[index].Value)]
				}
			}

			bulkOperation := BulkOperation{
				BulkID: getHash(group.DistinguishedName),
				Method: "POST",
				Path:   "/Users",
				Data:   newGroup,
			}
			bulkRequest.Operations = append(bulkRequest.Operations, bulkOperation)
		}
	}
	return bulkRequest
}

func getHash(text string) string {
	return base64.StdEncoding.EncodeToString([]byte(text))
}
