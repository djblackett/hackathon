#!/usr/bin/env bash
# ----------------------------------------------------------
# filerenamer.sh  -  thin Docker wrapper for Filename‑Fixer
#
# Usage:
#   ./filerenamer.sh  <input_dir>  <output_dir>  [extra flags]
#
# Optional:
#   Place a .env file next to this script, e.g.
#       OPENAI_API_KEY=sk-...
#       RELAY_URL=https://filerenamer-relay.fly.dev/suggest
#   Those vars will be loaded and passed into the container.
# ----------------------------------------------------------

set -euo pipefail

# 1. Resolve paths
IN=$(realpath "$1")
# OUT=$(realpath "$2")
shift 2

# 2. Find .env (same dir as script)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ENV_FILE="$SCRIPT_DIR/.env"

# 3. If .env exists, make a temporary --env-file compatible copy
ENV_ARG=""
if [[ -f "$ENV_FILE" ]]; then
  TMP_ENV=$(mktemp)
  # strip comments / blank lines
  grep -v '^\s*#' "$ENV_FILE" | grep -E '.+=' > "$TMP_ENV"
  ENV_ARG="--env-file $TMP_ENV"
fi

# 4. Run container
docker run --rm \
  -v "$IN":/in \
  -v "$OUT":/out \
  $ENV_ARG \
  djblackett/filename-fixer:v0.3 \
  --dir /in "$@"

#   docker run --rm \
#   --env-file "$ENV_FILE" \
#   -v "$IN":/in \
#   -v "$OUT":/out \
#   djblackett/filename-fixer:v0.1 \
#   --dir /in --out /out "$@"

# 5. Cleanup tmp env file
[[ -n "${TMP_ENV:-}" ]] && rm "$TMP_ENV"
