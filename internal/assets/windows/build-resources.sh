#!/bin/bash

# This script is used to build the Windows resources for the Windows installer.

# Check if windres is available
if ! command -v x86_64-w64-mingw32-windres &> /dev/null; then
    echo "x86_64-w64-mingw32-windres not found. Installing mingw-w64..."
    sudo apt-get update
    sudo apt-get -y install mingw-w64
else
    echo "mingw-w64 is already installed."
fi

# Check if output file exists
if [ -f "rsrc_windows_amd64.syso" ]; then
    read -r -p "File rsrc_windows_amd64.syso already exists. Do you want to overwrite it? (y/N) " response
    case "$response" in
        [yY][eE][sS]|[yY])
            # Build the Windows resources
            x86_64-w64-mingw32-windres -i app.rc -o rsrc_windows_amd64.syso --target=pe-x86-64
            echo "File has been overwritten."
            ;;
        *)
            echo "Operation cancelled."
            exit 0
            ;;
    esac
else
    # Build the Windows resources
    x86_64-w64-mingw32-windres -i app.rc -o rsrc_windows_amd64.syso --target=pe-x86-64
    echo "File has been created."
fi