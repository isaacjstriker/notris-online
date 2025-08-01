# Notris Online

[![Deploy to Render](https://github.com/isaacjstriker/devware/actions/workflows/deploy.yml/badge.svg)](https://github.com/isaacjstriker/devware/actions/workflows/deploy.yml)
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Live Demo](https://img.shields.io/badge/demo-live-brightgreen)](https://notris-online.onrender.com/)

A real-time multiplayer Tetris game built with Go and WebSockets, featuring user authentication, room-based matchmaking, and live gameplay synchronization.

View it now at https://notris-online.onrender.com/

## Description

Notris is a modern web-based Tetris implementation that allows players to compete against each other in real-time multiplayer matches. Players can create or join game rooms, ready up for matches, and play head-to-head Tetris with live opponent visibility and game state synchronization.

**Key Features:**
- 🎮 **Real-time Multiplayer**: Play Tetris against opponents with live game state synchronization
- 🏠 **Room System**: Create private or public rooms with customizable starting levels
- 👥 **User Management**: Secure registration and authentication with JWT tokens
- 📊 **Live Updates**: See your opponent's board, score, and progress in real-time
- 🔄 **Reconnection Handling**: Automatic reconnection and proper disconnect notifications
- ⏱️ **Match Timer**: Track game duration during multiplayer matches
- 🎯 **Responsive Design**: Clean, modern web interface that works across devices

## Why?

This project serves as a comprehensive demonstration of modern web application development, showcasing several key technical challenges:

**Real-time Communication**: Building a responsive multiplayer experience requires careful WebSocket management, message routing, and state synchronization between clients.

**Scalable Architecture**: The application demonstrates clean separation of concerns with a Go backend handling game logic and WebSocket connections, while a JavaScript frontend manages the user interface and real-time updates.

**User Experience**: Creating an engaging multiplayer game involves solving complex UX challenges like connection handling, room management, player notifications, and seamless game state transitions.

**Full-Stack Development**: The project integrates backend services (Go, PostgreSQL), real-time communication (WebSockets), and frontend technologies (HTML5 Canvas, JavaScript) into a cohesive application.

## Quick Start

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
   PORT=8080
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
   Navigate to `http://localhost:8080` to start playing!

## Usage

### Getting Started
1. **Register an account** or log in if you already have one
2. **Navigate to Multiplayer** from the main menu
3. **Create a room** or **join an existing room**
4. **Wait for another player** to join your room
5. **Ready up** when you're prepared to play
6. **Play Tetris** in real-time against your opponent!

### Game Controls
- **Arrow Keys**: Move and rotate pieces
- **Spacebar**: Hard drop piece
- **C**: Hold piece
- **Escape**: Pause game / Open menu

### Multiplayer Features
- **Room Browser**: See all available public rooms
- **Custom Rooms**: Create private rooms with specific settings
- **Real-time Sync**: Watch your opponent's board update live
- **Disconnect Handling**: Proper notifications when players leave
- **Match Timer**: Track how long each game lasts

## Contributing

We welcome contributions! Whether you want to fix bugs, add features, or improve documentation, here's how you can help:

### Getting Started
1. **Fork the repository** on GitHub
2. **Clone your fork** locally
3. **Create a feature branch**: `git checkout -b feature/amazing-feature`
4. **Make your changes** and test thoroughly
5. **Commit your changes**: `git commit -m "Add amazing feature"`
6. **Push to your fork**: `git push origin feature/amazing-feature`
7. **Open a Pull Request** with a clear description of your changes

### Development Guidelines
- **Code Style**: Follow standard Go conventions and gofmt formatting
- **Testing**: Add tests for new functionality where appropriate
- **Documentation**: Update README and code comments for significant changes
- **Commits**: Use clear, descriptive commit messages

### Areas for Contribution
- 🎨 **UI/UX Improvements**: Enhance the visual design and user experience
- 🎮 **Game Features**: Add new game modes, power-ups, or mechanics
- 🔧 **Performance**: Optimize WebSocket handling or game rendering
- 📱 **Mobile Support**: Improve touch controls and responsive design
- 🧪 **Testing**: Add unit tests and integration tests
- 📚 **Documentation**: Improve setup guides and API documentation

### Reporting Issues
Found a bug? Have a feature request? Please [open an issue](https://github.com/isaacjstriker/notris-online/issues) with:
- Clear description of the problem or suggestion
- Steps to reproduce (for bugs)
- Expected vs actual behavior
- Your environment details (OS, browser, etc.)

### About the developer
Hi! I'm Isaac, and I've been fascinated with programming for a long time. I’d write games in Python, or work on a website for fun here and there, and always enjoyed coding. After 4 years as a mechanic, I’ve realized I have a real skill for this programming stuff, and decided to properly learn how to do it by taking some in-depth CS and backend development courses. Now that I’ve completed those and have several real-world projects under my belt, I’m excited for a full-time career in software engineering.

I'm now intimately familiar with Go, which has quickly become my favorite language to write in.

---

Built with ❤️ using Go, WebSockets, and HTML5 Canvas
