package ldap

import (
	"crypto/tls"
	"fmt"

	"gopkg.in/ldap.v3"
)

// GetConnection provides connection to LDAP server
// Note that the connection is not bind yet with any user
func (c *Config) GetConnection() (*ldap.Conn, error) {
	tlsConfig := &tls.Config{InsecureSkipVerify: true}
	connection, err := ldap.DialTLS("tcp", fmt.Sprintf("%s:%d", c.Host, c.Port), tlsConfig)
	if err != nil {
		connection, err = ldap.Dial("tcp", fmt.Sprintf("%s:%d", c.Host, c.Port))
	}
	return connection, err
}
