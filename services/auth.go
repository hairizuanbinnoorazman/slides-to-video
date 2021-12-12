package services

type Auth struct {
	Secret     string
	ExpiryTime int
	Issuer     string
	HashKey    []byte
	BlockKey   []byte
	CookieName string
}
