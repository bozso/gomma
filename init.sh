root="$(realpath $1)"

export PATH="${PATH}:${root}/bin"
export PYTHONPATH="${PYTHONPATH}:${root}"
