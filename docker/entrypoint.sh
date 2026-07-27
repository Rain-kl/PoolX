#!/bin/sh
set -eu

umask 077

if [ ! -f "${FOAM_CONFIG_SOURCE}" ]; then
  echo "missing config: ${FOAM_CONFIG_SOURCE}" >&2
  echo "mount config.yaml to /run/foam/config.yaml" >&2
  exit 1
fi

cp "${FOAM_CONFIG_SOURCE}" /app/config.yaml
chown foam:foam /app/config.yaml
chmod 0600 /app/config.yaml

exec su-exec foam:foam "$@"

