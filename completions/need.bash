# Bash completion script for need
#
# Installation:
#   Copy this file to /etc/bash_completion.d/need
#   or add to your ~/.bashrc:
#     source /path/to/need.bash

_need_completions() {
    local cur prev opts
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"

    # Global options
    opts="--file -f --env -e --jobs -j --dry-run -n --verbose -v --quiet -q"
    opts+=" --color --progress --check-env --show-install --list-env"
    opts+=" --shell --keep --version -V --help -h"
    opts+=" --debug-lex --debug-parse --debug-var --debug-target"
    opts+=" --debug-recipe --debug-env --debug-cond --debug-include"
    opts+=" --debug-ast --debug-semantic --debug-eval --debug-plan"

    # Handle specific option arguments
    case "${prev}" in
        --file|-f)
            # Complete with Needfile-like files
            COMPREPLY=( $(compgen -f -- "${cur}") )
            return 0
            ;;
        --env|-e)
            # Complete with environment names from Needfile
            if [[ -f "Needfile" || -f "needfile" ]]; then
                local envs
                envs=$(grep -E '^\s*\.environment:\s*' Needfile needfile 2>/dev/null | \
                       sed -E 's/^.*\.environment:\s*//g' | \
                       grep -v '^$' || true)
                COMPREPLY=( $(compgen -W "${envs}" -- "${cur}") )
            fi
            return 0
            ;;
        --jobs|-j)
            # Complete with common job counts
            COMPREPLY=( $(compgen -W "1 2 4 8 16" -- "${cur}") )
            return 0
            ;;
        --color)
            COMPREPLY=( $(compgen -W "auto always never" -- "${cur}") )
            return 0
            ;;
        --progress)
            COMPREPLY=( $(compgen -W "auto always never" -- "${cur}") )
            return 0
            ;;
    esac

    # If current word starts with -, complete with options
    if [[ "${cur}" == -* ]]; then
        COMPREPLY=( $(compgen -W "${opts}" -- "${cur}") )
        return 0
    fi

    # Complete with targets from Needfile
    if [[ -f "Needfile" || -f "needfile" ]]; then
        local targets needfile
        needfile="Needfile"
        [[ -f "needfile" ]] && needfile="needfile"
        
        # Extract targets (lines that start with @ or have : without =)
        targets=$(grep -E '^(@[a-zA-Z_][a-zA-Z0-9_-]*|[a-zA-Z0-9_./-]+):' "${needfile}" 2>/dev/null | \
                  sed -E 's/:.*$//' | \
                  sed 's/^@//' || true)
        COMPREPLY=( $(compgen -W "${targets}" -- "${cur}") )
    fi

    return 0
}

complete -F _need_completions need
