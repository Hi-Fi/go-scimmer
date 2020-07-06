package scim

import (
	"encoding/json"
	"fmt"

	"github.com/hi-fi/go-scimmer/pkg/model"
	log "github.com/sirupsen/logrus"
)

func (c *Config) GenerateBulkRequest(idMap *model.IDMap, users []*model.User, groups []*model.Group) BulkRequest {
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
		}
		bulkRequest.Operations = append(bulkRequest.Operations, bulkOperation)
	}

	for _, group := range groups {
		if group.ScimID == "" {
			newGroup := newGroup(group)
			for index := range newGroup.Members {
				// Trying to first add from idMap
				if len(idMap.Mapping[model.EncodeText(newGroup.Members[index].Value)]) > 0 {
					newGroup.Members[index].Value = idMap.Mapping[model.EncodeText(newGroup.Members[index].Value)]
				} else {
					newGroup.Members[index].Value = fmt.Sprintf("bulkId:%s", model.EncodeText(newGroup.Members[index].Value))
				}
			}

			bulkOperation := BulkOperation{
				BulkID: model.EncodeText(group.DistinguishedName),
				Method: "POST",
				Path:   "/Groups",
				Data:   newGroup,
			}
			bulkRequest.Operations = append(bulkRequest.Operations, bulkOperation)
		}
	}
	if c.DryRun {
		pretty, _ := json.MarshalIndent(bulkRequest, "", "    ")
		log.Infof("BulkRequest: %v\n", string(pretty))
	}

	return bulkRequest
}
