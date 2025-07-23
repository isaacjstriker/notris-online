// This file will handle authentication logic on the frontend.

const TOKEN_KEY = 'devware_jwt';
const USERNAME_KEY = 'devware_username';

function saveAuthInfo(token, username) {
    localStorage.setItem(TOKEN_KEY, token);
    localStorage.setItem(USERNAME_KEY, username);
}

function clearAuthInfo() {
    localStorage.removeItem(TOKEN_KEY);
    localStorage.removeItem(USERNAME_KEY);
}

function getAuthToken() {
    return localStorage.getItem(TOKEN_KEY);
}

function getUsername() {
    return localStorage.getItem(USERNAME_KEY);
}

async function handleLogin(event) {
    event.preventDefault();
    const username = document.getElementById('login-username').value;
    const password = document.getElementById('login-password').value;
    const errorEl = document.getElementById('login-error');

    try {
        const data = await loginUser(username, password);
        saveAuthInfo(data.token, data.username);
        location.reload(); // Reload to update UI everywhere
    } catch (error) {
        errorEl.textContent = error.message;
    }
}

async function handleRegister(event) {
    event.preventDefault();
    const username = document.getElementById('register-username').value;
    const email = document.getElementById('register-email').value;
    const password = document.getElementById('register-password').value;
    const errorEl = document.getElementById('register-error');

    try {
        await registerUser(username, email, password);
        // Automatically log in after successful registration
        const data = await loginUser(username, password);
        saveAuthInfo(data.token, data.username);
        location.reload();
    } catch (error) {
        errorEl.textContent = error.message;
    }
}

function logout() {
    clearAuthInfo();
    location.reload();
}
