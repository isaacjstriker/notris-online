# Notris Online

[![Deploy to Render](https://github.com/isaacjstriker/devware/actions/workflows/deploy.yml/badge.svg)](https://github.com/isaacjstriker/devware/actions/workflows/deploy.yml)
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![Live Demo](https://img.shields.io/badge/demo-live-brightgreen)](https://notris-online.onrender.com/)

![](https://github.com/isaacjstriker/notris-online/blob/main/notris-test.gif)

Notris Online is a multiplayer, Tetris‑style web game built to explore **real‑time concurrency**, **WebSockets**, and **full‑stack development**.

Players can connect via the browser, play classic falling‑block gameplay, and keep their game state in sync with the server in real time. This project served as a capstone for my Boot.dev computer science coursework and a playground for experimenting with Go on the backend and JavaScript on the frontend.

---

## Why I built this

I wanted to:

- Recreate the classic Tetris experience in the browser
- Learn how to handle **real‑time multiplayer** with WebSockets
- Play with a **Go backend** and a **JavaScript frontend**
- Practice structuring a non‑trivial project end‑to‑end (game logic + networking + deployment)

The official Tetris implementations I found were either single‑player, closed, or full of ads. Building my own gave me full control and a great learning opportunity.

---

## Features

- **Real‑time multiplayer gameplay**
  - WebSocket connections keep clients and server in sync
  - Game state updates broadcast to players
- **Server‑authoritative game logic**
  - Core game rules run on the server, not just the client
  - Prevents simple cheating and keeps behavior consistent
- **Classic Tetris mechanics**
  - Falling pieces, line clears, game over conditions
- **Browser-based UI**
  - Rendered in HTML/CSS/JavaScript
  - Keyboard input for movement and rotation
- **Deployable full‑stack app**
  - Go backend
  - JavaScript frontend
  - Ready to host on a modern platform

---

## Tech Stack

- **Backend:** Go
- **Frontend:** JavaScript, HTML, CSS
- **Networking:** WebSockets
- **Build & Deployment:** Render

---

## Architecture (high level)

- **Client (browser)**
  - Renders the board and active piece
  - Captures user input (left, right, rotate, drop)
  - Sends input events to the server over WebSockets
  - Receives updated game state and redraws the board

- **Server (Go)**
  - Manages player sessions and game rooms
  - Maintains authoritative game state:
    - Board contents
    - Active piece
    - Tick timing and gravity
  - Applies game rules and broadcasts updates to connected clients
  - Handles concurrency (multiple connections) using Go’s goroutines/channels

---

## Running locally

> Exact commands may differ depending on how you set up the project; this is a typical Go + JS workflow.

### Prerequisites

- Go (1.20+ recommended)
- Node.js (if there is a build step; otherwise just a browser)

### Backend

```bash
git clone https://github.com/isaacjstriker/notris-online.git
cd notris-online

# Build and run the Go server
go build -o notris-server ./cmd/server
./notris-server
```

The server will start listening on the configured host/port (see `main.go`).

### Frontend

If the frontend is served by the Go backend, simply open the server’s URL in a browser, e.g.:

```text
http://localhost:8080
```

If the frontend is separate (e.g., static files or a JS dev server), follow the instructions in the relevant directory to start a dev server and point it to the WebSocket backend.

---

## Key learning points

Notris Online helped me practice:

- Designing and implementing **real‑time multiplayer** behavior
- Using **WebSockets** to keep clients and server in sync
- Managing **concurrent connections** in Go
- Structuring game logic in a way that’s testable and not tightly coupled to the UI
- Deploying and running a small full‑stack application

---

## Future improvements (ideas)

- Add **matchmaking** and proper rooms/lobbies
- Implement score tracking and a **high score leaderboard**
- Add more polished visuals and animations
- Improve server‑side validation and reconnection logic
- Add bots or AI opponents for single‑player practice

---

## License

This project is currently for educational and portfolio use. If you’re interested in using or extending it, feel free to reach out or open an issue.
