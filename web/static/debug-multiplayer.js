// Multiplayer Debug Helper
// Open browser console and run these commands to debug multiplayer issues

// Check WebSocket connection status
function checkWebSocketStatus() {
    if (window.multiplayerManager && window.multiplayerManager.ws) {
        const ws = window.multiplayerManager.ws;
        console.log('WebSocket State:', {
            readyState: ws.readyState,
            url: ws.url,
            states: {
                0: 'CONNECTING',
                1: 'OPEN',
                2: 'CLOSING',
                3: 'CLOSED'
            }[ws.readyState]
        });
        return ws.readyState === 1; // OPEN
    } else {
        console.log('No WebSocket connection found');
        return false;
    }
}

// Send test ready message
function testReady(isReady = true) {
    if (window.multiplayerManager && window.multiplayerManager.ws) {
        const message = {
            type: 'player_ready',
            room_id: window.multiplayerManager.currentRoom?.id,
            data: { ready: isReady }
        };
        console.log('Sending test ready message:', message);
        window.multiplayerManager.ws.send(JSON.stringify(message));
    } else {
        console.log('No WebSocket connection available');
    }
}

// Check current room state
function checkRoomState() {
    if (window.multiplayerManager && window.multiplayerManager.currentRoom) {
        const room = window.multiplayerManager.currentRoom;
        console.log('Current Room State:', {
            id: room.id,
            name: room.name,
            status: room.status,
            players: room.players?.map(p => ({
                username: p.username,
                is_ready: p.is_ready,
                status: p.status
            }))
        });
        return room;
    } else {
        console.log('No current room');
        return null;
    }
}

// Check authentication
function checkAuth() {
    const token = localStorage.getItem('devware_jwt');
    const username = localStorage.getItem('devware_username');
    console.log('Authentication Status:', {
        hasToken: !!token,
        username: username,
        tokenLength: token ? token.length : 0
    });
    return { token, username };
}

// Monitor WebSocket messages
function monitorWebSocket(enable = true) {
    if (!window.multiplayerManager || !window.multiplayerManager.ws) {
        console.log('No WebSocket to monitor');
        return;
    }

    const ws = window.multiplayerManager.ws;

    if (enable) {
        const originalSend = ws.send;
        ws.send = function (data) {
            console.log('WS SEND:', JSON.parse(data));
            return originalSend.call(this, data);
        };

        ws.addEventListener('message', function (event) {
            console.log('WS RECEIVE:', JSON.parse(event.data));
        });

        console.log('WebSocket monitoring enabled');
    }
}

// Quick test sequence
function runQuickTest() {
    console.log('=== Multiplayer Quick Test ===');
    checkAuth();
    checkWebSocketStatus();
    checkRoomState();

    if (window.multiplayerManager) {
        console.log('Multiplayer Manager:', {
            isReady: window.multiplayerManager.isReady,
            hasCurrentRoom: !!window.multiplayerManager.currentRoom,
            wsConnected: window.multiplayerManager.ws?.readyState === 1
        });
    }
}

// Export to window for easy access
window.debugMultiplayer = {
    checkWebSocketStatus,
    testReady,
    checkRoomState,
    checkAuth,
    monitorWebSocket,
    runQuickTest
};

console.log('Multiplayer Debug Helper loaded. Use:');
console.log('- debugMultiplayer.runQuickTest() - Run full diagnostic');
console.log('- debugMultiplayer.checkWebSocketStatus() - Check WebSocket');
console.log('- debugMultiplayer.testReady(true/false) - Send ready message');
console.log('- debugMultiplayer.monitorWebSocket(true) - Monitor all WS messages');
console.log('- debugMultiplayer.checkRoomState() - Check current room');
console.log('- debugMultiplayer.checkAuth() - Check authentication');
