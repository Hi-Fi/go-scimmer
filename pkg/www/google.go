package www

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hi-fi/go-scimmer/pkg/model"
)

type googleHeader struct {
	ChannelID         string `header:"X-Goog-Channel-ID"`
	ChannelToken      string `header:"X-Goog-Channel-Token"`
	ChannelExpiration string `header:"X-Goog-Channel-Expiration"`
	ResourceID        string `header:"X-Goog-Resource-ID"`
	ResourceURI       string `header:"X-Goog-Resource-URI"`
	ResourceState     string `header:"X-Goog-Resource-State"`
	MessageNumber     string `header:"X-Goog-Message-Number"`
}

type userBody struct {
	Kind         string `json:"kind"`
	ID           string `json:"id"`
	Etag         string `json:"etag"`
	PrimaryEmail string `json:"primaryEmail"`
}

func GsuiteWebhook(c *gin.Context) {
	var (
		header googleHeader
		body   userBody
	)

	c.BindHeader(&header)
	c.BindJSON(&body)

	users := ldapConfig.LoadUser(body.PrimaryEmail)
	modelConfig.EnrichUsersAndGroupsWithScimIDs(users, []*model.Group{})
	scimConfig.SyncIdentities(users, []*model.Group{}, modelConfig)

	c.String(http.StatusOK, "")
}
