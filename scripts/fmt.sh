#!/bin/bash

# Function to check and run acorn fmt in the directory if a matching file is found
check_and_run() {
    local dir="$1"
    shopt -s nullglob
    for acornfile in "${dir}"/*.Acornfile "${dir}"/Acornfile.* "${dir}/Acornfile"; do
        if [ -f "${acornfile}" ]; then
            echo "Running acorn fmt on ${acornfile}"
            (cd "${dir}" && acorn fmt "$(basename "${acornfile}")")
        fi
    done
    shopt -u nullglob
}

# Function to search directories recursively
search_dirs() {
    check_and_run "$1"
    for dir in "$1"/*; do
        if [ -d "${dir}" ]; then
            # Recursive call
            search_dirs "${dir}"
        fi
    done
}

# Start the search from the current directory
search_dirs "."
