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
        // Tab switching
        this.roomBrowserTab.addEventListener('click', () => this.showTab('browser'));
        this.createRoomTab.addEventListener('click', () => this.showTab('create'));
        this.roomLobbyTab.addEventListener('click', () => this.showTab('lobby'));

        // Room browser
        this.refreshRoomsBtn.addEventListener('click', () => this.refreshRooms());

        // Create room
        this.createRoomForm.addEventListener('submit', (e) => this.handleCreateRoom(e));

        // Lobby actions
        this.readyBtn.addEventListener('click', () => this.toggleReady());
        this.leaveRoomBtn.addEventListener('click', () => this.leaveRoom());

        // Back button
        this.backBtn.addEventListener('click', () => this.handleBackToMenu());
    }

    showTab(tabName) {
        // Hide all sections
        this.roomBrowserSection.classList.add('hidden');
        this.createRoomSection.classList.add('hidden');
        this.roomLobbySection.classList.add('hidden');

        // Remove active class from all tabs
        this.roomBrowserTab.classList.remove('active');
        this.createRoomTab.classList.remove('active');
        this.roomLobbyTab.classList.remove('active');

        // Show selected section and activate tab
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

            // Always use 'tetris' since that's the only game type we support
            console.log('Making API call to /rooms/tetris');
            const response = await apiCall('/rooms/tetris', 'GET');
            console.log('API response received:', response);

            // The API returns the rooms array directly, or null if no rooms exist
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

        // Close existing connection
        if (this.ws) {
            this.ws.close();
        }

        // WebSocket URL (using same host as current page)
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

            // Clear any reconnection attempts
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
            // Also handle error as disconnection after a brief delay
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
                // If we're in a game and opponent leaves, handle it appropriately
                if (this.currentRoom && window.isMultiplayer) {
                    const playerName = message.data.playerName || message.data.username || 'A player';
                    this.showNotification(`${playerName} left the match`, 'warning');
                    console.log('Displayed disconnect notification for:', playerName);
                    // Let the server handle sending match_ended message if needed
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
        // Start the multiplayer game
        if (this.currentRoom) {
            setTimeout(() => {
                window.startMultiplayerGame(this.currentRoom.id);
            }, 1500); // Small delay to show the notification
        }
    }

    handleMultiplayerGameStarted(message) {
        console.log('Multiplayer game started:', message.data);
        if (message.data && message.data.message) {
            this.showNotification(message.data.message, 'info');
        }
        // Initialize the actual game loop
        if (window.initializeMultiplayerGameplay) {
            window.initializeMultiplayerGameplay(message.data.starting_level || 1);
        }
    }

    handlePlayerGameState(message) {
        // Get current user ID from auth
        const currentUser = window.currentUser || JSON.parse(localStorage.getItem('currentUser'));
        if (!currentUser) return;

        const isOwnState = message.UserID === currentUser.id;
        const gameState = message.data;

        if (isOwnState) {
            // This is our own game state - render on player1 side
            if (window.renderMultiplayerGame) {
                window.renderMultiplayerGame(gameState, 'player1');
            }
            if (window.updateMultiplayerGameInfo) {
                window.updateMultiplayerGameInfo(gameState, 'player1');
            }
        } else {
            // This is opponent's game state - render on player2 side
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

        // Show notification that the game ended
        if (message.data && message.data.message) {
            this.showNotification(message.data.message, 'info');
        } else {
            this.showNotification('Game ended!', 'info');
        }

        // Clean up game state
        if (window.cleanupMultiplayerGame) {
            window.cleanupMultiplayerGame();
        }

        // Return to multiplayer lobby after a short delay
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

        // Update players list
        const playersHTML = this.currentRoom.players.map(player => `
            <div class="player-item">
                <span class="player-name">${this.escapeHtml(player.username)}</span>
                <span class="player-status ${player.is_ready ? '' : 'not-ready'}">
                    ${player.is_ready ? 'Ready' : 'Not Ready'}
                </span>
            </div>
        `).join('');

        this.lobbyPlayers.innerHTML = playersHTML;

        // Update ready button state
        const currentUser = getCurrentUser();
        if (currentUser) {
            const userPlayer = this.currentRoom.players.find(p => p.username === currentUser.username);
            if (userPlayer) {
                this.isReady = userPlayer.is_ready;
                this.readyBtn.textContent = this.isReady ? 'Not Ready' : 'Ready';
            }
        }

        // Update status
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

        // Show lobby tab
        this.roomLobbyTab.style.display = 'block';
    }

    async toggleReady() {
        if (!this.currentRoom) return;

        try {
            // Send ready state via WebSocket instead of REST API
            if (this.ws && this.ws.readyState === WebSocket.OPEN) {
                this.isReady = !this.isReady;

                this.ws.send(JSON.stringify({
                    type: 'player_ready',
                    room_id: this.currentRoom.id,
                    data: {
                        ready: this.isReady
                    }
                }));

                // Update button text immediately for better UX
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
            // Still disconnect locally
            this.disconnectFromRoom();
            this.showTab('browser');
        }
    }

    disconnectFromRoom(sendDisconnectMessage = true) {
        if (this.ws && sendDisconnectMessage && this.ws.readyState === WebSocket.OPEN) {
            try {
                // Get current user for user ID
                const currentUser = getCurrentUser();

                // Send disconnect message before closing
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

                // Give a brief moment for the message to send, then close
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
            // Close without sending message (probably already disconnected)
            this.ws.close();
            this.ws = null;
        }

        this.currentRoom = null;
        this.isReady = false;
        this.isHost = false;
        this.roomLobbyTab.style.display = 'none';
        this.clearReconnectInterval();
    } handleDisconnection() {
        // If we're in a multiplayer game when disconnected, show immediate feedback
        if (window.isMultiplayer && this.currentRoom) {
            // Show a notification that connection was lost
            const shouldReturnToMenu = confirm(
                'Connection to the multiplayer game was lost.\n\n' +
                'This could mean your opponent left or there was a network issue.\n\n' +
                'Click OK to return to the main menu, or Cancel to try reconnecting.'
            );

            if (shouldReturnToMenu) {
                // Clean up and return to main menu
                if (window.cleanupMultiplayerGame) {
                    window.cleanupMultiplayerGame();
                }
                if (window.showView) {
                    window.showView('mainMenu');
                }
                return;
            }
        }

        // Try to reconnect after a short delay
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

        // Show countdown
        this.showGameCountdown(() => {
            // Start the actual multiplayer game
            this.startMultiplayerGame();
        });
    }

    handlePlayerGameState(message) {
        console.log('=== Player game state received ===');
        console.log('User ID:', message.user_id, 'Room ID:', message.room_id);
        console.log('Game state data:', message.data);

        // Determine if this is the current player or opponent
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
            // This is the current player's own game state - render to player1 canvas
            console.log('Rendering current player game state to player1 canvas');
            if (window.renderMultiplayerGame) {
                window.renderMultiplayerGame(message.data, 'player1');
                window.updateMultiplayerGameInfo(message.data, 'player1');
            }
        } else {
            // This is an opponent's game state - render to player2 canvas
            console.log('Rendering opponent game state to player2 canvas');
            if (window.handleMultiplayerUpdate) {
                window.handleMultiplayerUpdate(message.data);
            }
        }
    }

    handleMultiplayerGameStarted(gameData) {
        console.log('Multiplayer game started:', gameData);
        // The game is now active - players can start playing
        // This message comes after the game WebSocket would be set up in the old approach

        this.showNotification(gameData.message || 'Game started! Use arrow keys to play.', 'success');

        // Initialize empty game states for both players
        if (window.handleMultiplayerUpdate) {
            // Initialize empty game boards for visual setup
            const emptyGameState = {
                board: GameUtils.createEmptyBoard(),
                score: 0,
                level: gameData.starting_level || GAME_CONFIG.DEFAULT_STARTING_LEVEL,
                lines: 0,
                gameOver: false,
                paused: false
            };

            // Set up both player displays
            window.handleMultiplayerUpdate(emptyGameState);
        }
    }

    handlePlayerInput(message) {
        console.log('Player input received:', message);
        // This would handle opponent input if we wanted to show it
        // For now, we'll just log it
    }

    showGameCountdown(callback) {
        let countdown = 3;

        // Create countdown overlay
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

        // First, ensure we cleanup any existing game state
        if (window.cleanupGame) {
            window.cleanupGame();
        }
        if (window.cleanupMultiplayerGame) {
            window.cleanupMultiplayerGame();
        }

        // Switch to multiplayer game view
        if (window.showView) {
            console.log('Calling showView(multiplayerGame)');
            window.showView('multiplayerGame');
        }

        // Use requestAnimationFrame to ensure DOM is fully rendered
        requestAnimationFrame(() => {
            requestAnimationFrame(() => {
                console.log('DOM should be ready, initializing multiplayer game...');

                // Debug: Check if view is properly visible
                const multiplayerView = document.getElementById('multiplayer-game-view');
                console.log('Multiplayer view state:', {
                    element: !!multiplayerView,
                    hidden: multiplayerView?.classList.contains('hidden'),
                    display: multiplayerView?.style.display,
                    offsetWidth: multiplayerView?.offsetWidth,
                    offsetHeight: multiplayerView?.offsetHeight
                });

                // Verify all required canvas elements exist
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

                // Start the game with multiplayer flag
                if (window.startMultiplayerGame) {
                    // Pass starting level from room settings if available
                    const startingLevel = this.currentRoom.settings?.starting_level || 1;

                    // Set up player names
                    this.setupMultiplayerGameView();

                    console.log('Calling window.startMultiplayerGame...');
                    try {
                        window.startMultiplayerGame(this.currentRoom.id, this.ws, startingLevel);
                    } catch (error) {
                        console.error('Error starting multiplayer game:', error);
                        alert('Failed to start multiplayer game. Please try again.');

                        // Return to multiplayer lobby on error
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
        // Set up player names and initial stats
        const currentUser = getCurrentUser();
        if (currentUser) {
            document.getElementById('player1-name').textContent = currentUser.username;
        }

        // Find opponent name
        if (this.currentRoom && this.currentRoom.players) {
            const opponent = this.currentRoom.players.find(p => p.username !== currentUser?.username);
            if (opponent) {
                document.getElementById('player2-name').textContent = opponent.username;
            }
        }

        // Reset stats
        document.getElementById('player1-score').textContent = '0';
        document.getElementById('player1-level').textContent = this.currentRoom.settings?.starting_level || '1';
        document.getElementById('player1-lines').textContent = '0';
        document.getElementById('player2-score').textContent = '0';
        document.getElementById('player2-level').textContent = this.currentRoom.settings?.starting_level || '1';
        document.getElementById('player2-lines').textContent = '0';
    }

    handleGameState(gameState) {
        console.log('Game state update received:', gameState);

        // This handles direct game state updates from other players
        // Call handleMultiplayerUpdate to render the opponent's gameplay
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

        // Only process updates from other players, not ourselves
        const currentUser = getCurrentUser();
        if (!currentUser) {
            console.warn('No current user found, cannot process player update');
            return;
        }

        console.log('Current user from getCurrentUser():', currentUser);

        // Extract user_id and data from the message structure
        const messageUserId = message.user_id || message.UserID;
        const gameStateData = message.data || message.GameState;

        console.log('Message user ID:', messageUserId, 'type:', typeof messageUserId);
        console.log('Current user ID:', currentUser.id, 'type:', typeof currentUser.id);

        // Check if this message has the required fields
        if (messageUserId && gameStateData) {
            console.log('Processing player update from user ID:', messageUserId);
            console.log('Current user ID:', currentUser.id);

            // Only process updates from other players (not ourselves)
            // Convert both to strings for comparison to avoid type issues
            if (String(messageUserId) !== String(currentUser.id)) {
                console.log('Processing opponent update! Calling handleMultiplayerUpdate...');
                console.log('GameState data:', JSON.stringify(gameStateData, null, 2));

                // Pass the game state data to the game renderer
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
        // This will be called by main.js to show the main menu
        if (window.showView) {
            window.showView('mainMenu');
        }
    }

    // Handle player finished notification
    handlePlayerFinished(message) {
        console.log('Player finished:', message.data);

        if (message.data && message.data.playerName) {
            const isMyself = message.data.playerName === getCurrentUser()?.username;

            if (isMyself) {
                // I finished - show my result
                this.showNotification(
                    `You finished in position ${message.data.position}!`,
                    'info'
                );
            } else {
                // Opponent finished - the match will end for everyone
                this.showNotification(
                    `${message.data.playerName} finished! Match ending for all players...`,
                    'warning'
                );

                // If we're currently in game, end our game too
                if (window.isMultiplayer && window.ws) {
                    console.log('Ending multiplayer game due to opponent finishing');
                    // The server will handle finishing us automatically
                    // Just wait for the game_complete message
                }
            }
        }
    }

    // Handle game completion with final results
    handleGameComplete(data) {
        console.log('Game completed:', data);

        if (data.results && Array.isArray(data.results)) {
            this.showGameCompleteModal(data.results);
        }
    }

    // Show game completion modal with results
    showGameCompleteModal(results) {
        // Create modal overlay
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

        // Create content
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

        // Store reference for cleanup
        this.gameCompleteModal = overlay;
    }

    // Close game complete modal
    closeGameCompleteModal() {
        if (this.gameCompleteModal) {
            this.gameCompleteModal.remove();
            this.gameCompleteModal = null;
        }
    }

    // Show notification
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

        // Set colors based on type
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
            default: // info
                notification.style.background = '#007bff';
                notification.style.color = 'white';
        }

        notification.textContent = message;
        document.body.appendChild(notification);

        // Auto-remove after 4 seconds
        setTimeout(() => {
            if (notification.parentNode) {
                notification.parentNode.removeChild(notification);
            }
        }, 4000);
    }

    // Format score for display
    formatScore(score) {
        return score.toLocaleString();
    }

    // Spectate a room
    async spectateRoom(roomId) {
        console.log('Requesting to spectate room:', roomId);

        try {
            // Connect to multiplayer if not already connected
            if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
                await this.connectToMultiplayer();
            }

            // Send spectate request
            this.ws.send(JSON.stringify({
                type: 'spectate_request',
                room_id: roomId,
                user_id: getCurrentUser()?.id || 0,
                data: {}
            }));

            // Show spectator view
            this.showSpectatorView(roomId);

        } catch (error) {
            console.error('Failed to spectate room:', error);
            this.showNotification('Failed to start spectating. Please try again.', 'error');
        }
    }

    // Handle spectate data from server
    handleSpectateData(data) {
        console.log('Received spectate data:', data);

        if (!this.currentSpectatingRoom) {
            return;
        }

        // Update spectator view with game data
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

    // Show spectator view
    showSpectatorView(roomId) {
        // Switch to a spectator tab/view
        this.currentSpectatingRoom = roomId;

        // Create spectator interface
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

        // Show spectator view (you'd need to create this UI element)
        this.showNotification('Spectating mode started! (UI in development)', 'info');
    }

    // Stop spectating
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

    // Handle room closed notification
    handleRoomClosed(message) {
        const reason = message.data?.reason || 'Room was closed';
        console.log(`Room ${message.room_id} was closed: ${reason}`);

        // If we're currently in this room, redirect to browser
        if (this.currentRoom && this.currentRoom.id === message.room_id) {
            alert(`${reason}. Returning to room browser.`);
            this.disconnectFromRoom();
            this.showTab('browser');
            this.refreshRooms(); // Refresh the room list
        }
    }

    // Handle match ended notification (when a player leaves during active game)
    handleMatchEnded(message) {
        const reason = message.data?.reason || 'unknown';
        const playerName = message.data?.playerName || 'A player';
        const customMessage = message.data?.message || `Match ended: ${playerName} left the game`;

        console.log(`Match ended in room ${message.room_id}: ${reason}`);

        // If we're currently in this room and in game, we need to handle this
        if (this.currentRoom && this.currentRoom.id === message.room_id) {
            // Show prominent notification
            alert(customMessage + '\n\nReturning to multiplayer lobby.');

            // Clean up multiplayer game state if we're in a game
            if (window.isMultiplayer && window.cleanupMultiplayerGame) {
                window.cleanupMultiplayerGame();
            }

            // If we're in game view, return to multiplayer view
            if (window.showView) {
                window.showView('multiplayer');
            }

            // Clean up general game state
            if (window.cleanupGame) {
                window.cleanupGame();
            }

            // Disconnect from room and return to browser
            this.disconnectFromRoom();
            this.showTab('browser');
            this.refreshRooms();
        } else {
            // Just show a notification if we're not directly affected
            this.showNotification(customMessage, 'warning');
        }
    }

    // Handle rooms updated notification (for cleanup)
    handleRoomsUpdated(message) {
        console.log('Rooms updated:', message.data);

        // If we're currently viewing the room browser, refresh the list
        if (!this.roomBrowserSection.classList.contains('hidden')) {
            this.refreshRooms();
        }

        // Show a notification if rooms were removed due to inactivity
        if (message.data?.reason === 'inactive_cleanup' && message.data?.removed_rooms?.length > 0) {
            const count = message.data.removed_rooms.length;
            console.log(`${count} inactive room(s) were automatically closed`);

            // Show a subtle notification without interrupting gameplay
            this.showNotification(`${count} inactive room${count > 1 ? 's' : ''} removed`, 'info');
        }
    }

    // Show a temporary notification
    showNotification(message, type = 'info') {
        // Create notification element
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

        // Auto-remove after 3 seconds
        setTimeout(() => {
            notification.style.opacity = '0';
            setTimeout(() => {
                if (notification.parentNode) {
                    notification.parentNode.removeChild(notification);
                }
            }, 300);
        }, 3000);
    }

    // Initialize multiplayer when view is shown
    initialize() {
        // Reset state
        this.disconnectFromRoom();
        this.showTab('browser');

        // Check if user is logged in
        const currentUser = getCurrentUser();
        if (!currentUser) {
            alert('Please log in to access multiplayer features.');
            if (window.showView) {
                window.showView('mainMenu');
            }
            return;
        }

        // Check if user has valid user_id (for users who logged in before the user_id fix)
        if (currentUser.id === null) {
            alert('Your login session is outdated. Please log out and log in again to access multiplayer features.');
            if (window.showView) {
                window.showView('mainMenu');
            }
            return;
        }

        // Connect to receive general room updates (not specific to a room)
        this.connectForRoomUpdates();

        // Initial room refresh
        this.refreshRooms();
    }

    // Connect to receive general room updates when browsing
    connectForRoomUpdates() {
        if (this.ws && this.ws.readyState === WebSocket.OPEN) {
            return; // Already connected
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
                // Only handle global room updates, not room-specific messages
                if (message.type === 'rooms_updated') {
                    this.handleRoomsUpdated(message);
                }
            } catch (error) {
                console.error('Failed to parse room update message:', error);
            }
        };

        this.ws.onclose = () => {
            console.log('Room updates connection closed');
            // Don't auto-reconnect for general room updates
        };

        this.ws.onerror = (error) => {
            console.error('Room updates WebSocket error:', error);
        };
    }
}

// Global multiplayer manager instance
let multiplayerManager = null;

// Initialize when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    multiplayerManager = new MultiplayerManager();
    // Make it globally accessible
    window.multiplayerManager = multiplayerManager;
});
