// Enhanced leaderboard system with multiple views and categories

let currentLeaderboardOptions = {
    period: 'all',
    category: 'score',
    limit: 15,
    includeAchievements: true
};

async function loadLeaderboard() {
    const contentDiv = document.getElementById('leaderboard-content');
    const gameType = 'tetris';

    contentDiv.innerHTML = '<div class="loading-spinner">Loading leaderboard...</div>';

    try {
        // Load both leaderboard and recent games
        const [entries, recentGames] = await Promise.all([
            getLeaderboard(gameType, currentLeaderboardOptions),
            getRecentGames(gameType, 5)
        ]);

        renderLeaderboard(entries);
        renderRecentGames(recentGames);
    } catch (error) {
        contentDiv.innerHTML = `<p class="error-message">Failed to load leaderboard: ${error.message}</p>`;
    }
}

function renderLeaderboard(entries) {
    const contentDiv = document.getElementById('leaderboard-content');

    if (!entries || entries.length === 0) {
        contentDiv.innerHTML = '<div class="empty-state">No scores yet. Be the first!</div>';
        return;
    }

    // Create table based on current category
    const table = document.createElement('table');
    table.className = 'leaderboard-table';

    let tableHTML = '';

    // Category-specific headers and data
    switch (currentLeaderboardOptions.category) {
        case 'speed':
            tableHTML = `
                <thead>
                    <tr>
                        <th>Rank</th>
                        <th>Player</th>
                        <th>Best Time</th>
                        <th>Best Score</th>
                        <th>Games</th>
                        <th>Achievements</th>
                    </tr>
                </thead>
                <tbody>
                    ${entries.map((entry, index) => `
                        <tr ${isCurrentUser(entry.username) ? 'class="current-user"' : ''}>
                            <td class="rank-cell">${getRankIcon(index + 1)}${index + 1}</td>
                            <td class="player-cell">${escapeHTML(entry.username)}</td>
                            <td class="stat-cell">${formatTime(entry.best_time)}</td>
                            <td class="score-cell">${formatScore(entry.best_score)}</td>
                            <td class="games-cell">${entry.games_played}</td>
                            <td class="achievements-cell">${formatAchievements(entry.achievements)}</td>
                        </tr>
                    `).join('')}
                </tbody>
            `;
            break;

        case 'efficiency':
            tableHTML = `
                <thead>
                    <tr>
                        <th>Rank</th>
                        <th>Player</th>
                        <th>Avg PPM</th>
                        <th>Lines Cleared</th>
                        <th>Games</th>
                        <th>Achievements</th>
                    </tr>
                </thead>
                <tbody>
                    ${entries.map((entry, index) => `
                        <tr ${isCurrentUser(entry.username) ? 'class="current-user"' : ''}>
                            <td class="rank-cell">${getRankIcon(index + 1)}${index + 1}</td>
                            <td class="player-cell">${escapeHTML(entry.username)}</td>
                            <td class="stat-cell">${entry.avg_ppm ? entry.avg_ppm.toFixed(1) : 'N/A'}</td>
                            <td class="lines-cell">${entry.total_lines || 0}</td>
                            <td class="games-cell">${entry.games_played}</td>
                            <td class="achievements-cell">${formatAchievements(entry.achievements)}</td>
                        </tr>
                    `).join('')}
                </tbody>
            `;
            break;

        case 'endurance':
            tableHTML = `
                <thead>
                    <tr>
                        <th>Rank</th>
                        <th>Player</th>
                        <th>Longest Game</th>
                        <th>Total Time</th>
                        <th>Games</th>
                        <th>Achievements</th>
                    </tr>
                </thead>
                <tbody>
                    ${entries.map((entry, index) => `
                        <tr ${isCurrentUser(entry.username) ? 'class="current-user"' : ''}>
                            <td class="rank-cell">${getRankIcon(index + 1)}${index + 1}</td>
                            <td class="player-cell">${escapeHTML(entry.username)}</td>
                            <td class="stat-cell">${formatTime(entry.best_time)}</td>
                            <td class="time-cell">${formatTime(entry.total_time)}</td>
                            <td class="games-cell">${entry.games_played}</td>
                            <td class="achievements-cell">${formatAchievements(entry.achievements)}</td>
                        </tr>
                    `).join('')}
                </tbody>
            `;
            break;

        default: // score
            tableHTML = `
                <thead>
                    <tr>
                        <th>Rank</th>
                        <th>Player</th>
                        <th>Best Score</th>
                        <th>Avg Score</th>
                        <th>Games</th>
                        <th>Achievements</th>
                    </tr>
                </thead>
                <tbody>
                    ${entries.map((entry, index) => `
                        <tr ${isCurrentUser(entry.username) ? 'class="current-user"' : ''}>
                            <td class="rank-cell">${getRankIcon(index + 1)}${index + 1}</td>
                            <td class="player-cell">${escapeHTML(entry.username)}</td>
                            <td class="score-cell">${formatScore(entry.best_score)}</td>
                            <td class="avg-cell">${entry.avg_score.toFixed(0)}</td>
                            <td class="games-cell">${entry.games_played}</td>
                            <td class="achievements-cell">${formatAchievements(entry.achievements)}</td>
                        </tr>
                    `).join('')}
                </tbody>
            `;
    }

    table.innerHTML = tableHTML;
    contentDiv.innerHTML = '';
    contentDiv.appendChild(table);
}

function renderRecentGames(games) {
    const recentDiv = document.getElementById('recent-games-list');

    if (!games || games.length === 0) {
        recentDiv.innerHTML = '<p class="empty-recent">No recent games</p>';
        return;
    }

    const gamesList = games.map(game => `
        <div class="recent-game-item">
            <div class="recent-game-player">${escapeHTML(game.additional_data?.username || 'Unknown')}</div>
            <div class="recent-game-score">${formatScore(game.score)}</div>
            <div class="recent-game-time">${formatTimeAgo(game.played_at)}</div>
            ${game.additional_data?.lines_cleared ? `<div class="recent-game-lines">${game.additional_data.lines_cleared} lines</div>` : ''}
        </div>
    `).join('');

    recentDiv.innerHTML = gamesList;
}

// Helper functions for formatting
function getRankIcon(rank) {
    switch (rank) {
        case 1: return '🥇 ';
        case 2: return '🥈 ';
        case 3: return '🥉 ';
        default: return '';
    }
}

function formatScore(score) {
    return score.toLocaleString();
}

function formatTime(seconds) {
    if (!seconds || seconds === 0) return 'N/A';
    const mins = Math.floor(seconds / 60);
    const secs = Math.floor(seconds % 60);
    return `${mins}:${secs.toString().padStart(2, '0')}`;
}

function formatTimeAgo(dateStr) {
    const date = new Date(dateStr);
    const now = new Date();
    const diffMs = now - date;
    const diffMins = Math.floor(diffMs / 60000);
    const diffHours = Math.floor(diffMs / 3600000);
    const diffDays = Math.floor(diffMs / 86400000);

    if (diffMins < 1) return 'Just now';
    if (diffMins < 60) return `${diffMins}m ago`;
    if (diffHours < 24) return `${diffHours}h ago`;
    return `${diffDays}d ago`;
}

function formatAchievements(achievements) {
    if (!achievements || achievements.length === 0) return '<span class="no-achievements">-</span>';

    const badgeMap = {
        'First Game': '🎮',
        'Getting Started': '🚀',
        'Dedicated Player': '⭐',
        'Tetris Master': '👑',
        'High Scorer': '🎯',
        'Score Champion': '🏆',
        'Legendary Score': '💎',
        'Speed Demon': '⚡',
        'Lightning Fast': '🔥',
        'First Tetris': '🧩',
        'Tetris Expert': '🎪',
        'Line Clearer': '📏',
        'Line Master': '🏁'
    };

    const badges = achievements.slice(0, 3).map(achievement =>
        `<span class="achievement-badge" title="${achievement}">${badgeMap[achievement] || '🏅'}</span>`
    ).join('');

    const extraCount = achievements.length > 3 ? `<span class="extra-achievements">+${achievements.length - 3}</span>` : '';

    return badges + extraCount;
}

function isCurrentUser(username) {
    const currentUser = localStorage.getItem('username');
    return currentUser && currentUser === username;
}

function escapeHTML(str) {
    const p = document.createElement('p');
    p.appendChild(document.createTextNode(str));
    return p.innerHTML;
}

let ws;
const canvas = document.getElementById('game-canvas');
const ctx = canvas.getContext('2d');
const nextPieceCanvas = document.getElementById('next-piece-canvas');
const nextPieceCtx = nextPieceCanvas.getContext('2d');
const holdPieceCanvas = document.getElementById('hold-piece-canvas');
const holdPieceCtx = holdPieceCanvas.getContext('2d');
const TILE_SIZE = 40; // Size of each block in pixels (400px canvas / 10 tiles = 40px per tile)

const COLORS = [
    '#000000', // 0: Background
    '#00FFFF', // 1: I piece (Cyan)
    '#FFFF00', // 2: O piece (Yellow)
    '#800080', // 3: T piece (Purple)
    '#00FF00', // 4: S piece (Green)
    '#FF0000', // 5: Z piece (Red)
    '#FFA500', // 6: L piece (Orange)
    '#0000FF', // 7: J piece (Blue)
];

function cleanupGame() {
    // Close existing WebSocket connection
    if (ws && ws.readyState !== WebSocket.CLOSED) {
        ws.close();
    }
    ws = null;

    // Remove event listeners
    document.removeEventListener('keydown', handleKeyPress);

    // Clear canvas if it exists
    if (typeof ctx !== 'undefined' && ctx) {
        ctx.fillStyle = '#000';
        ctx.fillRect(0, 0, canvas.width, canvas.height);
    }

    // Clear next piece canvas if it exists
    if (typeof nextPieceCtx !== 'undefined' && nextPieceCtx) {
        nextPieceCtx.fillStyle = '#000';
        nextPieceCtx.fillRect(0, 0, nextPieceCanvas.width, nextPieceCanvas.height);
    }

    // Clear hold piece canvas if it exists
    if (typeof holdPieceCtx !== 'undefined' && holdPieceCtx) {
        holdPieceCtx.fillStyle = '#000';
        holdPieceCtx.fillRect(0, 0, holdPieceCanvas.width, holdPieceCanvas.height);
    }

    console.log('Game cleaned up');
}

function startGame(gameType, startLevel = 1) {
    // Clean up any existing game first
    cleanupGame();

    const protocol = window.location.protocol === 'https' ? 'wss' : 'ws';
    const wsURL = `${protocol}://${window.location.host}/ws/game`;

    ws = new WebSocket(wsURL);

    ws.onopen = () => {
        console.log(`Connected to ${gameType} game server.`);
        document.addEventListener('keydown', handleKeyPress);

        // Send the starting level to the server
        if (startLevel > 1) {
            ws.send(JSON.stringify({ type: 'setLevel', level: startLevel }));
        }
    };

    ws.onmessage = (event) => {
        const gameState = JSON.parse(event.data);
        if (gameState.type === 'gameOver') {
            showGameOverScreen(gameState.score);
            ws.close();
        } else {
            renderGame(gameState);
            updateGameInfo(gameState);

            // Check for game over in normal game state
            if (gameState.gameOver) {
                showGameOverScreen(gameState.score, gameState.stats);
                ws.close();
            }
        }
    };

    ws.onclose = () => {
        console.log('Disconnected from game server.');
        // Cleanup is handled by cleanupGame() function
    };

    ws.onerror = (error) => {
        console.error('WebSocket Error:', error);
    };
}

function handleKeyPress(event) {
    if (!ws || ws.readyState !== WebSocket.OPEN) return;

    let action = '';
    switch (event.key) {
        case 'ArrowLeft':
            action = 'left';
            break;
        case 'ArrowRight':
            action = 'right';
            break;
        case 'ArrowDown':
            action = 'down';
            break;
        case 'ArrowUp':
            action = 'hardDrop';
            break;
        case ' ': // Spacebar for rotation
        case 'x':
        case 'X':
            action = 'rotate';
            break;
        case 'c':
        case 'C':
        case 'Shift': // Common hold keys in Tetris games
            action = 'hold';
            break;
        case 'Escape':
            // Toggle game menu
            const gameMenuOverlay = document.getElementById('game-menu-overlay');
            if (gameMenuOverlay.classList.contains('hidden')) {
                // Open game menu and pause
                gameMenuOverlay.classList.remove('hidden');

                // Hide restart button if in multiplayer mode
                const restartBtn = document.getElementById('restart-game-btn');
                if (restartBtn) {
                    if (window.isMultiplayer) {
                        restartBtn.style.display = 'none';
                    } else {
                        restartBtn.style.display = 'block';
                    }
                }

                if (ws && ws.readyState === WebSocket.OPEN) {
                    ws.send(JSON.stringify({ type: 'input', key: 'pause' }));
                }
            } else {
                // Close game menu and resume
                gameMenuOverlay.classList.add('hidden');
                if (ws && ws.readyState === WebSocket.OPEN) {
                    ws.send(JSON.stringify({ type: 'input', key: 'pause' }));
                }
            }
            event.preventDefault();
            return;
        default:
            return;
    }
    ws.send(JSON.stringify({ type: 'input', key: action }));
    event.preventDefault(); // Prevent arrow keys from scrolling the page
}

function renderGame(state) {
    // Clear canvas
    ctx.fillStyle = '#000';
    ctx.fillRect(0, 0, canvas.width, canvas.height);

    // Draw the game board
    if (state.board) {
        for (let row = 0; row < state.board.length; row++) {
            for (let col = 0; col < state.board[row].length; col++) {
                const tileValue = state.board[row][col];
                if (tileValue > 0) {
                    ctx.fillStyle = COLORS[tileValue];
                    ctx.fillRect(col * TILE_SIZE, row * TILE_SIZE, TILE_SIZE - 1, TILE_SIZE - 1);
                }
            }
        }
    }

    // Draw the ghost piece (outline of where current piece will land)
    if (state.ghostPiece && state.ghostPiece.shape && !state.paused) {
        ctx.strokeStyle = '#666666'; // Gray outline
        ctx.lineWidth = 2;
        for (let row = 0; row < state.ghostPiece.shape.length; row++) {
            for (let col = 0; col < state.ghostPiece.shape[row].length; col++) {
                if (state.ghostPiece.shape[row][col] === 1) {
                    const x = (state.ghostPiece.x + col) * TILE_SIZE;
                    const y = (state.ghostPiece.y + row) * TILE_SIZE;
                    ctx.strokeRect(x, y, TILE_SIZE - 1, TILE_SIZE - 1);
                }
            }
        }
    }
}

function updateGameInfo(state) {
    document.getElementById('score').textContent = state.score;
    document.getElementById('lines').textContent = state.lines;
    document.getElementById('level').textContent = state.level;

    // Update statistics if available
    if (state.stats) {
        const minutes = Math.floor(state.stats.timePlayed / 60);
        const seconds = state.stats.timePlayed % 60;
        const timeStr = `${minutes}:${seconds.toString().padStart(2, '0')}`;

        document.getElementById('time').textContent = timeStr;
        document.getElementById('ppm').textContent = `${state.stats.ppm.toFixed(1)} PPM`;
        document.getElementById('stat-single').textContent = state.stats.lineStats[0];
        document.getElementById('stat-double').textContent = state.stats.lineStats[1];
        document.getElementById('stat-triple').textContent = state.stats.lineStats[2];
        document.getElementById('stat-tetris').textContent = state.stats.lineStats[3];
    }

    // Render next piece
    renderNextPiece(state.nextPiece);

    // Render hold piece
    renderHoldPiece(state.holdPiece);

    // Render hold piece
    renderHoldPiece(state.holdPiece);
}

function showGameOverScreen(finalScore, stats = null) {
    // Submit score to leaderboard if user is logged in
    const token = getAuthToken();
    if (token) {
        const metadata = stats ? {
            time_played: stats.timePlayed,
            pieces_placed: stats.piecesPlaced,
            ppm: stats.ppm,
            line_stats: stats.lineStats
        } : {};

        submitScore('tetris', finalScore, metadata)
            .then(() => {
                console.log('Score submitted successfully!');
            })
            .catch(error => {
                console.error('Failed to submit score:', error);
            });
    }

    // Create game over overlay
    const overlay = document.createElement('div');
    overlay.style.cssText = `
        position: fixed;
        top: 0;
        left: 0;
        width: 100%;
        height: 100%;
        background: rgba(0, 0, 0, 0.8);
        display: flex;
        justify-content: center;
        align-items: center;
        z-index: 1000;
        font-family: 'Courier New', monospace;
        color: white;
    `;

    const gameOverContent = document.createElement('div');
    gameOverContent.style.cssText = `
        background: #000;
        border: 2px solid #fff;
        padding: 30px;
        text-align: center;
        max-width: 400px;
        width: 90%;
    `;

    let content = `
        <h2 style="margin-top: 0; color: #fff;">GAME OVER</h2>
        <p>Final Score: <strong>${finalScore}</strong></p>
    `;

    // Add authentication status message
    if (token) {
        content += `<p style="color: #00ff00; font-size: 0.9em;">✓ Score submitted to leaderboard!</p>`;
    } else {
        content += `<p style="color: #ffaa00; font-size: 0.9em;">Login to save your score to the leaderboard</p>`;
    }

    if (stats) {
        const minutes = Math.floor(stats.timePlayed / 60);
        const seconds = stats.timePlayed % 60;
        const timeStr = `${minutes}:${seconds.toString().padStart(2, '0')}`;

        content += `
            <div style="margin: 20px 0; text-align: left;">
                <div>Time Played: ${timeStr}</div>
                <div>Pieces Placed: ${stats.piecesPlaced}</div>
                <div>Pieces per Minute: ${stats.ppm.toFixed(1)}</div>
                <div style="margin-top: 10px;">Line Clears:</div>
                <div style="margin-left: 20px;">
                    <div>Singles: ${stats.lineStats[0]}</div>
                    <div>Doubles: ${stats.lineStats[1]}</div>
                    <div>Triples: ${stats.lineStats[2]}</div>
                    <div>Tetris: ${stats.lineStats[3]}</div>
                </div>
            </div>
        `;
    }

    content += `
        <div style="margin-top: 20px;">
            <button id="restart-game-btn" style="margin-right: 10px; padding: 8px 16px; background: #333; color: #fff; border: 1px solid #fff; font-family: 'Courier New', monospace;">Play Again</button>
            <button id="back-to-menu-from-gameover-btn" style="padding: 8px 16px; background: #333; color: #fff; border: 1px solid #fff; font-family: 'Courier New', monospace;">Main Menu</button>
        </div>
    `;

    gameOverContent.innerHTML = content;
    overlay.appendChild(gameOverContent);
    document.body.appendChild(overlay);

    // Add event listeners
    document.getElementById('restart-game-btn').addEventListener('click', () => {
        document.body.removeChild(overlay);
        // Start a new game with same level
        const lastLevel = localStorage.getItem('lastSelectedLevel') || 1;
        startGame('tetris', parseInt(lastLevel));
    });

    document.getElementById('back-to-menu-from-gameover-btn').addEventListener('click', () => {
        document.body.removeChild(overlay);
        if (typeof cleanupGame === 'function') {
            cleanupGame();
        }
        // This function should be available from main.js
        if (typeof showView === 'function') {
            showView('mainMenu');
        } else {
            // Fallback - reload the page
            window.location.reload();
        }
    });
}

function renderNextPiece(nextPiece) {
    // Clear the next piece canvas
    nextPieceCtx.fillStyle = '#000';
    nextPieceCtx.fillRect(0, 0, nextPieceCanvas.width, nextPieceCanvas.height);

    if (nextPiece) {
        const pieceSize = 20; // Smaller tiles for next piece display
        const offsetX = (nextPieceCanvas.width - nextPiece[0].length * pieceSize) / 2;
        const offsetY = (nextPieceCanvas.height - nextPiece.length * pieceSize) / 2;

        for (let row = 0; row < nextPiece.length; row++) {
            for (let col = 0; col < nextPiece[row].length; col++) {
                if (nextPiece[row][col] === 1) {
                    nextPieceCtx.fillStyle = '#fff';
                    nextPieceCtx.fillRect(
                        offsetX + col * pieceSize,
                        offsetY + row * pieceSize,
                        pieceSize - 1,
                        pieceSize - 1
                    );
                }
            }
        }
    }
}

function renderHoldPiece(holdPiece) {
    // Clear the hold piece canvas
    holdPieceCtx.fillStyle = '#000';
    holdPieceCtx.fillRect(0, 0, holdPieceCanvas.width, holdPieceCanvas.height);

    if (holdPiece) {
        const pieceSize = 20; // Smaller tiles for hold piece display
        const offsetX = (holdPieceCanvas.width - holdPiece[0].length * pieceSize) / 2;
        const offsetY = (holdPieceCanvas.height - holdPiece.length * pieceSize) / 2;

        for (let row = 0; row < holdPiece.length; row++) {
            for (let col = 0; col < holdPiece[row].length; col++) {
                if (holdPiece[row][col] === 1) {
                    holdPieceCtx.fillStyle = '#aaa'; // Slightly dimmed to show it's held
                    holdPieceCtx.fillRect(
                        offsetX + col * pieceSize,
                        offsetY + row * pieceSize,
                        pieceSize - 1,
                        pieceSize - 1
                    );
                }
            }
        }
    }
}

// Multiplayer game functionality
function startMultiplayerGame(roomId, multiplayerWs) {
    console.log('Starting multiplayer game for room:', roomId);

    // Clean up any existing game first
    cleanupGame();

    // Store multiplayer WebSocket reference
    window.multiplayerWs = multiplayerWs;
    window.isMultiplayer = true;
    window.currentRoomId = roomId;

    // Connect to game WebSocket
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsURL = `${protocol}//${window.location.host}/ws/game?room=${roomId}&multiplayer=true`;

    ws = new WebSocket(wsURL);

    ws.onopen = () => {
        console.log('Connected to multiplayer game server for room:', roomId);
        document.addEventListener('keydown', handleKeyPress);

        // Send multiplayer initialization
        ws.send(JSON.stringify({
            type: 'multiplayerInit',
            roomId: roomId
        }));
    };

    ws.onmessage = (event) => {
        const gameState = JSON.parse(event.data);

        if (gameState.type === 'multiplayerUpdate') {
            handleMultiplayerUpdate(gameState);
        } else if (gameState.type === 'gameOver') {
            handleMultiplayerGameOver(gameState);
        } else if (gameState.type === 'opponentFinished') {
            handleOpponentFinished(gameState);
        } else {
            // Regular game state update
            renderGame(gameState);
            updateGameInfo(gameState);

            // Send game state to other players
            if (window.multiplayerWs && window.multiplayerWs.readyState === WebSocket.OPEN) {
                window.multiplayerWs.send(JSON.stringify({
                    type: 'game_state',
                    data: {
                        score: gameState.score,
                        lines: gameState.lines,
                        level: gameState.level,
                        board: gameState.board
                    }
                }));
            }

            // Check for game over - CRITICAL: This needs to trigger multiplayer game over
            if (gameState.gameOver) {
                console.log('Game over detected in multiplayer game!');
                handleMultiplayerGameOver(gameState);
            }
        }
    }; ws.onclose = () => {
        console.log('Disconnected from multiplayer game server.');
        cleanupMultiplayerGame();
    };

    ws.onerror = (error) => {
        console.error('Multiplayer Game WebSocket Error:', error);
    };
}

function handleMultiplayerUpdate(gameState) {
    console.log('Multiplayer update:', gameState);
    // Update opponent board display if implemented
    if (gameState.opponents) {
        updateOpponentDisplays(gameState.opponents);
    }
}

function handleMultiplayerGameOver(gameState) {
    console.log('Multiplayer game over:', gameState);

    // Send final results to multiplayer system
    if (window.multiplayerWs && window.multiplayerWs.readyState === WebSocket.OPEN) {
        window.multiplayerWs.send(JSON.stringify({
            type: 'player_finished',
            data: {
                score: gameState.score,
                lines: gameState.lines,
                stats: gameState.stats,
                position: gameState.position || 1
            }
        }));
    }

    // Show multiplayer game over screen
    showMultiplayerGameOverScreen(gameState);
}

function handleOpponentFinished(gameState) {
    console.log('Opponent finished:', gameState);
    // Show notification that opponent finished
    showOpponentFinishedNotification(gameState.playerName, gameState.position);
}

function showMultiplayerGameOverScreen(gameState) {
    // Create multiplayer game over overlay
    const overlay = document.createElement('div');
    overlay.className = 'game-over-overlay';
    overlay.style.cssText = `
        position: fixed;
        top: 0;
        left: 0;
        width: 100%;
        height: 100%;
        background: rgba(0, 0, 0, 0.9);
        display: flex;
        align-items: center;
        justify-content: center;
        z-index: 10000;
        color: white;
        text-align: center;
    `;

    const content = document.createElement('div');
    content.innerHTML = `
        <h2>Game Over!</h2>
        <p>Your Score: ${formatScore(gameState.score)}</p>
        <p>Lines Cleared: ${gameState.lines || 0}</p>
        <p>Position: ${gameState.position || 'Unknown'}</p>
        <button onclick="returnToLobby()" style="margin: 10px; padding: 10px 20px; font-size: 16px;">Return to Lobby</button>
        <button onclick="returnToMenu()" style="margin: 10px; padding: 10px 20px; font-size: 16px;">Main Menu</button>
    `;

    overlay.appendChild(content);
    document.body.appendChild(overlay);
}

function showOpponentFinishedNotification(playerName, position) {
    const notification = document.createElement('div');
    notification.style.cssText = `
        position: fixed;
        top: 100px;
        right: 20px;
        background: #007bff;
        color: white;
        padding: 12px 20px;
        border-radius: 4px;
        z-index: 9999;
        font-size: 14px;
    `;
    notification.textContent = `${playerName} finished in position ${position}!`;

    document.body.appendChild(notification);

    setTimeout(() => {
        if (notification.parentNode) {
            notification.parentNode.removeChild(notification);
        }
    }, 3000);
}

function updateOpponentDisplays(opponents) {
    // This would update opponent board displays if we had them in the UI
    console.log('Updating opponent displays:', opponents);
}

function cleanupMultiplayerGame() {
    window.isMultiplayer = false;
    window.currentRoomId = null;
    window.multiplayerWs = null;
    cleanupGame();
}

function returnToLobby() {
    // Remove game over overlay
    const overlay = document.querySelector('.game-over-overlay');
    if (overlay) {
        overlay.remove();
    }

    cleanupMultiplayerGame();

    // Return to multiplayer lobby
    if (window.showView && window.multiplayerManager) {
        window.showView('multiplayer');
        window.multiplayerManager.showTab('lobby');
    }
}

function returnToMenu() {
    // Remove game over overlay
    const overlay = document.querySelector('.game-over-overlay');
    if (overlay) {
        overlay.remove();
    }

    cleanupMultiplayerGame();

    // Return to main menu
    if (window.showView) {
        window.showView('mainMenu');
    }
}

// Make functions globally available
window.startMultiplayerGame = startMultiplayerGame;
window.returnToLobby = returnToLobby;
window.returnToMenu = returnToMenu;
