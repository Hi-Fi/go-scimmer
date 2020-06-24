package scim

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/hi-fi/go-scimmer/pkg/model"
)

func (c *Config) SyncIdentities(users []*model.User, groups []*model.Group, idMap *model.IDMap) (updatedUsers []model.User, updatedGroups []model.Group, err error) {
	if c.BulkSupported {
		c.handleNewIdentitiesAsBulk(idMap, users, groups)
	} else {
		c.syncUsers(idMap, users)
		idMap.ExportIDMap()
		c.syncGroups(idMap, groups)
		idMap.ExportIDMap()
	}
	return
}

func (c *Config) syncUsers(idMap *model.IDMap, users []*model.User) {
	var wg sync.WaitGroup
	log.Printf("Dry run: %v", c.DryRun)
	for _, user := range users {
		if !c.DryRun {
			wg.Add(1)
			go c.postAndUpdateUser(newUser(user), idMap, &wg)
		} else if user.ScimID == "" {
			idMap.Mapping[user.DistinguishedName] = fmt.Sprintf("dry_run_%s", model.EncodeText(user.DistinguishedName))
		}
	}
	wg.Wait()
}

func (c *Config) postAndUpdateUser(user *User, idMap *model.IDMap, wg *sync.WaitGroup) {
	defer wg.Done()
	targetURL := fmt.Sprintf("%s/Users/%s", c.EndpointURL, user.ID)
	var (
		response *http.Response
		request  *http.Request
		err      error
	)
	httpClient := http.Client{
		Timeout: 30 * time.Second,
	}

	payload, err := json.Marshal(user)

	if err != nil {
		log.Printf("Payload marshalling failed. Error: %v", err)
		return
	}

	if user.ID == "" {
		request, err = http.NewRequest(http.MethodPost, targetURL, bytes.NewReader(payload))
	} else {
		request, err = http.NewRequest(http.MethodPut, targetURL, bytes.NewReader(payload))
	}

	if err != nil {
		log.Printf("Request creation failed. Error: %v", err)
		return
	}

	request.Header.Set("Content-Type", "application/json; charset=utf-8")
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))

	response, err = httpClient.Do(request)
	if err != nil {
		log.Printf("Request to %s failed. Error: %v", targetURL, err)
		return
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Printf("Reading of body failed. Error: %v", err)
		return
	}

	if response.StatusCode >= 400 {
		log.Printf("Request failed with status %s. Sent payload was %s and response body: %s", response.Status, string(payload), string(body))
		return
	}

	err = json.Unmarshal(body, user)
	if err != nil {
		log.Printf("Unmarshal of user from response failed. Error: %v", err)
		return
	}

	log.Printf("Handled user %s with target system ID %s", user.UserName, user.ID)
	idMap.MappingMutex.Lock()
	idMap.Mapping[user.distinguishedName] = user.ID
	idMap.MappingMutex.Unlock()

}

func (c *Config) handleNewIdentitiesAsBulk(idMap *model.IDMap, users []*model.User, groups []*model.Group) BulkRequest {
	bulkRequest := newBulkRequest()
	for _, user := range users {
		if user.ScimID == "" {
			newUser := newUser(user)
			bulkOperation := BulkOperation{
				BulkID: model.EncodeText(user.DistinguishedName),
				Method: "POST",
				Path:   "/Users",
				Data:   newUser,
			}
			bulkRequest.Operations = append(bulkRequest.Operations, bulkOperation)
		}
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
		fmt.Printf("BulkRequest: %v\n", string(pretty))
	}

	return bulkRequest
}

func getHash(text string) string {
	return base64.StdEncoding.EncodeToString([]byte(text))
}

func (c *Config) syncGroups(idMap *model.IDMap, groups []*model.Group) {
	var wg sync.WaitGroup
	log.Printf("Dry run: %v", c.DryRun)
	for _, group := range groups {
		if !c.DryRun {
			wg.Add(1)
			go c.postAndUpdateGroup(newGroup(group), idMap, &wg)
		} else if group.ScimID == "" {
			idMap.Mapping[group.DistinguishedName] = fmt.Sprintf("dry_run_%s", model.EncodeText(group.DistinguishedName))
		}
	}
	wg.Wait()
}

func (c *Config) postAndUpdateGroup(group *Group, idMap *model.IDMap, wg *sync.WaitGroup) {
	defer wg.Done()
	// Wait that all members are done to external system.
	membersDone := 0
	for membersDone < len(group.Members) {
		for index, member := range group.Members {
			idMap.MappingMutex.RLock()
			mappedValue := idMap.Mapping[member.Value]
			idMap.MappingMutex.RUnlock()
			if mappedValue != "" {
				group.Members[index].Value = mappedValue
				membersDone = membersDone + 1
				log.Printf("Mapped value for member %s to %s. %d done", member.Value, mappedValue, membersDone)
			}
		}
		if membersDone < len(group.Members) {
			//Wait a bit to allow missing members to be created
			log.Printf("%d members out of %d in group %s created to external system. Waiting...", membersDone, len(group.Members), group.DisplayName)
			time.Sleep(time.Second)
		}
	}
	targetURL := fmt.Sprintf("%s/Groups/%s", c.EndpointURL, group.ID)
	var (
		response *http.Response
		request  *http.Request
		err      error
		payload  []byte
	)
	httpClient := http.Client{
		Timeout: 30 * time.Second,
	}

	if group.ID == "" {
		payload, err = json.Marshal(group)
		log.Printf("Payload: %s", string(payload))

		if err != nil {
			log.Printf("Payload marshalling failed. Error: %v", err)
			return
		}

		request, err = http.NewRequest(http.MethodPost, targetURL, bytes.NewReader(payload))
	} else {
		// Groups can't be replaced. Only action possible is to update the members of the group
		payload, err = json.Marshal(newPatchOp("replace", Group{
			Members: group.Members,
		}))

		if err != nil {
			log.Printf("Payload marshalling failed. Error: %v", err)
			return
		}

		request, err = http.NewRequest(http.MethodPatch, targetURL, bytes.NewReader(payload))
	}

	if err != nil {
		log.Printf("Request creation failed. Error: %v", err)
		return
	}

	request.Header.Set("Content-Type", "application/json; charset=utf-8")
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))

	response, err = httpClient.Do(request)
	if err != nil {
		log.Printf("Request to %s failed. Error: %v", targetURL, err)
		return
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Printf("Reading of body failed. Error: %v", err)
		return
	}

	if response.StatusCode >= 400 {
		log.Printf("Request failed with status %s. Sent payload was %s and response body: %s", response.Status, string(payload), string(body))
		return
	}

	// Patch and Delete returns 204 when successfull.
	if response.StatusCode != 204 {
		err = json.Unmarshal(body, group)
		if err != nil {
			log.Printf("Unmarshal of group from response failed. Error: %v", err)
			return
		}
	}

	log.Printf("Handled group %s with target system ID %s", group.DisplayName, group.ID)
	idMap.MappingMutex.Lock()
	idMap.Mapping[group.distinguishedName] = group.ID
	idMap.MappingMutex.Unlock()
}
