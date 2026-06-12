package main

import (
	"blog_aggregator/internal/config"
	"blog_aggregator/internal/database"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"

	"context"
	"time"

	"github.com/google/uuid"
)

type State struct {
	database *database.Queries
	config   *config.Config
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

func HandlerUsers(s *State, cmd Command) error {
	users, err := s.database.GetUsers(context.Background())
	if err != nil {
		return err
	}

	for _, user := range users {
		if user.Name == s.config.CurrentUserName {
			fmt.Printf("* %s (current)\n", user.Name)
		} else {
			fmt.Printf("* %s\n", user.Name)
		}
	}

	return nil
}
func HandlerReset(s *State, cmd Command) error {
	err := s.database.ResetUsers(context.Background())
	if err != nil {
		return err
	}

	fmt.Println("Users table reset successfully.")
	return nil
}
func HandlerRegister(s *State, cmd Command) error {
	if len(cmd.args) == 0 {
		return errors.New("username is required")
	}

	user, err := s.database.CreateUser(
		context.Background(),
		database.CreateUserParams{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Name:      cmd.args[0],
		},
	)

	if err != nil {
		return err
	}

	err = s.config.SetUser(user.Name)
	if err != nil {
		return err
	}

	fmt.Println("User created:")
	fmt.Println(user)

	return nil
}
func HandlerLogin(s *State, cmd Command) error {
	if len(cmd.args) == 0 {
		return errors.New("username is required")
	}

	user, err := s.database.GetUser(
		context.Background(),
		cmd.args[0],
	)
	if err != nil {
		return err
	}

	err = s.config.SetUser(user.Name)
	if err != nil {
		return err
	}

	fmt.Printf("User '%s' has been set.\n", user.Name)
	return nil
}

func main() { // !!
	cfg, err := config.Read()
	if err != nil {
		fmt.Printf("ERROR: %v", err)
		os.Exit(1)
	}

	db, err := sql.Open("postgres", cfg.DBURL)
	if err != nil {
		log.Fatal(err)
	}
	dbQueries := database.New(db)

	state := &State{
		database: dbQueries,
		config:   &cfg,
	}

	commands := Commands{
		handlers: make(map[string]func(*State, Command) error),
	}
	commands.Register("login", HandlerLogin)
	commands.Register("register", HandlerRegister)
	commands.Register("reset", HandlerReset)
	commands.Register("users", HandlerUsers)

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
