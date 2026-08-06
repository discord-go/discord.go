package interactions

// ApplicationIntegrationType represents the type of integration.
type ApplicationIntegrationType int

const (
	ApplicationIntegrationTypeGuildInstall ApplicationIntegrationType = 0
	ApplicationIntegrationTypeUserInstall  ApplicationIntegrationType = 1
)
