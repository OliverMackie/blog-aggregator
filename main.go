package main

import (
	"database/sql"
	"fmt"
	"os"

	config "github.com/OliverMackie/blog-aggregator/internal"
	"github.com/OliverMackie/blog-aggregator/internal/database"
	_ "github.com/lib/pq"
)

func main() {
	con, err := config.Read()
	if err != nil {
		fmt.Println("Error reading config:", err)
		return
	}
	db, err := sql.Open("postgres", con.Db_url)
	dbQueries := database.New(db)
	state := State{dbQueries, con}
	commands := commands{make(map[string]func(*State, Command) error)}
	commands.register("login", handlerLogin)
	commands.register("register", handlerRegister)
	commands.register("reset", handlerReset)
	commands.register("users", handlerUsers)
	commands.register("agg", handlerAgg)
	commands.register("addfeed", middlewareLoggedIn(handlerAddFeed))
	commands.register("feeds", handlerFeeds)
	commands.register("follow", middlewareLoggedIn(handlerFollow))
	commands.register("following", handlerFollowing)
	commands.register("unfollow", middlewareLoggedIn(handlerUnfollow))
	commands.register("browse", handlerBrowse)
	Args := os.Args
	if len(Args) < 2 {
		fmt.Println("No command provided")
		os.Exit(1)
	}
	Args = Args[1:]
	com := Command{Args[0], Args[1:]}
	err = commands.run(&state, com)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}
