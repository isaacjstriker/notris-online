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

function registerUser(username, email, password) {
    return apiRequest('POST', '/register', { username, email, password });
}

function loginUser(username, password) {
    return apiRequest('POST', '/api/login', { username, password });
}

function getLeaderboard(gameType, options = {}) {
    const params = new URLSearchParams();

    if (options.limit) params.append('limit', options.limit);
    if (options.period) params.append('period', options.period);
    if (options.category) params.append('category', options.category);
    if (options.includeAchievements) params.append('include_achievements', 'true');

    const queryString = params.toString();
    const url = `/leaderboard/${gameType}${queryString ? '?' + queryString : ''}`;

    return apiRequest('GET', url);
}

function getRecentGames(gameType, limit = 10) {
    return apiRequest('GET', `/recent/${gameType}?limit=${limit}`);
}

function submitScore(gameType, score, metadata = {}) {
    return apiRequest('POST', '/scores', { game_type: gameType, score, metadata });
}

function apiCall(path, method, body = null) {
    console.log(`apiCall: ${method} ${path}`, body);
    return apiRequest(method, path, body);
}
