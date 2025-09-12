# Notris Online

[![Deploy to Render](https://github.com/isaacjstriker/devware/actions/workflows/deploy.yml/badge.svg)](https://github.com/isaacjstriker/devware/actions/workflows/deploy.yml)
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Live Demo](https://img.shields.io/badge/demo-live-brightgreen)](https://notris-online.onrender.com/)

A labor of love. Notris is a real-time multiplayer Tetris game built with Go and WebSockets. Key skills demonstrated include user authentication, room-based matchmaking, and live gameplay synchronization.

View it now at https://notris-online.onrender.com/

**Key Features:**
- **Real-time Multiplayer**: Play Tetris against opponents with live game state synchronization
- **Room System**: Create private or public rooms with customizable starting levels
- **User Management**: Secure registration and authentication with JWT tokens

## Capstone Project

I created this as my capstone project for Boot.dev, in which computer science and programming fundamentals are taught through the backend slant. I also really wanted to play Tetris online with my friends.

## Cloning

### Prerequisites

- [Go](https://go.dev/doc/install) 1.19 or later
- [PostgreSQL](https://www.postgresql.org/download/) database (local or cloud)

### Installation

1. **Clone the repository**
   ```bash
   git clone https://github.com/isaacjstriker/notris-online.git
   cd notris-online
   ```

2. **Set up environment variables**
   ```bash
   touch .env
   # Edit .env with your database connection details
   ```
   
   Required environment variables:
   ```env
   DATABASE_URL="postgres://username:password@localhost/notris-online?sslmode=disable"
   JWT_SECRET="your-secret-key-here"
   PORT=8888
   ```

3. **Install dependencies**
   ```bash
   go mod tidy
   ```

4. **Start the server**
   ```bash
   go run .
   ```

5. **Open your browser**
   Navigate to `http://localhost:8888` to start playing!

### Game Controls
- **Arrow Keys**: Move and rotate pieces
- **Spacebar**: Hard drop piece
- **C**: Hold piece
- **Escape**: Pause game / Open menu

## Contributing

You are more than welcome to contribute if you have a feature or bug fix in mind. Please give a clear description of your work when you push you open a pull request.

### About the developer
I'm a full stack web developer that will create projects as case studies, to determine areas for technical skill improvement. I love all programming languages, I do not discriminate (except for Assembly LOL).

---
