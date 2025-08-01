const ValidationUtils = {
    isValidRoomName: function (name) {
        return name && name.trim().length >= 3 && name.trim().length <= 50;
    }
};

const GameUtils = {
    createEmptyBoard: function () {
        return Array(20).fill().map(() => Array(10).fill(0));
    }
};

class MultiplayerManager {
    constructor() {
        this.ws = null;
        this.currentRoom = null;
        this.isReady = false;
        this.isHost = false;
        this.reconnectInterval = null;
        this.rooms = [];

        this.initializeElements();
        this.attachEventListeners();
    }

    initializeElements() {
        this.roomBrowserTab = document.getElementById('room-browser-tab');
        this.createRoomTab = document.getElementById('create-room-tab');
        this.roomLobbyTab = document.getElementById('room-lobby-tab');

        this.roomBrowserSection = document.getElementById('room-browser');
        this.createRoomSection = document.getElementById('create-room');
        this.roomLobbySection = document.getElementById('room-lobby');

        this.roomsList = document.getElementById('rooms-list');
        this.refreshRoomsBtn = document.getElementById('refresh-rooms-btn');

        this.createRoomForm = document.getElementById('create-room-form');
        this.roomNameInput = document.getElementById('room-name');
        this.privateRoomCheckbox = document.getElementById('private-room');
        this.startingLevelSelect = document.getElementById('starting-level');

        this.lobbyRoomName = document.getElementById('lobby-room-name');
        this.lobbyGameType = document.getElementById('lobby-game-type');
        this.lobbyPlayerCount = document.getElementById('lobby-player-count');
        this.lobbyPlayers = document.getElementById('lobby-players');
        this.readyBtn = document.getElementById('ready-btn');
        this.leaveRoomBtn = document.getElementById('leave-room-btn');
        this.lobbyStatus = document.getElementById('lobby-status');

        this.backBtn = document.getElementById('back-to-menu-from-multiplayer-btn');
    }

    attachEventListeners() {
        this.roomBrowserTab.addEventListener('click', () => this.showTab('browser'));
        this.createRoomTab.addEventListener('click', () => this.showTab('create'));
        this.roomLobbyTab.addEventListener('click', () => this.showTab('lobby'));

        this.refreshRoomsBtn.addEventListener('click', () => this.refreshRooms());

        this.createRoomForm.addEventListener('submit', (e) => this.handleCreateRoom(e));

        this.readyBtn.addEventListener('click', () => this.toggleReady());
        this.leaveRoomBtn.addEventListener('click', () => this.leaveRoom());

        this.backBtn.addEventListener('click', () => this.handleBackToMenu());
    }

    showTab(tabName) {
        this.roomBrowserSection.classList.add('hidden');
        this.createRoomSection.classList.add('hidden');
        this.roomLobbySection.classList.add('hidden');

        this.roomBrowserTab.classList.remove('active');
        this.createRoomTab.classList.remove('active');
        this.roomLobbyTab.classList.remove('active');

        switch (tabName) {
            case 'browser':
                this.roomBrowserSection.classList.remove('hidden');
                this.roomBrowserTab.classList.add('active');
                this.refreshRooms();
                break;
            case 'create':
                this.createRoomSection.classList.remove('hidden');
                this.createRoomTab.classList.add('active');
                break;
            case 'lobby':
                this.roomLobbySection.classList.remove('hidden');
                this.roomLobbyTab.classList.add('active');
                break;
        }
    }

    async refreshRooms() {
        try {
            console.log('Starting to refresh rooms...');
            this.roomsList.innerHTML = '<div class="loading">Loading rooms...</div>';

            console.log('Making API call to /rooms/tetris');
            const response = await apiCall('/rooms/tetris', 'GET');
            console.log('API response received:', response);

            this.rooms = Array.isArray(response) ? response : [];
            console.log('Processed rooms:', this.rooms);
            this.displayRooms();
        } catch (error) {
            console.error('Failed to fetch rooms:', error);
            console.error('Error details:', error.message, error.stack);
            this.roomsList.innerHTML = '<div class="loading">Failed to load rooms. Please try again.</div>';
        }
    }

    displayRooms() {
        if (this.rooms.length === 0) {
            this.roomsList.innerHTML = '<div class="loading">No rooms available. Create one to get started!</div>';
            return;
        }

        const roomsHTML = this.rooms.map(room => {
            const currentPlayers = room.settings?.current_players || 0;
            const isFull = currentPlayers >= room.max_players;
            const isActive = room.status === 'active';
            const startingLevel = room.settings?.starting_level || 1;

            let actionButton = '';
            if (isActive) {
                actionButton = `
                    <div class="room-actions">
                        <button onclick="event.stopPropagation(); multiplayerManager.spectateRoom('${room.id}')" 
                                class="btn-spectate" style="background: #28a745; color: white; border: none; padding: 5px 10px; border-radius: 3px; cursor: pointer; margin-left: 10px;">
                            Spectate
                        </button>
                    </div>
                `;
            } else if (!isFull) {
                actionButton = '<span class="join-hint">Click to join</span>';
            }

            return `
                <div class="room-item ${isActive ? '' : 'joinable'}" ${!isActive && !isFull ? `onclick="multiplayerManager.joinRoom('${room.id}')"` : ''}>
                    <div class="room-header">
                        <span class="room-name">${this.escapeHtml(room.name)}</span>
                        <span class="room-status ${isFull ? 'full' : isActive ? 'active' : ''}">${room.status}</span>
                    </div>
                    <div class="room-info">
                        ${room.game_type} • ${currentPlayers}/${room.max_players} players • Level ${startingLevel}
                        ${actionButton}
                    </div>
                </div>
            `;
        }).join('');

        this.roomsList.innerHTML = roomsHTML;
    }

    async handleCreateRoom(e) {
        e.preventDefault();
        logger.debug('Creating room: form submitted');

        const roomData = {
            name: this.roomNameInput.value.trim(),
            game_type: 'tetris',
            max_players: 2,
            is_private: this.privateRoomCheckbox.checked,
            settings: {
                starting_level: parseInt(this.startingLevelSelect.value, 10) || 1
            }
        };

        logger.debug('Creating room: room data', roomData);

        if (!ValidationUtils.isValidRoomName(roomData.name)) {
            alert('Please enter a valid room name (3-50 characters).');
            return;
        }

        try {
            logger.debug('Creating room: making API call to /rooms');
            const room = await apiCall('/rooms', 'POST', roomData);
            logger.info('Creating room: API call successful', { roomId: room.id });
            this.currentRoom = room;
            this.isHost = true;
            this.connectToRoom(room.id);
            this.showTab('lobby');
            this.updateLobbyDisplay();
        } catch (error) {
            logger.error('Creating room: API call failed', error);
            alert('Failed to create room. Please try again.');
        }
    }

    async joinRoom(roomId) {
        try {
            const response = await apiCall(`/room/${roomId}/join`, 'POST');
            const room = await apiCall(`/room/${roomId}`, 'GET');

            this.currentRoom = room;
            this.isHost = false;
            this.connectToRoom(roomId);
            this.showTab('lobby');
            this.updateLobbyDisplay();
        } catch (error) {
            console.error('Failed to join room:', error);
            alert('Failed to join room. Please try again.');
        }
    }

    connectToRoom(roomId) {
        const token = localStorage.getItem('devware_jwt');
        if (!token) {
            alert('Please log in to play multiplayer.');
            return;
        }

        if (this.ws) {
            this.ws.close();
        }

        const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const wsUrl = `${wsProtocol}//${window.location.host}/ws/room/${roomId}?token=${token}`;

        this.ws = new WebSocket(wsUrl);

        this.ws.onopen = () => {
            console.log('Connected to room:', roomId);
            this.clearReconnectInterval();
        };

        this.ws.onmessage = (event) => {
            try {
                const message = JSON.parse(event.data);
                this.handleWebSocketMessage(message);
            } catch (error) {
                console.error('Failed to parse WebSocket message:', error);
            }
        };

        this.ws.onclose = (event) => {
            console.log('Disconnected from room - Code:', event.code, 'Reason:', event.reason);

            this.clearReconnectInterval();

            // Handle disconnection based on the close code
            if (event.code === 1000) {
                // Normal closure - user intentionally left
                console.log('Normal disconnection');
            } else {
                // Abnormal closure - connection lost
                console.log('Abnormal disconnection - handling as connection loss');
                this.handleDisconnection();
            }
        };

        this.ws.onerror = (error) => {
            console.error('WebSocket error:', error);
            setTimeout(() => {
                if (this.ws && this.ws.readyState !== WebSocket.OPEN) {
                    this.handleDisconnection();
                }
            }, 1000);
        };
    }

    handleWebSocketMessage(message) {
        console.log('=== WebSocket message received ===');
        console.log('Message type:', message.type);
        console.log('Full message:', message);

        switch (message.type) {
            case 'connected':
                console.log('Connected to room:', message.data.room_id);
                break;
            case 'room_update':
                this.handleRoomUpdate(message.data.room);
                break;
            case 'game_start':
                this.handleGameStart(message);
                break;
            case 'multiplayer_game_started':
                this.handleMultiplayerGameStarted(message);
                break;
            case 'player_game_state':
                this.handlePlayerGameState(message);
                break;
            case 'multiplayer_game_ended':
                this.handleMultiplayerGameEnded(message);
                break;
            case 'player_joined':
                console.log('Player joined:', message.data.username);
                break;
            case 'player_left':
            case 'player_disconnected':
                console.log('Player left/disconnected:', message.data.playerName || message.data.username);
                if (this.currentRoom && window.isMultiplayer) {
                    const playerName = message.data.playerName || message.data.username || 'A player';
                    this.showNotification(`${playerName} left the match`, 'warning');
                    console.log('Displayed disconnect notification for:', playerName);
                }
                break;
            case 'room_closed':
                this.handleRoomClosed(message);
                break;
            case 'rooms_updated':
                this.handleRoomsUpdated(message);
                break;
            case 'match_ended':
                this.handleMatchEnded(message);
                break;
            case 'error':
                console.error('Server error:', message.data.error);
                alert(message.data.error);
                break;
            default:
                console.log('Unhandled message type:', message.type);
        }
    }

    handleRoomUpdate(room) {
        console.log('Received room update:', room);
        if (this.currentRoom && room.id === this.currentRoom.id) {
            this.currentRoom = room;
            this.updateLobbyDisplay();
        }
    }

    handleGameStart(message) {
        console.log('Game starting:', message.data);
        if (message.data && message.data.message) {
            this.showNotification(message.data.message, 'success');
        }
        if (this.currentRoom) {
            setTimeout(() => {
                window.startMultiplayerGame(this.currentRoom.id);
            }, 1500);
        }
    }

    handleMultiplayerGameStarted(message) {
        console.log('Multiplayer game started:', message.data);
        if (message.data && message.data.message) {
            this.showNotification(message.data.message, 'info');
        }
        if (window.initializeMultiplayerGameplay) {
            window.initializeMultiplayerGameplay(message.data.starting_level || 1);
        }
    }

    handlePlayerGameState(message) {
        const currentUser = window.currentUser || JSON.parse(localStorage.getItem('currentUser'));
        if (!currentUser) return;

        const isOwnState = message.UserID === currentUser.id;
        const gameState = message.data;

        if (isOwnState) {
            if (window.renderMultiplayerGame) {
                window.renderMultiplayerGame(gameState, 'player1');
            }
            if (window.updateMultiplayerGameInfo) {
                window.updateMultiplayerGameInfo(gameState, 'player1');
            }
        } else {
            if (window.renderMultiplayerGame) {
                window.renderMultiplayerGame(gameState, 'player2');
            }
            if (window.updateMultiplayerGameInfo) {
                window.updateMultiplayerGameInfo(gameState, 'player2');
            }
        }
    }

    handleMultiplayerGameEnded(message) {
        console.log('Multiplayer game ended:', message.data);

        if (message.data && message.data.message) {
            this.showNotification(message.data.message, 'info');
        } else {
            this.showNotification('Game ended!', 'info');
        }

        if (window.cleanupMultiplayerGame) {
            window.cleanupMultiplayerGame();
        }

        setTimeout(() => {
            if (window.showView) {
                window.showView('multiplayer');
            }
            this.showTab('browser');
        }, 2000);
    }

    updateLobbyDisplay() {
        if (!this.currentRoom) return;

        this.lobbyRoomName.textContent = this.currentRoom.name;
        this.lobbyGameType.textContent = this.currentRoom.game_type;
        this.lobbyPlayerCount.textContent = `${this.currentRoom.players.length}/${this.currentRoom.max_players}`;

        const playersHTML = this.currentRoom.players.map(player => `
            <div class="player-item">
                <span class="player-name">${this.escapeHtml(player.username)}</span>
                <span class="player-status ${player.is_ready ? '' : 'not-ready'}">
                    ${player.is_ready ? 'Ready' : 'Not Ready'}
                </span>
            </div>
        `).join('');

        this.lobbyPlayers.innerHTML = playersHTML;

        const currentUser = getCurrentUser();
        if (currentUser) {
            const userPlayer = this.currentRoom.players.find(p => p.username === currentUser.username);
            if (userPlayer) {
                this.isReady = userPlayer.is_ready;
                this.readyBtn.textContent = this.isReady ? 'Not Ready' : 'Ready';
            }
        }

        const allReady = this.currentRoom.players.length > 1 &&
            this.currentRoom.players.every(p => p.is_ready);

        if (this.currentRoom.status === 'playing') {
            this.lobbyStatus.textContent = 'Game in progress...';
        } else if (allReady) {
            this.lobbyStatus.textContent = 'All players ready! Starting game...';
        } else if (this.currentRoom.players.length < 2) {
            this.lobbyStatus.textContent = 'Waiting for more players to join...';
        } else {
            this.lobbyStatus.textContent = 'Waiting for all players to be ready...';
        }

        this.roomLobbyTab.style.display = 'block';
    }

    async toggleReady() {
        if (!this.currentRoom) return;

        try {
            if (this.ws && this.ws.readyState === WebSocket.OPEN) {
                this.isReady = !this.isReady;

                this.ws.send(JSON.stringify({
                    type: 'player_ready',
                    room_id: this.currentRoom.id,
                    data: {
                        ready: this.isReady
                    }
                }));

                this.readyBtn.textContent = this.isReady ? 'Not Ready' : 'Ready';
                console.log('Sent ready state:', this.isReady);
            } else {
                console.error('WebSocket not connected');
                alert('Connection lost. Please try rejoining the room.');
            }
        } catch (error) {
            console.error('Failed to toggle ready state:', error);
            alert('Failed to update ready state.');
        }
    }

    async leaveRoom() {
        if (!this.currentRoom) return;

        try {
            await apiCall(`/api/room/${this.currentRoom.id}/leave`, 'POST');
            this.disconnectFromRoom();
            this.showTab('browser');
        } catch (error) {
            console.error('Failed to leave room:', error);
            this.disconnectFromRoom();
            this.showTab('browser');
        }
    }

    disconnectFromRoom(sendDisconnectMessage = true) {
        if (this.ws && sendDisconnectMessage && this.ws.readyState === WebSocket.OPEN) {
            try {
                const currentUser = getCurrentUser();

                this.ws.send(JSON.stringify({
                    type: 'player_disconnect',
                    room_id: this.currentRoom?.id || '',
                    user_id: currentUser?.id || 0,
                    data: {
                        reason: 'user_left',
                        timestamp: Date.now()
                    }
                }));

                console.log('Sent disconnect message for room:', this.currentRoom?.id, 'user:', currentUser?.id);

                setTimeout(() => {
                    if (this.ws) {
                        this.ws.close(1000, 'User left room');
                        this.ws = null;
                    }
                }, 100);
            } catch (error) {
                console.error('Error sending disconnect message:', error);
                this.ws.close();
                this.ws = null;
            }
        } else if (this.ws) {
            this.ws.close();
            this.ws = null;
        }

        this.currentRoom = null;
        this.isReady = false;
        this.isHost = false;
        this.roomLobbyTab.style.display = 'none';
        this.clearReconnectInterval();
    } handleDisconnection() {
        if (window.isMultiplayer && this.currentRoom) {
            const shouldReturnToMenu = confirm(
                'Connection to the multiplayer game was lost.\n\n' +
                'This could mean your opponent left or there was a network issue.\n\n' +
                'Click OK to return to the main menu, or Cancel to try reconnecting.'
            );

            if (shouldReturnToMenu) {
                if (window.cleanupMultiplayerGame) {
                    window.cleanupMultiplayerGame();
                }
                if (window.showView) {
                    window.showView('mainMenu');
                }
                return;
            }
        }

        if (this.currentRoom) {
            this.reconnectInterval = setTimeout(() => {
                console.log('Attempting to reconnect...');
                this.connectToRoom(this.currentRoom.id);
            }, 3000);
        }
    }

    clearReconnectInterval() {
        if (this.reconnectInterval) {
            clearTimeout(this.reconnectInterval);
            this.reconnectInterval = null;
        }
    }

    handleGameStart(gameData) {
        console.log('Game starting!', gameData);

        this.showGameCountdown(() => {
            this.startMultiplayerGame();
        });
    }

    handlePlayerGameState(message) {
        console.log('=== Player game state received ===');
        console.log('User ID:', message.user_id, 'Room ID:', message.room_id);
        console.log('Game state data:', message.data);

        const currentUser = getCurrentUser();
        if (!currentUser) {
            console.error('No current user found for game state processing');
            return;
        }

        const isCurrentPlayer = message.user_id === currentUser.id;

        console.log('Current user ID:', currentUser.id);
        console.log('Message user ID:', message.user_id);
        console.log('Is current player:', isCurrentPlayer);

        if (isCurrentPlayer) {
            console.log('Rendering current player game state to player1 canvas');
            if (window.renderMultiplayerGame) {
                window.renderMultiplayerGame(message.data, 'player1');
                window.updateMultiplayerGameInfo(message.data, 'player1');
            }
        } else {
            console.log('Rendering opponent game state to player2 canvas');
            if (window.handleMultiplayerUpdate) {
                window.handleMultiplayerUpdate(message.data);
            }
        }
    }

    handleMultiplayerGameStarted(gameData) {
        console.log('Multiplayer game started:', gameData);

        this.showNotification(gameData.message || 'Game started! Use arrow keys to play.', 'success');

        if (window.handleMultiplayerUpdate) {
            const emptyGameState = {
                board: GameUtils.createEmptyBoard(),
                score: 0,
                level: gameData.starting_level || GAME_CONFIG.DEFAULT_STARTING_LEVEL,
                lines: 0,
                gameOver: false,
                paused: false
            };

            window.handleMultiplayerUpdate(emptyGameState);
        }
    }

    handlePlayerInput(message) {
        console.log('Player input received:', message);
    }

    showGameCountdown(callback) {
        let countdown = 3;

        const countdownOverlay = document.createElement('div');
        countdownOverlay.className = 'countdown-overlay';
        countdownOverlay.style.cssText = `
            position: fixed;
            top: 0;
            left: 0;
            width: 100%;
            height: 100%;
            background: rgba(0, 0, 0, 0.8);
            display: flex;
            align-items: center;
            justify-content: center;
            z-index: 10000;
            font-size: 72px;
            color: white;
            font-weight: bold;
        `;
        countdownOverlay.textContent = countdown;
        document.body.appendChild(countdownOverlay);

        const countdownInterval = setInterval(() => {
            countdown--;
            if (countdown > 0) {
                countdownOverlay.textContent = countdown;
            } else {
                countdownOverlay.textContent = 'GO!';
                setTimeout(() => {
                    document.body.removeChild(countdownOverlay);
                    callback();
                }, 500);
                clearInterval(countdownInterval);
            }
        }, 1000);
    }

    startMultiplayerGame() {
        console.log('Starting multiplayer game...');

        if (window.cleanupGame) {
            window.cleanupGame();
        }
        if (window.cleanupMultiplayerGame) {
            window.cleanupMultiplayerGame();
        }

        if (window.showView) {
            console.log('Calling showView(multiplayerGame)');
            window.showView('multiplayerGame');
        }

        requestAnimationFrame(() => {
            requestAnimationFrame(() => {
                console.log('DOM should be ready, initializing multiplayer game...');

                const multiplayerView = document.getElementById('multiplayer-game-view');
                console.log('Multiplayer view state:', {
                    element: !!multiplayerView,
                    hidden: multiplayerView?.classList.contains('hidden'),
                    display: multiplayerView?.style.display,
                    offsetWidth: multiplayerView?.offsetWidth,
                    offsetHeight: multiplayerView?.offsetHeight
                });

                const requiredCanvases = [
                    'player1-canvas', 'player2-canvas',
                    'player1-next-canvas', 'player2-next-canvas',
                    'player1-hold-canvas', 'player2-hold-canvas'
                ];

                const missingCanvases = requiredCanvases.filter(id => !document.getElementById(id));
                if (missingCanvases.length > 0) {
                    console.error('Missing canvas elements:', missingCanvases);
                    alert('Game setup error: Missing canvas elements. Please refresh and try again.');
                    return;
                }

                if (window.startMultiplayerGame) {
                    const startingLevel = this.currentRoom.settings?.starting_level || 1;

                    this.setupMultiplayerGameView();

                    console.log('Calling window.startMultiplayerGame...');
                    try {
                        window.startMultiplayerGame(this.currentRoom.id, this.ws, startingLevel);
                    } catch (error) {
                        console.error('Error starting multiplayer game:', error);
                        alert('Failed to start multiplayer game. Please try again.');

                        if (window.showView) {
                            window.showView('multiplayer');
                        }
                    }
                } else {
                    console.error('Multiplayer game start function not found');
                    alert('Game initialization error. Please refresh the page.');
                }
            });
        });
    }

    setupMultiplayerGameView() {
        const currentUser = getCurrentUser();
        if (currentUser) {
            document.getElementById('player1-name').textContent = currentUser.username;
        }

        if (this.currentRoom && this.currentRoom.players) {
            const opponent = this.currentRoom.players.find(p => p.username !== currentUser?.username);
            if (opponent) {
                document.getElementById('player2-name').textContent = opponent.username;
            }
        }

        document.getElementById('player1-score').textContent = '0';
        document.getElementById('player1-level').textContent = this.currentRoom.settings?.starting_level || '1';
        document.getElementById('player1-lines').textContent = '0';
        document.getElementById('player2-score').textContent = '0';
        document.getElementById('player2-level').textContent = this.currentRoom.settings?.starting_level || '1';
        document.getElementById('player2-lines').textContent = '0';
    }

    handleGameState(gameState) {
        console.log('Game state update received:', gameState);

        if (window.handleMultiplayerUpdate) {
            console.log('Calling handleMultiplayerUpdate from handleGameState');
            window.handleMultiplayerUpdate(gameState);
        } else {
            console.error('handleMultiplayerUpdate function not found');
        }
    }

    handlePlayerUpdate(message) {
        console.log('=== Player update received ===');
        console.log('Message:', JSON.stringify(message, null, 2));

        const currentUser = getCurrentUser();
        if (!currentUser) {
            console.warn('No current user found, cannot process player update');
            return;
        }

        console.log('Current user from getCurrentUser():', currentUser);

        const messageUserId = message.user_id || message.UserID;
        const gameStateData = message.data || message.GameState;

        console.log('Message user ID:', messageUserId, 'type:', typeof messageUserId);
        console.log('Current user ID:', currentUser.id, 'type:', typeof currentUser.id);

        if (messageUserId && gameStateData) {
            console.log('Processing player update from user ID:', messageUserId);
            console.log('Current user ID:', currentUser.id);

            if (String(messageUserId) !== String(currentUser.id)) {
                console.log('Processing opponent update! Calling handleMultiplayerUpdate...');
                console.log('GameState data:', JSON.stringify(gameStateData, null, 2));

                if (window.handleMultiplayerUpdate) {
                    window.handleMultiplayerUpdate(gameStateData);
                } else {
                    console.error('handleMultiplayerUpdate function not found');
                }
            } else {
                console.log('Ignoring update from self (user ID:', currentUser.id, ')');
            }
        } else {
            console.warn('Invalid player update message structure - missing user_id or data:', message);
            console.warn('Expected: user_id and data fields');
            console.warn('Received user_id:', messageUserId);
            console.warn('Received data:', gameStateData);
        }
    }

    handleBackToMenu() {
        this.disconnectFromRoom();
        if (window.showView) {
            window.showView('mainMenu');
        }
    }

    handlePlayerFinished(message) {
        console.log('Player finished:', message.data);

        if (message.data && message.data.playerName) {
            const isMyself = message.data.playerName === getCurrentUser()?.username;

            if (isMyself) {
                this.showNotification(
                    `You finished in position ${message.data.position}!`,
                    'info'
                );
            } else {
                this.showNotification(
                    `${message.data.playerName} finished! Match ending for all players...`,
                    'warning'
                );

                if (window.isMultiplayer && window.ws) {
                    console.log('Ending multiplayer game due to opponent finishing');
                }
            }
        }
    }

    handleGameComplete(data) {
        console.log('Game completed:', data);

        if (data.results && Array.isArray(data.results)) {
            this.showGameCompleteModal(data.results);
        }
    }

    showGameCompleteModal(results) {
        const overlay = document.createElement('div');
        overlay.className = 'game-complete-overlay';
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
        content.style.cssText = `
            background: #1a1a1a;
            padding: 30px;
            border-radius: 10px;
            max-width: 500px;
            width: 90%;
            box-shadow: 0 4px 20px rgba(0, 0, 0, 0.3);
        `;

        let resultsHTML = '<h2 style="margin-bottom: 20px; color: #ffd700;">🏆 Game Complete!</h2>';
        resultsHTML += '<div style="margin-bottom: 20px; text-align: left;">';

        results.forEach((result, index) => {
            const medal = index === 0 ? '🥇' : index === 1 ? '🥈' : index === 2 ? '🥉' : '🏅';
            resultsHTML += `
                <div style="padding: 10px; margin: 5px 0; background: #2a2a2a; border-radius: 5px; display: flex; justify-content: space-between; align-items: center;">
                    <span>${medal} ${result.position}. ${this.escapeHtml(result.username)}</span>
                    <span style="color: #00ff00; font-weight: bold;">${this.formatScore(result.score)}</span>
                </div>
            `;
        });

        resultsHTML += '</div>';
        resultsHTML += `
            <div style="margin-top: 20px;">
                <button onclick="multiplayerManager.closeGameCompleteModal(); multiplayerManager.showTab('browser');" 
                        style="margin: 0 10px; padding: 10px 20px; font-size: 16px; background: #007bff; color: white; border: none; border-radius: 5px; cursor: pointer;">
                    Browse Rooms
                </button>
                <button onclick="multiplayerManager.closeGameCompleteModal(); window.showView('mainMenu');" 
                        style="margin: 0 10px; padding: 10px 20px; font-size: 16px; background: #6c757d; color: white; border: none; border-radius: 5px; cursor: pointer;">
                    Main Menu
                </button>
            </div>
        `;

        content.innerHTML = resultsHTML;
        overlay.appendChild(content);
        document.body.appendChild(overlay);

        this.gameCompleteModal = overlay;
    }

    closeGameCompleteModal() {
        if (this.gameCompleteModal) {
            this.gameCompleteModal.remove();
            this.gameCompleteModal = null;
        }
    }

    showNotification(message, type = 'info') {
        const notification = document.createElement('div');
        notification.style.cssText = `
            position: fixed;
            top: 100px;
            right: 20px;
            padding: 12px 20px;
            border-radius: 4px;
            z-index: 9999;
            font-size: 14px;
            max-width: 300px;
            word-wrap: break-word;
        `;

        switch (type) {
            case 'success':
                notification.style.background = '#28a745';
                notification.style.color = 'white';
                break;
            case 'error':
                notification.style.background = '#dc3545';
                notification.style.color = 'white';
                break;
            case 'warning':
                notification.style.background = '#ffc107';
                notification.style.color = 'black';
                break;
            default:
                notification.style.background = '#007bff';
                notification.style.color = 'white';
        }

        notification.textContent = message;
        document.body.appendChild(notification);

        setTimeout(() => {
            if (notification.parentNode) {
                notification.parentNode.removeChild(notification);
            }
        }, 4000);
    }

    formatScore(score) {
        return score.toLocaleString();
    }

    async spectateRoom(roomId) {
        console.log('Requesting to spectate room:', roomId);

        try {
            if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
                await this.connectToMultiplayer();
            }

            this.ws.send(JSON.stringify({
                type: 'spectate_request',
                room_id: roomId,
                user_id: getCurrentUser()?.id || 0,
                data: {}
            }));

            this.showSpectatorView(roomId);

        } catch (error) {
            console.error('Failed to spectate room:', error);
            this.showNotification('Failed to start spectating. Please try again.', 'error');
        }
    }

    handleSpectateData(data) {
        console.log('Received spectate data:', data);

        if (!this.currentSpectatingRoom) {
            return;
        }

        const spectatorContent = document.querySelector('.spectator-content');
        if (spectatorContent) {
            let contentHTML = `
                <div class="spectator-info">
                    <h4>Room: ${this.escapeHtml(data.roomName || 'Unknown')}</h4>
                    <p>Game Type: ${data.gameType || 'Unknown'}</p>
                </div>
                <div class="spectator-players">
                    <h5>Players:</h5>
            `;

            if (data.playerInfo) {
                Object.values(data.playerInfo).forEach(player => {
                    contentHTML += `
                        <div class="spectator-player">
                            <span class="player-name">${this.escapeHtml(player.username)}</span>
                            <span class="player-score">Score: ${this.formatScore(player.score || 0)}</span>
                            <span class="player-status status-${player.status}">${player.status}</span>
                        </div>
                    `;
                });
            }

            contentHTML += '</div>';
            spectatorContent.innerHTML = contentHTML;
        }
    }

    showSpectatorView(roomId) {
        this.currentSpectatingRoom = roomId;

        const spectatorHTML = `
            <div class="spectator-view">
                <div class="spectator-header">
                    <h3>Spectating Game</h3>
                    <button onclick="multiplayerManager.stopSpectating()" class="btn-stop-spectate">Stop Spectating</button>
                </div>
                <div class="spectator-content">
                    <div class="loading">Loading game data...</div>
                </div>
            </div>
        `;

        this.showNotification('Spectating mode started! (UI in development)', 'info');
    }

    stopSpectating() {
        this.currentSpectatingRoom = null;
        this.showTab('browser');
        this.showNotification('Stopped spectating', 'info');
    }

    escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }

    handleRoomClosed(message) {
        const reason = message.data?.reason || 'Room was closed';
        console.log(`Room ${message.room_id} was closed: ${reason}`);

        if (this.currentRoom && this.currentRoom.id === message.room_id) {
            alert(`${reason}. Returning to room browser.`);
            this.disconnectFromRoom();
            this.showTab('browser');
            this.refreshRooms();
        }
    }

    handleMatchEnded(message) {
        const reason = message.data?.reason || 'unknown';
        const playerName = message.data?.playerName || 'A player';
        const customMessage = message.data?.message || `Match ended: ${playerName} left the game`;

        console.log(`Match ended in room ${message.room_id}: ${reason}`);

        if (this.currentRoom && this.currentRoom.id === message.room_id) {
            alert(customMessage + '\n\nReturning to multiplayer lobby.');

            if (window.isMultiplayer && window.cleanupMultiplayerGame) {
                window.cleanupMultiplayerGame();
            }

            if (window.showView) {
                window.showView('multiplayer');
            }

            if (window.cleanupGame) {
                window.cleanupGame();
            }

            this.disconnectFromRoom();
            this.showTab('browser');
            this.refreshRooms();
        } else {
            this.showNotification(customMessage, 'warning');
        }
    }

    handleRoomsUpdated(message) {
        console.log('Rooms updated:', message.data);

        if (!this.roomBrowserSection.classList.contains('hidden')) {
            this.refreshRooms();
        }

        if (message.data?.reason === 'inactive_cleanup' && message.data?.removed_rooms?.length > 0) {
            const count = message.data.removed_rooms.length;
            console.log(`${count} inactive room(s) were automatically closed`);

            this.showNotification(`${count} inactive room${count > 1 ? 's' : ''} removed`, 'info');
        }
    }

    showNotification(message, type = 'info') {
        const notification = document.createElement('div');
        notification.className = `notification notification-${type}`;
        notification.textContent = message;
        notification.style.cssText = `
            position: fixed;
            top: 20px;
            right: 20px;
            background: ${type === 'info' ? '#007bff' : '#dc3545'};
            color: white;
            padding: 12px 20px;
            border-radius: 4px;
            box-shadow: 0 2px 8px rgba(0,0,0,0.2);
            z-index: 10000;
            font-size: 14px;
            max-width: 300px;
            transition: opacity 0.3s ease;
        `;

        document.body.appendChild(notification);

        setTimeout(() => {
            notification.style.opacity = '0';
            setTimeout(() => {
                if (notification.parentNode) {
                    notification.parentNode.removeChild(notification);
                }
            }, 300);
        }, 3000);
    }

    initialize() {
        this.disconnectFromRoom();
        this.showTab('browser');

        const currentUser = getCurrentUser();
        if (!currentUser) {
            alert('Please log in to access multiplayer features.');
            if (window.showView) {
                window.showView('mainMenu');
            }
            return;
        }

        if (currentUser.id === null) {
            alert('Your login session is outdated. Please log out and log in again to access multiplayer features.');
            if (window.showView) {
                window.showView('mainMenu');
            }
            return;
        }

        this.connectForRoomUpdates();

        this.refreshRooms();
    }

    connectForRoomUpdates() {
        if (this.ws && this.ws.readyState === WebSocket.OPEN) {
            return;
        }

        const user = getCurrentUser();
        if (!user || !user.token) return;

        const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const wsUrl = `${wsProtocol}//${window.location.host}/ws/room/browse?user_id=${user.id}&token=${user.token}`;
        console.log('Connecting for room updates:', wsUrl);

        this.ws = new WebSocket(wsUrl);

        this.ws.onopen = () => {
            console.log('Connected for room updates');
        };

        this.ws.onmessage = (event) => {
            try {
                const message = JSON.parse(event.data);
                if (message.type === 'rooms_updated') {
                    this.handleRoomsUpdated(message);
                }
            } catch (error) {
                console.error('Failed to parse room update message:', error);
            }
        };

        this.ws.onclose = () => {
            console.log('Room updates connection closed');
        };

        this.ws.onerror = (error) => {
            console.error('Room updates WebSocket error:', error);
        };
    }
}

let multiplayerManager = null;

document.addEventListener('DOMContentLoaded', () => {
    multiplayerManager = new MultiplayerManager();
    window.multiplayerManager = multiplayerManager;
});
