package ldap

import (
	"testing"

	"github.com/hi-fi/go-scimmer/pkg/model"
)

func TestOpenConnection(t *testing.T) {
	LoadUsersAndGroups()
}

func TestScimIDUpdate(t *testing.T) {
	modelUsers := []model.User{
		{
			DistinguishedName: "cn=Test User,dc=example,dc=org",
			ScimID:            "tuser_scim_id",
			FirstName:         "Test",
			LastName:          "Test1",
			Username:          "test@test1.com",
			Email:             "test@test1.com",
		},
		{
			DistinguishedName: "cn=Testi Useri,dc=example,dc=org",
			ScimID:            "tuseri_scim_id",
			FirstName:         "Test",
			LastName:          "Test2",
			Username:          "test@test2.com",
			Email:             "test@test2.com",
		},
	}

	UpdateScimIDs(modelUsers)
}
