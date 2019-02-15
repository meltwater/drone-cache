#!/bin/sh

# Pre-run commands
if [[ $(echo $DEBUG) == "true" || $(echo $PLUGIN_DEBUG) == "true" ]]; then
    env
fi

# Hand off to the CMD
exec "$@"
