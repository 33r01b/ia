# Agent Rules

## Complexity constraints
- Do not create interfaces without clear need
- Avoid abstractions without real use case
- Deliver a working minimal v1 first
- Prefer minimal changes over large refactoring

## Execution Environment

The agent runs inside an isolated container without language runtimes.

Constraints:
- Go is NOT available
- Docker / docker-compose are NOT available
- Code execution is NOT possible

Filesystem:
- Do not create or modify files unless explicitly requested

## Git usage

You can run git commands in read-only mode.

Allowed:
- git status
- git diff
- git log
- git show

Forbidden unless explicitly instructed:
- git commit
- git push
- git rebase
- git reset
- git checkout (changing branches)

## Implications

- Do not attempt to run code
- Do not suggest running commands as verification
- Do not claim that code compiles or tests pass
- Rely only on static analysis of the codebase

## Verification policy

Since code execution is not available:

- Always provide a manual verification checklist
- Explicitly state assumptions
- Mark uncertain parts of the solution

## Response format

When providing a solution:

1. Assumptions
2. Proposed solution
3. Changes (minimal diff or code snippets)
4. Manual verification steps
5. Risks / uncertainties
