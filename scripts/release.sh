#!/usr/bin/env bash
set -euo pipefail

# scripts/release.sh - Automates building release binaries, packages, git tagging and GitHub release
VERSION="1.0.0"
TAG="v${VERSION}"

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT_DIR}"

echo "================================================================"
echo "🚀 INICIANDO PROCESSO DE RELEASE DO TETRIS ${TAG}"
echo "================================================================"

# 1. Build all binaries and packages
echo "📦 Executando build de todos os binários e pacotes..."
./scripts/build-release.sh

# 2. Check git status
if git rev-parse "${TAG}" >/dev/null 2>&1; then
    echo "⚠️  A tag ${TAG} já existe localmente."
else
    echo "🏷️  Criando tag git ${TAG}..."
    git tag -a "${TAG}" -m "Release ${TAG} - Idiomas (en, es, pt-br, fr, ita), pacotes Debian, RPM e macOS"
    echo "✅ Tag ${TAG} criada com sucesso!"
fi

echo ""
echo "================================================================"
echo "🎉 RELEASE PRONTO PARA ATUALIZAÇÃO NO GIT!"
echo "================================================================"
echo ""
echo "Para enviar para o GitHub, execute:"
echo "   git push origin main --tags"
echo ""
if command -v gh >/dev/null 2>&1; then
    echo "Para publicar o Release no GitHub com todos os binários e pacotes:"
    echo "   gh release create ${TAG} dist/tetris* dist/checksums.txt --title \"Tetris ${TAG}\" --notes \"Versão 1.0.0 oficial com suporte multilíngue (en, es, pt-br, fr, it) e pacotes de instalação para Debian (.deb), RedHat/Fedora (.rpm) e macOS/Linux (.tar.gz).\""
fi
echo "================================================================"
