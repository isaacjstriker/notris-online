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
    document.getElementById('back-to-menu-from-login-btn').addEventListener('click', () => showView('mainMenu'));
    document.getElementById('back-to-menu-from-register-btn').addEventListener('click', () => showView('mainMenu'));

    document.getElementById('login-form').addEventListener('submit', handleLogin);
    document.getElementById('register-form').addEventListener('submit', handleRegister);

    document.getElementById('singleplayer-btn').addEventListener('click', () => {
        showView('levelSelect');
    });

    document.getElementById('multiplayer-btn').addEventListener('click', () => {
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
            // Small delay to ensure the view is fully rendered before starting the game
            setTimeout(() => {
                // Try initializing canvases with retry logic first
                if (initializeSingleplayerCanvasesWithRetry()) {
                    startGame('tetris', startLevel);
                } else {
                    console.error('Failed to initialize canvases, cannot start game');
                    alert('Failed to start the game. Please refresh the page and try again.');
                }
            }, 100);
        });
    });

    document.getElementById('leaderboard-btn').addEventListener('click', async () => {
        showView('leaderboard');
        await loadLeaderboard();
    });

    document.querySelectorAll('.tab-btn').forEach(btn => {
        btn.addEventListener('click', async () => {
            document.querySelectorAll('.tab-btn').forEach(b => b.classList.remove('active'));
            btn.classList.add('active');
            currentLeaderboardOptions.period = btn.dataset.period;
            await loadLeaderboard();
        });
    });

    updateAuthUI();
    showView('mainMenu');

    window.addEventListener('beforeunload', (event) => {
        if (window.isMultiplayer && window.multiplayerWs && window.multiplayerWs.readyState === WebSocket.OPEN) {
            if (window.multiplayerManager && typeof window.multiplayerManager.disconnectFromRoom === 'function') {
                window.multiplayerManager.disconnectFromRoom(true);
            } else {
                const currentUser = getCurrentUser();
                const roomId = window.multiplayerManager?.currentRoom?.id || '';

                window.multiplayerWs.send(JSON.stringify({
                    type: 'player_disconnect',
                    room_id: roomId,
                    user_id: currentUser?.id || 0,
                    data: { reason: 'page_unload' }
                }));

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

    document.addEventListener('visibilitychange', () => {
        if (document.hidden && window.isMultiplayer && window.multiplayerWs && window.multiplayerWs.readyState === WebSocket.OPEN) {
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
            }, 5000);
        }
    });

    // Password Toggle Functionality
    function initializePasswordToggles() {
        const passwordToggleBtns = document.querySelectorAll('.password-toggle-btn');

        passwordToggleBtns.forEach(btn => {
            btn.addEventListener('click', (e) => {
                e.preventDefault();

                const targetId = btn.getAttribute('data-target');
                const passwordInput = document.getElementById(targetId);
                const toggleIcon = btn.querySelector('.password-toggle-icon');

                if (passwordInput.type === 'password') {
                    passwordInput.type = 'text';
                    toggleIcon.textContent = '🙈'; // Hide icon
                    btn.setAttribute('aria-label', 'Hide password');
                } else {
                    passwordInput.type = 'password';
                    toggleIcon.textContent = '👁️'; // Show icon
                    btn.setAttribute('aria-label', 'Show password');
                }
            });
        });
    }

    // Initialize password toggles
    initializePasswordToggles();
});