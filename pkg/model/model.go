package model

// User defines internal user object
type User struct {
	DistinguishedName string
	Username          string
	Email             string
	FirstName         string
	LastName          string
	ScimID            string
}

// Group defines internal group object
type Group struct {
	CommonName        string
	DistinguishedName string
	ScimID            string
	Members           []string
}
