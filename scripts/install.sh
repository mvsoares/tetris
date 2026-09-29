#!/usr/bin/env bash
set -euo pipefail

# Tetris Installer Script
# Usage: curl -fsSL https://raw.githubusercontent.com/mvsoares/tetris/main/scripts/install.sh | bash

VERSION="1.0.0"
REPO="mvsoares/tetris"
INSTALL_DIR="/usr/local/bin"

echo "🎮 Tetris Installer v${VERSION}"
echo "================================="

OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

case "${ARCH}" in
    x86_64|amd64)
        PKG_ARCH="amd64"
        RPM_ARCH="x86_64"
        ;;
    aarch64|arm64)
        PKG_ARCH="arm64"
        RPM_ARCH="aarch64"
        ;;
    *)
        echo "❌ Arquitetura não suportada: ${ARCH}"
        exit 1
        ;;
esac

TMP_DIR="$(mktemp -d)"
trap 'rm -rf "${TMP_DIR}"' EXIT

cd "${TMP_DIR}"

if [ "${OS}" = "darwin" ]; then
    echo "🍎 Detectado macOS (${ARCH})..."
    TAR_NAME="tetris_${VERSION}_darwin_${PKG_ARCH}.tar.gz"
    URL="https://github.com/${REPO}/releases/download/v${VERSION}/${TAR_NAME}"
    
    echo "⬇️  Baixando ${TAR_NAME}..."
    curl -fsSL -o "${TAR_NAME}" "${URL}"
    tar -xzf "${TAR_NAME}"
    
    echo "📦 Instalando em ${INSTALL_DIR}..."
    if [ -w "${INSTALL_DIR}" ]; then
        mv tetris "${INSTALL_DIR}/tetris"
        chmod +x "${INSTALL_DIR}/tetris"
    else
        sudo mv tetris "${INSTALL_DIR}/tetris"
        sudo chmod +x "${INSTALL_DIR}/tetris"
    fi

elif [ "${OS}" = "linux" ]; then
    if command -v apt-get >/dev/null 2>&1 && [ -f /etc/debian_version ]; then
        echo "🐧 Detectado Debian/Ubuntu (${PKG_ARCH})..."
        DEB_NAME="tetris_${VERSION}_${PKG_ARCH}.deb"
        URL="https://github.com/${REPO}/releases/download/v${VERSION}/${DEB_NAME}"
        
        echo "⬇️  Baixando ${DEB_NAME}..."
        curl -fsSL -o "${DEB_NAME}" "${URL}"
        
        echo "📦 Instalando via dpkg..."
        sudo dpkg -i "${DEB_NAME}" || sudo apt-get install -f -y

    elif (command -v dnf >/dev/null 2>&1 || command -v rpm >/dev/null 2>&1) && [ -f /etc/redhat-release -o -f /etc/fedora-release ]; then
        echo "🐧 Detectado Fedora/RHEL/CentOS (${RPM_ARCH})..."
        RPM_NAME="tetris-${VERSION}-1.${RPM_ARCH}.rpm"
        URL="https://github.com/${REPO}/releases/download/v${VERSION}/${RPM_NAME}"
        
        echo "⬇️  Baixando ${RPM_NAME}..."
        curl -fsSL -o "${RPM_NAME}" "${URL}"
        
        echo "📦 Instalando via rpm/dnf..."
        if command -v dnf >/dev/null 2>&1; then
            sudo dnf install -y "${RPM_NAME}"
        else
            sudo rpm -Uvh "${RPM_NAME}"
        fi

    else
        echo "🐧 Detectado Linux genérico (${PKG_ARCH})..."
        TAR_NAME="tetris_${VERSION}_linux_${PKG_ARCH}.tar.gz"
        URL="https://github.com/${REPO}/releases/download/v${VERSION}/${TAR_NAME}"
        
        echo "⬇️  Baixando ${TAR_NAME}..."
        curl -fsSL -o "${TAR_NAME}" "${URL}"
        tar -xzf "${TAR_NAME}"
        
        echo "📦 Instalando em ${INSTALL_DIR}..."
        if [ -w "${INSTALL_DIR}" ]; then
            mv tetris "${INSTALL_DIR}/tetris"
            chmod +x "${INSTALL_DIR}/tetris"
        else
            sudo mv tetris "${INSTALL_DIR}/tetris"
            sudo chmod +x "${INSTALL_DIR}/tetris"
        fi
    fi
else
    echo "❌ Sistema operacional não suportado: ${OS}"
    exit 1
fi

echo ""
echo "✅ Tetris v${VERSION} instalado com sucesso!"
echo "🕹️  Para jogar, digite: tetris"
