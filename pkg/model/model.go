package model

// User defines internal user object
type User struct {
	Username  string
	Email     string
	FirstName string
	LastName  string
	ScimID    string
}

// Group defines internal group object
type Group struct {
	Name    string
	ScimID  string
	Members []string
}
