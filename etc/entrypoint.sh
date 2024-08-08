#!/bin/bash

if [ -z "$1" ]; then
    echo "No command specified"
    exit 1
fi

# Create the /alloc/logs folder if it doesn't exist
install -d -m 0755 /alloc/logs || exit 1

# Create the persistent data folder if it doesn't exist
install -d -m 0755 /data || exit 1

# Run the command
set -e
umask 022
exec "$@"
