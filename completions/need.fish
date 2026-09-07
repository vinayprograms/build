# Fish completion script for need
#
# Installation:
#   Copy this file to ~/.config/fish/completions/need.fish
#   or to /usr/share/fish/vendor_completions.d/need.fish

# Disable file completions by default
complete -c need -f

# Helper function to get Needfile path
function __need_needfile
    if test -f "Needfile"
        echo "Needfile"
    else if test -f "needfile"
        echo "needfile"
    end
end

# Helper function to get targets from Needfile
function __need_targets
    set -l needfile (__need_needfile)
    if test -n "$needfile"
        grep -E '^(@[a-zA-Z_][a-zA-Z0-9_-]*|[a-zA-Z0-9_./-]+):' "$needfile" 2>/dev/null | \
            sed -E 's/:.*$//' | \
            sed 's/^@//'
    end
end

# Helper function to get environments from Needfile
function __need_environments
    set -l needfile (__need_needfile)
    if test -n "$needfile"
        grep -E '^\s*\.environment:\s*' "$needfile" 2>/dev/null | \
            sed -E 's/^.*\.environment:\s*//g' | \
            grep -v '^$'
    end
end

# File flag
complete -c need -s f -l file -d "Use alternate Needfile" -r -F

# Environment flag
complete -c need -s e -l env -d "Use named environment" -xa "(__need_environments)"

# Jobs flag
complete -c need -s j -l jobs -d "Parallel jobs" -xa "1 2 4 8 16"

# Boolean flags
complete -c need -s n -l dry-run -d "Show what would execute"
complete -c need -s v -l verbose -d "Verbose output"
complete -c need -s q -l quiet -d "Suppress non-error output"
complete -c need -l check-env -d "Verify environment requirements"
complete -c need -l show-install -d "Show install instructions"
complete -c need -l list-env -d "List available environments"
complete -c need -l shell -d "Open shell in sandbox environment"
complete -c need -l keep -d "Keep sandbox running after build"
complete -c need -s V -l version -d "Show version"
complete -c need -s h -l help -d "Show help"

# Color and progress flags
complete -c need -l color -d "Color output mode" -xa "auto always never"
complete -c need -l progress -d "Progress output mode" -xa "auto always never"

# Debug flags
complete -c need -l debug-lex -d "Dump lexer tokens"
complete -c need -l debug-parse -d "Dump parser scope validation"
complete -c need -l debug-var -d "Dump variable parsing"
complete -c need -l debug-target -d "Dump target parsing"
complete -c need -l debug-recipe -d "Dump recipe parsing"
complete -c need -l debug-env -d "Dump environment parsing"
complete -c need -l debug-cond -d "Dump conditional parsing"
complete -c need -l debug-include -d "Dump include parsing"
complete -c need -l debug-ast -d "Dump full AST with error recovery"
complete -c need -l debug-semantic -d "Dump semantic analysis"
complete -c need -l debug-eval -d "Dump variable evaluation"
complete -c need -l debug-plan -d "Dump build planning"

# Target completions (when not completing an option)
complete -c need -n "not __fish_seen_subcommand_from --" -xa "(__need_targets)"
