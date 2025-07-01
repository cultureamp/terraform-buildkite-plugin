<!-- markdownlint-disable line-length-->
# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2025-07-01

### <!-- 0 -->🚀 Features

- Feat: init project by @xphir in [#1](https://github.com/cultureamp/terraform-buildkite-plugin/pull/1)

### <!-- 1 -->🐛 Bug Fixes

- Fix: update job conditions to require changes for markdown, cspell, action, scripts, and golang on non-main branches by @xphir in [#14](https://github.com/cultureamp/terraform-buildkite-plugin/pull/14)
- Fix: update job conditions to run on non-main branches for markdown, cspell, action, scripts, and golang by @xphir in [#12](https://github.com/cultureamp/terraform-buildkite-plugin/pull/12)
- Fix: add a placeholder CHANGELOG.md file as the changelog generator needs it to exist to properly generate changelogs. by @xphir
- Fix: remove hook due to ci compatibility issues by @xphir in [#9](https://github.com/cultureamp/terraform-buildkite-plugin/pull/9)
- Fix: resolve incompatible goreleaser versions by @xphir in [#8](https://github.com/cultureamp/terraform-buildkite-plugin/pull/8)
- Fix: correct workflow name casing and disable changelog generation by @xphir in [#2](https://github.com/cultureamp/terraform-buildkite-plugin/pull/2)

### <!-- 10 -->💼 Other

- Build(deps): bump marocchino/sticky-pull-request-comment from 2.9.2 to 2.9.3 by @dependabot[bot] in [#4](https://github.com/cultureamp/terraform-buildkite-plugin/pull/4)
- Build(deps): bump peter-evans/create-pull-request from ba864ad40c29a20a464f75f942160a3213edfbd1 to a59c52d55daab9f6a666e8741848b3cf101395a7 by @dependabot[bot] in [#6](https://github.com/cultureamp/terraform-buildkite-plugin/pull/6)
- Initial commit by @xphir

### <!-- 2 -->🚜 Refactor

- Refactor: consolidate changelog generation step to release workflow by @xphir in [#10](https://github.com/cultureamp/terraform-buildkite-plugin/pull/10)

### <!-- 6 -->🧪 Testing

- Test: improve test tools and actions by @xphir in [#15](https://github.com/cultureamp/terraform-buildkite-plugin/pull/15)

### <!-- 7 -->⚙️ Miscellaneous Tasks

- Ci: add initial release workflow for pull request validation by @xphir in [#25](https://github.com/cultureamp/terraform-buildkite-plugin/pull/25)
- Ci: remove unused release job outputs and update markdown job conditions by @xphir
- Ci: update conditions for markdown, cspell, action, scripts, and golang jobs to include pre-release branch by @xphir
- Ci: update GitHub token usage in changelog generation step by @xphir in [#24](https://github.com/cultureamp/terraform-buildkite-plugin/pull/24)
- Ci: update changelog generation to include version tag by @xphir
- Ci: update commit message and title for pre-release PR to reflect changelog updates by @xphir
- Ci: add CHANGELOG.md to ignorePaths in cspell configuration by @xphir
- Ci: remove following brace by @xphir in [#23](https://github.com/cultureamp/terraform-buildkite-plugin/pull/23)
- Ci: add specialised token for the pull request action by @xphir
- Ci: enhance pre-release PR body with detailed release preview and changelog by @xphir
- Ci: add "bot" label to dependabot updates for gomod and github-actions by @xphir
- Ci: flip the incorrect branch vs base names by @xphir in [#21](https://github.com/cultureamp/terraform-buildkite-plugin/pull/21)
- Ci: rename release job to pre-release and add pre-release workflow by @xphir in [#20](https://github.com/cultureamp/terraform-buildkite-plugin/pull/20)
- Ci: rename lint-github-actions to lint-actions for consistency by @xphir
- Ci: add .semrel to .gitignore to exclude semantic release files by @xphir
- Ci: add-dependabot-labels by @xphir in [#19](https://github.com/cultureamp/terraform-buildkite-plugin/pull/19)
- Ci: improve cliff template to conform to markdownlint by @xphir in [#18](https://github.com/cultureamp/terraform-buildkite-plugin/pull/18)
- Ci: improve git-cliff configuration and add just changelog command by @xphir in [#17](https://github.com/cultureamp/terraform-buildkite-plugin/pull/17)

## New Contributors

- @xphir made their first contribution in [#25](https://github.com/cultureamp/terraform-buildkite-plugin/pull/25)
- @dependabot[bot] made their first contribution in [#4](https://github.com/cultureamp/terraform-buildkite-plugin/pull/4)

