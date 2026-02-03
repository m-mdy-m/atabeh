#!/usr/bin/env bash
# scripts/release.sh
# Usage:   ./scripts/release.sh [version]
#          ./scripts/release.sh 1.2.3        # non-interactive
#          ./scripts/release.sh              # interactive prompt

set -euo pipefail

RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'; NC='\033[0m'
die()  { echo -e "${RED}ERROR:${NC} $*" >&2; exit 1; }
info() { echo -e "${GREEN}ℹ${NC}  $*"; }
warn() { echo -e "${YELLOW}⚠${NC}  $*"; }

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$REPO_ROOT"
VERSION="${1:-}"
if [ -z "$VERSION" ]; then
  echo -n "Release version (without 'v'): "
  read -r VERSION
fi
VERSION="${VERSION#v}"
if ! [[ "$VERSION" =~ ^[0-9]+\.[0-9]+\.[0-9]+([-][a-zA-Z0-9.]+)?$ ]]; then
  die "Version must look like 1.2.3 or 1.2.3-rc1.  Got: '$VERSION'"
fi

TAG="v${VERSION}"
info "Target version: ${TAG}"
if [ -n "$(git status --porcelain)" ]; then
  warn "Working tree is dirty. Showing diff:"
  git status -s
  echo -n "Continue anyway? [y/N] "
  read -r yn
  [[ "$yn" =~ ^[Yy] ]] || die "Aborted — commit or stash first."
fi
if git rev-parse "$TAG" >/dev/null 2>&1; then
  die "Tag $TAG already exists."
fi
info "Running tests …"
make test || die "Tests failed. Fix them before releasing."
info "Cross-compiling …"
export VERSION GIT_COMMIT="$(git rev-parse --short HEAD)" BUILD_DATE="$(date -u '+%Y-%m-%dT%H:%M:%SZ')"

LDFLAGS="-X 'github.com/m-mdy-m/atabeh/cmd/cli.Version=${VERSION}' \
         -X 'github.com/m-mdy-m/atabeh/cmd/cli.GitCommit=${GIT_COMMIT}' \
         -X 'github.com/m-mdy-m/atabeh/cmd/cli.BuildDate=${BUILD_DATE}' \
         -s -w"

mkdir -p bin/release

declare -A TARGETS=(
  ["linux-amd64"]="linux:amd64"
  ["linux-arm64"]="linux:arm64"
  ["darwin-arm64"]="darwin:arm64"
  ["windows-amd64"]="windows:amd64"
)

for label in "${!TARGETS[@]}"; do
  IFS=':' read -r goos goarch <<< "${TARGETS[$label]}"
  out="bin/release/atabeh-${label}"
  [[ "$goos" == "windows" ]] && out+=".exe"

  info "  → $out"
  GOOS="$goos" GOARCH="$goarch" CGO_ENABLED=0 \
    go build -ldflags "$LDFLAGS" -o "$out" ./cmd/atabeh
done

info "Binaries written to bin/release/"
info "Creating tag $TAG …"
git tag -a "$TAG" -m "Release $TAG"
echo ""
echo -e "${GREEN}✓ Release $TAG is ready.${NC}"
echo ""
echo "  To publish:"
echo "    git push origin main"
echo "    git push origin $TAG"
echo ""
echo "  Binaries:"
ls -lh bin/release/