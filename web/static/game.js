// This file will contain placeholder game logic and leaderboard rendering.

async function loadLeaderboard() {
    const contentDiv = document.getElementById('leaderboard-content');
    const gameType = 'tetris'; // Hardcoded to tetris

    contentDiv.innerHTML = '<p>Loading...</p>';

    try {
        const entries = await getLeaderboard(gameType);
        if (!entries || entries.length === 0) {
            contentDiv.innerHTML = '<p>No scores yet. Be the first!</p>';
            return;
        }

        const table = document.createElement('table');
        table.innerHTML = `
            <thead>
                <tr>
                    <th>Rank</th>
                    <th>Player</th>
                    <th>Best Score</th>
                    <th>Games Played</th>
                    <th>Average Score</th>
                </tr>
            </thead>
            <tbody>
                ${entries.map((entry, index) => `
                    <tr>
                        <td>${index + 1}</td>
                        <td>${escapeHTML(entry.username)}</td>
                        <td>${entry.best_score}</td>
                        <td>${entry.games_played}</td>
                        <td>${entry.avg_score.toFixed(1)}</td>
                    </tr>
                `).join('')}
            </tbody>
        `;
        contentDiv.innerHTML = '';
        contentDiv.appendChild(table);
    } catch (error) {
        contentDiv.innerHTML = `<p class="error-message">Failed to load leaderboard: ${error.message}</p>`;
    }
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
