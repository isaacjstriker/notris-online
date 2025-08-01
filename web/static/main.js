document.addEventListener('DOMContentLoaded', () => {
    const views = {
        mainMenu: document.getElementById('main-menu-view'),
        game: document.getElementById('game-view'),
        login: document.getElementById('login-view'),
        register: document.getElementById('register-view'),
        leaderboard: document.getElementById('leaderboard-view'),
        levelSelect: document.getElementById('level-select-view'),
        multiplayer: document.getElementById('multiplayer-view'),
        multiplayerGame: document.getElementById('multiplayer-game-view'),
    };

    const authNav = {
        loginBtn: document.getElementById('login-nav-btn'),
        registerBtn: document.getElementById('register-nav-btn'),
        userInfo: document.getElementById('user-info'),
        usernameDisplay: document.getElementById('username-display'),
        logoutBtn: document.getElementById('logout-btn'),
    };

    function showView(viewName) {
        Object.values(views).forEach(view => {
            if (view) {
                view.classList.add('hidden');
                view.style.display = 'none';
            }
        });

        const mainHeader = document.getElementById('main-header');
        if (viewName === 'game' || viewName === 'multiplayerGame') {
            mainHeader.style.display = 'none';
        } else {
            mainHeader.style.display = 'flex';
        }

        if (viewName === 'multiplayerGame') {
            document.body.classList.add('multiplayer-active');
        } else {
            document.body.classList.remove('multiplayer-active');
        }

        let targetView = null;
        if (viewName === 'mainMenu' || viewName === 'main-menu') {
            targetView = views.mainMenu;
        } else if (viewName === 'game') {
            targetView = views.game;
        } else if (viewName === 'multiplayerGame') {
            targetView = views.multiplayerGame;
        } else if (viewName === 'login') {
            targetView = views.login;
        } else if (viewName === 'register') {
            targetView = views.register;
        } else if (viewName === 'leaderboard') {
            targetView = views.leaderboard;
        } else if (viewName === 'levelSelect') {
            targetView = views.levelSelect;
        } else if (viewName === 'multiplayer') {
            // Check authentication before showing multiplayer view
            const currentUser = getCurrentUser();
            if (!currentUser || !currentUser.token) {
                alert('Please log in to access multiplayer features.');
                showView('login');
                return;
            }

            targetView = views.multiplayer;
            if (window.multiplayerManager) {
                window.multiplayerManager.initialize();
            }
        }

        if (targetView) {
            targetView.classList.remove('hidden');
            targetView.style.display = '';
        }
    }

    window.showView = showView;

    function updateAuthUI() {
        const token = getAuthToken();
        const username = getUsername();
        if (token && username) {
            authNav.loginBtn.classList.add('hidden');
            authNav.registerBtn.classList.add('hidden');
            authNav.userInfo.classList.remove('hidden');
            authNav.usernameDisplay.textContent = username;
        } else {
            authNav.loginBtn.classList.remove('hidden');
            authNav.registerBtn.classList.remove('hidden');
            authNav.userInfo.classList.add('hidden');
        }
    }

    window.updateAuthUI = updateAuthUI;

    authNav.loginBtn.addEventListener('click', () => showView('login'));
    authNav.registerBtn.addEventListener('click', () => showView('register'));
    authNav.logoutBtn.addEventListener('click', logout);

    document.getElementById('show-register-btn').addEventListener('click', () => showView('register'));
    document.getElementById('show-login-btn').addEventListener('click', () => showView('login'));

    document.getElementById('game-menu-btn').addEventListener('click', () => {
        document.getElementById('game-menu-overlay').classList.remove('hidden');
        if (typeof ws !== 'undefined' && ws && ws.readyState === WebSocket.OPEN) {
            ws.send(JSON.stringify({ type: 'input', key: 'pause' }));
        }
    });

    document.getElementById('resume-game-btn').addEventListener('click', () => {
        document.getElementById('game-menu-overlay').classList.add('hidden');
        if (typeof ws !== 'undefined' && ws && ws.readyState === WebSocket.OPEN) {
            ws.send(JSON.stringify({ type: 'input', key: 'pause' }));
        }
    });

    document.getElementById('restart-game-btn').addEventListener('click', () => {
        document.getElementById('game-menu-overlay').classList.add('hidden');
        const lastLevel = parseInt(localStorage.getItem('lastSelectedLevel')) || 1;
        startGame('tetris', lastLevel);
    });

    document.getElementById('back-to-main-menu-btn').addEventListener('click', () => {
        // Clean up the game before returning to menu
        if (typeof cleanupGame === 'function') {
            cleanupGame();
        }
        document.getElementById('game-menu-overlay').classList.add('hidden');
        showView('mainMenu');
    });

    document.getElementById('multiplayer-menu-btn').addEventListener('click', () => {
        document.getElementById('multiplayer-game-menu-overlay').classList.remove('hidden');
    });

    document.getElementById('resume-multiplayer-game-btn').addEventListener('click', () => {
        document.getElementById('multiplayer-game-menu-overlay').classList.add('hidden');
    });

    document.getElementById('leave-multiplayer-game-btn').addEventListener('click', () => {
        // Send disconnect message before leaving
        if (window.multiplayerManager && typeof window.multiplayerManager.disconnectFromRoom === 'function') {
            window.multiplayerManager.disconnectFromRoom(true);
        }

        if (typeof cleanupMultiplayerGame === 'function') {
            cleanupMultiplayerGame();
        }
        document.getElementById('multiplayer-game-menu-overlay').classList.add('hidden');
        showView('mainMenu');
    });

    document.getElementById('back-to-multiplayer-menu-btn').addEventListener('click', () => {
        // Send disconnect message before leaving
        if (window.multiplayerManager && typeof window.multiplayerManager.disconnectFromRoom === 'function') {
            window.multiplayerManager.disconnectFromRoom(true);
        }

        if (typeof cleanupMultiplayerGame === 'function') {
            cleanupMultiplayerGame();
        }
        document.getElementById('multiplayer-game-menu-overlay').classList.add('hidden');
        showView('multiplayer');
    });

    document.getElementById('back-to-menu-from-leaderboard-btn').addEventListener('click', () => showView('mainMenu'));
    document.getElementById('back-to-menu-from-level-select-btn').addEventListener('click', () => showView('mainMenu'));
    document.getElementById('back-to-menu-from-multiplayer-btn').addEventListener('click', () => showView('mainMenu'));

    document.getElementById('login-form').addEventListener('submit', handleLogin);
    document.getElementById('register-form').addEventListener('submit', handleRegister);

    document.getElementById('singleplayer-btn').addEventListener('click', () => {
        showView('levelSelect');
    });

    document.getElementById('multiplayer-btn').addEventListener('click', () => {
        // Check if user is logged in before allowing access to multiplayer
        const currentUser = getCurrentUser();
        if (!currentUser || !currentUser.token) {
            alert('Please log in to access multiplayer features.');
            showView('login');
            return;
        }

        showView('multiplayer');
    });

    document.querySelectorAll('.level-btn').forEach(button => {
        button.addEventListener('click', () => {
            const startLevel = parseInt(button.dataset.level);
            localStorage.setItem('lastSelectedLevel', startLevel);
            showView('game');
            startGame('tetris', startLevel);
        });
    });

    document.getElementById('leaderboard-btn').addEventListener('click', async () => {
        showView('leaderboard');
        await loadLeaderboard();
    });

    // Leaderboard filter tabs
    document.querySelectorAll('.tab-btn').forEach(btn => {
        btn.addEventListener('click', async () => {
            // Update active tab
            document.querySelectorAll('.tab-btn').forEach(b => b.classList.remove('active'));
            btn.classList.add('active');
            currentLeaderboardOptions.period = btn.dataset.period;
            await loadLeaderboard();
        });
    });

    document.querySelectorAll('.category-btn').forEach(btn => {
        btn.addEventListener('click', async () => {
            document.querySelectorAll('.category-btn').forEach(b => b.classList.remove('active'));
            btn.classList.add('active');
            currentLeaderboardOptions.category = btn.dataset.category;
            await loadLeaderboard();
        });
    });

    updateAuthUI();
    showView('mainMenu');

    window.addEventListener('beforeunload', (event) => {
        if (window.isMultiplayer && window.multiplayerWs && window.multiplayerWs.readyState === WebSocket.OPEN) {
            // Use the multiplayer manager's disconnect method if available
            if (window.multiplayerManager && typeof window.multiplayerManager.disconnectFromRoom === 'function') {
                window.multiplayerManager.disconnectFromRoom(true);
            } else {
                // Fallback to direct WebSocket message
                const currentUser = getCurrentUser();
                const roomId = window.multiplayerManager?.currentRoom?.id || '';

                window.multiplayerWs.send(JSON.stringify({
                    type: 'player_disconnect',
                    room_id: roomId,
                    user_id: currentUser?.id || 0,
                    data: { reason: 'page_unload' }
                }));

                // Also try to close the connection gracefully
                try {
                    window.multiplayerWs.close(1000, 'Page unload');
                } catch (e) {
                    console.log('Error closing WebSocket:', e);
                }
            }

            event.preventDefault();
            event.returnValue = 'You are currently in a multiplayer game. Are you sure you want to leave?';
            return event.returnValue;
        }
    });

    // Also handle page visibility changes and window blur/focus for additional disconnect detection
    document.addEventListener('visibilitychange', () => {
        if (document.hidden && window.isMultiplayer && window.multiplayerWs && window.multiplayerWs.readyState === WebSocket.OPEN) {
            // Page is hidden - could indicate tab switching or browser minimizing
            // Don't disconnect immediately, but send a heartbeat to detect if connection is lost
            setTimeout(() => {
                if (document.hidden && window.multiplayerWs && window.multiplayerWs.readyState === WebSocket.OPEN) {
                    try {
                        const currentUser = getCurrentUser();
                        const roomId = window.multiplayerManager?.currentRoom?.id || '';

                        window.multiplayerWs.send(JSON.stringify({
                            type: 'heartbeat',
                            room_id: roomId,
                            user_id: currentUser?.id || 0,
                            data: { timestamp: Date.now() }
                        }));
                    } catch (e) {
                        console.log('Heartbeat failed:', e);
                    }
                }
            }, 5000); // 5 second delay before heartbeat
        }
    });
});