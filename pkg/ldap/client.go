package ldap

import (
	"crypto/tls"
	"fmt"

	"gopkg.in/ldap.v3"
)

// GetConnection provides connection to LDAP server
// Note that the connection is not bind yet with any user
func GetConnection(config Config) (*ldap.Conn, error) {
	tlsConfig := &tls.Config{InsecureSkipVerify: true}
	connection, err := ldap.DialTLS("tcp", fmt.Sprintf("%s:%d", config.Host, config.Port), tlsConfig)
	if err != nil {
		connection, err = ldap.Dial("tcp", fmt.Sprintf("%s:%d", config.Host, config.Port))
	}
	return connection, err
}
