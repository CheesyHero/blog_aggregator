# Gator

Gator is a command-line RSS feed aggregator written in Go. It allows users to register accounts, follow RSS feeds, aggregate posts from those feeds, and browse articles directly from the terminal.

## Features

* User registration and login
* RSS feed management
* Follow and unfollow feeds
* Continuous feed aggregation
* Store posts in PostgreSQL
* Browse recent posts from followed feeds
* Persistent configuration
* Multi-user support

## Requirements

* Go 1.24+
* PostgreSQL
* Goose
* SQLC

## Installation

Clone the repository:

```bash
git clone https://github.com/CheesyHero/blog_aggregator.git
cd blog_aggregator
```

Install dependencies:

```bash
go mod download
```

Run database migrations:

```bash
goose -dir sql/schema postgres "<connection_string>" up
```

Generate database code:

```bash
sqlc generate
```

Build the application:

```bash
go build -o gator
```

## Configuration

Create a `.gatorconfig.json` file in your home directory:

```json
{
  "db_url": "postgres://postgres:postgres@localhost:5432/gator?sslmode=disable",
  "current_user_name": ""
}
```

## Commands

### Register a User

```bash
./gator register <username>
```

### Login

```bash
./gator login <username>
```

### List Users

```bash
./gator users
```

### Add a Feed

```bash
./gator addfeed "Boot.dev Blog" "https://blog.boot.dev/index.xml"
```

### List Feeds

```bash
./gator feeds
```

### Follow a Feed

```bash
./gator follow <feed_url>
```

### Unfollow a Feed

```bash
./gator unfollow <feed_url>
```

### View Followed Feeds

```bash
./gator following
```

### Aggregate Feeds

Run continuously and collect new posts:

```bash
./gator agg 1m
```

Examples:

```bash
./gator agg 30s
./gator agg 5m
```

### Browse Posts

Browse recent posts from followed feeds:

```bash
./gator browse
```

Specify a custom limit:

```bash
./gator browse 10
```

## Example Feeds

Boot.dev Blog

```
https://www.boot.dev/blog/index.xml
```

Hacker News

```
https://news.ycombinator.com/rss
```

TechCrunch

```
https://techcrunch.com/feed/
```

## Project Structure

```
blog_aggregator/
├── internal/
│   ├── config/
│   └── database/
├── sql/
│   ├── queries/
│   └── schema/
├── main.go
├── sqlc.yaml
└── README.md
```
