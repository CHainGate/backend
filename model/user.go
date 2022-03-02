package model

// TODO: make field "not optional"
type User struct {
	Id                uint
	Name              string
	Email             string `gorm:"unique"`
	Password          []byte
	IsActive          bool
	EmailVerification EmailVerification `gorm:"foreignKey:Email;references:Email"`
}

type EmailVerification struct {
	Id               uint
	Email            string
	VerificationCode uint64
}
