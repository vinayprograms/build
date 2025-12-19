# Fish completion script for build
#
# Installation:
#   Copy this file to ~/.config/fish/completions/build.fish
#   or to /usr/share/fish/vendor_completions.d/build.fish

# Disable file completions by default
complete -c build -f

# Helper function to get Buildfile path
function __build_buildfile
    if test -f "Buildfile"
        echo "Buildfile"
    else if test -f "buildfile"
        echo "buildfile"
    end
end

# Helper function to get targets from Buildfile
function __build_targets
    set -l buildfile (__build_buildfile)
    if test -n "$buildfile"
        grep -E '^(@[a-zA-Z_][a-zA-Z0-9_-]*|[a-zA-Z0-9_./-]+):' "$buildfile" 2>/dev/null | \
            sed -E 's/:.*$//' | \
            sed 's/^@//'
    end
end

# Helper function to get environments from Buildfile
function __build_environments
    set -l buildfile (__build_buildfile)
    if test -n "$buildfile"
        grep -E '^\s*\.environment:\s*' "$buildfile" 2>/dev/null | \
            sed -E 's/^.*\.environment:\s*//g' | \
            grep -v '^$'
    end
end

# File flag
complete -c build -s f -l file -d "Use alternate Buildfile" -r -F

# Environment flag
complete -c build -s e -l env -d "Use named environment" -xa "(__build_environments)"

# Jobs flag
complete -c build -s j -l jobs -d "Parallel jobs" -xa "1 2 4 8 16"

# Boolean flags
complete -c build -s n -l dry-run -d "Show what would execute"
complete -c build -s v -l verbose -d "Verbose output"
complete -c build -s q -l quiet -d "Suppress non-error output"
complete -c build -l check-env -d "Verify environment requirements"
complete -c build -l show-install -d "Show install instructions"
complete -c build -l list-env -d "List available environments"
complete -c build -l shell -d "Open shell in sandbox environment"
complete -c build -l keep -d "Keep sandbox running after build"
complete -c build -s V -l version -d "Show version"
complete -c build -s h -l help -d "Show help"

# Color and progress flags
complete -c build -l color -d "Color output mode" -xa "auto always never"
complete -c build -l progress -d "Progress output mode" -xa "auto always never"

# Debug flags
complete -c build -l debug-lex -d "Dump lexer tokens"
complete -c build -l debug-parse -d "Dump parser scope validation"
complete -c build -l debug-var -d "Dump variable parsing"
complete -c build -l debug-target -d "Dump target parsing"
complete -c build -l debug-recipe -d "Dump recipe parsing"
complete -c build -l debug-env -d "Dump environment parsing"
complete -c build -l debug-cond -d "Dump conditional parsing"
complete -c build -l debug-include -d "Dump include parsing"
complete -c build -l debug-ast -d "Dump full AST with error recovery"
complete -c build -l debug-semantic -d "Dump semantic analysis"
complete -c build -l debug-eval -d "Dump variable evaluation"
complete -c build -l debug-plan -d "Dump build planning"

# Target completions (when not completing an option)
complete -c build -n "not __fish_seen_subcommand_from --" -xa "(__build_targets)"
