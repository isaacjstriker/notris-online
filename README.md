# Notris Online

[![Deploy to Render](https://github.com/isaacjstriker/devware/actions/workflows/deploy.yml/badge.svg)](https://github.com/isaacjstriker/devware/actions/workflows/deploy.yml)
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![Live Demo](https://img.shields.io/badge/demo-live-brightgreen)](https://notris-online.onrender.com/)

![](https://github.com/isaacjstriker/notris-online/blob/main/notris-test.gif)

A Tetris game with multiplayer, leaderboards, and all plays in the browser. Built with JS and Go utilizing WebSockets and concurrency.

## Ok... but why? Tetris is everywhere!

Fair point, but I wanted to play it anytime, anywhere, from the browser. Also, the official Tetris browser game is RIDDLED with ads, and doesn't have multiplayer. And I wanted to test the experience I've gained through my CS course.

## Quick Start

All you need is an internet connection! Play online now at: https://notris-online.onrender.com/

**Key Features:**
- **Real-time Multiplayer**: Play Tetris against opponents with live game state synchronization
- **Room System**: Create private or public rooms with customizable starting levels
- **User Management**: Login to store your high scores across all your devices

### Prerequisites

- [Go](https://go.dev/doc/install) 1.19 or later
- [PostgreSQL](https://www.postgresql.org/download/) database (local or cloud)

### Installation (as easy as...)

1. **Clone the repository**
   ```bash
   git clone https://github.com/isaacjstriker/notris-online.git
   cd notris-online
   ```

2. **Install dependencies**
   ```bash
   go mod tidy
   ```

3. **Start the server**
   ```bash
   go run .
   ```
   Navigate to `http://localhost:8888` to start playing!

## Contributing

You are more than welcome to contribute if you have a feature or bug fix in mind. Please give a clear description of your work when you open a pull request.

---
