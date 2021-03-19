#!/bin/bash -e

ROOT="$(dirname "$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )")"

$ROOT/deploy/up.sh
