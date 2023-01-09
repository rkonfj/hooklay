package internal

import "errors"

type Authenticator struct {
	token string
}

func NewAuthenticator(token string) *Authenticator {
	return &Authenticator{
		token: token,
	}
}

func (a *Authenticator) Authenticate(token string) error {
	if token != a.token {
		return errors.New("401")
	}
	return nil
}
