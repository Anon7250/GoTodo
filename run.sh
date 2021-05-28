#!/bin/bash
PWD="$(pwd)"
if [ "$(expr substr $(uname -s) 1 6)" == "CYGWIN" ]; then
  PWD="$(cygpath -m -w "$PWD")"
fi
docker run -it --mount type=bind,source="$PWD",target=/work -w /work --rm golang "$@"
