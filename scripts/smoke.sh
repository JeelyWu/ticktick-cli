#!/usr/bin/env bash
set -euo pipefail

expect_contains() {
  local output="$1"
  local expected="$2"

  if [[ "${output}" != *"${expected}"* ]]; then
    echo "expected output to contain: ${expected}" >&2
    exit 1
  fi
}

./bin/tick version >/dev/null
expect_contains "$(./bin/tick --help)" "TickTick CLI"
expect_contains "$(./bin/tick auth --help)" "login"
expect_contains "$(./bin/tick project --help)" "ls"
expect_contains "$(./bin/tick task --help)" "add"
expect_contains "$(./bin/tick quick add --help)" "Create a task"
expect_contains "$(./bin/tick config --help)" "set"
