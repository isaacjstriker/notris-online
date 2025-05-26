# Devware

Devware is a CLI-based game suite and leaderboard system, built as my capstone project for Boot.dev.

## Technologies Used

- **Go**: Main programming language for CLI, game orchestration, and database interaction.
- **Lua**: Used for writing individual game logic, embedded and executed from Go.
- **PostgreSQL**: Stores online leaderboard data.
- **Goose**: Handles database migrations.
- **sqlc**: Generates type-safe Go code from SQL queries.
- **gopher-lua**: Embeds Lua scripting in Go.

## Dependencies

- [Go](https://golang.org/) 1.18+
- [gopher-lua](https://github.com/yuin/gopher-lua)
- [Goose](https://github.com/pressly/goose)
- [sqlc](https://github.com/kyleconroy/sqlc)
- [PostgreSQL](https://www.postgresql.org/)

Install Go dependencies with:

```sh
go mod tidy
```

## Project Structure

- main.go — CLI entry point
- games/ — Contains Lua scripts and Go packages for each game
- db/ — Database migrations and generated code (planned)
- README.md — This file

## About

This project is my capstone for Boot.dev, emphasizing skills in Go and database-backed application development. I used Lua for game logic as it is the language I am most comfortable scripting games with.