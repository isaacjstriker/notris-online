// Multiplayer functionality for room management and WebSocket communication

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
        // Tab elements
        this.roomBrowserTab = document.getElementById('room-browser-tab');
        this.createRoomTab = document.getElementById('create-room-tab');
        this.roomLobbyTab = document.getElementById('room-lobby-tab');

        // Section elements
        this.roomBrowserSection = document.getElementById('room-browser');
        this.createRoomSection = document.getElementById('create-room');
        this.roomLobbySection = document.getElementById('room-lobby');

        // Room browser elements
        this.roomsList = document.getElementById('rooms-list');
        this.refreshRoomsBtn = document.getElementById('refresh-rooms-btn');

        // Create room elements
        this.createRoomForm = document.getElementById('create-room-form');
        this.roomNameInput = document.getElementById('room-name');
        this.privateRoomCheckbox = document.getElementById('private-room');

        // Lobby elements
        this.lobbyRoomName = document.getElementById('lobby-room-name');
        this.lobbyGameType = document.getElementById('lobby-game-type');
        this.lobbyPlayerCount = document.getElementById('lobby-player-count');
        this.lobbyPlayers = document.getElementById('lobby-players');
        this.readyBtn = document.getElementById('ready-btn');
        this.leaveRoomBtn = document.getElementById('leave-room-btn');
        this.lobbyStatus = document.getElementById('lobby-status');

        // Back button
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
            this.roomsList.innerHTML = '<div class="loading">Loading rooms...</div>';

            // Always use 'tetris' since that's the only game type we support
            const response = await apiCall('/rooms/tetris', 'GET');

            // The API returns the rooms array directly, not wrapped in a response object
            this.rooms = response || [];
            this.displayRooms();
        } catch (error) {
            console.error('Failed to fetch rooms:', error);
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

            return `
                <div class="room-item" onclick="multiplayerManager.joinRoom('${room.id}')">
                    <div class="room-header">
                        <span class="room-name">${this.escapeHtml(room.name)}</span>
                        <span class="room-status ${isFull ? 'full' : ''}">${room.status}</span>
                    </div>
                    <div class="room-info">
                        ${room.game_type} • ${currentPlayers}/${room.max_players} players
                    </div>
                </div>
            `;
        }).join('');

        this.roomsList.innerHTML = roomsHTML;
    }

    async handleCreateRoom(e) {
        e.preventDefault();
        console.log('Creating room: form submitted');

        const roomData = {
            name: this.roomNameInput.value.trim(),
            game_type: 'tetris',
            max_players: 2,
            is_private: this.privateRoomCheckbox.checked,
            settings: {}
        };

        console.log('Creating room: room data', roomData);

        if (!roomData.name) {
            alert('Please enter a room name.');
            return;
        }

        try {
            console.log('Creating room: making API call to /rooms');
            const room = await apiCall('/rooms', 'POST', roomData);
            console.log('Creating room: API call successful', room);
            this.currentRoom = room;
            this.isHost = true;
            this.connectToRoom(room.id);
            this.showTab('lobby');
            this.updateLobbyDisplay();
        } catch (error) {
            console.error('Creating room: API call failed', error);
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

        this.ws.onclose = () => {
            console.log('Disconnected from room');
            this.handleDisconnection();
        };

        this.ws.onerror = (error) => {
            console.error('WebSocket error:', error);
        };
    }

    handleWebSocketMessage(message) {
        switch (message.type) {
            case 'room_update':
                this.currentRoom = message.data;
                this.updateLobbyDisplay();
                break;
            case 'game_start':
                this.handleGameStart(message.data);
                break;
            case 'game_state':
                this.handleGameState(message.data);
                break;
            case 'player_joined':
                console.log('Player joined:', message.data.username);
                break;
            case 'player_left':
                console.log('Player left:', message.data.username);
                break;
            case 'room_closed':
                this.handleRoomClosed(message);
                break;
            case 'rooms_updated':
                this.handleRoomsUpdated(message);
                break;
            case 'error':
                console.error('Server error:', message.data.error);
                alert(message.data.error);
                break;
        }
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
        const currentUser = this.getCurrentUser();
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
            await apiCall(`/room/${this.currentRoom.id}/ready`, 'POST');
            // The room update will come via WebSocket
        } catch (error) {
            console.error('Failed to toggle ready state:', error);
            alert('Failed to update ready state.');
        }
    }

    async leaveRoom() {
        if (!this.currentRoom) return;

        try {
            await apiCall(`/room/${this.currentRoom.id}/leave`, 'POST');
            this.disconnectFromRoom();
            this.showTab('browser');
        } catch (error) {
            console.error('Failed to leave room:', error);
            // Still disconnect locally
            this.disconnectFromRoom();
            this.showTab('browser');
        }
    }

    disconnectFromRoom() {
        if (this.ws) {
            this.ws.close();
            this.ws = null;
        }
        this.currentRoom = null;
        this.isReady = false;
        this.isHost = false;
        this.roomLobbyTab.style.display = 'none';
        this.clearReconnectInterval();
    }

    handleDisconnection() {
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
        // This will be implemented when integrating with the game engine
        // For now, just show an alert
        alert('Game is starting!');
    }

    handleGameState(gameState) {
        console.log('Game state update:', gameState);
        // This will be implemented when integrating with the game engine
    }

    handleBackToMenu() {
        this.disconnectFromRoom();
        // This will be called by main.js to show the main menu
        if (window.showView) {
            window.showView('mainMenu');
        }
    }

    getCurrentUser() {
        const token = localStorage.getItem('devware_jwt');
        const username = localStorage.getItem('devware_username');
        return token && username ? { token, username } : null;
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
        if (!this.getCurrentUser()) {
            alert('Please log in to access multiplayer features.');
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

        const user = this.getCurrentUser();
        if (!user) return;

        const wsUrl = `ws://localhost:8080/ws/room/browse?user_id=${user.id}`;
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
}

// Global multiplayer manager instance
let multiplayerManager = null;

// Initialize when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    multiplayerManager = new MultiplayerManager();
    // Make it globally accessible
    window.multiplayerManager = multiplayerManager;
});
