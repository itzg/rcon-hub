package rcon_hub

import (
	"github.com/gliderlabs/ssh"
	"github.com/spf13/viper"
	"io"
	"log"
)

type SshServer struct {
	config *Config
}

func NewSshServer(config *Config) *SshServer {
	return &SshServer{
		config:config,
	}
}

func (s *SshServer) ListenAndServe() error {
	auth := NewAuth(s.config)

	hostKeyResolver := NewHostKeyResolver(s.config)

	bind := viper.GetString(ConfigBind)

	log.Printf("Accepting SSH connections at %s", bind)
	return ssh.ListenAndServe(bind,
		func(session ssh.Session) {
			log.Printf("I: New session for user=%s from=%s\n", session.User(), session.RemoteAddr())
			shell := NewShell(session, s.config)
			shell.SetPrompt("> ")

			commander := NewCommander(s.config, shell)

			for {
				line, err := shell.Read()

				if err != nil {
					if err == io.EOF {
						err = commander.HandleEof()
						if err == io.EOF {
							return
						} else if err != nil {
							endSessionWithError(session, shell, err)
							return
						}
					} else {
						endSessionWithError(session, shell, err)
						return
					}
				}

				err = commander.Process(line)
				if err != nil {
					if err == io.EOF {
						return
					} else {
						endSessionWithError(session, shell, err)
					}
				}
			}
		},
		ssh.PasswordAuth(auth.PasswordHandler),
		hostKeyResolver.ResolveOption(),
	)
}

func endSessionWithError(s ssh.Session, shell *Shell, err error) {
	_ = shell.OutputLine("")
	_ = shell.OutputLine(err.Error())
	_ = s.Exit(1)
}
