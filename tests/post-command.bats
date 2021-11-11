#!/usr/bin/env bats

load '/usr/local/lib/bats/load.bash'

@test "post-command: as command" {
  export BUILDKITE_PLUGIN_TERRAFORM_POSTCOMMAND="echo hello from command"

  run $PWD/hooks/post-command

  assert_success
  assert_output --partial "hello from command"
}

@test "post-command: as file" {
  export BUILDKITE_PLUGIN_TERRAFORM_POSTCOMMAND="tests/fixtures/post-command-sample.sh"

  run $PWD/hooks/post-command

  assert_success
  assert_output --partial "hello from post-command file"
}
