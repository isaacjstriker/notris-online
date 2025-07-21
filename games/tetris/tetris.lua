-- Tetris Game Configuration and Scripting
-- This file provides Lua-based configuration for the Tetris game

tetris = {}

-- Game configuration
tetris.config = {
    board_width = 10,
    board_height = 20,
    starting_level = 1,
    lines_per_level = 10,
    base_drop_speed = 1000, -- milliseconds
    min_drop_speed = 50,
    
    -- Scoring system
    scoring = {
        single_line = 40,
        double_line = 100,
        triple_line = 300,
        tetris = 1200,
        soft_drop_points = 1,
        hard_drop_points = 2
    },
    
    -- Piece spawn probabilities (can be modified for different gameplay)
    piece_weights = {
        I = 1.0,  -- Line piece
        O = 1.0,  -- Square piece
        T = 1.0,  -- T piece
        S = 1.0,  -- S piece
        Z = 1.0,  -- Z piece
        J = 1.0,  -- J piece
        L = 1.0   -- L piece
    }
}

-- Calculate drop speed based on level
function tetris.get_drop_speed(level)
    local speed = tetris.config.base_drop_speed - ((level - 1) * 50)
    if speed < tetris.config.min_drop_speed then
        speed = tetris.config.min_drop_speed
    end
    return speed
end

-- Calculate score for cleared lines
function tetris.calculate_line_score(lines_cleared, level)
    local base_score = 0
    
    if lines_cleared == 1 then
        base_score = tetris.config.scoring.single_line
    elseif lines_cleared == 2 then
        base_score = tetris.config.scoring.double_line
    elseif lines_cleared == 3 then
        base_score = tetris.config.scoring.triple_line
    elseif lines_cleared == 4 then
        base_score = tetris.config.scoring.tetris
    end
    
    return base_score * (level + 1)
end

-- Generate weighted random piece
function tetris.get_random_piece()
    local total_weight = 0
    for piece, weight in pairs(tetris.config.piece_weights) do
        total_weight = total_weight + weight
    end
    
    local random_value = math.random() * total_weight
    local current_weight = 0
    
    for piece, weight in pairs(tetris.config.piece_weights) do
        current_weight = current_weight + weight
        if random_value <= current_weight then
            return piece
        end
    end
    
    return "I" -- Default fallback
end

-- Calculate level progression
function tetris.calculate_level(lines_cleared)
    return math.floor(lines_cleared / tetris.config.lines_per_level) + 1
end

-- Achievement system
tetris.achievements = {
    {
        name = "First Steps",
        description = "Clear your first line",
        check = function(stats)
            return stats.lines >= 1
        end
    },
    {
        name = "Tetris Master",
        description = "Clear 4 lines at once",
        check = function(stats)
            return stats.max_lines_at_once >= 4
        end
    },
    {
        name = "Speed Demon",
        description = "Reach level 10",
        check = function(stats)
            return stats.level >= 10
        end
    },
    {
        name = "Marathon Runner",
        description = "Clear 100 lines",
        check = function(stats)
            return stats.lines >= 100
        end
    },
    {
        name = "High Scorer",
        description = "Score over 50,000 points",
        check = function(stats)
            return stats.score >= 50000
        end
    }
}

-- Check for achievements
function tetris.check_achievements(stats)
    local earned = {}
    for i, achievement in ipairs(tetris.achievements) do
        if achievement.check(stats) then
            table.insert(earned, achievement)
        end
    end
    return earned
end

-- Custom game mode configurations
tetris.game_modes = {
    classic = {
        name = "Classic",
        description = "Traditional Tetris gameplay",
        starting_level = 1,
        ghost_piece = false
    },
    sprint = {
        name = "40-Line Sprint",
        description = "Clear 40 lines as fast as possible",
        starting_level = 1,
        target_lines = 40,
        ghost_piece = true
    },
    marathon = {
        name = "Marathon",
        description = "Endless play with increasing difficulty",
        starting_level = 1,
        ghost_piece = true,
        level_cap = 15
    }
}

-- Helper function to get game mode configuration
function tetris.get_mode_config(mode_name)
    return tetris.game_modes[mode_name] or tetris.game_modes.classic
end

-- Difficulty modifiers
tetris.difficulty = {
    easy = {
        drop_speed_multiplier = 1.5,
        line_clear_delay = 500,
        preview_pieces = 3
    },
    normal = {
        drop_speed_multiplier = 1.0,
        line_clear_delay = 300,
        preview_pieces = 1
    },
    hard = {
        drop_speed_multiplier = 0.7,
        line_clear_delay = 100,
        preview_pieces = 1
    }
}

-- Tetris piece definitions (for reference)
tetris.pieces = {
    I = {
        name = "I-Piece",
        color = "cyan",
        rotations = 2,
        shapes = {
            {{1,1,1,1}},
            {{1},{1},{1},{1}}
        }
    },
    O = {
        name = "O-Piece", 
        color = "yellow",
        rotations = 1,
        shapes = {
            {{1,1},{1,1}}
        }
    },
    T = {
        name = "T-Piece",
        color = "purple", 
        rotations = 4,
        shapes = {
            {{0,1,0},{1,1,1}},
            {{1,0},{1,1},{1,0}},
            {{1,1,1},{0,1,0}},
            {{0,1},{1,1},{0,1}}
        }
    }
    -- Additional pieces would be defined here
}

return tetris