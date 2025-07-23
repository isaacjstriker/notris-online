# Dev Ware

Dev Ware is a CLI-based game suite and leaderboard system, built as my capstone project for Boot.dev. The name comes form the ecclectic 'Wario Ware' franchise, from which I took much inspiration. I even created sketch ideas of games on sticky notes, similar to the developers behind 'Wario Ware!'

## Why? (The Motivation)

This project serves as a capstone for the [Boot.dev](https://boot.dev) backend developer curriculum. The primary goal was to build a complete, deployable application that demonstrates proficiency in several key areas of backend development:

-   **Go Proficiency:** Writing clean, modular, and robust Go code for the core application logic.
-   **Database Integration:** Designing a database schema, writing SQL queries, and managing migrations for a PostgreSQL database.
-   **Scripting Engine Embedding:** Integrating a Lua scripting engine (`gopher-lua`) to allow for flexible and decoupled game logic.
-   **Cloud Deployment:** Preparing the application for a cloud environment and deploying it using modern platform-as-a-service (PaaS) tools like Render and Supabase.
-   **Security:** Implementing secure practices for user authentication (password hashing with bcrypt), session management (JWT), and code scanning (`gosec`).

## Technologies & Tools

This project integrates a variety of modern technologies and tools to create a robust backend application.

| Category              | Technology / Tool                                                              | Purpose                                                              |
| --------------------- | ------------------------------------------------------------------------------ | -------------------------------------------------------------------- |
| **Languages**         | Go, Lua, SQL                                                                   | Core application logic, game scripting, and database queries.        |
| **Database**          | PostgreSQL                                                                     | Relational database for storing user and score data.                 |
| **Go Libraries**      | `gopher-lua`, `go-keyboard`, `godotenv`, `pq`                                  | Lua embedding, terminal input, config loading, and DB driver.        |
| **Tooling**           | `sqlc`, `goose`, `gosec`                                                       | Type-safe query generation, DB migrations, and security scanning.    |
| **Platform & Services** | [Supabase](https://supabase.com), [Render](https://render.com), [GitHub](https://github.com) | Managed PostgreSQL hosting, cloud application deployment, and source control. |

## Quick Start

Follow these steps to get the application running locally.

### 1. Prerequisites

-   [Go](https://go.dev/doc/install) (version 1.18 or later)
-   [PostgreSQL](https://www.postgresql.org/download/) installed locally, or a free database from [Supabase](https://supabase.com).
-   [Goose](https://github.com/pressly/goose) for database migrations. Install it with:
    ```sh
    go install github.com/pressly/goose/v3/cmd/goose@latest
    ```

### 2. Clone the Repository

```sh
git clone https://github.com/isaacjstriker/devware.git
cd devware
```

### 3. Configure Your Environment

The application is configured using a `.env` file.

1.  **Create the `.env` file:**
    ```sh
    touch .env
    ```
2.  **Add your Database URL:** Open the `.env` file and add your PostgreSQL connection string.
    ```properties
    # Example for a local PostgreSQL database
    DATABASE_URL="host=localhost user=postgres password=your_password dbname=devware sslmode=disable"
    ```
3.  **JWT Secret (Automatic):** The `JWT_SECRET` is required for security. The application will automatically generate a secure key and add it to your `.env` file the first time you run it.

### 4. Run Database Migrations

With your `DATABASE_URL` set in the `.env` file, run the migrations to create the necessary tables.

```sh
goose -dir "sql/schema" postgres "${DATABASE_URL}" up
```

### 5. Run the Application

Install Go dependencies and run the main program.

```sh
go mod tidy
go run .
```
If the `JWT_SECRET` was missing, the program will add it to your `.env` file and exit. Simply run `go run .` again to start the application.

## Usage

Once the application is running, you will be greeted by the main menu.

-   Use the **Up/Down arrow keys** to navigate the menu.
-   Press **Enter** to select an option.
-   Press **'q'** to quit the application from the main menu.

You can play games as a guest, but to save your scores to the leaderboard, you must register an account and log in.

## Project Structure

```
.
├── games/                # Game packages (Go) and scripts (Lua)
│   ├── breakout/
│   ├── tetris/
│   └── typing/
├── internal/             # Core application logic not meant for reuse
│   ├── auth/             # User authentication and session management
│   ├── config/           # Environment configuration loading
│   ├── database/         # Database connection and query execution
│   └── types/            # Shared data structures
├── ui/                   # Reusable UI components (e.g., menu)
├── .gitignore
├── go.mod                # Go module dependencies
├── go.sum
├── main.go               # Application entry point
└── README.md             # This file
```

## Contributing

Contributions are welcome! If you have ideas for new games, features, or improvements, please follow these steps:

1.  **Fork** the repository.
2.  Create a new **branch** for your feature (`git checkout -b feature/new-game`).
3.  **Commit** your changes (`git commit -m 'Add new game: Space Invaders'`).
4.  **Push** to the branch (`git push origin feature/new-game`).
5.  Open a **Pull Request**.
