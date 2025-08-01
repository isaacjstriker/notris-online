// This file will handle authentication logic on the frontend.

const TOKEN_KEY = 'devware_jwt';
const USERNAME_KEY = 'devware_username';
const USER_ID_KEY = 'devware_user_id';

function saveAuthInfo(token, username, userID) {
    localStorage.setItem(TOKEN_KEY, token);
    localStorage.setItem(USERNAME_KEY, username);
    if (userID !== undefined && userID !== null) {
        localStorage.setItem(USER_ID_KEY, userID.toString());
    }
}

function clearAuthInfo() {
    localStorage.removeItem(TOKEN_KEY);
    localStorage.removeItem(USERNAME_KEY);
    localStorage.removeItem(USER_ID_KEY);
}

function getAuthToken() {
    return localStorage.getItem(TOKEN_KEY);
}

function getUserID() {
    const userID = localStorage.getItem(USER_ID_KEY);
    return userID ? parseInt(userID) : null;
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
        saveAuthInfo(data.token, data.username, data.user_id);
        location.reload(); // Reload to update UI everywhere
    } catch (error) {
        errorEl.textContent = error.message;
    }
}

async function handleRegister(event) {
    event.preventDefault();

    const errorEl = document.getElementById('register-error');
    errorEl.textContent = '';

    const formData = new FormData(event.target);
    const userData = {
        username: formData.get('username'),
        email: formData.get('email'),
        password: formData.get('password'),
        confirmPassword: formData.get('confirmPassword')
    };

    // Basic validation
    if (!userData.username || userData.username.length < 3) {
        errorEl.textContent = 'Username must be at least 3 characters long';
        return;
    }

    if (!userData.email) {
        errorEl.textContent = 'Email is required';
        return;
    }

    if (!userData.password || userData.password.length < 6) {
        errorEl.textContent = 'Password must be at least 6 characters long';
        return;
    }

    if (userData.password !== userData.confirmPassword) {
        errorEl.textContent = 'Passwords do not match';
        return;
    }

    try {
        const response = await fetch('/api/register', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                username: userData.username,
                email: userData.email,
                password: userData.password
            }),
        });

        const result = await response.json();

        if (response.ok) {
            // Registration successful, now log in the user
            console.log('Registration successful, attempting login...');

            const loginResponse = await fetch('/api/login', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    username: userData.username,
                    password: userData.password
                }),
            });

            const loginResult = await loginResponse.json();

            if (loginResponse.ok) {
                // Login successful after registration
                saveAuthInfo(loginResult.token, loginResult.username, loginResult.user_id);
                updateAuthUI();
                showView('mainMenu');
            } else {
                // Registration worked but login failed
                errorEl.textContent = 'Registration successful, but login failed. Please try logging in manually.';
            }
        } else {
            // Registration failed
            errorEl.textContent = result.error || 'Registration failed';
        }
    } catch (error) {
        errorEl.textContent = 'Registration failed: ' + error.message;
    }
}

function getCurrentUser() {
    const token = localStorage.getItem('devware_jwt');
    const username = localStorage.getItem('devware_username');
    const user_id = localStorage.getItem('devware_user_id');

    if (!token || !username) {
        return null;
    }

    if (!user_id) {
        return {
            id: null,
            username: username,
            token: token
        };
    }

    return { id: user_id, username: username, token: token };
}

window.getCurrentUser = getCurrentUser;

function getCurrentUser() {
    const token = getAuthToken();
    const username = getUsername();
    let userID = getUserID();

    // If user_id is missing but we have token and username, user might have logged in before the user_id fix
    if (token && username && !userID) {
        console.log('User logged in but no user_id stored. This user may need to log in again.');
        // For now, return a user object with id: null to indicate the issue
        return { token, username, id: null };
    }

    return token && username && userID ? { token, username, id: userID } : null;
}

// Make getCurrentUser globally accessible
window.getCurrentUser = getCurrentUser;

function logout() {
    clearAuthInfo();
    location.reload();
}
