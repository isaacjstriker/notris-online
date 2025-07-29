// This file will contain placeholder game logic and leaderboard rendering.

async function loadLeaderboard() {
    const contentDiv = document.getElementById('leaderboard-content');
    const gameType = 'tetris'; // Hardcoded to tetris

    contentDiv.innerHTML = '<p>Loading...</p>';

    try {
        const entries = await getLeaderboard(gameType);
        if (entries.length === 0) {
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
const TILE_SIZE = 30; // Size of each block in pixels

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
            alert(`Game Over! Final Score: ${gameState.score}`);
            ws.close();
        } else {
            renderGame(gameState);
            updateGameInfo(gameState);
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
            action = 'rotate';
            break;
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
}

function updateGameInfo(state) {
    document.getElementById('score-display').textContent = state.score;
    document.getElementById('lines-display').textContent = state.lines;
    document.getElementById('level-display').textContent = state.level;

    // Render next piece
    renderNextPiece(state.nextPiece);
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
