#/bin/sh


run() {
  flag="$1"
  file="$2"
  echo "=+=+=+=+=+=+=+= $flag =+=+=+=+=+=+=+=+="
  build "$flag" -f "$file"
  printf "Press enter to continue..."
  read ans
}

run --debug-lex "$1"
run --debug-var "$1"
run --debug-target "$1"
run --debug-recipe "$1"
run --debug-env "$1"
run --debug-cond "$1"
run --debug-include "$1"
run --debug-ast "$1"
run --debug-semantic "$1"
run --debug-eval "$1"
run --debug-plan "$1"
run "-v" "$1"
