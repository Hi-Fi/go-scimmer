package ldap

import (
	"fmt"
	"log"

	"github.com/hi-fi/go-scimmer/pkg/model"
	"gopkg.in/ldap.v3"
)

// LoadUsersAndGroups Connects to LDAP and loads all users and groups
func (c *Config) LoadUsersAndGroups() ([]*model.User, []*model.Group) {
	l, err := openBindedLDAPConnection(c)
	if err != nil {
		// error in ldap bind
		log.Fatal(err)
	}
	defer l.Close()

	return loadUsers(l, c), loadGroups(l, c)

}

// UpdateScimIDs updates external system IDs to given internal attribute
// Attribute has to exist before update
func UpdateScimIDs(users []model.User, l *ldap.Conn) {
	err := l.Bind("cn=admin,dc=example,dc=org", "admin")
	if err != nil {
		// error in ldap bind
		log.Fatal(err)
	}

	for _, user := range users {
		req := ldap.NewModifyRequest(user.DistinguishedName, nil)
		req.Replace("description", []string{user.ScimID})
		if err = l.Modify(req); err != nil {
			log.Fatalf("Failed to modify DN: %s\n", err)
		}
	}
}

func loadUsers(bindedConnection *ldap.Conn, c *Config) (users []*model.User) {
	fetchedAttributes := []string{"mail", "givenName", "sn", "uid"}
	if len(c.ActiveUserAttribute) > 0 {
		fetchedAttributes = append(fetchedAttributes, c.ActiveUserAttribute)
	}
	searchRequest := ldap.NewSearchRequest(
		c.BaseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		c.UserFilter,
		fetchedAttributes,
		nil,
	)

	sr, err := bindedConnection.Search(searchRequest)
	if err != nil {
		log.Fatal(err)
	}

	for _, entry := range sr.Entries {
		user := &model.User{}
		user.DistinguishedName = entry.DN
		for _, attribute := range entry.Attributes {
			switch attributeName := attribute.Name; attributeName {
			case "mail":
				user.Email = attribute.Values[0]
			case "givenName":
				user.FirstName = attribute.Values[0]
			case "sn":
				user.LastName = attribute.Values[0]
			case c.ActiveUserAttribute:
				user.Active = (attribute.Values[0] == c.ActiveUserValue)
			}
			if attribute.Name == c.UsernameAttribute {
				user.Username = attribute.Values[0]
			}
		}
		users = append(users, user)
	}

	return users
}

func loadGroups(bindedConnection *ldap.Conn, c *Config) (groups []*model.Group) {
	searchRequest := ldap.NewSearchRequest(
		c.BaseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		c.GroupFilter,
		[]string{"cn", "memberUid", "member"},
		nil,
	)

	sr, err := bindedConnection.Search(searchRequest)
	if err != nil {
		log.Fatal(err)
	}

	for _, entry := range sr.Entries {
		group := &model.Group{}
		group.DistinguishedName = entry.DN
		for _, attribute := range entry.Attributes {
			switch attributeName := attribute.Name; attributeName {
			case "cn":
				group.CommonName = attribute.Values[0]
			case "memberUid": // PosixGroups, which make members visible with uids.
				group.Members = posixGroupMembersToDN(attribute.Values, bindedConnection, c)
			case "member":
				group.Members = attribute.Values
			}
		}
		groups = append(groups, group)
	}

	return groups
}

func posixGroupMembersToDN(members []string, bindedConnection *ldap.Conn, config *Config) []string {
	dnMembers := []string{}

	for _, uid := range members {
		searchRequest := ldap.NewSearchRequest(
			"dc=example,dc=org",
			ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
			fmt.Sprintf("(&(%s)(uid=%s))", config.UserFilter, uid),
			[]string{"dn"},
			nil,
		)

		sr, err := bindedConnection.Search(searchRequest)
		if err != nil || len(sr.Entries) != 1 {
			log.Fatal(err)
		}
		dnMembers = append(dnMembers, sr.Entries[0].DN)
	}
	return dnMembers
}

func openBindedLDAPConnection(c *Config) (*ldap.Conn, error) {
	conn, err := c.GetConnection()
	if err != nil {
		return conn, err
	}
	err = conn.Bind(c.UserDN, c.Password)
	return conn, err
}
