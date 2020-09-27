package model

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"strings"

	"gopkg.in/yaml.v2"
)

// LoadIDMap loads ID mapping from given file as YAML
func (idMap *IDMap) LoadIDMap() error {
	idMap.Mapping = make(map[string]MappedId)

	data, err := ioutil.ReadFile(idMap.FilePath)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(data, idMap.Mapping)
	if err != nil {
		return err
	}
	return nil
}

// ExportIDMap exports ID mapping to given file as YAML
func (idMap *IDMap) ExportIDMap() error {
	idMap.MappingMutex.RLock()
	marshalled, err := yaml.Marshal(idMap.Mapping)
	idMap.MappingMutex.RUnlock()
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(idMap.FilePath, marshalled, 0644)
	if err != nil {
		return err
	}
	return nil
}

// EnrichUsersAndGroupsWithScimIDs load ID mapping and adds new users and groups to it
func (idMap *IDMap) EnrichUsersAndGroupsWithScimIDs(users []*User, groups []*Group) {
	idMap.LoadIDMap()
	for _, user := range users {
		mapping := idMap.Mapping[user.DistinguishedName]
		user.ScimID = mapping.ScimID
		user.Checksum = mapping.Checksum
		user.UpdatedAt = mapping.UpdatedAt
	}
	for _, group := range groups {
		mapping := idMap.Mapping[group.DistinguishedName]
		group.ScimID = mapping.ScimID
		group.Checksum = mapping.Checksum
		group.UpdatedAt = mapping.UpdatedAt
	}
}

// EncodeText encodes given string with base64
func EncodeText(text string) string {
	return base64.StdEncoding.EncodeToString([]byte(text))
}

// DecodeText decodes given base64 encoded string. If decoding fails, returns given string as is
func DecodeText(encodedText string) string {
	decoded, err := base64.StdEncoding.DecodeString(encodedText)
	if err != nil {
		return encodedText
	}
	return string(decoded)
}

// CalculateUserChecksum generates sha256 hash from user
func CalculateUserChecksum(user *User) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(fmt.Sprintf("%s_%s_%s_%s", user.FirstName, user.LastName, user.Email, user.Username))))
}

// CalculateGroupChecksum generates sha256 hash from user
func CalculateGroupChecksum(group *Group) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(fmt.Sprintf("%s_%s", group.CommonName, strings.Join(group.Members[:], ",")))))
}
