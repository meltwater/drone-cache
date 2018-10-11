#!/bin/sh

# Enable strict mode, see: http://redsymbol.net/articles/unofficial-bash-strict-mode/
set -euo pipefail
IFS=$'\n\t'

echo 'Setting up pre-commit hooks...'
(set -x; rm -f .git/hooks/pre-commit \
    && ln -s `pwd`/scripts/ensure_formatted.sh .git/hooks/pre-commit)

echo 'Done.'
exit 0
