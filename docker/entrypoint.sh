#!/bin/sh
set -eu

umask 077

# Copy config file if it exists; otherwise rely on FOAM_* environment variables.
if [ -f "${FOAM_CONFIG_SOURCE:-}" ]; then
  cp "${FOAM_CONFIG_SOURCE}" /app/config.yaml
  chown foam:foam /app/config.yaml
  chmod 0600 /app/config.yaml
fi

exec su-exec foam:foam "$@"
