package config

const (
	reductionKeyPrefix                       = "REDUCTION_"
	ReductionKeyLocalAccessExpirationSeconds = reductionKeyPrefix + "LOCAL_ACCESS_EXPIRATION_SECONDS"
	reductionSessionKeyPrefix                = reductionKeyPrefix + "SESSION_"
	ReductionKeySessionCookiePath            = reductionSessionKeyPrefix + "COOKIE_PATH"
	ReductionKeySessionSecure                = reductionSessionKeyPrefix + "SECURE"
	ReductionKeySessionHMACKey               = reductionSessionKeyPrefix + "HMAC_KEY"
)
