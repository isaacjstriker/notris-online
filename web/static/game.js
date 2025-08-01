const GAME_CONFIG = {
    TILE_SIZE: 40,
    SMALL_TILE_SIZE: 20,
    BOARD_WIDTH: 10,
    BOARD_HEIGHT: 20,
    CANVAS_SETUP_RETRY_DELAY: 100,
    CANVAS_SETUP_MAX_ATTEMPTS: 5,
    GAME_OVER_DISPLAY_DELAY: 3000,
    RECONNECTION_DELAY: 3000,
    NOTIFICATION_DELAY: 1000,
    DEFAULT_STARTING_LEVEL: 1,
    RECENT_GAMES_LIMIT: 5,
    WS_RECONNECT_ATTEMPTS: 3,
    WS_HEARTBEAT_INTERVAL: 30000,
    MODAL_Z_INDEX: 1000
};

const PIECE_COLORS = [
    '#000000',
    '#00FFFF',
    '#0000FF',
    '#FFA500',
    '#FFFF00',
    '#FF0000',
    '#800080',
    '#00FF00'
];

const COLORS = PIECE_COLORS;

const logger = {
    info: (msg, data) => console.log(msg, data || ''),
    debug: (msg, data) => console.log(msg, data || ''),
    error: (msg, data) => console.error(msg, data || ''),
    warn: (msg, data) => console.warn(msg, data || '')
};

let currentLeaderboardOptions = {
    period: 'all',
    limit: 15,
    includeAchievements: true
};

async function loadLeaderboard() {
    const contentDiv = document.getElementById('leaderboard-content');
    const gameType = 'tetris';

    contentDiv.innerHTML = '<div class="loading-spinner">Loading leaderboard...</div>';

    try {
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

    const table = document.createElement('table');
    table.className = 'leaderboard-table';

    const tableHTML = `
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
let canvas, ctx, nextPieceCanvas, nextPieceCtx, holdPieceCanvas, holdPieceCtx;

function initializeSingleplayerCanvases() {
    canvas = document.getElementById('game-canvas');
    ctx = canvas.getContext('2d');
    nextPieceCanvas = document.getElementById('next-piece-canvas');
    nextPieceCtx = nextPieceCanvas.getContext('2d');
    holdPieceCanvas = document.getElementById('hold-piece-canvas');
    holdPieceCtx = holdPieceCanvas.getContext('2d');

    if (!canvas || !ctx || !nextPieceCanvas || !nextPieceCtx || !holdPieceCanvas || !holdPieceCtx) {
        console.error('Failed to initialize singleplayer canvases');
        return false;
    }

    return true;
}

function cleanupGame() {
    if (typeof ws !== 'undefined' && ws) {
        ws.close();
        ws = null;
    }

    document.removeEventListener('keydown', handleKeyPress);

    if (canvas && ctx) {
        ctx.fillStyle = '#000';
        ctx.fillRect(0, 0, canvas.width, canvas.height);
    }

    if (nextPieceCanvas && nextPieceCtx) {
        nextPieceCtx.fillStyle = '#000';
        nextPieceCtx.fillRect(0, 0, nextPieceCanvas.width, nextPieceCanvas.height);
    }

    if (holdPieceCanvas && holdPieceCtx) {
        holdPieceCtx.fillStyle = '#000';
        holdPieceCtx.fillRect(0, 0, holdPieceCanvas.width, holdPieceCanvas.height);
    }
}

function startGame(gameType, startLevel = GAME_CONFIG.DEFAULT_STARTING_LEVEL) {
    cleanupGame();

    // Initialize canvases for singleplayer
    if (!initializeSingleplayerCanvases()) {
        console.error('Cannot start game - canvas initialization failed');
        alert('Failed to initialize game. Please refresh the page and try again.');
        return;
    }

    const protocol = window.location.protocol === 'https' ? 'wss' : 'ws';
    const wsURL = `${protocol}://${window.location.host}/ws/game`;

    ws = new WebSocket(wsURL);

    ws.onopen = () => {
        logger.info(`Connected to ${gameType} game server`);
        document.addEventListener('keydown', handleKeyPress);

        // Immediately update the level display to match the starting level
        document.getElementById('level').textContent = startLevel;

        if (startLevel > GAME_CONFIG.DEFAULT_STARTING_LEVEL) {
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

            if (gameState.gameOver) {
                showGameOverScreen(gameState.score, gameState.stats);
                ws.close();
            }

            if (window.isMultiplayer && window.multiplayerWs && window.multiplayerWs.readyState === WebSocket.OPEN) {
                broadcastGameStateToOpponents(gameState);
            }
        }
    };

    ws.onclose = () => {
        logger.info('Disconnected from game server');
    };

    ws.onerror = (error) => {
        logger.error('WebSocket Error', error);
    };
}

function broadcastGameStateToOpponents(gameState) {
    if (!window.multiplayerWs || window.multiplayerWs.readyState !== WebSocket.OPEN) {
        return;
    }

    try {
        const message = {
            type: 'game_state',
            room_id: window.currentRoomId || '',
            user_id: getCurrentUser()?.id || 0,
            data: {
                board: gameState.board,
                score: gameState.score,
                level: gameState.level,
                lines: gameState.lines,
                gameOver: gameState.gameOver,
                paused: gameState.paused,
                timestamp: Date.now()
            }
        };

        logger.debug('Broadcasting game state to opponents', { messageType: message.type });
        window.multiplayerWs.send(JSON.stringify(message));
    } catch (error) {
        logger.error('Error broadcasting game state', error);
    }
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
        case ' ':
        case 'x':
        case 'X':
            action = 'rotate';
            break;
        case 'c':
        case 'C':
        case 'Shift':
            action = 'hold';
            break;
        case 'Escape':
            const gameMenuOverlay = document.getElementById('game-menu-overlay');
            if (gameMenuOverlay.classList.contains('hidden')) {
                gameMenuOverlay.classList.remove('hidden');

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
    event.preventDefault();
}

function handleMultiplayerKeyPress(event) {
    if (!window.multiplayerWs || window.multiplayerWs.readyState !== WebSocket.OPEN) return;

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
        case ' ':
        case 'x':
        case 'X':
            action = 'rotate';
            break;
        case 'c':
        case 'C':
        case 'Shift':
            action = 'hold';
            break;
        case 'Escape':
            const multiplayerGameMenuOverlay = document.getElementById('multiplayer-game-menu-overlay');
            if (multiplayerGameMenuOverlay && multiplayerGameMenuOverlay.classList.contains('hidden')) {
                multiplayerGameMenuOverlay.classList.remove('hidden');
                window.multiplayerWs.send(JSON.stringify({
                    type: 'game_input',
                    room_id: window.currentRoomId,
                    data: {
                        action: 'pause'
                    }
                }));
            } else if (multiplayerGameMenuOverlay) {
                multiplayerGameMenuOverlay.classList.add('hidden');
                window.multiplayerWs.send(JSON.stringify({
                    type: 'game_input',
                    room_id: window.currentRoomId,
                    data: {
                        action: 'pause'
                    }
                }));
            }
            event.preventDefault();
            return;
        default:
            return;
    }

    window.multiplayerWs.send(JSON.stringify({
        type: 'game_input',
        room_id: window.currentRoomId,
        data: {
            action: action
        }
    }));
    event.preventDefault();
}

function renderGame(state) {
    ctx.fillStyle = '#000';
    ctx.fillRect(0, 0, canvas.width, canvas.height);

    if (state.board) {
        for (let row = 0; row < state.board.length; row++) {
            for (let col = 0; col < state.board[row].length; col++) {
                const tileValue = state.board[row][col];
                if (tileValue > 0) {
                    ctx.fillStyle = PIECE_COLORS[tileValue];
                    ctx.fillRect(col * GAME_CONFIG.TILE_SIZE, row * GAME_CONFIG.TILE_SIZE, GAME_CONFIG.TILE_SIZE - 1, GAME_CONFIG.TILE_SIZE - 1);
                }
            }
        }
    }

    if (state.ghostPiece && state.ghostPiece.shape && !state.paused) {
        ctx.strokeStyle = '#666666';
        ctx.lineWidth = 2;
        for (let row = 0; row < state.ghostPiece.shape.length; row++) {
            for (let col = 0; col < state.ghostPiece.shape[row].length; col++) {
                if (state.ghostPiece.shape[row][col] === 1) {
                    const x = (state.ghostPiece.x + col) * GAME_CONFIG.TILE_SIZE;
                    const y = (state.ghostPiece.y + row) * GAME_CONFIG.TILE_SIZE;
                    ctx.strokeRect(x, y, GAME_CONFIG.TILE_SIZE - 1, GAME_CONFIG.TILE_SIZE - 1);
                }
            }
        }
    }
}

function updateGameInfo(state) {
    document.getElementById('score').textContent = state.score;
    document.getElementById('lines').textContent = state.lines;
    document.getElementById('level').textContent = state.level;

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

    renderNextPiece(state.nextPiece);

    renderHoldPiece(state.holdPiece);
}

function showGameOverScreen(finalScore, stats = null) {
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
                logger.info('Score submitted successfully');
            })
            .catch(error => {
                logger.error('Failed to submit score', error);
            });
    }

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

    document.getElementById('restart-game-btn').addEventListener('click', () => {
        document.body.removeChild(overlay);
        const lastLevel = localStorage.getItem('lastSelectedLevel') || 1;
        startGame('tetris', parseInt(lastLevel));
    });

    document.getElementById('back-to-menu-from-gameover-btn').addEventListener('click', () => {
        document.body.removeChild(overlay);
        if (typeof cleanupGame === 'function') {
            cleanupGame();
        }
        if (typeof showView === 'function') {
            showView('mainMenu');
        } else {
            window.location.reload();
        }
    });
}

function renderNextPiece(nextPiece) {
    nextPieceCtx.fillStyle = '#000';
    nextPieceCtx.fillRect(0, 0, nextPieceCanvas.width, nextPieceCanvas.height);

    if (nextPiece) {
        const pieceSize = GAME_CONFIG.SMALL_TILE_SIZE;
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
    holdPieceCtx.fillStyle = '#000';
    holdPieceCtx.fillRect(0, 0, holdPieceCanvas.width, holdPieceCanvas.height);

    if (holdPiece) {
        const pieceSize = GAME_CONFIG.SMALL_TILE_SIZE;
        const offsetX = (holdPieceCanvas.width - holdPiece[0].length * pieceSize) / 2;
        const offsetY = (holdPieceCanvas.height - holdPiece.length * pieceSize) / 2;

        for (let row = 0; row < holdPiece.length; row++) {
            for (let col = 0; col < holdPiece[row].length; col++) {
                if (holdPiece[row][col] === 1) {
                    holdPieceCtx.fillStyle = '#aaa';
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

function setupMultiplayerCanvasesWithRetry(maxAttempts = GAME_CONFIG.CANVAS_SETUP_MAX_ATTEMPTS) {
    let attempts = 0;

    const attemptSetup = () => {
        try {
            logger.debug(`Attempting to set up canvases (attempt ${attempts + 1})`);

            setupMultiplayerCanvases();

            if (!window.player1Canvas || !window.player2Canvas) {
                if (attempts < maxAttempts - 1) {
                    attempts++;
                    logger.debug(`Canvas setup failed, retrying in ${GAME_CONFIG.CANVAS_SETUP_RETRY_DELAY}ms...`);
                    setTimeout(attemptSetup, GAME_CONFIG.CANVAS_SETUP_RETRY_DELAY);
                } else {
                    throw new Error(`Failed to set up canvases after ${maxAttempts} attempts`);
                }
            } else {
                logger.debug('Canvas setup successful');
            }
        } catch (error) {
            logger.error('Canvas setup error', error);
            throw error;
        }
    };

    attemptSetup();
}

function initializeMultiplayerState(roomId, multiplayerWs) {
    try {
        cleanupGame();

        window.multiplayerWs = multiplayerWs;
        window.isMultiplayer = true;
        window.currentRoomId = roomId;

        window.gameState = null;
        window.opponentGameState = null;

        window.multiplayerGameStartTime = Date.now();
        window.multiplayerTimerInterval = setInterval(updateMultiplayerTimer, 1000);

        return true;
    } catch (error) {
        logger.error('Failed to initialize multiplayer state', error);
        return false;
    }
}

function updateMultiplayerTimer() {
    if (!window.multiplayerGameStartTime) return;

    const elapsed = Math.floor((Date.now() - window.multiplayerGameStartTime) / 1000);
    const minutes = Math.floor(elapsed / 60);
    const seconds = elapsed % 60;
    const timeStr = `${minutes}:${seconds.toString().padStart(2, '0')}`;

    const timerElement = document.getElementById('match-timer');
    if (timerElement) {
        timerElement.textContent = timeStr;
    }
}

function sendMultiplayerGameStart(roomId) {
    try {
        if (window.multiplayerWs && window.multiplayerWs.readyState === WebSocket.OPEN) {
            const message = {
                type: 'start_multiplayer_game',
                room_id: roomId
            };

            window.multiplayerWs.send(JSON.stringify(message));
            logger.debug('Sent start_multiplayer_game message via room WebSocket');
            return true;
        } else {
            logger.warn('Cannot send multiplayer game start - WebSocket not ready');
            return false;
        }
    } catch (error) {
        logger.error('Failed to send multiplayer game start message', error);
        return false;
    }
}

function startMultiplayerGame(roomId, multiplayerWs, startingLevel = GAME_CONFIG.DEFAULT_STARTING_LEVEL) {
    logger.info('Starting multiplayer game for room', { roomId });

    try {
        if (!initializeMultiplayerState(roomId, multiplayerWs)) {
            throw new Error('Failed to initialize multiplayer state');
        }

        setupMultiplayerCanvasesWithRetry();

        document.addEventListener('keydown', handleMultiplayerKeyPress);

        if (!sendMultiplayerGameStart(roomId)) {
            throw new Error('Failed to send game start message');
        }

        logger.info('Multiplayer game setup complete - using room WebSocket for all communication');
    } catch (error) {
        logger.error('Failed to start multiplayer game', error);
        cleanupGame();
        throw error;
    }
}

function setupMultiplayerCanvases() {
    console.log('Setting up multiplayer canvases...');
    console.log('Current view visibility check:', {
        multiplayerGameView: document.getElementById('multiplayer-game-view'),
        isHidden: document.getElementById('multiplayer-game-view')?.classList.contains('hidden'),
        displayStyle: document.getElementById('multiplayer-game-view')?.style.display
    });

    window.player1Canvas = document.getElementById('player1-canvas');
    if (!window.player1Canvas) {
        console.error('Player 1 canvas not found!');
        return;
    }
    window.player1Ctx = window.player1Canvas.getContext('2d');

    window.player1NextCanvas = document.getElementById('player1-next-canvas');
    window.player1NextCtx = window.player1NextCanvas?.getContext('2d');

    window.player1HoldCanvas = document.getElementById('player1-hold-canvas');
    window.player1HoldCtx = window.player1HoldCanvas?.getContext('2d');

    window.player2Canvas = document.getElementById('player2-canvas');
    if (!window.player2Canvas) {
        console.error('Player 2 canvas not found! Checking if element exists in DOM...');
        console.error('DOM search result:', document.querySelector('#player2-canvas'));
        console.error('Current view visibility:', {
            multiplayerGame: document.getElementById('multiplayer-game-view')?.classList.contains('hidden'),
            multiplayerGameView: !!document.getElementById('multiplayer-game-view')
        });
        return;
    }
    window.player2Ctx = window.player2Canvas.getContext('2d');

    window.player2NextCanvas = document.getElementById('player2-next-canvas');
    window.player2NextCtx = window.player2NextCanvas?.getContext('2d');

    window.player2HoldCanvas = document.getElementById('player2-hold-canvas');
    window.player2HoldCtx = window.player2HoldCanvas?.getContext('2d');

    console.log('Multiplayer canvases set up successfully');
    console.log('Player 1 canvas size:', window.player1Canvas.width, 'x', window.player1Canvas.height);
    console.log('Player 2 canvas size:', window.player2Canvas.width, 'x', window.player2Canvas.height);

    console.log('Testing canvas rendering...');
    // Test that both canvases are displaying correctly by briefly rendering a dummy piece
    try {
        window.player1Ctx.fillStyle = '#ff0000';
        window.player1Ctx.fillRect(0, 0, 50, 50);
        console.log('Player 1 canvas test render successful');

        window.player2Ctx.fillStyle = '#00ff00';
        window.player2Ctx.fillRect(0, 0, 50, 50);
        console.log('Player 2 canvas test render successful');

        // Clear test renders after a moment
        setTimeout(() => {
            if (window.player1Ctx && window.player1Canvas) {
                window.player1Ctx.fillStyle = '#000';
                window.player1Ctx.fillRect(0, 0, window.player1Canvas.width, window.player1Canvas.height);
            }
            if (window.player2Ctx && window.player2Canvas) {
                window.player2Ctx.fillStyle = '#000';
                window.player2Ctx.fillRect(0, 0, window.player2Canvas.width, window.player2Canvas.height);
            }
        }, GAME_CONFIG.NOTIFICATION_DELAY);
    } catch (error) {
        console.error('Canvas test render failed:', error);
    }
}

function renderMultiplayerGame(gameState, player) {
    console.log(`=== renderMultiplayerGame called for ${player} ===`);
    console.log('Game state board size:', gameState?.board?.length);

    const canvas = player === 'player1' ? window.player1Canvas : window.player2Canvas;
    const ctx = player === 'player1' ? window.player1Ctx : window.player2Ctx;
    const nextCanvas = player === 'player1' ? window.player1NextCanvas : window.player2NextCanvas;
    const nextCtx = player === 'player1' ? window.player1NextCtx : window.player2NextCtx;
    const holdCanvas = player === 'player1' ? window.player1HoldCanvas : window.player2HoldCanvas;
    const holdCtx = player === 'player1' ? window.player1HoldCtx : window.player2HoldCtx;

    if (!canvas || !ctx) {
        console.error('Canvas or context not found for player:', player);
        console.error('Canvas:', canvas, 'Context:', ctx);
        console.error('Available canvas elements:', {
            player1Canvas: !!window.player1Canvas,
            player2Canvas: !!window.player2Canvas,
            player1Ctx: !!window.player1Ctx,
            player2Ctx: !!window.player2Ctx
        });
        return;
    }

    logger.debug(`Rendering to ${player} canvas`, { width: canvas.width, height: canvas.height });

    try {
        const tileSize = canvas.width / GAME_CONFIG.BOARD_WIDTH;
        logger.debug(`Tile size for ${player}`, { tileSize });

        ctx.fillStyle = '#000';
        ctx.fillRect(0, 0, canvas.width, canvas.height);
        logger.debug(`Cleared ${player} canvas`);

        if (gameState.board && Array.isArray(gameState.board)) {
            console.log(`Drawing board for ${player}, board size: ${gameState.board.length}x${gameState.board[0]?.length}`);
            let tilesDrawn = 0;

            for (let row = 0; row < gameState.board.length; row++) {
                if (gameState.board[row] && Array.isArray(gameState.board[row])) {
                    for (let col = 0; col < gameState.board[row].length; col++) {
                        const tileValue = gameState.board[row][col];
                        if (tileValue > 0 && tileValue < COLORS.length) {
                            ctx.fillStyle = COLORS[tileValue];
                            ctx.fillRect(col * tileSize, row * tileSize, tileSize - 1, tileSize - 1);
                            tilesDrawn++;
                        }
                    }
                }
            }
            console.log(`Drew ${tilesDrawn} tiles for ${player}`);
        } else {
            console.warn(`No valid board data for ${player}:`, gameState.board);
        }

        // Draw the ghost piece (only for player1 - the local player)
        if (player === 'player1' && gameState.ghostPiece && gameState.ghostPiece.shape && !gameState.paused) {
            ctx.strokeStyle = '#666666';
            ctx.lineWidth = 2;
            for (let row = 0; row < gameState.ghostPiece.shape.length; row++) {
                for (let col = 0; col < gameState.ghostPiece.shape[row].length; col++) {
                    if (gameState.ghostPiece.shape[row][col] === 1) {
                        const x = (gameState.ghostPiece.x + col) * tileSize;
                        const y = (gameState.ghostPiece.y + row) * tileSize;
                        ctx.strokeRect(x, y, tileSize - 1, tileSize - 1);
                    }
                }
            }
        }

        // Render next piece (only for player1)
        if (player === 'player1' && gameState.nextPiece && nextCtx && nextCanvas) {
            renderPieceToCanvas(gameState.nextPiece, nextCtx, nextCanvas);
        }

        // Render hold piece (only for player1)
        if (player === 'player1' && gameState.holdPiece && holdCtx && holdCanvas) {
            renderPieceToCanvas(gameState.holdPiece, holdCtx, holdCanvas);
        }

        console.log(`Successfully completed rendering for ${player}`);
    } catch (error) {
        console.error('Error in renderMultiplayerGame for', player, ':', error);
        console.error('Error stack:', error.stack);
        console.error('Game state:', gameState);
    }
}

function renderPieceToCanvas(piece, ctx, canvas) {
    ctx.fillStyle = '#000';
    ctx.fillRect(0, 0, canvas.width, canvas.height);

    if (!piece || piece.length === 0) return;

    const pieceSize = Math.min(canvas.width, canvas.height) / 4;
    const offsetX = (canvas.width - piece[0].length * pieceSize) / 2;
    const offsetY = (canvas.height - piece.length * pieceSize) / 2;

    for (let row = 0; row < piece.length; row++) {
        for (let col = 0; col < piece[row].length; col++) {
            if (piece[row][col] > 0) {
                ctx.fillStyle = COLORS[piece[row][col]];
                ctx.fillRect(
                    offsetX + col * pieceSize,
                    offsetY + row * pieceSize,
                    pieceSize - 1,
                    pieceSize - 1
                );
            }
        }
    }
}

function updateMultiplayerGameInfo(gameState, player) {
    document.getElementById(`${player}-score`).textContent = gameState.score.toLocaleString();
    document.getElementById(`${player}-level`).textContent = gameState.level;
    document.getElementById(`${player}-lines`).textContent = gameState.lines;
}

function cleanupMultiplayerGame() {
    document.removeEventListener('keydown', handleKeyPress);
    document.removeEventListener('keydown', handleMultiplayerKeyPress);
    window.isMultiplayer = false;
    window.multiplayerWs = null;
    window.currentRoomId = null;
    window.opponentGameState = null;

    if (window.multiplayerGameStateInterval) {
        clearInterval(window.multiplayerGameStateInterval);
        window.multiplayerGameStateInterval = null;
    }

    if (window.multiplayerTimerInterval) {
        clearInterval(window.multiplayerTimerInterval);
        window.multiplayerTimerInterval = null;
    }
    window.multiplayerGameStartTime = null;

    if (window.player1Ctx) {
        window.player1Ctx.clearRect(0, 0, window.player1Canvas.width, window.player1Canvas.height);
    }
    if (window.player2Ctx) {
        window.player2Ctx.clearRect(0, 0, window.player2Canvas.width, window.player2Canvas.height);
    }
}

function handleMultiplayerUpdate(gameState) {
    console.log('=== handleMultiplayerUpdate called ===');
    console.log('Received gameState:', JSON.stringify(gameState, null, 2));
    console.log('Canvas elements check:');
    console.log('player2Canvas:', window.player2Canvas);
    console.log('player2Ctx:', window.player2Ctx);

    if (!window.player2Canvas || !window.player2Ctx) {
        console.log('Player2 canvas missing, attempting to re-setup...');
        setupMultiplayerCanvases();

        if (!window.player2Canvas || !window.player2Ctx) {
            console.error('Failed to setup player2 canvas after retry');
            console.error('Multiplayer game view exists:', !!document.getElementById('multiplayer-game-view'));
            console.error('Player2 canvas element exists:', !!document.getElementById('player2-canvas'));
            return;
        }
        console.log('Successfully re-setup player2 canvas');
    }

    if (gameState && gameState.board) {
        console.log('Rendering opponent board with score:', gameState.score);
        console.log('Opponent board dimensions:', gameState.board.length, 'x', gameState.board[0]?.length);

        const opponentGameState = {
            board: gameState.board,
            score: gameState.score || 0,
            level: gameState.level || 1,
            lines: gameState.lines || 0,
            currentPiece: null,
            nextPiece: null,
            holdPiece: null,
            ghostPiece: null,
            gameOver: false,
            paused: false
        };

        window.opponentGameState = opponentGameState;

        try {
            console.log('About to render multiplayer game for player2');
            renderMultiplayerGame(opponentGameState, 'player2');
            updateMultiplayerGameInfo(opponentGameState, 'player2');
            console.log('Successfully rendered opponent game state');

            console.log('Verifying player2 canvas content...');
            const imageData = window.player2Ctx.getImageData(0, 0, window.player2Canvas.width, window.player2Canvas.height);
            const hasContent = Array.from(imageData.data).some(pixel => pixel !== 0);
            console.log('Player2 canvas has visual content:', hasContent);

        } catch (error) {
            console.error('Error rendering opponent game state:', error);
            console.error('Error stack:', error.stack);
        }
    } else {
        console.warn('Invalid or missing game state data:', gameState);
        console.warn('Expected: gameState.board, got:', gameState?.board);
    }
}

window.cleanupMultiplayerGame = cleanupMultiplayerGame;

window.handleMultiplayerUpdate = handleMultiplayerUpdate;

function handleMultiplayerGameOver(gameState) {
    console.log('Multiplayer game over:', gameState);

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

    showMultiplayerGameOverScreen(gameState);
}

function handleOpponentFinished(gameState) {
    console.log('Opponent finished:', gameState);
    showOpponentFinishedNotification(gameState.playerName, gameState.position);
}

function showMultiplayerGameOverScreen(gameState) {
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
    }, GAME_CONFIG.GAME_OVER_DISPLAY_DELAY);
}

function updateOpponentDisplays(opponents) {
    console.log('Updating opponent displays:', opponents);
}

function returnToLobby() {
    const overlay = document.querySelector('.game-over-overlay');
    if (overlay) {
        overlay.remove();
    }

    cleanupMultiplayerGame();

    if (window.showView && window.multiplayerManager) {
        window.showView('multiplayer');
        window.multiplayerManager.showTab('lobby');
    }
}

function returnToMenu() {
    const overlay = document.querySelector('.game-over-overlay');
    if (overlay) {
        overlay.remove();
    }

    cleanupMultiplayerGame();

    if (window.showView) {
        window.showView('mainMenu');
    }
}

window.startMultiplayerGame = startMultiplayerGame;
window.returnToLobby = returnToLobby;
window.returnToMenu = returnToMenu;
