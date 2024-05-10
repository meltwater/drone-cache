#!/usr/bin/env bash

#Function to check if package is installed or not
#args: $1: Name of the Package
function check_package_installed() {
  LOCAL_PACKAGE_NAME=$1
  echo "Checking if $LOCAL_PACKAGE_NAME is installed or not..."
  brew list $LOCAL_PACKAGE_NAME
  if [ "$?" -eq 1 ];then
    echo "Installing $LOCAL_PACKAGE_NAME package..."
    brew install $LOCAL_PACKAGE_NAME
  fi
}

function create_git_template() {
    cd $BASEDIR
    mkdir -p ~/.git_template/hooks
    git config --global init.templatedir ${GIT_TEMPLATE}
    git config --global --add $GIT_LEAKS true
    git config --global --add $GIT_LEAKS_PRE_COMMIT true
    find hooks/ -type f -exec cp "{}" ~/.git_template/hooks \;
    #cp -f hooks/* ~/.git_template/hooks
    cat ~/.gitconfig
}

GIT_TEMPLATE="~/.git_template"
GIT_LEAKS=hook.pre-push.gitleaks
GIT_LEAKS_PRE_COMMIT=hook.pre-commit.gitleaks

pushd `dirname $0` && BASEDIR=$(pwd -L) && popd

echo This script will install hooks that run scripts that could be updated without notice.

while true; do
    read -p "Do you wish to install these hooks?" yn
    case $yn in
        [Yy]* ) check_package_installed "gitleaks";
                break;;
        [Nn]* ) exit;;
        * ) echo "Please answer yes or no.";;
    esac
done

create_git_template