-- name: CreateChallengeScore :one
INSERT INTO challenge_scores (
    user_id, 
    total_score, 
    games_played, 
    total_duration, 
    avg_accuracy, 
    perfect_games, 
    results_json,
    created_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetChallengeScoresByUser :many
SELECT * FROM challenge_scores 
WHERE user_id = ? 
ORDER BY created_at DESC 
LIMIT ?;

-- name: GetTopChallengeScores :many
SELECT 
    cs.*,
    u.username,
    u.email
FROM challenge_scores cs
JOIN users u ON cs.user_id = u.id
ORDER BY cs.total_score DESC 
LIMIT ?;

-- name: GetUserBestChallengeScore :one
SELECT * FROM challenge_scores 
WHERE user_id = ? 
ORDER BY total_score DESC 
LIMIT 1;

-- name: GetChallengeScoreStats :one
SELECT 
    COUNT(*) as total_challenges,
    AVG(total_score) as avg_score,
    MAX(total_score) as best_score,
    AVG(avg_accuracy) as overall_avg_accuracy
FROM challenge_scores 
WHERE user_id = ?;