package rcon_hub

import (
	"bufio"
	"fmt"
	"github.com/james4k/rcon"
	"io"
	"log"
	"strings"
)

type Commander struct {
	connections map[string]Connection
	shell       *Shell

	remoteConsole *rcon.RemoteConsole
}

func NewCommander(config *Config, shell *Shell) *Commander {
	return &Commander{
		connections: config.Connections,
		shell: shell,
	}
}

func (c *Commander) Process(command string) error {
	if command == "" {
		return nil
	}

	if c.remoteConsole != nil {
		return c.forwardCommand(command)
	}

	parts := strings.Split(command, " ")

	switch strings.ToLower(parts[0]) {
	case "list":
		if err := c.list(); err != nil {
			return err
		}
	case "connect":
		if err := c.connect(parts[1:]); err != nil {
			return err
		}
	case "exit","quit":
		return io.EOF
	case "help","?":
		if err := c.showHelp(); err != nil {
			return err
		}
	default:
		if err := c.showHelp(); err != nil {
			return err
		}
	}

	return nil
}

var commands = []string {
	"list",
	"connect [connection]",
	"exit|quit",
	"help|?",
}

func (c *Commander) showHelp() error {
	if err := c.shell.OutputLine("Available commands:"); err != nil {
		return err
	}
	for _, command := range commands {
		if err := c.shell.OutputLine(command); err != nil {
			return err
		}
	}
	return nil
}

func (c *Commander) list() error {
	for k,v := range c.connections {
		if err := c.shell.OutputLine(fmt.Sprintf("%s -> %s", k, v.Address)); err != nil {
			return err
		}
	}

	return nil
}

func (c *Commander) connect(args []string) error {
	if len(args) <= 0 {
		if err := c.shell.OutputLine("Missing args"); err != nil {
			return err
		}
		return nil
	}

	if len(args) != 1 {
		if err := c.shell.OutputLine("Too many args"); err != nil {
			return err
		}
		return nil
	}

	name := args[0]
	connection, exists := c.connections[name]
	if !exists {
		if err := c.shell.OutputLine("Unknown connection"); err != nil {
			return err
		}
		return nil
	}

	remoteConsole, rconErr := rcon.Dial(connection.Address, connection.Password)
	if rconErr != nil {
		if err := c.shell.OutputLine(fmt.Sprintf("Failed to connect: %s", rconErr)); err != nil {
			return err
		}
		return nil
	}

	c.remoteConsole = remoteConsole
	if err := c.shell.OutputLine("Connected!"); err != nil {
		return err
	}
	if err := c.shell.OutputLine("Use Control-D to disconnect"); err != nil {
		return err
	}
	c.shell.SetPrompt(fmt.Sprintf("%s> ", name))

	go c.readFromRemoteConsole()

	return nil
}

func (c *Commander) disconnect() error {
	if c.remoteConsole == nil {
		if err := c.shell.Bell(); err != nil {
			return err
		}
		if err := c.shell.OutputLine("Not connected"); err !=  nil {
			return err
		}
	}

	if err := c.remoteConsole.Close(); err != nil {
		return err
	}

	c.remoteConsole = nil
	c.shell.SetPrompt("> ")

	return nil
}

func (c *Commander) forwardCommand(command string) error {
	_, err := c.remoteConsole.Write(command)
	if err != nil {
		return err
	}

	return nil
}

func (c *Commander) readFromRemoteConsole() {
	for c.remoteConsole != nil {
		response, _, remoteErr := c.remoteConsole.Read()
		if remoteErr != nil {
			if strings.Contains(remoteErr.Error(), "use of closed network connection") {
				return
			}
			if err := c.shell.OutputLine(fmt.Sprintf("Remote error: %s", remoteErr)); err !=  nil {
				return
			}
			return
		}

		if err := c.shell.OutputLine(""); err !=  nil {
			log.Printf("E: Local error in remote console read: %s", err)
			return
		}
		scanner := bufio.NewScanner(strings.NewReader(response))
		for scanner.Scan() {
			if err := c.shell.OutputLine(scanner.Text()); err !=  nil {
				log.Printf("E: Local error in remote console read: %s", err)
				return
			}
		}
		if err := c.shell.Refresh(); err != nil {
			log.Printf("E: Local error in remote console read: %s", err)
			return
		}
	}

}

func (c *Commander) HandleEof() error {
	if c.remoteConsole != nil {
		if err := c.disconnect(); err != nil {
			return err
		}
		return nil
	} else {
		return io.EOF
	}
}
