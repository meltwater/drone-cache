#!/bin/bash

#Helper script to be used as a pre-commit hook.

echo "This hook checks for any secrets getting pushed as part of commit. If you feel that scan is false positive. \
Then add the exclusion in .gitleaksignore file. For more info visit: https://github.com/zricethezav/gitleaks"

GIT_LEAKS=$(git config --bool hook.pre-push.gitleaks)

echo "INFO: Scanning Commits information for any GIT LEAKS"
gitleaks detect -s ./ --log-level=debug --log-opts=-1 -v
STATUS=$?
if [ $STATUS != 0  ]; then
    echo "WARNING: GIT LEAKS has detected sensitive information in your changes. Please remove them or add them (IF NON-SENSITIVE) in .gitleaksignore file."
    exit $STATUS
else
    exit 0
fi