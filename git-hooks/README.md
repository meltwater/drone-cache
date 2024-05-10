This document explains on how to install certain git hooks globally for all repositories in your machine.

Step 1: git clone https://github.com/drone-plugins/drone-meltwater-cache.git
Step 2: cd git-hooks
Step 3: Run install.sh

"install.sh" script will create .git_template in the user directory and will put the git hook and its dependent scripts in it. Along with the .git_template folder, it will add 2 sections "init" and "hooks boolean" in the .gitconfig file in the same user's root directory.
After running "install.sh" if you create/clone a new git repository then all the hooks will get install automatically for the git repository. In case of existing git repository copy the contents of ~/.git_template/hooks into the .git/hooks directory of existing git repository.