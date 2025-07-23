local breakout = {}

-- Default game configuration
breakout.config = {
    paddle_size = 8,
    ball_speed_x = 0.5, -- Horizontal speed
    ball_speed_y = -0.25, -- Vertical speed (negative is up)
    lives = 3,
    brick_score = 10
}

-- Difficulty modifiers
breakout.difficulty = {
    easy = {
        paddle_size_multiplier = 1.5,
        ball_speed_multiplier = 0.8,
    },
    normal = {
        paddle_size_multiplier = 1.0,
        ball_speed_multiplier = 1.0,
    },
    hard = {
        paddle_size_multiplier = 0.7,
        ball_speed_multiplier = 1.3,
    }
}

-- Brick layouts
breakout.layouts = {
    standard = {
        {1,1,1,1,1,1,1,1,1,1},
        {1,1,1,1,1,1,1,1,1,1},
        {1,1,1,1,1,1,1,1,1,1},
        {1,1,1,1,1,1,1,1,1,1},
    }
}

return breakout