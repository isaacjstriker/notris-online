# Multiplayer Testing Guide

## Overview
This guide helps you test the comprehensive multiplayer Tetris functionality that has been implemented, including:

1. **Basic Multiplayer**: Room creation, joining, ready system
2. **Automatic Game Start**: When all players are ready
3. **Game Integration**: Multiplayer Tetris gameplay
4. **Game Results & Scoring**: Final standings and results
5. **Reconnection Handling**: Disconnect/reconnect during games
6. **Spectator Mode**: Watch ongoing games
7. **Room Cleanup**: Automatic cleanup of inactive rooms

## Prerequisites

1. **Build and Start the Server**:
   ```bash
   cd /home/ijstr/github.com/isaacjstriker/devware
   go build -o devware main.go
   ./devware
   ```

2. **Create Test Accounts**:
   - You'll need at least 2 user accounts for multiplayer testing
   - Register accounts through the web interface or API

## Testing Scenarios

### 1. Basic Multiplayer Flow

**Test Steps**:
1. **First Browser Tab**:
   - Open `http://localhost:8080`
   - Login with User 1
   - Navigate to Multiplayer section
   - Create a new room (name: "Test Room")

2. **Second Browser Tab**:
   - Open `http://localhost:8080` in a new tab/window
   - Login with User 2
   - Navigate to Multiplayer → Browse Rooms
   - You should see "Test Room" in the list
   - Click to join the room

3. **Both Tabs**:
   - Both players should now be in the lobby
   - Player names should be visible
   - Status should show "Not Ready" for both players

**Expected Results**:
- ✅ Room appears in browser for second player
- ✅ Both players can see each other in the lobby
- ✅ Room status updates in real-time

### 2. Ready System & Auto-Start

**Test Steps**:
1. **In both browser tabs**:
   - Click the "Ready" button
   - Status should change to "Ready" next to your name
   - When both players are ready, game should auto-start

**Expected Results**:
- ✅ Ready status updates immediately in both tabs
- ✅ When both ready: "All players ready! Starting game..." message
- ✅ Game starts automatically with countdown (3-2-1-GO)
- ✅ Tetris game interface appears

**Debug Info**:
- Check browser console for WebSocket messages
- Look for `player_ready` messages being sent
- Check for `room_update` messages being received

### 3. Multiplayer Game Play

**Test Steps**:
1. **During active game**:
   - Play Tetris normally in both tabs
   - Use arrow keys to move/rotate pieces
   - Try to score points and clear lines

2. **Finish the game**:
   - Let one player finish first (game over)
   - The other player should be notified
   - Eventually both games should end

**Expected Results**:
- ✅ Game controls work normally
- ✅ Score updates in real-time
- ✅ When first player finishes: notification appears
- ✅ Final results screen shows standings and scores

### 4. Spectator Mode

**Test Steps**:
1. **Create a game in progress**:
   - Follow steps 1-2 from basic flow
   - Start the game (both players ready)

2. **Third Browser Tab**:
   - Open new tab, login with User 3
   - Navigate to Multiplayer → Browse Rooms
   - Find the active room (status: "active")
   - Click "Spectate" button

**Expected Results**:
- ✅ "Spectate" button appears for active games
- ✅ Spectator can view game information
- ✅ Spectator receives real-time updates

### 5. Reconnection Handling

**Test Steps**:
1. **Start a multiplayer game**:
   - Get two players in a game
   - Start playing

2. **Simulate disconnection**:
   - Close one browser tab (simulating disconnect)
   - Other player should be notified of disconnection

3. **Reconnect**:
   - Reopen the closed tab
   - Login again and rejoin the same room
   - Should be able to resume

**Expected Results**:
- ✅ Disconnection notification appears
- ✅ 30-second reconnection timer starts
- ✅ Successful reconnection restores game state
- ✅ If timeout: player marked as finished with 0 score

### 6. Room Cleanup Testing

**Test Steps**:
1. **Create multiple test rooms**:
   - Create 3-4 rooms but don't start games
   - Leave them idle for 5+ minutes

2. **Check cleanup**:
   - Rooms should automatically disappear from browse list
   - Check server logs for cleanup messages

**Expected Results**:
- ✅ Rooms older than 5 minutes are cleaned up
- ✅ Players in cleaned rooms get notifications
- ✅ Cleanup happens every minute (check logs)

## Debugging Common Issues

### Ready Button Not Working
**Symptoms**: Clicking ready doesn't change status
**Debug Steps**:
1. Check browser console for WebSocket errors
2. Verify WebSocket connection is open: `this.ws.readyState === 1`
3. Look for `player_ready` messages in Network tab
4. Check server logs for JWT validation errors

**Common Fixes**:
- Refresh the page to re-establish WebSocket connection
- Check that JWT token is valid in localStorage
- Verify user is properly authenticated

### Game Not Auto-Starting
**Symptoms**: Both players ready but game doesn't start
**Debug Steps**:
1. Check server logs for "Auto-starting game" messages
2. Verify room has exactly 2 players
3. Check that both players have `is_ready: true`
4. Look for `game_start` WebSocket messages

**Common Fixes**:
- Ensure both players are actually ready (green status)
- Check that room isn't already in 'active' state
- Verify WebSocket hub is processing messages

### WebSocket Connection Issues
**Symptoms**: Real-time updates not working
**Debug Steps**:
1. Check Network tab for WebSocket connection
2. Look for 404 or 401 errors on `/ws/room/{id}`
3. Verify JWT token is included in WebSocket URL
4. Check for CORS issues

**Common Fixes**:
- Refresh browser to re-establish connection
- Clear localStorage and re-login
- Check server logs for WebSocket upgrade errors

### Game Integration Issues
**Symptoms**: Game doesn't start after countdown
**Debug Steps**:
1. Check that `startMultiplayerGame` function exists
2. Verify WebSocket connection to `/ws/game`
3. Look for game state messages
4. Check for JavaScript errors in console

## Expected Server Log Messages

When testing, you should see these log messages:

```
Client {id} connected to room {roomId}
Room {roomId} ready check: {ready}/{total} players ready
Auto-starting game in room {roomId}: {ready}/{total} players ready
Cleaned up {count} inactive rooms
Player {username} reconnected to room {roomId}
Game completed for room {roomId} with {count} players
```

## Performance Testing

### Load Testing
- Test with multiple concurrent rooms
- Test with many spectators on single game
- Monitor WebSocket connection limits

### Stress Testing
- Rapid room creation/deletion
- Fast ready/unready toggling
- Multiple disconnections/reconnections

## API Endpoints for Manual Testing

You can also test with direct API calls:

```bash
# Create room
curl -X POST http://localhost:8080/api/rooms \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name": "Test Room", "game_type": "tetris", "max_players": 2}'

# Join room
curl -X POST http://localhost:8080/api/room/{roomId}/join \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"

# Get rooms list
curl http://localhost:8080/api/rooms
```

## Success Criteria

The multiplayer system is working correctly when:

- ✅ 2+ players can join rooms and see each other
- ✅ Ready system works and updates in real-time
- ✅ Games auto-start when all players ready
- ✅ Multiplayer Tetris gameplay functions properly
- ✅ Game results are calculated and displayed correctly
- ✅ Reconnection works within 30-second window
- ✅ Spectator mode allows viewing active games
- ✅ Inactive rooms are cleaned up automatically
- ✅ All real-time updates work via WebSocket
- ✅ No memory leaks or connection issues

## Troubleshooting Quick Reference

| Issue | Check | Solution |
|-------|-------|-----------|
| Can't see rooms | Authentication | Re-login, check JWT |
| Ready not working | WebSocket connection | Refresh page |
| Game won't start | Both players ready? | Verify ready status |
| No real-time updates | WebSocket errors | Check console logs |
| Connection lost | Server running? | Restart server |
| Game won't load | JavaScript errors | Check browser console |

---

This comprehensive testing should validate all multiplayer features. If any issues persist, check the browser console and server logs for specific error messages.
