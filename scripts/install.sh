#!/usr/bin/env bash
set -euo pipefail

REPO="${REPO:-JeelyWu/ticktick-cli}"
BINARY="${BINARY:-tick}"

usage() {
  cat <<'EOF'
Install tick from GitHub Releases.

Environment variables:
  VERSION      Release tag to install, for example v0.1.0. Defaults to latest.
  INSTALL_DIR  Destination directory. Defaults to /usr/local/bin when writable,
               otherwise $HOME/.local/bin.
  REPO         GitHub repository in owner/name form. Defaults to JeelyWu/ticktick-cli.
EOF
}

need_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "missing required command: $1" >&2
    exit 1
  fi
}

default_install_dir() {
  if [ -d /usr/local/bin ] && [ -w /usr/local/bin ]; then
    printf '%s\n' /usr/local/bin
    return
  fi
  printf '%s\n' "${HOME}/.local/bin"
}

normalize_os() {
  case "$(uname -s)" in
    Darwin) printf '%s\n' darwin ;;
    Linux) printf '%s\n' linux ;;
    *)
      echo "unsupported operating system: $(uname -s)" >&2
      exit 1
      ;;
  esac
}

normalize_arch() {
  case "$(uname -m)" in
    x86_64) printf '%s\n' amd64 ;;
    arm64 | aarch64) printf '%s\n' arm64 ;;
    *)
      echo "unsupported architecture: $(uname -m)" >&2
      exit 1
      ;;
  esac
}

resolve_version() {
  if [ -n "${VERSION:-}" ]; then
    case "${VERSION}" in
      v*) printf '%s\n' "${VERSION}" ;;
      *) printf 'v%s\n' "${VERSION}" ;;
    esac
    return
  fi

  local api_url
  local tag

  api_url="https://api.github.com/repos/${REPO}/releases/latest"
  tag="$(
    curl -fsSL "${api_url}" |
      sed -n 's/.*"tag_name":[[:space:]]*"\([^"]*\)".*/\1/p' |
      head -n 1
  )"

  if [ -z "${tag}" ]; then
    echo "failed to resolve the latest release tag from ${api_url}" >&2
    exit 1
  fi

  printf '%s\n' "${tag}"
}

verify_archive() {
  local checksum_file="$1"
  local archive_name="$2"
  local checksum_entry="$3"

  grep "  ${archive_name}\$" "${checksum_file}" >"${checksum_entry}" || {
    echo "checksum for ${archive_name} not found in $(basename "${checksum_file}")" >&2
    exit 1
  }

  if command -v sha256sum >/dev/null 2>&1; then
    (cd "$(dirname "${checksum_file}")" && sha256sum -c "$(basename "${checksum_entry}")")
    return
  fi

  if command -v shasum >/dev/null 2>&1; then
    (cd "$(dirname "${checksum_file}")" && shasum -a 256 -c "$(basename "${checksum_entry}")")
    return
  fi

  echo "missing checksum tool: sha256sum or shasum" >&2
  exit 1
}

main() {
  local os
  local arch
  local tag
  local version
  local archive_name
  local checksum_name
  local release_base
  local install_dir
  local tmpdir
  local archive_path
  local checksum_path
  local checksum_entry
  local extract_dir

  if [ "${1:-}" = "--help" ] || [ "${1:-}" = "-h" ]; then
    usage
    exit 0
  fi

  need_cmd curl
  need_cmd tar
  need_cmd install
  need_cmd grep
  need_cmd sed

  os="$(normalize_os)"
  arch="$(normalize_arch)"
  tag="$(resolve_version)"
  version="${tag#v}"
  archive_name="${BINARY}_${version}_${os}_${arch}.tar.gz"
  checksum_name="${BINARY}_${version}_checksums.txt"
  release_base="https://github.com/${REPO}/releases/download/${tag}"
  install_dir="${INSTALL_DIR:-$(default_install_dir)}"

  tmpdir="$(mktemp -d)"
  trap 'rm -rf "${tmpdir}"' EXIT

  archive_path="${tmpdir}/${archive_name}"
  checksum_path="${tmpdir}/${checksum_name}"
  checksum_entry="${tmpdir}/${archive_name}.sha256"
  extract_dir="${tmpdir}/extract"

  mkdir -p "${install_dir}" "${extract_dir}"

  curl -fsSL "${release_base}/${archive_name}" -o "${archive_path}"
  curl -fsSL "${release_base}/${checksum_name}" -o "${checksum_path}"
  verify_archive "${checksum_path}" "${archive_name}" "${checksum_entry}"

  tar -xzf "${archive_path}" -C "${extract_dir}"

  if [ ! -f "${extract_dir}/${BINARY}" ]; then
    echo "binary ${BINARY} not found in ${archive_name}" >&2
    exit 1
  fi

  install -m 0755 "${extract_dir}/${BINARY}" "${install_dir}/${BINARY}"
  printf 'installed %s %s to %s\n' "${BINARY}" "${tag}" "${install_dir}/${BINARY}"
}

main "$@"
