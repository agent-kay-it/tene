# tene CLI UX & AI-Safe Permission Model — PRD

> **Sprint ID**: `cli-ux-permission-model`
> Master Plan reference: [master-plan.md](master-plan.md)

---

## 1. Personas

### P1 — Indie Hacker "Kay" (Primary)

- 1인 개발자, Cursor + Claude Code 매일 사용
- Stripe / OpenAI / Resend 등 7-15개 secret 운영
- 멘탈 모델: "tene 는 .env 의 안전한 대체재"
- Pain: 매번 `tene set` 할 때 keychain 풀고, `tene list` 할 때 또 풀고, AI 가 키 이름 모르니 매번 사용자가 알려줘야 함
- Wins: vault 가 있는 머신에서는 default 로 작업 흐름이 안 끊김. AI 가 키 이름은 알지만 값은 절대 못 봄.

### P2 — Backend Engineer "Sam" (Secondary)

- 팀 SaaS 의 CI/CD 운영
- tene 를 `--no-keychain` 으로 호출 (Docker container 안에서)
- Pain: 매 명령 password stdin 흘리기 (`echo $PW | tene ...`)
- Wins: list 같은 metadata 명령은 password 없이도 동작 → CI 스크립트 단순화

### P3 — First-time User "Jin" (Tertiary)

- tene 처음 써봄
- Pain: `tene init` 후 "이거 매번 password 묻는 거 맞아? 너무 귀찮은데"
- Wins: init 직후 onboarding 메시지가 "list/env list 는 password 안 묻고, get/run 만 묻습니다" 라고 명확히 알려줌

---

## 2. Job Stories (JTBD format)

### JS1 (metadata read)

**When** I'm pair-programming with Claude Code and I want to remind myself which API keys this project has,
**I want to** run `tene list` and see a one-line answer instantly,
**so I can** keep my flow without dropping back into typing a password.

### JS2 (AI assist key naming)

**When** Claude is about to add a new feature that needs an OpenAI key,
**I want** Claude itself to be able to run `tene list --json` and check the canonical key name,
**so** it doesn't invent `OPENAI_KEY` when my vault has `OPENAI_API_KEY`.

### JS3 (AI safety, value protection)

**When** Claude tries to actually read the value (e.g., `tene get OPENAI_API_KEY`) inside a non-TTY tool result,
**I want** tene to refuse with STDOUT_SECRET_BLOCKED,
**so** plaintext never enters the AI context window without my explicit opt-in.

### JS4 (CI/CD metadata)

**When** my GitHub Actions workflow needs to verify that the prod environment has all the required keys,
**I want** to run `tene list --json --env prod --no-keychain` without piping a password,
**so** my pipeline script stays simple and the password env var is only needed where actually required (run / get).

### JS5 (Discoverability)

**When** I forget which tene commands need a password and which don't,
**I want** to run `tene permissions` and see the full table,
**so** I don't have to grep CHANGELOG or read docs.

### JS6 (First-run education)

**When** I'm finishing `tene init`,
**I want** the output to tell me, in one sentence, the permission model AND the preview default,
**so** I'm not surprised the next time `tene list` doesn't ask for a password and also shows partial value.

### JS7 (Q2 — preview opt-out for privacy-first user)

**When** I'm a privacy-first user who doesn't want even partial secret values stored in vault.db,
**I want to** run `tene config preview.enabled=false` immediately after `tene init`,
**so that** even if my vault.db is exfiltrated, no plaintext fragments are exposed.

### JS8 (F8 — audit log readability)

**When** I'm trying to recall which tene commands I ran yesterday,
**I want to** run `tene audit tail -n 50` or `tene audit show --since 1d --filter cli.secretread`,
**so** I can trace my own activity without writing raw SQL against vault.db.

### JS9 (F8 — audit log housekeeping)

**When** my vault has been in use for a year and `audit_log` grew to 80 MB,
**I want** tene to print a one-line stderr warning on the next command suggesting `tene audit prune --older-than 30d`,
**so** I can keep vault.db lean without losing the most recent forensic data.

---

## 3. User Stories (feature-aligned)

### F1 — Vault Metadata Read API + Schema v2 Migration (Q2 decision)

- As a tene internals contributor, I want a vault method that returns `{name, version, updated_at, preview}` per secret so I can power a no-decrypt read path that even shows partial values.
- As an existing tene user, I want my v1 vault.db to auto-upgrade to v2 on first command without losing data.
- As a security-conscious user, I want a `tene config preview.enabled=false` switch to disable the preview column entirely.
- DoD:
  - `Vault.ListSecretMetadata(env)` returns `[]domain.VaultKeyMeta` (now with `Preview` field), never touches `encrypted_value`.
  - Schema migrates v1 → v2 via `ALTER TABLE secrets ADD COLUMN preview TEXT DEFAULT ''` (idempotent).
  - `tene migrate fill-previews` command (one-time) populates preview for existing secrets after one master-password unlock.
  - `crypto.DerivePreview(plaintext, front, back) string` helper; hard cap `front+back ≤ 8`.
  - On `tene set` / `tene import`, preview is auto-derived from plaintext before encryption — unless `preview.enabled=false`, then empty.

### F2 — Permission Tier Model

- As a tene security auditor, I want each CLI command to declare its permission tier in a single declarative table so I can review the security surface in 30 seconds.
- DoD: new `internal/auth/permissions.go` exports `PermLevel` enum (`PermMetaRead`, `PermSecretWrite`, `PermSecretRead`) and a `CommandTier` map. `root.go` PreRunE hooks consult this map.

### F3 — `list` Reads Preview Column Directly (Q2 decision)

- As an indie hacker, I want `tene list` to work without unlocking the vault and still show partial value previews.
- As a `--json` consumer, I want the JSON shape to always include `preview` as a string (possibly empty when `preview.enabled=false`).
- As an existing user whose vault still has v1 schema or v2 schema with empty preview rows (pre-migration), I want list to gracefully show empty string for those rows.
- DoD:
  - `tene list` returns in <15ms on a 100-entry vault (no master-key derivation, no Argon2id cost).
  - JSON keys: `name`, `preview` (string, possibly ""), `version`, `updatedAt` all present, type-stable.
  - Text mode: 3-column table `NAME | PREVIEW | UPDATED`. PREVIEW column shows literal `""` (or `-`) when preview empty/disabled.
  - When `preview.enabled=false`, list output is functionally equivalent to "Q2 option C form": name + env + version + updated only.

### F4 — Audit Logging by Permission Tier

- As a security-curious operator, I want every CLI invocation to record which permission tier was activated in `audit_log`.
- DoD: `audit_log.action` includes the tier prefix (`cli.metaread`, `cli.secretwrite`, `cli.secretread`). Existing entries (`secret.read`, `vault.init`, etc.) preserved.

### F5 — `tene permissions` Info Command

- As a curious user, I want to run `tene permissions` and see the full tier table. (Q1 final: `tene permissions` is the only command surface; the previously considered `tene info --permissions` alternative was rejected.)
- DoD: new command prints a table with columns `Command | Tier | Requires Password?`. JSON shape: `{ commands: [{ name, tier, requiresUnlock }] }`.

### F6 — Keychain Fallback UX Polish

- As a Linux user without libsecret installed, I want a one-time stderr notice on first command after fallback, saying "Using file-keystore at ~/.tene/keyfile".
- DoD: notice appears exactly once per session (state stored in `.tene/.fallback-warned` sentinel file), suppressed by `--quiet`.

### F7 — `tene init` First-Run UX Update

- As a first-time user, I want the `tene init` success output to include a one-line permission model summary AND a 1-line note about preview default.
- DoD: post-init output includes 3 hint lines (matching plan.md F7 step 1 exactly):
  - "Run `tene permissions` to see which commands need your password."
  - "`tene list` shows last 4 chars of each value by default (no prefix exposed)."
  - "Disable previews with `tene config preview.enabled=false` — or run `tene config preview.front=N` to also show first N chars (opt-in; exposes API key prefix)."
  - Total init output increase: ≤ 3 lines.
  - Rationale: default `front=0, back=4` (Q2 final, 2026-05-20). The "first-4 + last-4" wording from earlier drafts is obsolete and must NOT appear in init output.

### F8 — Audit Log Management (Q3 decision — NEW)

- As an active user, I want `tene audit tail -n 50` to show the recent 50 audit entries in a readable table.
- As a user investigating my own behavior, I want `tene audit show --since 1d --filter cli.secretread` to filter by time + action prefix.
- As a user maintaining a long-lived vault, I want `tene audit prune --older-than 30d` to manually delete old entries (with confirmation prompt).
- As a user who hasn't noticed the audit log growing, I want a one-line stderr warning on a normal `tene` command when audit_log size crosses 50 MB (default), suggesting prune.
- As a config-tuner, I want `tene config audit.warnAtMB=100` to raise the warning threshold.
- DoD:
  - `tene audit tail|show|prune` subcommands implemented.
  - `tene audit prune` requires explicit confirmation (`--force` to skip).
  - Threshold warning fires at most once per 24h per machine (sentinel `~/.tene/.audit-warned-<dir-hash>`).
  - Auto-deletion never happens — manual prune only.
  - `tene audit` itself is `PermMetaRead` tier (no password needed).

---

## 4. Non-Functional Requirements

### NFR-01 — Security

- All **14 invariants** from `master-plan.md §8` MUST hold throughout the sprint (**8 existing + 5 Q2 + 1 F8**). Source of truth: `master-plan.md §8` (Korean text) ↔ `design.md §10` (numbered list) ↔ `state JSON securityInvariants[]` (14 entries). Any divergence is a critical defect.
- Audit log retention unchanged by default (indefinite). Manual prune via `tene audit prune` only.
- No new network endpoint; tene remains zero-server.
- **Preview trade-off (Q2, 2026-05-20 default 강화)**: vault.db now contains plaintext preview substrings for each secret by default. **Default exposure: back 4 only (`front=0, back=4`)** — API key prefix (`sk-`, `ghp_`, `AKIA…`) hidden, so service identification (OpenAI/Stripe/GitHub/AWS) is NOT possible from a leaked vault.db. Users can opt-in to prefix exposure via `tene config preview.front=N` (max front+back ≤ 8) when they explicitly want the visual cue. Document explicitly in `SECURITY.md`, README, and `tene init` output. Full opt-out via `tene config preview.enabled=false`.
- **Audit log size threshold (F8)**: warn at 50 MB by default. Configurable via `tene config audit.warnAtMB=N`. Auto-deletion never happens.

### NFR-02 — Performance

- `tene list` no-decrypt path p50 < 15ms, p99 < 40ms (vault.db with 100 secrets in 5 environments).
- `tene list --json` decrypt path p50 < 80ms (existing baseline).
- No regression on `tene get`, `tene run`, `tene export` (existing benchmarks).

### NFR-03 — Compatibility

- Backward compatibility: 100% behavior-preserving on existing tests. v1→v2 schema migration is automatic and idempotent.
- `go.mod` minimum version unchanged (currently Go 1.22+, check at PR time).
- CLI flags: no existing flag renamed or removed. New flags additive only.
- JSON output: all existing keys present; new keys additive. `preview` field type is **string** (possibly empty `""`) — NOT null (Q2 decision overrides earlier null plan).
- Old v1 vault.db opens successfully on new binary. `secrets.preview` is empty until `tene migrate fill-previews` runs.
- New v2 vault.db opening on OLD binary: SQLite ignores unknown columns on SELECT *, but `INSERT INTO secrets` with old binary writes NULL preview — acceptable degraded state (rare scenario, documented).

### NFR-04 — Observability

- Audit log entry added for every CLI invocation (currently only specific events like `secret.read`, `vault.init`).
- `audit_log.action` format: `cli.<tier>.<verb>` (e.g., `cli.metaread.list`, `cli.secretwrite.set`).
- No external telemetry. Audit log stays in local vault.db.

### NFR-05 — Cross-Platform

- macOS Keychain Access, Linux libsecret/GNOME Keyring/KWallet, Windows CredManager — all supported via `go-keyring`. No new platform-specific code.
- File-keystore fallback for Linux servers without keyring daemon — already works.

### NFR-06 — Documentation

- README.md gets a "Permission Tiers" section (<150 words).
- CHANGELOG.md `[Unreleased]` gets entries for each feature, marked `### Added`.
- `apps/web/content/blog/*.mdx` — optional blog post (separate PDCA cycle, not blocking this sprint).
- `tene --help` Long text gets the tier table appended.

---

## 5. Pre-mortem — "If this sprint fails, why?"

### Failure Mode A — Permission tier is inconsistently applied

- **Symptom**: User runs `tene set ...` expecting password-free behavior; gets a password prompt; reads CHANGELOG and finds tier-table says `set` is `PermSecretWrite` (requires unlock).
- **Why fails**: Mismatch between user mental model ("tene UX is now password-free") and implemented reality (only metadata-tier).
- **Mitigation**: Crystal clear docs. `tene permissions` command. Hero message in `tene init`. README section.

### Failure Mode B — `--no-keychain` users still get full prompt frequency

- **Symptom**: CI user on `--no-keychain` runs `tene list` in a script. We removed the unlock requirement, so OK. But `tene set ...` in the next script step still prompts.
- **Why fails**: We can't remove unlock from write paths without breaking crypto invariant.
- **Mitigation**: Document this clearly. `TENE_MASTER_PASSWORD` env var path is the supported answer (already exists).

### Failure Mode C — JSON consumer breaks because `preview` is now `null`

- **Symptom**: Existing user has a `jq` pipeline expecting `preview` to be a string. After update, it's null on no-keychain machines.
- **Why fails**: We didn't communicate the JSON shape change.
- **Mitigation**: CHANGELOG `### Changed` entry. Test case asserting both `"preview": null` AND `"preview": "sk_t****x"` are valid.

### Failure Mode D — tene-cloud build breaks at F1

- **Symptom**: F1 additive change to `VaultKeyMeta` accidentally renames a JSON tag.
- **Why fails**: Cross-repo silent break.
- **Mitigation**: Pre-flight `grep -rn "VaultKeyMeta" ../tene-cloud/` BEFORE writing F1. Run `cd tene-cloud && go build ./...` as part of F1 acceptance.

### Failure Mode E — F5 over-engineering

- **Symptom**: `tene permissions` becomes a config-driven runtime engine instead of a static table.
- **Why fails**: Lost focus on user value (just a table).
- **Mitigation**: F5 implementation budget is hard-capped at 80 LOC. If it grows beyond that, revert and inline as a printf table.

### Failure Mode F — Security review fails

- **Symptom**: Reviewer asks "what stops a malicious binary from calling `Vault.ListSecretMetadata` and dumping all key names to a remote server?"
- **Why fails**: Answer is "name plaintext is acceptable threat surface". We need to write that down.
- **Mitigation**: `docs/security.md` or `SECURITY.md` update — explicitly document the threat model: "We protect secret VALUES at rest and in transit. Secret NAMES are considered low-sensitivity metadata, plaintext-stored, and intentionally readable without unlock for UX."

### Failure Mode G (Q2 decision — UPDATED 2026-05-20) — Preview plaintext column leads to a real incident

- **Symptom**: A user posts their `~/projects/foo/.tene/vault.db` in a GitHub issue accidentally. Attacker tries to extract previews and identify which services the user has keys for.
- **Default (front=0, back=4) mitigation** — preview = `…aBcD` 형태, **prefix 가 없어서 OpenAI/Stripe/GitHub/AWS 어떤 service 인지 모름**. 공격자가 얻는 정보는 "이 환경에 N 개의 secret 이 있고 각 secret 의 last-4 chars" 뿐. Service identification 없이는 사회공학 공격을 만들기 어려움.
- **Opt-in (front>0) 사용자에게 발생 가능한 시나리오**: 사용자가 `preview.front=4` 로 설정 → 누군가 추출한 `sk-proj…aBc1` 으로 OpenAI key 식별 후 사회공학 시도 ("hey, your OpenAI key got rotated, DM me"). 이 시나리오는 사용자가 명시적으로 visibility 를 선택했을 때만 발생.
- **Why this still matters**: 사용자가 trade-off 모르고 opt-in 했거나, vault.db 를 너무 쉽게 다루게 될 수 있음.
- **Mitigation**:
  1. **Default 가 안전** — front=0 이라 OOTB 로는 service identification 불가.
  2. `SECURITY.md` section "What does my vault.db reveal if leaked?" — default 와 opt-in 시나리오를 각각 명시.
  3. `tene config preview.front=N` 명령에 **명시적 confirm 프롬프트** 추가 (plan.md F1 step 7 과 동일 문구, plan 이 source of truth):
     ```
     WARNING: setting preview.front > 0 will expose API key prefixes (sk-, ghp_, AKIA...) in vault.db.
     This makes service identification possible if vault.db leaks. Continue? [y/N]
     ```
     `--force` 플래그로 스킵 가능 (F8 `audit prune --force` 와 동일 컨벤션).
  4. `tene init` 출력에 1-line 안내 (default 가 safe 임을 강조).
  5. `tene config preview.enabled=false` 으로 완전 비활성 (privacy-first 사용자).
  6. README "Security" section 에 trade-off matrix 게시 (front=0 vs front=4 비교표).
  7. vault.db 파일 권한 0600 강제 (existing).
  8. `.gitignore` 에 `.tene/` 강제 (existing).
  9. Threat model boundary: "tene protects against passive attackers (file at rest); tene does NOT protect against active attackers who already have your vault.db and your machine."

### Failure Mode H (F8 — NEW) — Audit log warning becomes spam

- **Symptom**: Warning fires on every command after 50 MB threshold → user trains themselves to ignore stderr → misses important warnings.
- **Why fails**: Warning frequency too high.
- **Mitigation**: warn at most once per 24h per machine via sentinel `~/.tene/.audit-warned-<dir-hash>` with timestamp. `--quiet` flag suppresses entirely. After successful `tene audit prune`, sentinel reset.

### Failure Mode I (F1 — NEW) — Schema migration races

- **Symptom**: Two concurrent `tene` commands on the same vault.db trigger migration simultaneously → SQLite locked errors, or partial preview population.
- **Why fails**: No serialization on migration boundary.
- **Mitigation**: `migrate()` uses `BEGIN IMMEDIATE TRANSACTION` to acquire write lock. `ALTER TABLE` is idempotent (`IF NOT EXISTS` semantics emulated by checking `PRAGMA table_info` first). `fill-previews` is opportunistic — if it fails partway, next run resumes from where it left off.

---

## 6. GTM / Release Notes Angle

### Headline draft

> "tene 0.x — CLI without the password fatigue (without losing the security)"

### One-paragraph hook

> tene 가 vault 안의 key 이름 목록을 보여주려고 매번 master password 를 묻던 시절은 끝. 이번 릴리즈부터 `tene list`, `tene env list`, `tene permissions` 같은 metadata-tier 명령은 unlock 없이 즉답한다. 동시에 `tene get`, `tene run`, `tene export` 같은 secret-value 접근은 기존 정책을 그대로 유지: AI agent 가 secret 값을 stdout 으로 받는 경로는 명시적 opt-in 없이는 여전히 막혀 있다.

### Blog candidate (optional, separate PDCA)

- Slug: `tene-permission-tiers`
- Category: `tools`
- Tags: `tene` `security` `harness-engineering`
- Angle: "Three tiers, not two — why password fatigue is a UX problem, not a security one"

### Cross-share

- HN: `agent-kay` 댓글 시 자연 노출
- Reddit: r/cursor, r/ClaudeAI (AI 가 키 이름 학습할 수 있다는 점 강조)
- Daily.dev: AI-Safe Secrets squad 포스트

---

## 7. Open Questions — RESOLVED (2026-05-20)

| Q | Decision | Owner |
|---|---|---|
| **Q1**: `tene permissions` 명령 이름 OK? | ✅ 채택. `tene permissions` (top-level subcommand). | user |
| **Q2**: JSON `preview` 처리 — null vs absent vs always-string? | ⚠️ 보안 trade-off 명시적 수락. **vault DB 평문 preview 컬럼 신설**. JSON 은 항상 string 타입. schema v2 migration 필요. **Default (2026-05-20 강화): `front=0, back=4`** — prefix 노출 차단 (sk-, ghp_, AKIA 식별 불가). 사용자가 `tene config preview.front=N` 으로 opt-in 시 prefix 추가 노출. 완전 opt-out 은 `preview.enabled=false`. | user |
| **Q3**: audit log 2배 증가 부담? | ✅ 수락. **추가로 F8 (Audit Log Management) sprint 에 신설** — `tene audit tail/show/prune` + 50MB 임계값 stderr 경고. 자동 삭제 절대 금지. | user |
| **Q4**: `tene init` 출력에 hint 한 줄 vs 전체 표? | ✅ A 채택. hint 한 줄 (preview 안내 1 라인 추가로 총 2 라인 증가). | user |

이 결정들은 master-plan.md §1 RISK + §8 Security Invariants + design.md §0 D6 에 반영. 추가 open question 없음.

---

> **Next phase**: Plan (`plan.md`) — feature 별 step-by-step 구현 단위
