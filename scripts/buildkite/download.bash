#!/usr/bin/env bash

set -euo pipefail

# Download logic based on that used by https://github.com/monebag/monorepo-diff-buildkite-plugin
# Used under the terms of that license.

# log <level> <message> [exit_code]
#
# Prints a formatted message to stdout or stderr, styled by level.
#
# Arguments:
#   <level>      One of: info, warn, error, fatal
#   <message>    The log message to print
#   [exit_code]  (fatal only) Optional exit code, defaults to 1
log() {
  local level="$1"
  shift

  local color reset
  reset=$(tput sgr0 2>/dev/null || true)

  case "$level" in
  info)
    color=$(tput setaf 6 2>/dev/null || true)
    ;; # cyan
  warn)
    color=$(tput setaf 3 2>/dev/null || true)
    ;; # yellow
  error)
    color=$(tput setaf 1 2>/dev/null || true)
    ;; # red
  fatal)
    color=$(tput setaf 1 2>/dev/null || true)
    local message="$1"
    local code="${2:-1}"
    printf '%s\n' "[${color}FATAL${reset}] $message" >&2
    exit "$code"
    ;;
  *)
    color=""
    ;;
  esac

  printf '%s\n' "[${color}${level^^}${reset}] $*" >&2
}

# check_cmd <command>
# Returns 0 if the given command exists in PATH, otherwise returns 1.
check_cmd() {
  local cmd="$1"
  command -v "$cmd" >/dev/null 2>&1
}

# need_cmd <command>
# Exits fatally if the given command is not available in PATH.
# Useful for pre-flight checks in scripts.
need_cmd() {
  local cmd="$1"
  if ! check_cmd "$cmd"; then
    log fatal "need '$cmd' (command not found)"
  fi
}

# _parse_architecture <os> <arch>
# Normalizes architecture string for known OS/arch combinations.
# Outputs "<os>_<normalized_arch>" or exits fatally if unsupported.
_parse_architecture() {
  local -r ostype="$1"
  local arch="$2"
  case "$arch" in
  arm64 | arm | armhf | aarch64 | aarch64_be | armv6l | armv7l | armv8l | arm64e)
    arch="arm64"
    ;;
  amd64 | xx86 | x86pc | i386 | i686 | i686-64 | x64 | x86_64 | x86_64h | athlon)
    arch="amd64"
    ;;
  *)
    log fatal "unsupported architecture \"$arch\"" 2
    ;;
  esac

  echo "${ostype}_${arch}"
}

# downloader <url> <output>
# Downloads a file using curl or wget, preferring curl if available.
# Also accepts --check to ensure a downloader is installed.
_downloader() {
  local -r url="$1"
  local -r output="$2"

  local download_client
  if check_cmd curl; then
    download_client=curl
  elif check_cmd wget; then
    download_client=wget
  else
    download_client='curl or wget'
  fi

  if [ "$url" = --check ]; then
    need_cmd "$download_client"
  elif [ "$download_client" = curl ]; then
    curl -sSfL "$url" -o "$output"
  elif [ "$download_client" = wget ]; then
    wget "$url" -O "$output"
  else
    log fatal "Unknown downloader"
  fi
}

# get_version <plugin-name>
# Extracts the version suffix for the given Buildkite plugin from BUILDKITE_PLUGINS.
# Example:
#   BUILDKITE_PLUGINS='{"cultureamp/terraform-buildkite-plugin#v1.2.3"}'
#   get_version terraform-buildkite-plugin → v1.2.3
_get_version() {
  local plugin_name="$1"
  local plugins="${BUILDKITE_PLUGINS:-}"
  echo "$plugins" | sed -nE "s/.*${plugin_name}#(v?[0-9][^\" ]*).*/\1/p"
}

# _get_download_url <repo> <executable> <arch> <version>
# Builds the download URL for the given repo, executable, architecture, and version.
# If version is empty, returns the latest URL.
_get_download_url() {
  local repo="$1"
  local executable="$2"
  local arch="$3"
  local version="$4"

  if [[ -z "$version" ]]; then
    echo "${repo}/releases/latest/download/${executable}_${arch}"
  else
    echo "${repo}/releases/download/${version#v}/${executable}_${arch}"
  fi
}

# download_binary_and_run <executable> <repo> [-- binary args...]
# Downloads the appropriate binary for the current system and executes it.
# If a version is found in BUILDKITE_PLUGINS, it is used; otherwise, the latest release is downloaded.
# Any extra arguments are passed directly to the binary.
download_binary_and_run() {
  local -r executable="$1"
  local -r repo="$2"
  shift 2

  local os arch architecture version url

  os="$(uname -s | tr '[:upper:]' '[:lower:]')"
  arch="$(uname -m)"
  architecture="$(_parse_architecture "$os" "$arch")" || return 1

  version="$(_get_version "$executable")" || return 1
  url="$(_get_download_url "$repo" "$executable" "$architecture" "$version")"

  log info "Downloading $url"
  if ! _downloader "$url" "$executable"; then
    log fatal "failed to download $url"
  fi

  chmod +x "$executable"
  exec "./$executable" "$@"
}
