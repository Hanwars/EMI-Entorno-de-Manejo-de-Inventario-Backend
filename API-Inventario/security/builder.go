package security

import "time"

type Builder interface {
	CreateToken(userID int32, role int32, name string, duration time.Duration) (string, error)
	VerifyToken(token string) (*Payload, error)
}
