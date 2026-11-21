# HANDOFF - living (DIRECT-WRITE PROTOCOL)

## People
- **Human** - sole git author and contributor. Commits, pushes, publishes. Approves everything external.
- **Planner (OpenCode)** - owns everything except `web/**`; writes specs, reviews, wires backend/routes.
- **Frontend builder (Devin)** - desktop app on the same machine; owns `web/**` **only**; writes files directly into the project folder.

## Project location (for Devin - Windows)
The project lives inside WSL Ubuntu. From Windows/Devin, open:

```
\\wsl.localhost\Ubuntu\home\solution\edj
```
Alternative: `\\wsl$\Ubuntu\home\solution\edj`

Linux path (for me): `/home/solution/edj`

## Flow
1. Planner freezes a TASK in `TASKS/` and the human sends it to the frontend builder.
2. Frontend builder reads the frozen docs from the project folder and writes **only** `web/**` files directly there.
3. Frontend builder must **not** stage, commit, push, merge, branch, create PRs, or edit any file outside `web/`.
4. Planner reviews the written files against `SPEC.md`, `UIUX.md`, `MOTION.md`, `ASSETS.md` and writes `REVIEWS/TASK-###.md`: **PASS** or **CHANGE REQUEST**.
5. Human commits. Human is the only author in git history.

## Rules
- No AI agent runs git write commands, ever.
- No references to or credits for any other project/product in code, docs, UI, or commits.
- If a task is ambiguous, the builder asks the human - never guess.
- Do not run a build or install packages; the planner handles compilation.
- Legacy `handoffs/` and patch format are no longer required under direct-write; a builder may still drop a summary there if useful.
