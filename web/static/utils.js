import { GAME_CONFIG } from './config.js';
import { logger } from './logger.js';

export const CanvasUtils = {
    setupWithRetry: async function (canvasSelector, maxAttempts = GAME_CONFIG.CANVAS_SETUP_MAX_ATTEMPTS) {
        for (let attempt = 0; attempt < maxAttempts; attempt++) {
            const canvas = document.querySelector(canvasSelector);
            if (canvas) {
                logger.debug(`Canvas setup successful for ${canvasSelector} on attempt ${attempt + 1}`);
                return canvas;
            }

            if (attempt < maxAttempts - 1) {
                logger.debug(`Canvas setup failed for ${canvasSelector}, retrying in ${GAME_CONFIG.CANVAS_SETUP_RETRY_DELAY}ms... (attempt ${attempt + 1})`);
                await new Promise(resolve => setTimeout(resolve, GAME_CONFIG.CANVAS_SETUP_RETRY_DELAY));
            }
        }

        logger.error(`Failed to set up canvas ${canvasSelector} after ${maxAttempts} attempts`);
        throw new Error(`Canvas ${canvasSelector} not found after ${maxAttempts} attempts`);
    },

    clear: function (canvas, ctx) {
        ctx.clearRect(0, 0, canvas.width, canvas.height);
    },

    calculateTileSize: function (canvas) {
        return canvas.width / GAME_CONFIG.BOARD_WIDTH;
    }
};

export const WebSocketUtils = {
    create: function (url, onOpen, onMessage, onClose, onError) {
        const ws = new WebSocket(url);

        ws.onopen = (event) => {
            logger.info(`WebSocket connected to ${url}`);
            if (onOpen) onOpen(event);
        };

        ws.onmessage = (event) => {
            try {
                const message = JSON.parse(event.data);
                logger.debug('WebSocket message received', { type: message.type, url });
                if (onMessage) onMessage(message);
            } catch (error) {
                logger.error('Failed to parse WebSocket message', { error, data: event.data });
            }
        };

        ws.onclose = (event) => {
            logger.info(`WebSocket disconnected from ${url}`, { code: event.code, reason: event.reason });
            if (onClose) onClose(event);
        };

        ws.onerror = (error) => {
            logger.error(`WebSocket error for ${url}`, error);
            if (onError) onError(error);
        };

        return ws;
    },

    send: function (ws, message) {
        if (ws && ws.readyState === WebSocket.OPEN) {
            try {
                ws.send(JSON.stringify(message));
                logger.debug('WebSocket message sent', { type: message.type });
                return true;
            } catch (error) {
                logger.error('Failed to send WebSocket message', { error, message });
                return false;
            }
        } else {
            logger.warn('Attempted to send message on closed WebSocket', {
                readyState: ws?.readyState,
                messageType: message.type
            });
            return false;
        }
    }
};

export const DOMUtils = {
    show: function (selector) {
        const element = typeof selector === 'string' ? document.querySelector(selector) : selector;
        if (element) {
            element.classList.remove('hidden');
        }
    },

    hide: function (selector) {
        const element = typeof selector === 'string' ? document.querySelector(selector) : selector;
        if (element) {
            element.classList.add('hidden');
        }
    },

    toggle: function (selector, force) {
        const element = typeof selector === 'string' ? document.querySelector(selector) : selector;
        if (element) {
            if (force !== undefined) {
                element.classList.toggle('hidden', !force);
            } else {
                element.classList.toggle('hidden');
            }
        }
    },

    querySelector: function (selector, required = false) {
        const element = document.querySelector(selector);
        if (!element && required) {
            logger.error(`Required element not found: ${selector}`);
        }
        return element;
    }
};

export const GameUtils = {
    createEmptyBoard: function () {
        return Array(GAME_CONFIG.BOARD_HEIGHT).fill().map(() => Array(GAME_CONFIG.BOARD_WIDTH).fill(0));
    },

    formatTime: function (seconds) {
        const minutes = Math.floor(seconds / 60);
        const remainingSeconds = seconds % 60;
        return `${minutes.toString().padStart(2, '0')}:${remainingSeconds.toString().padStart(2, '0')}`;
    },

    formatScore: function (score) {
        return score.toLocaleString();
    }
};

export const ValidationUtils = {
    isValidRoomName: function (name) {
        return name && name.trim().length >= 3 && name.trim().length <= 50;
    },

    isValidStartingLevel: function (level) {
        return Number.isInteger(level) && level >= 1 && level <= 20;
    },

    isValidUsername: function (username) {
        return username && username.trim().length >= 3 && username.trim().length <= 20;
    }
};
