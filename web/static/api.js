// This file will contain functions for making API calls to the backend.

const API_BASE_URL = '/api';

async function apiRequest(method, path, body = null) {
    const options = {
        method,
        headers: {
            'Content-Type': 'application/json',
        },
    };

    const token = getAuthToken();
    if (token) {
        options.headers['Authorization'] = `Bearer ${token}`;
    }

    if (body) {
        options.body = JSON.stringify(body);
    }

    try {
        const response = await fetch(`${API_BASE_URL}${path}`, options);
        const data = await response.json();
        if (!response.ok) {
            throw new Error(data.error || 'An unknown error occurred.');
        }
        return data;
    } catch (error) {
        console.error(`API Error (${method} ${path}):`, error);
        throw error;
    }
}

// --- API Functions ---

function registerUser(username, email, password) {
    return apiRequest('POST', '/register', { username, email, password });
}

function loginUser(username, password) {
    return apiRequest('POST', '/login', { username, password });
}

function getLeaderboard(gameType, limit = 15) {
    return apiRequest('GET', `/leaderboard/${gameType}?limit=${limit}`);
}

// Submit a game score
function submitScore(gameType, score, metadata = {}) {
    return apiRequest('POST', '/scores', { game_type: gameType, score, metadata });
}
