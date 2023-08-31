#!/bin/bash

# Function to search directories recursively
search_dirs() {
    for dir in "$1"/*; do
        if [ -d "${dir}" ]; then
            if [ -f "${dir}/Acornfile" ]; then
                echo "Running acorn fmt in ${dir}"
                (cd "${dir}" && acorn fmt)
            fi
            # Recursive call
            search_dirs "${dir}"
        fi
    done
}

# Start the search from the current directory
search_dirs "."

