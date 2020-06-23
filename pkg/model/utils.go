package model

import (
	"encoding/base64"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// LoadIDMap loads ID mapping from given file as YAML
func (idMap *IDMap) LoadIDMap() error {
	idMap.Mapping = make(map[string]string)

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
	marshalled, err := yaml.Marshal(idMap.Mapping)
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
		user.ScimID = idMap.Mapping[user.DistinguishedName]
	}
	for _, group := range groups {
		group.ScimID = idMap.Mapping[group.DistinguishedName]
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
