#!/usr/bin/env bash
### every exit != 0 fails the script
set -e

# Enable verbose mode if DEBUG is set
verbose=""
if [[ -n $DEBUG ]]; then
    verbose="-v"
fi

# Loop through the provided directories
for var in "$@"; do
    echo "Fixing permissions for: $var"

    # Make all .sh and .desktop files executable
    find "$var" -type f -name '*.sh' -exec chmod $verbose a+x {} +
    find "$var" -type f -name '*.desktop' -exec chmod $verbose a+x {} +

    # Update permissions for all files and directories
    chgrp -R 0 "$var"
    chmod -R $verbose a+rw "$var"
    find "$var" -type d -exec chmod $verbose a+x {} +
done
