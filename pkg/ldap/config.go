package ldap

// Config ldap basic configuration for connection
type Config struct {
	Host string
	Port int
	Username string
	Password string
}