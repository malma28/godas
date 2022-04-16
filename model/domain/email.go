package domain

type EmailVerificationSend struct {
	FromName     string
	FromEmail    string
	FromPassword string
	ToEmail      string

	Title string

	Host string
	Port string
}

type EmailVerification struct {
	Email      string `json:"email" bson:"email"`
	Code       string `json:"code" bson:"code"`
	Expiration int64  `json:"expiration" bson:"expiration"`
	Cooldown   int64  `json:"cooldown" bson:"cooldown"`
}
