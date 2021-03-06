#!/usr/bin/env sh

if [ -n "$1" ] && [ ${0:0:4} = "/bin" ]; then
  ROOT_DIR=$1/..
else
  ROOT_DIR="$( cd "$( dirname "$0" )" && pwd )/.."
fi

. $ROOT_DIR/scripts/common.sh

mockery -recursive=true -name=CentrifugoInterface -dir=${ROOT_DIR}/internal/ -output ${ROOT_DIR}/internal/mocks
mockery -recursive=true -name=DocumentGeneratorInterface -dir=${ROOT_DIR}/internal/ -output ${ROOT_DIR}/internal/mocks