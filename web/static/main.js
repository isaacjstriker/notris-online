// This file will handle the main application flow, view switching, and event listeners.

document.addEventListener('DOMContentLoaded', () => {
    const views = {
        mainMenu: document.getElementById('main-menu-view'),
        game: document.getElementById('game-view'),
        login: document.getElementById('login-view'),
        register: document.getElementById('register-view'),
        leaderboard: document.getElementById('leaderboard-view'),
    };

    const authNav = {
        loginBtn: document.getElementById('login-nav-btn'),
        registerBtn: document.getElementById('register-nav-btn'),
        userInfo: document.getElementById('user-info'),
        usernameDisplay: document.getElementById('username-display'),
        logoutBtn: document.getElementById('logout-btn'),
    };

    function showView(viewName) {
        Object.values(views).forEach(view => view.classList.add('hidden'));
        if (views[viewName]) {
            views[viewName].classList.remove('hidden');
        }
    }

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
    document.getElementById('back-to-menu-btn').addEventListener('click', () => showView('mainMenu'));
    document.getElementById('back-to-menu-from-leaderboard-btn').addEventListener('click', () => showView('mainMenu'));

    // Forms
    document.getElementById('login-form').addEventListener('submit', handleLogin);
    document.getElementById('register-form').addEventListener('submit', handleRegister);

    // Game Menu
    document.querySelectorAll('.menu-btn[data-game]').forEach(button => {
        button.addEventListener('click', () => {
            const gameType = button.dataset.game;
            showView('game');
            // Start the actual game now
            startGame(gameType);
        });
    });

    // Leaderboard
    document.getElementById('leaderboard-btn').addEventListener('click', async () => {
        showView('leaderboard');
        await loadLeaderboard();
    });

    document.getElementById('leaderboard-game-select').addEventListener('change', loadLeaderboard);


    // --- Initialization ---
    updateAuthUI();
    showView('mainMenu');
});
