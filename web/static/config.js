export const GAME_CONFIG = {
    TILE_SIZE: 40,
    SMALL_TILE_SIZE: 20,
    BOARD_WIDTH: 10,
    BOARD_HEIGHT: 20,

    CANVAS_SETUP_RETRY_DELAY: 100,
    CANVAS_SETUP_MAX_ATTEMPTS: 5,
    GAME_OVER_DISPLAY_DELAY: 3000,
    RECONNECTION_DELAY: 3000,
    NOTIFICATION_DELAY: 1000,

    DEFAULT_STARTING_LEVEL: 1,
    RECENT_GAMES_LIMIT: 5,

    WS_RECONNECT_ATTEMPTS: 3,
    WS_HEARTBEAT_INTERVAL: 30000,

    MODAL_Z_INDEX: 1000
};

export const PIECE_COLORS = [
    '#000000', // 0: Empty (Black)
    '#00FFFF', // 1: I piece (Cyan)
    '#0000FF', // 2: J piece (Blue)
    '#FFA500', // 3: L piece (Orange)
    '#FFFF00', // 4: O piece (Yellow)
    '#FF0000', // 5: Z piece (Red)
    '#800080', // 6: S piece (Purple)
    '#00FF00'  // 7: T piece (Green)
];

export const MESSAGE_TYPES = {
    GAME_STATE: 'game_state',
    GAME_INPUT: 'game_input',
    GAME_START: 'game_start',
    GAME_END: 'game_end',

    ROOM_UPDATE: 'room_update',
    PLAYER_JOINED: 'player_joined',
    PLAYER_LEFT: 'player_left',
    PLAYER_UPDATE: 'player_update',

    START_MULTIPLAYER_GAME: 'start_multiplayer_game',
    MULTIPLAYER_INIT: 'multiplayer_init',
    PLAYER_GAME_STATE: 'player_game_state',
    MATCH_ENDED: 'match_ended',

    READY: 'ready',
    ERROR: 'error'
};

export const WS_STATES = {
    CONNECTING: 0,
    OPEN: 1,
    CLOSING: 2,
    CLOSED: 3
};
