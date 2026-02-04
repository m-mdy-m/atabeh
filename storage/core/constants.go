package core

const (
	TableConfigs       = "configs"
	TableProfiles      = "profiles"
	TableSubscriptions = "subscriptions"
)

const (
	ConfigColID        = "id"
	ConfigColProfileID = "profile_id"
	ConfigColName      = "name"
	ConfigColProtocol  = "protocol"
	ConfigColServer    = "server"
	ConfigColPort      = "port"
	ConfigColUUID      = "uuid"
	ConfigColPassword  = "password"
	ConfigColMethod    = "method"
	ConfigColTransport = "transport"
	ConfigColSecurity  = "security"
	ConfigColExtra     = "extra"
	ConfigColSource    = "source"
	ConfigColLastPing  = "last_ping"
	ConfigColIsAlive   = "is_alive"
	ConfigColCreatedAt = "created_at"
	ConfigColUpdatedAt = "updated_at"
)

const (
	ProfileColID          = "id"
	ProfileColName        = "name"
	ProfileColSource      = "source"
	ProfileColType        = "type"
	ProfileColConfigCount = "config_count"
	ProfileColAliveCount  = "alive_count"
	ProfileColLastSynced  = "last_synced_at"
	ProfileColCreatedAt   = "created_at"
	ProfileColUpdatedAt   = "updated_at"
)

const (
	SubscriptionColURL = "url"
)
