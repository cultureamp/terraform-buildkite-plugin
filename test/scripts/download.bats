#!/usr/bin/env bats

bats_load_library bats-support
bats_load_library bats-assert
bats_load_library bats-file

load '../../scripts/buildkite/download.bash'

#
# Tests for top-level docker bootstrap command. The rest of the plugin runs in Go.
#

# Uncomment the following line to debug stub failures
# export [stub_command]_STUB_DEBUG=/dev/tty
#export DOCKER_STUB_DEBUG=/dev/tty

setup() {
  export BUILDKITE_PLUGIN_TEST_MODE=true
}

teardown() {
  unset BUILDKITE_PLUGIN_TERRAFORM_BUILDKITE_PLUGIN_TEST_MODE
  rm ./terraform-buildkite-plugin || true
}

create_script() {
  cat >"$1" <<EOM
set -euo pipefail

echo "executing $1:\$@"

EOM
}

@test "Downloads and runs the command for the current architecture" {
  local architecture="$(get_architecture)"

  function downloader() {
    echo "$@"
    create_script $2
  }
  export -f downloader

  run download_binary_and_run

  unset downloader

  assert_success
  assert_line --regexp "https://github.com/xphir/terraform-buildkite-plugin/releases/latest/download/terraform-buildkite-plugin_${architecture} terraform-buildkite-plugin"
  assert_line --regexp "executing terraform-buildkite-plugin"
}
