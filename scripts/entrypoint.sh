#!/bin/sh

# Pre-run commands
if [ ${DEBUG} == "true" ]; then
    env
fi

# Hand off to the CMD
exec "$@"
