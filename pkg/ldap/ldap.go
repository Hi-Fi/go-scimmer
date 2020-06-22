package ldap

import (
	"log"

	"github.com/hi-fi/go-scimmer/pkg/model"
	"gopkg.in/ldap.v3"
)

// LoadUsersAndGroups Connects to LDAP and loads all users and groups
func LoadUsersAndGroups() ([]model.User, []model.Group) {
	l, err := openLDAPConnection()
	if err != nil {
		// error in ldap bind
		log.Fatal(err)
	}
	defer l.Close()
	err = l.Bind("cn=admin,dc=example,dc=org", "admin")
	if err != nil {
		// error in ldap bind
		log.Fatal(err)
	}

	return loadUsers(l), loadGroups(l)

}

func UpdateScimIDs(users []model.User) {
	l, err := openLDAPConnection()
	if err != nil {
		// error in ldap bind
		log.Fatal(err)
	}
	defer l.Close()
	err = l.Bind("cn=admin,dc=example,dc=org", "admin")
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

func loadUsers(l *ldap.Conn) (users []model.User) {
	searchRequest := ldap.NewSearchRequest(
		"dc=example,dc=org",
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		"(|(objectclass=user)(objectclass=person)(objectclass=inetOrgPerson)(objectclass=organizationalPerson))",
		[]string{"mail", "givenName", "sn"},
		nil,
	)

	sr, err := l.Search(searchRequest)
	if err != nil {
		log.Fatal(err)
	}

	for _, entry := range sr.Entries {
		user := model.User{}
		user.DistinguishedName = entry.DN
		for _, attribute := range entry.Attributes {
			switch attributeName := attribute.Name; attributeName {
			case "mail":
				user.Email = attribute.Values[0]
			case "givenName":
				user.FirstName = attribute.Values[0]
			case "sn":
				user.LastName = attribute.Values[0]
			}
		}
		users = append(users, user)
	}

	log.Printf("%v", users)
	return users
}

func loadGroups(l *ldap.Conn) (groups []model.Group) {
	searchRequest := ldap.NewSearchRequest(
		"dc=example,dc=org",
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		"(|(objectclass=group)(objectclass=groupOfNames)(objectclass=groupOfUniqueNames))",
		[]string{"cn", "memberUid", "member"},
		nil,
	)

	sr, err := l.Search(searchRequest)
	if err != nil {
		log.Fatal(err)
	}

	for _, entry := range sr.Entries {
		group := model.Group{}
		group.DistinguishedName = entry.DN
		for _, attribute := range entry.Attributes {
			switch attributeName := attribute.Name; attributeName {
			case "cn":
				group.CommonName = attribute.Values[0]
			case "memberUid":
				group.Members = attribute.Values
			case "member":
				group.Members = attribute.Values
			}
		}
		groups = append(groups, group)
	}

	log.Printf("%v", groups)
	return groups
}

func openLDAPConnection() (*ldap.Conn, error) {
	config := &Config{
		Host:     "localhost",
		Port:     389,
		Username: "jsaarinen",
		Password: "testi",
	}
	return GetConnection(*config)
}
