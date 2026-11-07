#!/bin/bash

# Simple and dumb script with no error-handling or edge cases yet.
# If no args, output to terminal

if [ $# -eq 0 ]; then
  echo "Usage: tunnel <host> <port>"
  exit 1
fi

# Assign args

host=${1%%:*}
port=${2##*:}