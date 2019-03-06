#!/bin/sh

# Pre-run commands
if [ "${DEBUG-}" = "true" ] || [ "${PLUGIN_DEBUG-}" = "true" ]; then
    env
fi

# Hand off to the CMD
exec su-exec appuser "$@"
