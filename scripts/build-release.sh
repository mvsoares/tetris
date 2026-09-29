#!/usr/bin/env bash
set -euo pipefail

VERSION="1.0.0"
COMMIT="$(git rev-parse --short HEAD 2>/dev/null || echo "none")"
BUILD_DATE="$(date -u +%Y-%m-%dT%H:%M:%SZ)"

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DIST_DIR="${ROOT_DIR}/dist"
NFPM_BIN="${HOME}/go/bin/nfpm"

if ! command -v "${NFPM_BIN}" >/dev/null 2>&1; then
    if command -v nfpm >/dev/null 2>&1; then
        NFPM_BIN="nfpm"
    else
        echo "⬇️  Instalando nfpm..."
        go install github.com/goreleaser/nfpm/v2/cmd/nfpm@v2.41.3
    fi
fi

echo "================================================================"
echo "🏗️  BUILDING TETRIS RELEASE v${VERSION} (${COMMIT})"
echo "================================================================"

rm -rf "${DIST_DIR}"
mkdir -p "${DIST_DIR}" "${DIST_DIR}/bin"

LDFLAGS="-s -w -X 'tetris/internal/version.Version=${VERSION}' -X 'tetris/internal/version.Commit=${COMMIT}' -X 'tetris/internal/version.Date=${BUILD_DATE}'"

PLATFORMS=(
    "linux/amd64"
    "linux/arm64"
    "darwin/amd64"
    "darwin/arm64"
    "windows/amd64"
)

for PLATFORM in "${PLATFORMS[@]}"; do
    GOOS="${PLATFORM%/*}"
    GOARCH="${PLATFORM#*/}"
    OUTPUT_NAME="tetris"
    if [ "${GOOS}" = "windows" ]; then
        OUTPUT_NAME="tetris.exe"
    fi
    TARGET_DIR="${DIST_DIR}/bin/${GOOS}_${GOARCH}"
    mkdir -p "${TARGET_DIR}"
    
    echo "🔨 Compilando para ${GOOS}/${GOARCH}..."
    CGO_ENABLED=0 GOOS="${GOOS}" GOARCH="${GOARCH}" go build \
        -ldflags "${LDFLAGS}" \
        -o "${TARGET_DIR}/${OUTPUT_NAME}" \
        "${ROOT_DIR}/cmd/tetris"
done

echo ""
echo "📦 Criando pacotes Debian (.deb) e RPM (.rpm)..."

cd "${ROOT_DIR}"

# Linux amd64
VERSION="${VERSION}" ARCH="amd64" BINARY_SRC="dist/bin/linux_amd64/tetris" \
    envsubst < packaging/nfpm.yaml | "${NFPM_BIN}" package -f - -p deb -t "${DIST_DIR}/tetris_${VERSION}_amd64.deb"
VERSION="${VERSION}" ARCH="amd64" BINARY_SRC="dist/bin/linux_amd64/tetris" \
    envsubst < packaging/nfpm.yaml | "${NFPM_BIN}" package -f - -p rpm -t "${DIST_DIR}/tetris-${VERSION}-1.x86_64.rpm"

# Linux arm64
VERSION="${VERSION}" ARCH="arm64" BINARY_SRC="dist/bin/linux_arm64/tetris" \
    envsubst < packaging/nfpm.yaml | "${NFPM_BIN}" package -f - -p deb -t "${DIST_DIR}/tetris_${VERSION}_arm64.deb"
VERSION="${VERSION}" ARCH="arm64" BINARY_SRC="dist/bin/linux_arm64/tetris" \
    envsubst < packaging/nfpm.yaml | "${NFPM_BIN}" package -f - -p rpm -t "${DIST_DIR}/tetris-${VERSION}-1.aarch64.rpm"

echo ""
echo "📦 Criando arquivos tar.gz e zip para distribuição..."

# Linux tarballs
cd "${DIST_DIR}"
tar -czf "tetris_${VERSION}_linux_amd64.tar.gz" -C "${DIST_DIR}/bin/linux_amd64" tetris -C "${ROOT_DIR}" models README.md LICENSE
tar -czf "tetris_${VERSION}_linux_arm64.tar.gz" -C "${DIST_DIR}/bin/linux_arm64" tetris -C "${ROOT_DIR}" models README.md LICENSE

# macOS tarballs
tar -czf "tetris_${VERSION}_darwin_amd64.tar.gz" -C "${DIST_DIR}/bin/darwin_amd64" tetris -C "${ROOT_DIR}" models README.md LICENSE
tar -czf "tetris_${VERSION}_darwin_arm64.tar.gz" -C "${DIST_DIR}/bin/darwin_arm64" tetris -C "${ROOT_DIR}" models README.md LICENSE

# Windows zip
(
    cd "${DIST_DIR}/bin/windows_amd64"
    zip -q -r "${DIST_DIR}/tetris_${VERSION}_windows_amd64.zip" tetris.exe
)
zip -q -u "${DIST_DIR}/tetris_${VERSION}_windows_amd64.zip" -j "${ROOT_DIR}/README.md" "${ROOT_DIR}/LICENSE"

echo ""
echo "🔒 Gerando checksums SHA-256..."
cd "${DIST_DIR}"
sha256sum tetris*.deb tetris*.rpm tetris*.tar.gz tetris*.zip > checksums.txt

# Also copy local release binary to repo root
cp "${DIST_DIR}/bin/linux_amd64/tetris" "${ROOT_DIR}/tetris"

echo ""
echo "================================================================"
echo "✅ RELEASE BUILD COMPLETED SUCCESSFULLY!"
echo "📁 Arquivos gerados em: ${DIST_DIR}"
echo "================================================================"
ls -lh "${DIST_DIR}"/*.{deb,rpm,tar.gz,zip,txt}
