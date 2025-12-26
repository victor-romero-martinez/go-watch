#!/usr/bin/env bash

set -e

APP_NAME="go-watch"
VERSION="${1:-dev}"
DIST_DIR="dist"
OUT_DIR="release"

mkdir -p "$OUT_DIR"

echo "📦 Packaging $APP_NAME version $VERSION"

for BIN in "$DIST_DIR"/*; do
  FILE=$(basename "$BIN")

  # Windows
  if [[ "$FILE" == *.exe ]]; then
    PLATFORM=${FILE#"$APP_NAME-"}
    PLATFORM=${PLATFORM%.exe}

    ZIP_NAME="$APP_NAME-$VERSION-$PLATFORM.zip"

    echo "→ $ZIP_NAME"
    zip -j "$OUT_DIR/$ZIP_NAME" "$BIN"

  # Linux / macOS
  else
    PLATFORM=${FILE#"$APP_NAME-"}
    TAR_NAME="$APP_NAME-$VERSION-$PLATFORM.tar.gz"

    echo "→ $TAR_NAME"
    tar -czf "$OUT_DIR/$TAR_NAME" -C "$DIST_DIR" "$FILE"
  fi
done

echo "✅ Packages generated in ./$OUT_DIR"
