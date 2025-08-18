#!/bin/sh

set -e

# default command, expects 'go42x' executable to be available in $PATH
if [ "$1" = 'app' ]; then
  exec go42x "${@:2}"
fi

# if arbitrary command was passed, execute it instead of default one
exec "$@"
