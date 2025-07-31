// This file will handle the main application flow, view switching, and event listeners.

document.addEventListener('DOMContentLoaded', () => {
    const views = {
        mainMenu: document.getElementById('main-menu-view'),
        game: document.getElementById('game-view'),
        login: document.getElementById('login-view'),
        register: document.getElementById('register-view'),
        leaderboard: document.getElementById('leaderboard-view'),
        levelSelect: document.getElementById('level-select-view'),
    };

    const authNav = {
        loginBtn: document.getElementById('login-nav-btn'),
        registerBtn: document.getElementById('register-nav-btn'),
        userInfo: document.getElementById('user-info'),
        usernameDisplay: document.getElementById('username-display'),
        logoutBtn: document.getElementById('logout-btn'),
    };

    function showView(viewName) {
        // Hide all views
        Object.values(views).forEach(view => {
            if (view) {
                view.classList.add('hidden');
                view.style.display = 'none';
            }
        });

        // Show/hide main header based on view
        const mainHeader = document.getElementById('main-header');
        if (viewName === 'game') {
            mainHeader.style.display = 'none';
        } else {
            mainHeader.style.display = 'flex';
        }

        // Show the requested view
        let targetView = null;
        if (viewName === 'mainMenu' || viewName === 'main-menu') {
            targetView = views.mainMenu;
        } else if (viewName === 'game') {
            targetView = views.game;
        } else if (viewName === 'login') {
            targetView = views.login;
        } else if (viewName === 'register') {
            targetView = views.register;
        } else if (viewName === 'leaderboard') {
            targetView = views.leaderboard;
        } else if (viewName === 'levelSelect') {
            targetView = views.levelSelect;
        }

        if (targetView) {
            targetView.classList.remove('hidden');
            targetView.style.display = '';
        }
    }    // Make showView globally accessible for game over screen
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

    // --- Event Listeners ---

    // Navigation
    authNav.loginBtn.addEventListener('click', () => showView('login'));
    authNav.registerBtn.addEventListener('click', () => showView('register'));
    authNav.logoutBtn.addEventListener('click', logout);

    document.getElementById('show-register-btn').addEventListener('click', () => showView('register'));
    document.getElementById('show-login-btn').addEventListener('click', () => showView('login'));

    // Game menu functionality
    document.getElementById('game-menu-btn').addEventListener('click', () => {
        document.getElementById('game-menu-overlay').classList.remove('hidden');
        // Pause the game when menu is opened
        if (typeof ws !== 'undefined' && ws && ws.readyState === WebSocket.OPEN) {
            ws.send(JSON.stringify({ type: 'input', key: 'pause' }));
        }
    });

    document.getElementById('resume-game-btn').addEventListener('click', () => {
        document.getElementById('game-menu-overlay').classList.add('hidden');
        // Resume the game
        if (typeof ws !== 'undefined' && ws && ws.readyState === WebSocket.OPEN) {
            ws.send(JSON.stringify({ type: 'input', key: 'pause' }));
        }
    });

    document.getElementById('restart-game-btn').addEventListener('click', () => {
        document.getElementById('game-menu-overlay').classList.add('hidden');
        // Get the last selected level and restart
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
    document.getElementById('back-to-menu-from-leaderboard-btn').addEventListener('click', () => showView('mainMenu'));
    document.getElementById('back-to-menu-from-level-select-btn').addEventListener('click', () => showView('mainMenu'));

    // Forms
    document.getElementById('login-form').addEventListener('submit', handleLogin);
    document.getElementById('register-form').addEventListener('submit', handleRegister);

    // Game Menu
    document.getElementById('singleplayer-btn').addEventListener('click', () => {
        showView('levelSelect');
    });

    document.getElementById('multiplayer-btn').addEventListener('click', () => {
        alert('Multiplayer is coming soon!');
    });

    // Level Selection
    document.querySelectorAll('.level-btn').forEach(button => {
        button.addEventListener('click', () => {
            const startLevel = parseInt(button.dataset.level);
            localStorage.setItem('lastSelectedLevel', startLevel);
            showView('game');
            startGame('tetris', startLevel);
        });
    });

    // Leaderboard
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

            // Update current options and reload
            currentLeaderboardOptions.period = btn.dataset.period;
            await loadLeaderboard();
        });
    });

    document.querySelectorAll('.category-btn').forEach(btn => {
        btn.addEventListener('click', async () => {
            // Update active category
            document.querySelectorAll('.category-btn').forEach(b => b.classList.remove('active'));
            btn.classList.add('active');

            // Update current options and reload
            currentLeaderboardOptions.category = btn.dataset.category;
            await loadLeaderboard();
        });
    });

    // --- Initialization ---
    updateAuthUI();
    showView('mainMenu');
});