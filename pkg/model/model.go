package model

import (
	"sync"
	"time"
)

// User defines internal user object
type User struct {
	DistinguishedName string
	Username          string
	Email             string
	FirstName         string
	LastName          string
	ScimID            string
	Active            bool
	UpdatedAt         time.Time
	Checksum          string
}

// Group defines internal group object
type Group struct {
	CommonName        string
	DistinguishedName string
	ScimID            string
	Members           []string
	UpdatedAt         time.Time
	Checksum          string
}

// IDMap contains mapping between local and external IDs
type IDMap struct {
	FilePath     string
	Mapping      map[string]MappedId
	MappingMutex sync.RWMutex
}

// MappedId stores metadata to help see if updates are needed
type MappedId struct {
	ScimID    string
	UpdatedAt time.Time
	Checksum  string
	Active    bool
}
