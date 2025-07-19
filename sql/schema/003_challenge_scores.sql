CREATE TABLE challenge_scores (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    total_score INTEGER NOT NULL,
    games_played INTEGER NOT NULL,
    total_duration REAL NOT NULL,
    avg_accuracy REAL NOT NULL,
    perfect_games INTEGER NOT NULL,
    results_json TEXT NOT NULL, -- JSON array of individual game results
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE INDEX idx_challenge_scores_user_id ON challenge_scores(user_id);
CREATE INDEX idx_challenge_scores_total_score ON challenge_scores(total_score DESC);
CREATE INDEX idx_challenge_scores_created_at ON challenge_scores(created_at DESC);