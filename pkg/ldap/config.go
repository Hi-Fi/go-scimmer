package ldap

// Config ldap basic configuration for connection
type Config struct {
	Host                string
	Port                int
	UserDN              string
	Password            string
	ActiveUserAttribute string
	ActiveUserValue     string
	GroupFilter         string
	UserFilter          string
	BaseDN              string
	UsernameAttribute   string
}
