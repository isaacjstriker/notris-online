math.randomseed(os.time())

words = {"gopher", "banana", "isaac", "computer", "monitor", "developer",
    "keyboard", "mouse", "screen", "internet", "program", "function",
    "variable", "loop", "array", "slice", "channel", "goroutine", "struct",
    "interface"}

function get_random_words(n)
    local selected = {}
    local used = {}
    local dict_size = #words
    for i = 1, n do
        local idx
        repeat
            idx = math.random(dict_size)
        until not used[idx]
        used[idx] = true
        table.insert(selected, words[idx])
    end
    return selected
end

function check_word(input, target)
    return input == target
end