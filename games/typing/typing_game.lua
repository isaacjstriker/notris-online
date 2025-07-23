math.randomseed(os.time())

-- Word list for the typing game
local words = {
    "hello", "world", "programming", "computer", "keyboard", "mouse", "screen",
    "typing", "speed", "accuracy", "challenge", "game", "score", "words",
    "practice", "skills", "improvement", "focus", "concentration", "rhythm",
    "language", "software", "development", "coding", "algorithm", "function",
    "variable", "constant", "loop", "condition", "array", "string", "integer",
    "boolean", "object", "method", "class", "interface", "package", "import",
    "export", "return", "break", "continue", "switch", "case", "default",
    "public", "private", "protected", "static", "final", "abstract", "virtual"
}

-- Shuffle function for randomizing words
local function shuffle(tbl)
    math.randomseed(os.time())
    for i = #tbl, 2, -1 do
        local j = math.random(i)
        tbl[i], tbl[j] = tbl[j], tbl[i]
    end
end

-- Trim whitespace from string
local function trim(s)
    return s:match("^%s*(.-)%s*$")
end

-- Display game results
local function display_results(stats)
    go_println("\n" .. string.rep("=", 50))
    go_println("TYPING CHALLENGE RESULTS")
    go_println(string.rep("=", 50))
    
    go_println(string.format("Time: %.1f seconds", stats.total_time))
    go_println(string.format("Words Typed: %d", stats.words_typed))
    go_println(string.format("Correct Words: %d", stats.correct_words))
    go_println(string.format("Accuracy: %.1f%%", stats.accuracy))
    go_println(string.format("Speed: %.1f WPM", stats.wpm))
    go_println(string.format("Final Score: %d points", stats.score))
    
    go_println("\n--- Performance Feedback ---")
    -- Performance feedback
    if stats.accuracy >= 95 then
        go_println("PERFECT! Amazing accuracy!")
    elseif stats.accuracy >= 85 then
        go_println("Great job! Excellent accuracy!")
    elseif stats.accuracy >= 70 then
        go_println("Good work! Keep practicing!")
    else
        go_println("Keep practicing to improve your accuracy!")
    end
    
    if stats.wpm >= 60 then
        go_println("Lightning fast typing!")
    elseif stats.wpm >= 40 then
        go_println("Excellent typing speed!")
    elseif stats.wpm >= 25 then
        go_println("Good typing speed!")
    else
        go_println("Focus on building your speed!")
    end
    
    go_println(string.rep("=", 50))
    go_println("Press Enter to continue...")
    go_read_line()
end

-- Main typing game function
function run_typing_game()
    -- Initialize game state
    local stats = {
        score = 0,
        wpm = 0,
        accuracy = 0,
        words_typed = 0,
        correct_words = 0,
        total_time = 0
    }
    
    -- Shuffle words and select subset
    shuffle(words)
    local game_words = {}
    for i = 1, math.min(20, #words) do
        game_words[i] = words[i]
    end
    
    -- Game introduction
    go_println("\n--- TYPING SPEED CHALLENGE ---")
    go_println(string.rep("=", 50))
    go_println(string.format("[INFO] You will type %d words as fast and accurately as possible.", #game_words))
    go_println("[INFO] Type each word exactly as shown and press Enter.")
    go_println("[RULE] The game ends if you make a mistake.")
    go_println("\nReady? Press Enter to start...")
    go_read_line()
    
    local start_time = go_current_time()
    
    -- Main game loop
    for i, word in ipairs(game_words) do
        go_println(string.format("\n[%d/%d] Type: %s", i, #game_words, word))
        go_print("> ")
        
        local user_input = trim(go_read_line())
        stats.words_typed = stats.words_typed + 1
        
        if user_input == word then
            go_println("[OK] Correct!")
            stats.correct_words = stats.correct_words + 1
        else
            go_println(string.format("[ERROR] Incorrect! You typed: '%s', expected: '%s'", user_input, word))
            go_println("Game Over!")
            break
        end
        
        -- Brief pause for user experience
        go_sleep(0.2)
    end
    
    local end_time = go_current_time()
    stats.total_time = end_time - start_time
    
    -- Calculate final statistics
    if stats.words_typed > 0 then
        stats.accuracy = (stats.correct_words / stats.words_typed) * 100
    end
    
    if stats.total_time > 0 then
        stats.wpm = stats.correct_words / (stats.total_time / 60.0)
    end
    
    -- Calculate score (WPM * Accuracy percentage)
    stats.score = math.floor(stats.wpm * (stats.accuracy / 100) * 10)
    
    -- Display results
    display_results(stats)
    
    -- Return stats table to Go
    return stats
end