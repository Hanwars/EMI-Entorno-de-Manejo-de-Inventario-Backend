package security

import (
	"fmt"
	"time"

	"github.com/aead/chacha20poly1305"
	"github.com/o1egl/paseto"
)

type PasetoBuilder struct {
	paseto      *paseto.V2
	symetricKey []byte
}

func NewPasetoBuilder(key string) (Builder, error) {
	if len(key) != chacha20poly1305.KeySize {
		return nil, fmt.Errorf("tamaño de la llave es invalido:%d carácteres", chacha20poly1305.KeySize)
	}
	builder := &PasetoBuilder{
		paseto:      paseto.NewV2(),
		symetricKey: []byte(key),
	}
	return builder, nil
}

/*
CreateToken(email string, duration time.Duration) (string, error)
VerifyToken(token string) (*Payload, error)
*/
func (builder *PasetoBuilder) CreateToken(userID int32, role int32, name string, duration time.Duration) (string, error) {
	payload, err := NewPayload(userID, role, name, duration)
	if err != nil {
		return "", err
	}
	return builder.paseto.Encrypt(builder.symetricKey, payload, nil)
}
func (builder *PasetoBuilder) VerifyToken(token string) (*Payload, error) {
	payload := &Payload{}
	err := builder.paseto.Decrypt(token, builder.symetricKey, payload, nil)
	if err != nil {
		return nil, ErrorInvalidToken
	}
	err = payload.Valid()
	if err != nil {
		return nil, err
	}
	return payload, nil
}
