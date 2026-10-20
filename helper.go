package main

import (
	"context"
	"database/sql"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	config "github.com/OliverMackie/blog-aggregator/internal"
	"github.com/OliverMackie/blog-aggregator/internal/database"
)

type State struct {
	db     *database.Queries
	config *config.Config
}

type Command struct {
	Name string
	Args []string
}

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func handlerLogin(state *State, command Command) error {
	if len(command.Args) < 1 {
		return fmt.Errorf("username argument is required")
	}

	_, err := state.db.GetUser(context.Background(), command.Args[0])
	if err != nil {
		return fmt.Errorf("error retrieving user: %v", err)
	}

	err = state.config.Setuser(command.Args[0])
	if err != nil {
		return fmt.Errorf("error setting user: %v", err)
	}
	fmt.Printf("User %s has been set\n", command.Args[0])
	return nil
}

func handlerRegister(state *State, command Command) error {
	if len(command.Args) < 1 {
		return fmt.Errorf("name argument is required")
	}
	_, err := state.db.CreateUser(context.Background(), database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      command.Args[0],
	})
	if err != nil {
		return fmt.Errorf("error creating user: %v", err)
	}
	state.config.Setuser(command.Args[0])
	fmt.Printf("User %s has been created\nData: %s\n", command.Args[0], state.config.Current_user_name)

	return nil
}

func handlerReset(state *State, command Command) error {
	err := state.db.Reset(context.Background())
	if err != nil {
		return fmt.Errorf("error resetting database: %v", err)
	}
	fmt.Println("Database has been reset")
	return nil
}

func handlerUsers(state *State, command Command) error {
	users, err := state.db.GetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("error retrieving users: %v", err)
	}
	for _, user := range users {
		if state.config.Current_user_name == user.Name {
			fmt.Printf("* %s (current)\n", user.Name)
		} else {
			fmt.Printf("* %s\n", user.Name)
		}
	}
	return nil
}

func handlerAgg(state *State, command Command) error {
	if len(command.Args) < 1 {
		return fmt.Errorf("time between reps argument is required")
	}
	timeBetweenRequests, err := time.ParseDuration(command.Args[0])
	if err != nil {
		return fmt.Errorf("error parsing time duration: %v", err)
	}
	ticker := time.NewTicker(timeBetweenRequests)
	defer ticker.Stop()

	for ; ; <-ticker.C {
		err = scrapeFeeds(state)
		if err != nil {
			return fmt.Errorf("error scraping feeds: %v", err)
		}
	}
}

func handlerAddFeed(state *State, command Command, user database.User) error {
	if len(command.Args) < 2 {
		return fmt.Errorf("name and url arguments are required")
	}

	feed, err := state.db.CreateFeed(context.Background(), database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      command.Args[0],
		Url:       command.Args[1],
		UserID:    user.ID,
	})
	if err != nil {
		return fmt.Errorf("error creating feed: %v", err)
	}
	_, err = state.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		FeedID:    feed.ID})
	if err != nil {
		return fmt.Errorf("error creating feed follow: %v", err)
	}
	fmt.Printf("ID: %s\nCreatedAt: %s\nUpdatedAt: %s\nName: %s\nUrl: %s\nUserID: %s\n", feed.ID, feed.CreatedAt, feed.UpdatedAt, feed.Name, feed.Url, feed.UserID)
	return nil
}

func handlerFeeds(state *State, command Command) error {
	feeds, err := state.db.GetFeeds(context.Background())
	if err != nil {
		return fmt.Errorf("error retrieving feeds: %v", err)
	}
	for _, feed := range feeds {
		fmt.Printf("Name: %s\nUrl: %s\n", feed.Name, feed.Url)
		user, err := state.db.GetUserName(context.Background(), feed.UserID)
		if err != nil {
			return fmt.Errorf("error retrieving user for feed: %v", err)
		}
		fmt.Printf("User: %s\n", user.Name)
	}
	return nil
}

func handlerFollow(state *State, command Command, user database.User) error {
	if len(command.Args) < 1 {
		return fmt.Errorf("url argument is required")
	}
	feedurl, err := state.db.GetFeedUrl(context.Background(), command.Args[0])
	if err != nil {
		return fmt.Errorf("error retrieving feed: %v", err)
	}

	_, err = state.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		FeedID:    feedurl.ID})
	if err != nil {
		return fmt.Errorf("error creating feed: %v", err)
	}
	fmt.Printf("Name: %s\nUser: %s\n", feedurl.Name, user.Name)
	return nil
}

func handlerFollowing(state *State, command Command) error {
	following, err := state.db.GetFeedFollowsForUser(context.Background(), state.config.Current_user_name)
	if err != nil {
		return fmt.Errorf("error retrieving following feeds: %v", err)
	}
	for _, follow := range following {
		fmt.Printf("Name: %s\n", follow.FeedName)
	}
	return nil
}

type commands struct {
	commandmap map[string]func(*State, Command) error
}

func (c *commands) run(s *State, cmd Command) error {
	if handler, ok := c.commandmap[cmd.Name]; ok {
		return handler(s, cmd)
	}
	return fmt.Errorf("unknown command: %s", cmd.Name)
}

func handlerUnfollow(state *State, command Command, user database.User) error {
	if len(command.Args) < 1 {
		return fmt.Errorf("url argument is required")
	}
	err := state.db.DeleteFeedFollow(context.Background(), database.DeleteFeedFollowParams{
		Name: user.Name,
		Url:  command.Args[0],
	})
	if err != nil {
		return fmt.Errorf("error unfollowing feed: %v", err)
	}
	return nil
}

func handlerBrowse(state *State, command Command) error {
	user_name := state.config.Current_user_name
	limit := int32(0)
	if len(command.Args) < 1 {
		limit = 2
	} else {
		limitvar, err := strconv.ParseInt(command.Args[0], 10, 32)
		if err != nil {
			return fmt.Errorf("error parsing limit: %v", err)
		}
		limit = int32(limitvar)
	}
	posts, err := state.db.GetPostsForUser(context.Background(), database.GetPostsForUserParams{
		Name:  user_name,
		Limit: limit,
	})
	if err != nil {
		return fmt.Errorf("error retrieving posts for user: %v", err)
	}
	for _, post := range posts {
		fmt.Printf("Title: %s\nUrl: %s\nDescription: %s\nPublishedAt: %s\n", post.Title, post.Url, post.Description.String, post.PublishedAt.Time)
	}
	return nil
}

func (c *commands) register(name string, f func(*State, Command) error) {
	c.commandmap[name] = f
}

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	rss := &RSSFeed{}
	client := http.Client{}
	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}
	req.Header.Set("User-Agent", "gator")
	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error fetching feed: %v", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	if err := xml.Unmarshal(body, &rss); err != nil {
		return nil, fmt.Errorf("error unmarshalling XML: %v", err)
	}

	rss.Channel.Title = html.UnescapeString(rss.Channel.Title)
	rss.Channel.Description = html.UnescapeString(rss.Channel.Description)
	for i, item := range rss.Channel.Item {
		rss.Channel.Item[i].Title = html.UnescapeString(item.Title)
		rss.Channel.Item[i].Description = html.UnescapeString(item.Description)
	}

	return rss, nil
}

func middlewareLoggedIn(handler func(s *State, c Command, u database.User) error) func(s *State, c Command) error {
	return func(s *State, c Command) error {
		user, err := s.db.GetUser(context.Background(), s.config.Current_user_name)
		if err != nil {
			return fmt.Errorf("error retrieving user: %v", err)
		}
		return handler(s, c, user)
	}
}

func scrapeFeeds(state *State) error {
	feed, err := state.db.GetFeedToFetch(context.Background())
	if err != nil {
		return fmt.Errorf("error retrieving feed to fetch: %v", err)
	}
	err = state.db.MarkFeedFetched(context.Background(), feed.ID)
	if err != nil {
		return fmt.Errorf("error marking feed as fetched: %v", err)
	}
	fetchedFeed, err := fetchFeed(context.Background(), feed.Url)
	if err != nil {
		return fmt.Errorf("error fetching feed: %v", err)
	}

	for _, item := range fetchedFeed.Channel.Item {
		valid := true
		if item.Description == "" {
			valid = false
		}
		validt := true
		if item.PubDate == "" {
			validt = false
		}
		var timepub time.Time
		formats := []string{
			time.RFC1123Z,
			time.RFC1123,
			time.RFC822Z,
			time.RFC822,
			time.RFC3339}
		for _, format := range formats {
			timepub, err = time.Parse(format, item.PubDate)
			if err == nil {
				break
			}
		}
		if err != nil {
			validt = false
		}

		_, err = state.db.CreatePost(context.Background(), database.CreatePostParams{
			ID:          uuid.New(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Title:       item.Title,
			Url:         item.Link,
			Description: sql.NullString{String: item.Description, Valid: valid},
			PublishedAt: sql.NullTime{Time: timepub, Valid: validt},
			FeedID:      feed.ID,
		})
		if err != nil {
			if strings.Contains(err.Error(), "duplicate key value") {
				return nil
			} else {
				log.Println(err)
			}
		}
	}
	fmt.Printf("Fetched feed: %s\n", feed.Name)
	return nil
}
