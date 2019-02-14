#!/bin/sh

# Enable strict mode, see: http://redsymbol.net/articles/unofficial-bash-strict-mode/
set -euo pipefail
IFS=$'\n\t'

echo 'Setting up test environemnt...'
(set -x; docker-compose up -d)

echo 'Testing...'
(set -x; go test ./...)

echo 'Done.'
exit 0
