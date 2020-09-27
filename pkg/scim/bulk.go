package scim

import (
	"encoding/json"
	"fmt"

	"github.com/hi-fi/go-scimmer/pkg/model"
	log "github.com/sirupsen/logrus"
)

func (c *Config) generateBulkRequest(idMap *model.IDMap, users []*model.User, groups []*model.Group) BulkRequest {
	bulkRequest := newBulkRequest()
	for _, user := range users {
		scimUser := newUser(user)
		bulkOperation := BulkOperation{
			Data: scimUser,
		}
		if user.ScimID == "" {
			bulkOperation.BulkID = model.EncodeText(user.DistinguishedName)
			bulkOperation.Method = "POST"
			bulkOperation.Path = "/Users"
		} else {
			bulkOperation.Method = "PUT"
			bulkOperation.Path = fmt.Sprintf("/Users/%s", user.ScimID)
		}
		bulkRequest.Operations = append(bulkRequest.Operations, bulkOperation)
	}

	for _, group := range groups {
		scimGroup := newGroup(group)
		for index := range scimGroup.Members {
			// Trying to first add from idMap
			if len(idMap.Mapping[model.EncodeText(scimGroup.Members[index].Value)].ScimID) > 0 {
				scimGroup.Members[index].Value = idMap.Mapping[model.EncodeText(scimGroup.Members[index].Value)].ScimID
			} else {
				scimGroup.Members[index].Value = fmt.Sprintf("bulkId:%s", model.EncodeText(scimGroup.Members[index].Value))
			}
		}

		var bulkOperation BulkOperation
		if group.ScimID == "" {
			bulkOperation = BulkOperation{
				BulkID: model.EncodeText(group.DistinguishedName),
				Method: "POST",
				Path:   "/Groups",
				Data:   scimGroup,
			}
		} else {
			bulkOperation = BulkOperation{
				Method: "PATCH",
				Path:   fmt.Sprintf("/Groups/%s", group.ScimID),
				Data: newPatchOp("replace", Group{
					Members: scimGroup.Members,
				}),
			}
		}
		bulkRequest.Operations = append(bulkRequest.Operations, bulkOperation)
	}

	if c.DryRun {
		pretty, _ := json.MarshalIndent(bulkRequest, "", "    ")
		log.Infof("BulkRequest: %v\n", string(pretty))
	}

	return bulkRequest
}
