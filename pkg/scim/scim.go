package scim

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/hi-fi/go-scimmer/pkg/model"
	log "github.com/sirupsen/logrus"
)

func (c *Config) SyncIdentities(users []*model.User, groups []*model.Group, idMap *model.IDMap) (err error) {
	if c.BulkSupported {
		c.generateBulkRequest(idMap, users, groups)
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
	log.Debugf("Dry run: %v", c.DryRun)
	for _, user := range users {
		if !c.DryRun {
			wg.Add(1)
			go c.postAndUpdateUser(newUser(user), idMap, &wg)
		} else if user.ScimID == "" {
			idMap.Mapping[user.DistinguishedName] = model.MappedId{
				ScimID:    fmt.Sprintf("dry_run_%s", model.EncodeText(user.DistinguishedName)),
				Checksum:  model.CalculateUserChecksum(user),
				UpdatedAt: time.Now(),
			}
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
		log.Errorf("Payload marshalling failed. Error: %v", err)
		return
	}

	if user.ID == "" {
		if !user.Active && !c.UploadDisabled {
			log.Debugf("User %s disabled, not uploading it", user.distinguishedName)
			return
		}
		request, err = http.NewRequest(http.MethodPost, targetURL, bytes.NewReader(payload))
	} else {
		if !user.Active && c.DeleteDisabled {
			request, err = http.NewRequest(http.MethodDelete, targetURL, nil)
		} else {
			request, err = http.NewRequest(http.MethodPut, targetURL, bytes.NewReader(payload))
		}
	}

	if err != nil {
		log.Errorf("Request creation failed. Error: %v", err)
		return
	}

	request.Header.Set("Content-Type", "application/json; charset=utf-8")
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))

	response, err = httpClient.Do(request)
	if err != nil {
		log.Errorf("Request to %s failed. Error: %v", targetURL, err)
		return
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Errorf("Reading of body failed. Error: %v", err)
		return
	}

	if response.StatusCode >= 400 {
		log.Errorf("Request failed with status %s. Sent payload was %s and response body: %s", response.Status, string(payload), string(body))
		return
	}

	err = json.Unmarshal(body, user)
	if err != nil {
		log.Errorf("Unmarshal of user from response failed. Error: %v", err)
		return
	}

	log.Debugf("Handled user %s with target system ID %s", user.UserName, user.ID)
	idMap.MappingMutex.Lock()
	idMap.Mapping[user.distinguishedName] = model.MappedId{
		ScimID:    user.ID,
		UpdatedAt: time.Now(),
		// This is needed in the list to handle group memberships correctly
		Active: user.Active,
	}
	idMap.MappingMutex.Unlock()

}

func (c *Config) syncGroups(idMap *model.IDMap, groups []*model.Group) {
	var wg sync.WaitGroup
	log.Debugf("Dry run: %v", c.DryRun)
	for _, group := range groups {
		if !c.DryRun {
			wg.Add(1)
			go c.postAndUpdateGroup(newGroup(group), idMap, &wg)
		} else if group.ScimID == "" {
			mapping := idMap.Mapping[group.DistinguishedName]
			mapping.ScimID = fmt.Sprintf("dry_run_%s", model.EncodeText(group.DistinguishedName))
		}
	}
	wg.Wait()
}

func (c *Config) postAndUpdateGroup(group *Group, idMap *model.IDMap, wg *sync.WaitGroup) {
	defer wg.Done()
	// Wait that all members are done to external system.
	membersDone := 0
	newMembers := []Member{}
	for membersDone < len(group.Members) {
		for _, member := range group.Members {
			idMap.MappingMutex.RLock()
			mappedValue := idMap.Mapping[member.Value].ScimID
			active := idMap.Mapping[member.Value].Active
			idMap.MappingMutex.RUnlock()
			if !active && (c.UploadDisabled || c.DeleteDisabled) {
				membersDone = membersDone + 1
				log.Debugf("Not including member %s to %s as member status is inactive. %d done", member.Value, group.DisplayName, membersDone)
			} else if mappedValue != "" {
				newMembers = append(newMembers, Member{
					DisplayName: member.DisplayName,
					Ref:         member.Ref,
					Value:       mappedValue,
				})
				membersDone = membersDone + 1
				log.Debugf("Mapped value for member %s to %s. %d done", member.Value, mappedValue, membersDone)
			}
		}
		if membersDone < len(group.Members) {
			//Wait a bit to allow missing members to be created
			log.Debugf("%d members out of %d in group %s validated for sync. Waiting a moment to make a new check...", membersDone, len(group.Members), group.DisplayName)
			time.Sleep(time.Second)
			// reset, as we go through whole list again
			membersDone = 0
			newMembers = []Member{}
		}
	}
	group.Members = newMembers
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
		log.Tracef("Payload: %s", string(payload))

		if err != nil {
			log.Errorf("Payload marshalling failed. Error: %v", err)
			return
		}

		request, err = http.NewRequest(http.MethodPost, targetURL, bytes.NewReader(payload))
	} else {
		// Groups can't be replaced. Only action possible is to update the members of the group
		payload, err = json.Marshal(newPatchOp("replace", Group{
			Members: group.Members,
		}))

		if err != nil {
			log.Errorf("Payload marshalling failed. Error: %v", err)
			return
		}

		request, err = http.NewRequest(http.MethodPatch, targetURL, bytes.NewReader(payload))
	}

	if err != nil {
		log.Errorf("Request creation failed. Error: %v", err)
		return
	}

	request.Header.Set("Content-Type", "application/json; charset=utf-8")
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))

	response, err = httpClient.Do(request)
	if err != nil {
		log.Errorf("Request to %s failed. Error: %v", targetURL, err)
		return
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Errorf("Reading of body failed. Error: %v", err)
		return
	}

	if response.StatusCode >= 400 {
		log.Errorf("Request failed with status %s. Sent payload was %s and response body: %s", response.Status, string(payload), string(body))
		return
	}

	// Patch and Delete returns 204 when successfull.
	if response.StatusCode != 204 {
		err = json.Unmarshal(body, group)
		if err != nil {
			log.Errorf("Unmarshal of group from response failed. Error: %v", err)
			return
		}
	}

	log.Debugf("Handled group %s with target system ID %s", group.DisplayName, group.ID)
	idMap.MappingMutex.Lock()
	mapping := idMap.Mapping[group.distinguishedName]
	mapping.ScimID = group.ID
	mapping.UpdatedAt = time.Now()
	idMap.MappingMutex.Unlock()
}
