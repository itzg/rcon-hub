package rcon_hub

import (
	"bufio"
	"fmt"
	"github.com/james4k/rcon"
	"io"
	"log"
	"strings"
	"sync"
)

type Commander struct {
	// connections contains the pre-defined connections
	connections map[string]Connection
	shell       *Shell

	consolesMu sync.Mutex
	// consoles contains currently connected consoles
	consoles map[string]*rcon.RemoteConsole
	// activeConsole is the console available for interaction
	activeConsole     *rcon.RemoteConsole
	activeConsoleName string
	maxNameLen        int
}

func NewCommander(config *Config, shell *Shell) *Commander {
	return &Commander{
		connections: config.Connections,
		shell:       shell,
		consoles:    make(map[string]*rcon.RemoteConsole),
	}
}

func (c *Commander) Process(command string) error {
	if command == "" {
		return nil
	}

	if c.activeConsole != nil {
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
	case "disconnect":
		if err := c.disconnect(parts[1:]); err != nil {
			return err
		}
	case "exit", "quit":
		c.closeAllConnections()
		return io.EOF
	case "help", "?":
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

func (c *Commander) HandleEof() error {
	if c.activeConsole != nil {
		if err := c.detach(); err != nil {
			return err
		}
		return nil
	} else {
		c.closeAllConnections()
		return io.EOF
	}
}

var commands = []string{
	"list",
	"connect <connection | host:port>",
	"disconnect <connection | host:port>",
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
	if err := c.shell.OutputLine("Configured connections:"); err != nil {
		return err
	}
	for k, v := range c.connections {
		if err := c.shell.OutputLine(fmt.Sprintf("  %s -> %s", k, v.Address)); err != nil {
			return err
		}
	}

	c.consolesMu.Lock()
	if len(c.consoles) > 0 {
		if err := c.shell.OutputLine("Connected consoles:"); err != nil {
			return err
		}
		for k, v := range c.consoles {
			if err := c.shell.OutputLine(fmt.Sprintf("  %s -> %s", k, v.RemoteAddr())); err != nil {
				return err
			}
		}
	}
	c.consolesMu.Unlock()

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

	c.activeConsole = remoteConsole
	if err := c.shell.OutputLine("Connected!"); err != nil {
		return err
	}
	if err := c.shell.OutputLine("Use Control-D to detach"); err != nil {
		return err
	}
	c.shell.SetPrompt(fmt.Sprintf("%s> ", name))

	if len(name) > c.maxNameLen {
		c.maxNameLen = len(name)
	}
	c.activeConsoleName = name
	c.consolesMu.Lock()
	c.consoles[name] = remoteConsole
	c.consolesMu.Unlock()

	go c.readFromRemoteConsole(name, remoteConsole)

	return nil
}

func (c *Commander) disconnect(args []string) error {
	if len(args) <= 0 {
		if err := c.shell.OutputLine("Missing connection name to disconnect"); err != nil {
			return err
		}
		return nil
	}

	name := args[0]
	console, exists := c.consoles[name]
	if !exists {
		if err := c.shell.OutputLine(fmt.Sprintf("Connection %s is not attached", name)); err != nil {
			return err
		}
		return nil
	}
	if err := console.Close(); err != nil {
		return err
	}
	delete(c.consoles, name)

	return nil
}

func (c *Commander) detach() error {
	if c.activeConsole == nil {
		if err := c.shell.Bell(); err != nil {
			return err
		}
		if err := c.shell.OutputLine("Not connected"); err != nil {
			return err
		}
	}

	if c.connections[c.activeConsoleName].AutoDisconnect {
		if err := c.activeConsole.Close(); err != nil {
			if err := c.shell.OutputLine(fmt.Sprintf("Failed to disconnect: %s", err)); err != nil {
				log.Printf("E: failed to output to shell: %s", err)
			}
		}
	} else {
		if err := c.shell.OutputLine(fmt.Sprintf("Detached. Use 'disconnect %s' to stop receiving", c.activeConsoleName)); err != nil {
			log.Printf("E: failed to output to shell: %s", err)
		}
	}

	c.activeConsole = nil
	c.activeConsoleName = ""
	c.shell.SetPrompt("> ")

	return nil
}

func (c *Commander) forwardCommand(command string) error {
	_, err := c.activeConsole.Write(command)
	if err != nil {
		return err
	}

	return nil
}

func (c *Commander) readFromRemoteConsole(name string, console *rcon.RemoteConsole) {
	for {
		response, _, remoteErr := console.Read()
		if remoteErr != nil {
			c.removeConnectedConsole(name)

			if strings.Contains(remoteErr.Error(), "use of closed network connection") {
				_ = c.shell.OutputLine(fmt.Sprintf("Disconnected from %s", name))
				return
			}
			if err := c.shell.OutputLine(fmt.Sprintf("Remote error: %s", remoteErr)); err != nil {
				return
			}
			return
		}

		scanner := bufio.NewScanner(strings.NewReader(response))
		for scanner.Scan() {
			if console == c.activeConsole {
				if err := c.shell.OutputLine(""); err != nil {
					log.Printf("E:failed to output to shell: %s", err)
					return
				}
				if err := c.shell.OutputLine(scanner.Text()); err != nil {
					log.Printf("E: failed to output to shell: %s", err)
					return
				}
			} else {
				if err := c.shell.OutputLine(fmt.Sprintf("%*s | %s", c.maxNameLen, name, scanner.Text())); err != nil {
					log.Printf("E:failed to output to shell: %s", err)
					return
				}
			}
		}
		if err := c.shell.Refresh(); err != nil {
			log.Printf("E: Local error in remote console read: %s", err)
			return
		}
	}

}

func (c *Commander) removeConnectedConsole(name string) {
	c.consolesMu.Lock()
	delete(c.consoles, name)
	c.consolesMu.Unlock()
}

func (c *Commander) closeAllConnections() {
	for _, console := range c.consoles {
		if err := console.Close(); err != nil {
			log.Printf("E: failed to close console connection to %s", console.RemoteAddr())
		}
	}
}
