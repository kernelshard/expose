#!/bin/sh
# Expose installation script
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/kernelshard/expose/main/install.sh | sh
#
# Customizations:
#   VERSION=v0.4.1 curl -fsSL ... | sh             # Install specific version
#   BINDIR=$HOME/bin curl -fsSL ... | sh           # Install into custom directory
#   USE_SUDO=1 curl -fsSL ... | sh                 # Force install to /usr/local/bin with sudo
#
set -eu

REPO="${EXPOSE_REPO:-kernelshard/expose}"
BINARY_NAME="expose"

# Setup color formatting if output is a terminal
if [ -t 1 ] && [ -z "${NO_COLOR:-}" ]; then
  RED='\033[0;31m'
  GREEN='\033[0;32m'
  BLUE='\033[0;34m'
  YELLOW='\033[0;33m'
  BOLD='\033[1m'
  RESET='\033[0m'
else
  RED=''
  GREEN=''
  BLUE=''
  YELLOW=''
  BOLD=''
  RESET=''
fi

log_info() {
  printf "${BLUE}==>${RESET} %s\n" "$*"
}

log_success() {
  printf "${GREEN}==>${RESET} ${BOLD}%s${RESET}\n" "$*"
}

log_warn() {
  printf "${YELLOW}warning:${RESET} %s\n" "$*"
}

log_err() {
  printf "${RED}error:${RESET} %s\n" "$*" >&2
}

detect_os() {
  os="$(uname -s)"
  case "$os" in
    Darwin) echo "Darwin" ;;
    Linux)  echo "Linux" ;;
    *)
      log_err "Unsupported operating system: $os"
      log_err "Expose shell installation supports macOS (Darwin) and Linux."
      log_err "Pre-built binaries for Windows are available at: https://github.com/${REPO}/releases"
      exit 1
      ;;
  esac
}

detect_arch() {
  arch="$(uname -m)"
  case "$arch" in
    x86_64|amd64) echo "x86_64" ;;
    arm64|aarch64) echo "arm64" ;;
    *)
      log_err "Unsupported architecture: $arch"
      log_err "Supported architectures are x86_64 and arm64."
      exit 1
      ;;
  esac
}

download_file() {
  url="$1"
  dest="$2"
  if command -v curl >/dev/null 2>&1; then
    curl -fsSL "$url" -o "$dest"
  elif command -v wget >/dev/null 2>&1; then
    wget -qO "$dest" "$url"
  else
    log_err "Neither curl nor wget was found. Please install either curl or wget to continue."
    exit 1
  fi
}

resolve_version() {
  if [ -n "${VERSION:-}" ]; then
    case "$VERSION" in
      v*) echo "$VERSION" ;;
      *)  echo "v$VERSION" ;;
    esac
    return
  fi

  # Attempt 1: Query redirect URL on GitHub Releases (no API rate limiting)
  latest_url=$(curl -sIL -o /dev/null -w '%{url_effective}' "https://github.com/${REPO}/releases/latest" 2>/dev/null || true)
  tag=$(basename "$latest_url" 2>/dev/null || true)
  if [ -n "$tag" ] && [ "$tag" != "latest" ] && [ "$tag" != "releases" ]; then
    echo "$tag"
    return
  fi

  # Attempt 2: Fall back to GitHub Releases API
  tag=""
  if command -v curl >/dev/null 2>&1; then
    tag=$(curl -sSL "https://api.github.com/repos/${REPO}/releases/latest" 2>/dev/null | grep '"tag_name":' | sed -E 's/.*"tag_name":[[:space:]]*"([^"]+)".*/\1/' || true)
  elif command -v wget >/dev/null 2>&1; then
    tag=$(wget -qO- "https://api.github.com/repos/${REPO}/releases/latest" 2>/dev/null | grep '"tag_name":' | sed -E 's/.*"tag_name":[[:space:]]*"([^"]+)".*/\1/' || true)
  fi

  if [ -n "${tag:-}" ]; then
    echo "$tag"
    return
  fi

  log_err "Failed to determine latest release version for ${REPO}."
  log_err "You can specify a version explicitly using: VERSION=v0.4.1 curl -fsSL ... | sh"
  exit 1
}

compute_sha256() {
  target="$1"
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$target" | awk '{print $1}'
  elif command -v shasum >/dev/null 2>&1; then
    shasum -a 256 "$target" | awk '{print $1}'
  else
    echo ""
  fi
}

choose_bindir() {
  if [ -n "${BINDIR:-}" ]; then
    echo "$BINDIR"
    return
  fi

  # If user explicitly requested sudo elevation
  if [ "${USE_SUDO:-0}" = "1" ]; then
    echo "/usr/local/bin"
    return
  fi

  # If /usr/local/bin is directly writable or user is root
  if [ -w "/usr/local/bin" ] || [ "$(id -u)" = "0" ]; then
    echo "/usr/local/bin"
    return
  fi

  # If sudo credentials are valid without password prompt
  if command -v sudo >/dev/null 2>&1 && sudo -n true 2>/dev/null; then
    echo "/usr/local/bin"
    return
  fi

  # Default non-root fallback without password prompt
  echo "${HOME}/.local/bin"
}

install_binary() {
  src="$1"
  target_dir="$2"
  target_file="${target_dir}/${BINARY_NAME}"

  # Attempt direct installation without elevation
  if mkdir -p "$target_dir" 2>/dev/null && [ -w "$target_dir" ]; then
    cp "$src" "$target_file"
    chmod 755 "$target_file"
    return 0
  fi

  # Elevate permissions with sudo
  if command -v sudo >/dev/null 2>&1; then
    log_info "Elevating permissions with sudo to install into ${target_dir}..."
    sudo mkdir -p "$target_dir"
    sudo cp "$src" "$target_file"
    sudo chmod 755 "$target_file"
    return 0
  fi

  log_err "Cannot install to ${target_dir}: insufficient permissions and sudo is not installed."
  return 1
}

main() {
  OS="$(detect_os)"
  ARCH="$(detect_arch)"
  TAG="$(resolve_version)"
  TARGET_DIR="$(choose_bindir)"

  log_info "Installing Expose ${TAG} (${OS}/${ARCH})..."

  ARCHIVE="expose_${OS}_${ARCH}.tar.gz"
  BASE_URL="https://github.com/${REPO}/releases/download/${TAG}"
  ARCHIVE_URL="${BASE_URL}/${ARCHIVE}"
  CHECKSUM_URL="${BASE_URL}/checksums.txt"

  TMP_DIR=$(mktemp -d 2>/dev/null || mktemp -d -t 'expose-install')
  cleanup() {
    rm -rf "$TMP_DIR"
  }
  trap cleanup EXIT INT TERM

  log_info "Downloading ${ARCHIVE}..."
  download_file "$ARCHIVE_URL" "${TMP_DIR}/${ARCHIVE}"

  log_info "Verifying SHA-256 checksum..."
  if download_file "$CHECKSUM_URL" "${TMP_DIR}/checksums.txt" 2>/dev/null; then
    expected_sum=$(grep "[[:space:]]${ARCHIVE}\$" "${TMP_DIR}/checksums.txt" 2>/dev/null | awk '{print $1}' || true)
    if [ -n "$expected_sum" ]; then
      actual_sum=$(compute_sha256 "${TMP_DIR}/${ARCHIVE}")
      if [ -n "$actual_sum" ]; then
        if [ "$expected_sum" != "$actual_sum" ]; then
          log_err "Checksum verification failed!"
          log_err "Expected: ${expected_sum}"
          log_err "Actual:   ${actual_sum}"
          exit 1
        fi
        log_info "Checksum verified: ${actual_sum}"
      else
        log_warn "Neither sha256sum nor shasum found; skipping checksum verification."
      fi
    else
      log_warn "Archive ${ARCHIVE} not found in checksums.txt; skipping checksum verification."
    fi
  else
    log_warn "Could not download checksums.txt; skipping checksum verification."
  fi

  log_info "Extracting archive..."
  tar -xzf "${TMP_DIR}/${ARCHIVE}" -C "$TMP_DIR"

  if [ ! -f "${TMP_DIR}/${BINARY_NAME}" ]; then
    log_err "Binary '${BINARY_NAME}' not found inside archive."
    exit 1
  fi

  log_info "Installing binary to ${TARGET_DIR}..."
  if ! install_binary "${TMP_DIR}/${BINARY_NAME}" "$TARGET_DIR"; then
    log_err "Installation failed."
    exit 1
  fi

  # Verify executable
  INSTALLED_BIN="${TARGET_DIR}/${BINARY_NAME}"
  if [ -x "$INSTALLED_BIN" ]; then
    version_out=$("$INSTALLED_BIN" --version 2>/dev/null || echo "$TAG")
    log_success "${version_out} successfully installed at ${INSTALLED_BIN}!"
  else
    log_err "Installed binary at ${INSTALLED_BIN} is not executable."
    exit 1
  fi

  # Check PATH
  case ":$PATH:" in
    *":$TARGET_DIR:"*) ;;
    *)
      printf "\n"
      log_warn "${TARGET_DIR} is not in your \$PATH."
      log_warn "To use 'expose', add this line to your shell profile (~/.zshrc or ~/.bashrc):"
      printf "\n    export PATH=\"%s:\$PATH\"\n\n" "$TARGET_DIR"
      ;;
  esac

  printf "\nRun 'expose --help' to get started!\n"
}

main "$@"
