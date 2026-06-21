---
name: polish-sdk-pr
description: Complete the human-judgement half of an SDK regeneration PR. Run locally in Claude Code inside an SDK repo (sdk-node / sdk-python / sdk-go / sdk-php) after the api-released pipeline opens a draft "Regenerate from api-vX.Y.Z" PR labeled needs-polish. Writes examples, helpers, README refresh, CHANGELOG, and a migration guide (breaking only), confirms the version bump, then flips the PR to ready-for-review. Never merges, never publishes.
disable-model-invocation: true
---

# polish-sdk-pr — finish an SDK regen PR

## Hard rules (read first — these are absolute)

- **NEVER merge.** Review and merge are human acts.
- **NEVER publish.** Publishing is a separate human-triggered `workflow_dispatch`.
- **NEVER edit generated code.** Files under the generated dir (`src/generated/` | `tracore/generated/` | `internal/generated/` | `src/Generated/`) carry a `DO NOT EDIT` header and are owned by codegen. Extend only the hand-written ergonomic layer.
- **Stay on the open draft PR's branch.** Do not open new PRs, retarget, rebase onto unrelated work, or switch branches.
- **Do not touch unrelated files.** Only touch examples, helpers, README, CHANGELOG, the migration guide, and the version manifest — nothing else.
- Match the repo's existing style, naming, and idioms. Read neighbouring code before writing.
- No emojis anywhere — code, comments, examples, README, CHANGELOG, PR notes.

If any step cannot be completed safely, stop and leave a note in the PR rather than guessing or working around a rule above.

## What this skill is

The `api-released` pipeline (in `tracore-io/tracore`) regenerates the client and opens a **draft** PR titled `Regenerate from api-vX.Y.Z`, labeled `needs-polish`. CI has done all the free mechanical work — codegen, version pin, mock tests. This skill does the half a machine should not: prose and judgement. It runs **locally on the maintainer's Claude subscription**, inside the SDK repo.

This file is the **canonical source**. It is synced read-only into each SDK repo at `.claude/skills/polish-sdk-pr/SKILL.md` by `sync-sdk-skill.yml`. Edit only the copy at `tooling/sdk-skills/polish-sdk-pr/SKILL.md` in the monorepo; the synced copies must never be hand-edited (the next sync would overwrite the change and the review trail would be lost).

## Trigger and context

You are invoked locally, by a maintainer, inside an SDK repo that already has the freshly-opened draft PR checked out (or available to check out). Your inputs:

- **The draft PR** titled `Regenerate from api-vX.Y.Z`, labeled `needs-polish`. Its body carries the canonical §8.3 checklist (CI half already checked) and the release context.
- **The dispatch payload** that opened the PR, surfaced in the PR body. The fields you rely on:
  - `api_version` — the pinned spec version this regen targets (e.g. `0.5.0`); matches `.api-version` and the tag `api-v0.5.0`.
  - `breaking` — boolean. Drives the version bump and whether a migration guide is required.
  - `diff_summary` — a markdown changelog of what changed in the spec since the previous release. This is your worklist for examples, CHANGELOG, and the migration guide.

If the PR body lacks `diff_summary` (it was truncated past the dispatch size cap), read the full changelog from the `api-vX.Y.Z` GitHub Release notes. The release workflow publishes the same notes to both `tracore-io/openapi` (the public spec release the PR was generated from) and the `tracore-io/tracore` monorepo; the truncation note in the PR body links straight to it.

## 0. Orient

Before changing anything:

1. Find the open draft PR titled `Regenerate from api-vX.Y.Z` (the one labeled `needs-polish`). If none, stop and say so.
2. Read its body: the §8.3 checklist (CI half checked) and `diff_summary`. Recover the diff from the release notes if it was truncated.
3. Read `.api-version` and the manifest to confirm the pinned API version and the default version bump CI wrote.
4. Skim the generated dir to learn the new/changed surface, and the hand-written ergonomic layer to learn the existing patterns (resource grouping, polling helper, pagination helper, typed errors, top-level `extract`).
5. Note the `breaking` flag. It makes the migration-guide step required and changes the version-bump expectation (step 5).

## 1. Examples for new and changed endpoints

For every endpoint added or changed in `diff_summary`, add or update a runnable example in the repo's examples location (match where existing examples live). Mirror the ergonomic contract — call through the resource group (`client.workspaces`, `.schemas`, `.documents`, `.runs`, `.webhooks`, `.environments`, `.user`), not the raw generated client. Keep examples idiomatic for the language. Do not invent endpoints that are not in the spec.

## 2. Extend helpers only where the pattern fits

If a change naturally extends an existing helper (polling, pagination, typed-error normalization, the top-level `extract` convenience), extend it following the established pattern. Do **not** invent new abstractions, new helper families, or speculative generality. When the existing pattern does not fit cleanly, leave a brief note in the PR rather than forcing it — helper ergonomics for scaffold repos are deferred until they graduate to published.

## 3. Refresh the README if examples went stale

If the README's quickstart or examples reference endpoints, fields, or signatures the regen changed, update them so they type-check / run against the current generated client and `api.tracore.io`. Leave the README untouched if nothing it shows changed.

## 4. CHANGELOG entry from diff_summary

Add a CHANGELOG entry for this version derived from `diff_summary`: group as Added / Changed / Removed / Fixed, human-readable, referencing the pinned `api-vX.Y.Z`. Keep it terse and factual.

## 5. Confirm or adjust the version bump

CI wrote a **default** bump. Reconcile it against reality using this rule:

- **Breaking release:** MAJOR bump — **but** while the SDK is still `0.x`, a breaking change maps to a **MINOR** bump instead (the `0.x` breaking lever). The MAJOR rule takes effect only at `>= 1.0`. (See the repo's CONTRIBUTING.)
- **Additive release:** MINOR bump.
- **Helper-only / doc-only change** with no new generated surface: you may refine MINOR down to PATCH.

Confirm the manifest version field and the pinned `tracore api_version` agree with `.api-version`.

## 6. Migration guide (breaking releases only)

Skip this entirely for additive releases. For a breaking release, add a migration guide in the repo's docs/migration location with a concrete **before / after** snippet for each breaking change drawn from `diff_summary`. Use this template per change:

```markdown
### <short title of the breaking change>

<one sentence: what changed and why a caller must act.>

Before (api-v<PREV>):

    <old SDK call / type — the form that no longer works>

After (api-v<THIS>):

    <new SDK call / type — the replacement>
```

Keep the before/after in the SDK's own surface (the ergonomic layer), not raw HTTP, so a reader can copy the fix directly.

## 7. Flip the PR to ready

Only after steps 1-6:

1. Check the boxes in the PR body's "To do" half, each with a brief note on what was done (or why it was skipped — e.g. "no README examples affected").
2. Swap the label `needs-polish` -> `ready-for-review`.
3. Mark the draft PR **ready for review**.
4. Leave a short summary comment listing what changed and the confirmed version bump.

Then stop. A human reviews and merges; a human triggers publish. **The skill never merges and never publishes.**

## Worked example (additive: one new endpoint)

Suppose `diff_summary` for `api-v0.6.0` contains:

```
### Added
- GET /workspaces/{slug}/usage — current-period page + extraction counters
```

`breaking` is `false`. A complete polish pass looks like:

- **Examples (step 1):** add an example calling it through the resource group — for sdk-node, `examples/get-usage.ts`:

      const usage = await client.workspaces.getUsage("my-workspace");
      console.log(usage.pages, usage.extractions);

  Match how the neighbouring `examples/` files import the client and read env vars; do not call the raw generated client.

- **Helpers (step 2):** none — a single GET fits the existing resource-group pattern with no new abstraction. Leave helpers untouched.
- **README (step 3):** untouched — the quickstart does not show usage counters.
- **CHANGELOG (step 4):**

      ## 0.6.0 — api-v0.6.0
      ### Added
      - `client.workspaces.getUsage(slug)` — current-period page and extraction counters.

- **Version bump (step 5):** additive -> MINOR. Confirm manifest = `0.6.0`, `.api-version` = `0.6.0`, pinned `tracore api_version` = `0.6.0`.
- **Migration guide (step 6):** skipped — additive release.
- **Flip (step 7):** check the four "To do" boxes with one-line notes ("added examples/get-usage.ts", "no helper change", "README unaffected", "CHANGELOG 0.6.0 added", "confirmed MINOR -> 0.6.0", "migration guide N/A — additive"); swap `needs-polish` -> `ready-for-review`; mark ready; comment the summary. Do not merge.

For a **breaking** release the same flow applies, except the version bump follows the breaking rule (MAJOR, or MINOR while `0.x`) and step 6 produces a migration guide using the before/after template above.

## Canonical draft-PR checklist (spec section 8.3)

CI fills the left ("Done") half before the PR opens; this skill fills the right ("To do") half.

```
Done by CI (free, no AI):        To do (polish skill, local subscription):
[x] .api-version pinned          [ ] examples for new/changed endpoints
[x] spec downloaded              [ ] extend helpers if the pattern fits
[x] codegen ran                  [ ] refresh README if examples are stale
[x] mock tests pass              [ ] CHANGELOG entry from diff_summary
[x] default version bump set     [ ] confirm/adjust version bump
                                 [ ] migration guide (breaking releases only)
```
