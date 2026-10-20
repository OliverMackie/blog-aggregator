# Blog Aggregator CLI
Provides a CLI for scraping RSS feeds from the internet, and saving them to a local database to be browsed.

## Prerequesites
Requires Postgres and Go to be installed on your device to function.

## Installation
``` go install https://github.com/OliverMackie/blog-aggregator ```

### Setting up psql
In order for the code to have access to your database there must be a file called .gatorconfig.json in your home directory.
This file should contain ```{db_url:"protocol://username:password@host:port/database?sslmode=disable"}``` Replacing the protocol, username, password, host, port and database with the relevant information concerning your specific psql install and setup.

### Database Schema
This project uses goose migrations to control the database schema. Install goose using this command: ```go install github.com/pressly/goose/v3/cmd/goose@latest``` Once it is installed cd to the blog-aggragator/sql/schema and run the command ```goose postgres '<connection_string>' up``` Where your connection_string is the same one stored in the db_url field. 

#### Commands
You can run a command by typing in the command line ```blog-aggregator <command name>```
The available commands are:
login, takes a registered name string and sets it as the current_user_name.
register, takes a non-registered name string and registers it, then sets it as the current_user_name.
reset, resets the database. Takes no arguments.
users, lists all the users from the users table, showing which one is logged in. Takes no arguments.
agg, takes a time string e.g. (10s) and fetches the oldest feed from the feeds table and creates a post from it.
addfeed, takes a name string and a url string and creates a feed record for that url (associated with the logged in user).
feeds, lists all feeds from the feeds table and the user who added them. Takes no arguments.
follow, takes a url and allows the current user to follow a feed created by a different user.
following, lists all the feeds the current user is following. Takes no arguments.
unfollow, takes a url and allows the user to unfollow the associated feed.
browse, prints the RSS data from the posts of the feeds followed by the user, from most recently to least recently posted. Optionally takes a limit of how many posts to print (default 2).