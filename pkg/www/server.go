package www

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/hi-fi/go-scimmer/pkg/ldap"
	"github.com/hi-fi/go-scimmer/pkg/model"
	"github.com/hi-fi/go-scimmer/pkg/scim"
)

var (
	ldapConfig  ldap.Config
	scimConfig  scim.Config
	modelConfig *model.IDMap
)

func (c *Config) StartServer(ldap ldap.Config, scim scim.Config, model *model.IDMap) {
	ldapConfig = ldap
	scimConfig = scim
	modelConfig = model

	route := gin.Default()
	v1API(route)
	route.Run(fmt.Sprintf(":%d", c.Port))
}
