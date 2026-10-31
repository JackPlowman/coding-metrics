# Local GitHub Actions

Repository workflows keep their triggers, runners, matrices, permissions, and
checkout steps. The composite actions in this directory contain the steps for
code checks, CodeQL analysis, dependency review, PR labelling, label syncing, and
cache cleanup.

Check out the repository before calling an action with
`uses: ./.github/actions/<action-name>`. Declare token permissions in the calling
job and pass `github.token` through the `github-token` input when required.
Composite actions cannot declare job permissions or access the `secrets` context.

Keep third-party action references pinned to commit SHAs and specify `shell` for
every composite `run` step. Validate workflow and action changes with actionlint,
zizmor, yamllint, Prettier, and `pinact run --verify --check`.
