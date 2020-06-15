package scim

type user struct {
	userName    string
	displayName string
	id          string
	schema      string
}

func newUser() user {
	user := user{}
	user.schema = "urn:ietf:params:scim:schemas:core:2.0:User"
	return user
}

type email struct {
	value     string
	emailType string
	primary   bool
}

type group struct {
	displayName string
	members     []member
	schema      string
}

func newGroup() group {
	group := group{}
	group.schema = "urn:ietf:params:scim:schemas:core:2.0:Group"
}

type member struct {
	value       string
	ref         string
	displayName string
}
