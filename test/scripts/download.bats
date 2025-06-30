#!/usr/bin/env bats

# Load helper libraries
bats_load_library bats-support
bats_load_library bats-assert
bats_load_library bats-file

# Load the script under test
load '../../scripts/buildkite/download.bash'

#
# Setup & teardown
#

setup() {
  export BUILDKITE_PLUGIN_TEST_MODE=true
}

teardown() {
  unset BUILDKITE_PLUGIN_TEST_MODE
  rm -f ./terraform-buildkite-plugin
}

#
# Utility: Create a stub executable
#
create_script() {
  local script_path="$1"
  cat >"$script_path" <<EOM
#!/usr/bin/env bash
set -euo pipefail
echo "executing $script_path:\$@"
EOM
  chmod +x "$script_path"
}

#
# Tests: download_binary_and_run (integration-style)
#
@test "download_binary_and_run downloads and runs the command for the current architecture" {
  local -r executable="terraform-buildkite-plugin"
  local -r repo="https://github.com/cultureamp/terraform-buildkite-plugin"
  local -r expected_architecture="linux_amd64"

  # Stub _downloader to simulate a successful binary download
  _downloader() {
    local url="$1"
    local output="$2"
    echo "$url $output"
    create_script "$output"
  }
  export -f _downloader

  # Stub _parse_architecture to always return a known value
  _parse_architecture() {
    echo "$expected_architecture"
  }
  export -f _parse_architecture

  run download_binary_and_run "$executable" "$repo"

  unset -f _downloader
  unset -f _parse_architecture

  assert_success
  assert_line --regexp "https://github.com/cultureamp/terraform-buildkite-plugin/releases/latest/download/${executable}_${expected_architecture} $executable"
  assert_line --regexp "executing $executable"
}

#
# Tests: _parse_architecture
#

@test "_parse_architecture normalizes known arch (x86_64 → amd64)" {
  run _parse_architecture linux x86_64
  assert_success
  assert_output "linux_amd64"
}

@test "_parse_architecture normalizes known arch (aarch64 → arm64)" {
  run _parse_architecture darwin aarch64
  assert_success
  assert_output "darwin_arm64"
}

@test "_parse_architecture fails on unknown arch" {
  run _parse_architecture linux totally_made_up
  assert_failure
  assert_line --partial "unsupported architecture"
}

#
# Tests: check_cmd / need_cmd
#

@test "check_cmd returns 0 for existing command" {
  run check_cmd echo
  assert_success
}

@test "check_cmd returns non-zero for missing command" {
  run check_cmd definitely-not-a-real-cmd
  assert_failure
}

@test "need_cmd logs fatal for missing command" {
  run need_cmd definitely-not-a-real-cmd
  assert_failure
  assert_line --partial "need 'definitely-not-a-real-cmd'"
}

#
# Tests: _get_version
#

@test "_get_version extracts plugin version correctly" {
  export BUILDKITE_PLUGINS='{"cultureamp/terraform-buildkite-plugin#v1.2.3"}'

  run _get_version terraform-buildkite-plugin
  assert_success
  assert_output "v1.2.3"
}

#
# Tests: _get_download_url
#

@test "_get_download_url returns latest URL when version is empty" {
  run _get_download_url "https://github.com/org/repo" "tool" "linux_amd64" ""
  assert_success
  assert_output "https://github.com/org/repo/releases/latest/download/tool_linux_amd64"
}

@test "_get_download_url returns versioned URL with v-prefix stripped" {
  run _get_download_url "https://github.com/org/repo" "tool" "linux_amd64" "v2.0.0"
  assert_success
  assert_output "https://github.com/org/repo/releases/download/2.0.0/tool_linux_amd64"
}
