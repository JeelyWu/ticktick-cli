#!/usr/bin/env bash
set -euo pipefail

./bin/tick version >/dev/null
./bin/tick --help | grep -q "TickTick CLI"
./bin/tick auth --help | grep -q "login"
./bin/tick project --help | grep -q "ls"
./bin/tick task --help | grep -q "add"
./bin/tick quick add --help | grep -q "Create a task"
./bin/tick config --help | grep -q "set"
