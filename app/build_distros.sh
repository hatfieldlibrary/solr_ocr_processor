#!/bin/bash

PLATFORMS="darwin/amd64"
PLATFORMS="$PLATFORMS windows/amd64"
PLATFORMS="$PLATFORMS linux/amd64"
PLATFORMS="$PLATFORMS freebsd/amd64"
PLATFORMS="$PLATFORMS netbsd/amd64"
PLATFORMS="$PLATFORMS openbsd/amd64"
PLATFORMS="$PLATFORMS solaris/amd64"

type setopt >/dev/null 2>&1

SCRIPT_NAME=`basename "$0"`
FAILURES=""
PREFIX="processor"
BIN_PATH="bin"
DISTROS="assets/build"

for PLATFORM in $PLATFORMS; do
  GOOS=${PLATFORM%/*}
  GOARCH=${PLATFORM#*/}
  OUTPUT_DIR="${PREFIX}-${GOOS}-${GOARCH}"
  BIN_FILENAME="${PREFIX}-${GOOS}-${GOARCH}"
  CMD="mkdir "${BIN_PATH}/${OUTPUT_DIR}
  echo "${CMD}"
  if [[ "${GOOS}" == "windows" ]]; then BIN_FILENAME="${BIN_FILENAME}.exe"; fi
  CMD="GOOS=${GOOS} GOARCH=${GOARCH} go build -o ${BIN_PATH}/${OUTPUT_DIR}/${BIN_FILENAME} main.go
    && cp -n ${DISTROS}/config.yml ${BIN_PATH}/${OUTPUT_DIR}/
    && tar -cf ${BIN_PATH}/${OUTPUT_DIR}.tar ${BIN_PATH}/${OUTPUT_DIR}/${BIN_FILENAME}"
  echo "${CMD}"
  eval $CMD || FAILURES="${FAILURES} ${PLATFORM}"
done

if [[ "${FAILURES}" != "" ]]; then
  echo ""
  echo "${SCRIPT_NAME} failed on: ${FAILURES}"
  exit 1
fi
