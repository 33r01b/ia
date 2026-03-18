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

## Shell usage

The agent is allowed to use shell utilities for repository inspection and controlled refactoring.

Allowed read-only tools:
- find
- grep / rg
- sort
- uniq
- sed
- awk
- cat
- head / tail
- wc

Allowed refactoring tools:
- cp
- mv
- mkdir
- rm (targeted only)
- sed -i
- perl (text refactoring only)

Usage rules:
- Prefer minimal and targeted changes
- Prefer repository-wide refactoring only when explicitly justified
- Explain planned file operations before running them
- List affected files for non-trivial refactoring
- Prefer simple commands over complex one-liners when possible

Restrictions:
- Do not execute project code via shell
- Do not run build tools, package managers, or compilers
- Do not use perl or shell scripting as a substitute for normal program execution
- Do not use filesystem-changing commands outside the project workspace
- Do not perform bulk destructive operations unless explicitly instructed

Before performing destructive or mass-refactoring operations:
- Explain what will be changed
- List affected files or file patterns
- State assumptions and risks
- Do not proceed without confirmation

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
