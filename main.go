package main

import (
	"blog_aggregator/internal/config"
	"errors"
	"fmt"
	"os"
)

type State struct {
	config *config.Config
}
type Command struct {
	name string
	args []string
}
type Commands struct {
	handlers map[string]func(*State, Command) error
}

func (c *Commands) Run(s *State, cmd Command) error {
	handler, ok := c.handlers[cmd.name]
	if !ok {
		return fmt.Errorf("Unknown Command: %s", cmd.name)
	}

	return handler(s, cmd)
}
func (c *Commands) Register(name string, f func(*State, Command) error) {

	c.handlers[name] = f
}

func HandlerLogin(s *State, cmd Command) error {
	if len(cmd.args) == 0 {
		return errors.New("Improper Arguments: Lengths must be greater than 0")
	}

	s.config.SetUser(cmd.args[0])

	fmt.Printf("User '%s' has been set.", s.config.CurrentUserName)
	return nil
}

func main() {
	cfg, err := config.Read()
	if err != nil {
		fmt.Printf("ERROR: %v", err)
		os.Exit(1)
	}

	state := &State{
		config: &cfg,
	}

	commands := Commands{
		handlers: make(map[string]func(*State, Command) error),
	}
	commands.Register("login", HandlerLogin)

	args := os.Args
	if len(args) < 2 {
		fmt.Println("not enough arguments")
		os.Exit(1)
	}

	cmd := Command{
		name: args[1],
		args: args[2:],
	}

	err = commands.Run(state, cmd)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
