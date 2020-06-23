package scim

// Config SCIM connction details
type Config struct {
	EndpointURL   string
	Token         string
	DryRun        bool
	BulkSupported bool
}
