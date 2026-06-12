package main

import (
	"blog_aggregator/internal/config"
	"blog_aggregator/internal/database"
	"database/sql"
	"encoding/xml"
	"errors"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"os"

	"context"
	"time"

	"github.com/google/uuid"

	_ "github.com/lib/pq"
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

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}
type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
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
func handlerAgg(s *State, cmd Command) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("usage: agg <time_between_reqs>")
	}

	timeBetweenRequests, err := time.ParseDuration(cmd.args[0])
	if err != nil {
		return err
	}

	fmt.Printf("Collecting feeds every %s\n", timeBetweenRequests)

	ticker := time.NewTicker(timeBetweenRequests)
	defer ticker.Stop()

	for ; ; <-ticker.C {
		err := scrapeFeeds(s)
		if err != nil {
			fmt.Printf("error scraping feeds: %v\n", err)
		}
	}
}
func handlerAddFeed(s *State, cmd Command, user database.User) error {
	if len(cmd.args) != 2 {
		return fmt.Errorf("usage: addfeed <name> <url>")
	}

	feed, err := s.database.CreateFeed(context.Background(), database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      cmd.args[0],
		Url:       cmd.args[1],
		UserID:    user.ID,
	})
	if err != nil {
		return err
	}

	feedFollow, err := s.database.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		FeedID:    feed.ID,
	})
	if err != nil {
		return err
	}

	fmt.Println("Feed created:")
	fmt.Printf("Name: %s\n", feed.Name)
	fmt.Printf("URL: %s\n", feed.Url)
	fmt.Printf("User: %s\n", user.Name)

	fmt.Println("Feed followed:")
	fmt.Printf("User: %s\n", feedFollow.UserName)
	fmt.Printf("Feed: %s\n", feedFollow.FeedName)

	return nil
}

func handlerFeeds(s *State, cmd Command) error {
	feeds, err := s.database.GetFeeds(context.Background())
	if err != nil {
		return err
	}

	for _, feed := range feeds {
		fmt.Printf("Name: %s\n", feed.Name)
		fmt.Printf("URL: %s\n", feed.Url)
		fmt.Printf("User: %s\n\n", feed.UserName)
	}

	return nil
}
func handlerFollow(s *State, cmd Command, user database.User) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("usage: follow <url>")
	}

	feed, err := s.database.GetFeedByURL(context.Background(), cmd.args[0])
	if err != nil {
		return err
	}

	feedFollow, err := s.database.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		FeedID:    feed.ID,
	})
	if err != nil {
		return err
	}

	fmt.Printf("%s is now following %s\n", feedFollow.UserName, feedFollow.FeedName)

	return nil
}
func handlerFollowing(s *State, cmd Command, user database.User) error {
	feedFollows, err := s.database.GetFeedFollowsForUser(context.Background(), user.Name)
	if err != nil {
		return err
	}

	for _, feedFollow := range feedFollows {
		fmt.Println(feedFollow.FeedName)
	}

	return nil
}
func handlerUnfollow(s *State, cmd Command, user database.User) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("usage: unfollow <url>")
	}

	feed, err := s.database.GetFeedByURL(context.Background(), cmd.args[0])
	if err != nil {
		return err
	}

	err = s.database.DeleteFeedFollow(context.Background(), database.DeleteFeedFollowParams{
		UserID: user.ID,
		FeedID: feed.ID,
	})
	if err != nil {
		return err
	}

	fmt.Printf("%s unfollowed %s\n", user.Name, feed.Name)

	return nil
}
func middlewareLoggedIn(handler func(*State, Command, database.User) error) func(*State, Command) error {
	return func(s *State, cmd Command) error {
		user, err := s.database.GetUser(context.Background(), s.config.CurrentUserName)
		if err != nil {
			return err
		}

		return handler(s, cmd, user)
	}
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
	commands.Register("agg", handlerAgg)
	commands.Register("feeds", handlerFeeds)
	commands.Register("addfeed", middlewareLoggedIn(handlerAddFeed))
	commands.Register("follow", middlewareLoggedIn(handlerFollow))
	commands.Register("following", middlewareLoggedIn(handlerFollowing))
	commands.Register("unfollow", middlewareLoggedIn(handlerUnfollow))

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

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, feedURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "gator")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode > 299 {
		return nil, fmt.Errorf("bad status code: %d", res.StatusCode)
	}

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var feed RSSFeed
	if err := xml.Unmarshal(data, &feed); err != nil {
		return nil, err
	}

	feed.Channel.Title = html.UnescapeString(feed.Channel.Title)
	feed.Channel.Description = html.UnescapeString(feed.Channel.Description)

	for i := range feed.Channel.Item {
		feed.Channel.Item[i].Title = html.UnescapeString(feed.Channel.Item[i].Title)
		feed.Channel.Item[i].Description = html.UnescapeString(feed.Channel.Item[i].Description)
	}

	return &feed, nil
}

func scrapeFeeds(s *State) error {
	feed, err := s.database.GetNextFeedToFetch(context.Background())
	if err != nil {
		return err
	}

	err = s.database.MarkFeedFetched(context.Background(), feed.ID)
	if err != nil {
		return err
	}

	fmt.Printf("Fetching feed: %s (%s)\n", feed.Name, feed.Url)

	rssFeed, err := fetchFeed(context.Background(), feed.Url)
	if err != nil {
		return err
	}

	for _, item := range rssFeed.Channel.Item {
		fmt.Printf("- %s\n", item.Title)
	}

	return nil
}
