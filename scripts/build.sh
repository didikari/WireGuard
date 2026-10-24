#!/usr/bin/env bash
set -euo pipefail

VERSION=${VERSION:-"1.0.0"}
COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "release")
BUILD_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

PLATFORMS=(
    "darwin/amd64"
    "darwin/arm64"
    "linux/amd64"
    "linux/arm64"
)

OUTPUT_DIR="dist"
rm -rf "${OUTPUT_DIR}"
mkdir -p "${OUTPUT_DIR}"

echo "🔨 Building ezwg v${VERSION} (${COMMIT})..."

LDFLAGS="-s -w -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${BUILD_DATE}"

for PLATFORM in "${PLATFORMS[@]}"; do
    GOOS="${PLATFORM%/*}"
    GOARCH="${PLATFORM#*/}"
    BIN_NAME="ezwg"
    ARCHIVE_NAME="ezwg_${VERSION}_${GOOS}_${GOARCH}"
    TARGET_DIR="${OUTPUT_DIR}/${ARCHIVE_NAME}"

    mkdir -p "${TARGET_DIR}"
    echo "  -> Compiling for ${GOOS}/${GOARCH}..."

    CGO_ENABLED=0 GOOS="${GOOS}" GOARCH="${GOARCH}" go build \
        -ldflags "${LDFLAGS}" \
        -trimpath \
        -o "${TARGET_DIR}/${BIN_NAME}" \
        ./cmd/ezwg

    # Include README and install helper in archive
    cp README.md "${TARGET_DIR}/" 2>/dev/null || true

    # Create tar.gz
    tar -czf "${OUTPUT_DIR}/${ARCHIVE_NAME}.tar.gz" -C "${OUTPUT_DIR}" "${ARCHIVE_NAME}"
    rm -rf "${TARGET_DIR}"
done

# Generate SHA256 Checksums
echo "🔐 Generating checksums..."
cd "${OUTPUT_DIR}"
shasum -a 256 *.tar.gz > checksums.txt
cd ..

echo "✅ Build complete! Release archives located in ${OUTPUT_DIR}/:"
ls -lh "${OUTPUT_DIR}"
