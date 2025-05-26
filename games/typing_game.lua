math.randomseed(os.time())

words = {"gopher", "banana", "isaac", "computer", "monitor", "developer"}

function get_words()
    return words
end

function check_word(input, target)
    return input == target
end