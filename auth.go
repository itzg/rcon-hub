package rcon_hub

import "github.com/gliderlabs/ssh"

type Auth struct {
	users map[string]User
}

func NewAuth(config *Config) *Auth {
	return &Auth{
		users: config.Users,
	}
}

func (a *Auth) PasswordHandler(ctx ssh.Context, password string) bool {
	user, exists := a.users[ctx.User()]
	if !exists {
		return false
	}

	return user.Password == password
}

