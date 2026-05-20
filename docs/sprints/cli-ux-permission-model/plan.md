# tene CLI UX & AI-Safe Permission Model — Implementation Plan

> **Sprint ID**: `cli-ux-permission-model`
> Master Plan reference: [master-plan.md](master-plan.md)
> PRD reference: [prd.md](prd.md)

---

## 0. Scope Reminder

Sprint 분석 결과 가설을 뒤집은 5가지 사실 (master-plan §1 WHY):
1. `secrets.name` 컬럼이 평문이므로 metadata-only read path 구현 가능
2. `go-keyring` 으로 macOS/Linux/Windows keychain 이 이미 다 지원됨
3. `STDOUT_SECRET_BLOCKED` + `--unsafe-stdout` + `TENE_ALLOW_STDOUT_SECRETS` 가 이미 구현됨
4. `pkg/domain/VaultKeyMeta` 가 이미 존재 (cloud push 용) — 재사용
5. `env list`, `env switch`, `env create`, `env delete` 는 이미 password-free

**추가 결정 (2026-05-20, Q1-Q4 사용자 답변)**:
- **Q2 영향**: `list` 가 unlock 없이 **partial value preview 도 보여주는** 방향으로 reshape. vault DB schema v1 → v2 (preview 컬럼 평문 추가). `set`/`import` 시 자동 derive. `tene config preview.enabled=false` 로 opt-out. **이 결정은 sprint 의 LOC 추산을 ~700 늘림**.
- **Q3 영향**: 신규 F8 (Audit Log Management) 추가. `tene audit tail/show/prune` + 50MB 임계값 경고. 자동 삭제 금지.

따라서 sprint 의 실제 작업은:
(a) **schema v2 migration + preview derive** (F1, 기존보다 무거워짐)
(b) declarative permission tier 표현 (F2)
(c) `list` 가 preview 컬럼 직접 읽음 (F3, vault unlock 분기 제거)
(d) audit / docs / discoverability 정비 (F4, F5, F7)
(e) keychain fallback UX (F6)
(f) audit log 관리 (F8, 신규)

---

## 1. Feature Implementation Plan

### F1 — Vault Metadata Read API + Schema v2 Migration (Q2 reshape)

**Goal**: vault 에서 key name + version + updated_at + **preview** 를 읽는 API. `encrypted_value` 컬럼은 절대 SELECT 안 함. Schema v1 → v2 migration 으로 preview 평문 컬럼 추가. 기존 secret 의 preview 는 lazy/명시 채움.

#### Files

| Action | Path | LOC est. |
|---|---|---:|
| Modify | `internal/vault/vault.go` | +60 (ListSecretMetadata, migrate v2, BackfillPreviews) |
| Modify | `internal/vault/schema.go` | +25 (schema v2 SQL, migration step) |
| Modify | `pkg/domain/vault_key_metadata.go` | +3 (`Preview string` 필드 additive) |
| Add | `pkg/crypto/preview.go` | +35 (DerivePreview helper) |
| Add | `pkg/crypto/preview_test.go` | +90 (edge cases) |
| Add | `internal/cli/migrate.go` | +120 (`tene migrate fill-previews` cmd) |
| Add | `internal/cli/config.go` | +180 (`tene config <k>=<v>` for preview.enabled / preview.front / preview.back / audit.warnAtMB) |
| Add | `internal/vault/migration_test.go` | +150 (v1→v2 cases) |
| Add | `internal/vault/metadata_test.go` | +80 |
| Modify | `internal/cli/set.go` | +15 (derive + store preview on encrypt) |
| Modify | `internal/cli/import_cmd.go` | +10 (same for batch) |

Total LOC est. ~770.

#### Steps

1. **Pre-flight** (필수, 코드 작성 전):
   - `grep -rn "VaultKeyMeta" /Users/popup-kay/Documents/GitHub/agentkay/tene-cloud/`
   - `grep -rn "encrypted_value\|\"secrets\"" /Users/popup-kay/Documents/GitHub/agentkay/tene-cloud/`
   - `cd /Users/popup-kay/Documents/GitHub/agentkay/tene-cloud && go build ./... && go test ./...` baseline 녹색 확인
2. **`pkg/crypto/preview.go`**:
   ```go
   // DerivePreview returns "abcd…wxyz" where front+back chars are exposed.
   // front, back must each be 0-8. Combined max is 8 (hard cap).
   // For values shorter than front+back+1, returns "*****".
   func DerivePreview(plaintext string, front, back int) string
   ```
   - Edge cases: empty plaintext → "", front=0 → "…wxyz", front+back > len → "*****" (no leak).
   - Hard validation: `front+back > 8` → error or clamp to 8.
3. **`internal/vault/schema.go`** schema v2:
   ```sql
   -- v1 schema unchanged.
   -- v2 adds: ALTER TABLE secrets ADD COLUMN preview TEXT NOT NULL DEFAULT '';
   ```
   - `currentSchemaVersion = 2`
   - Migration step (in `migrate()` after `getSchemaVersion`):
     ```go
     if version == 1 {
         // BEGIN IMMEDIATE for serialization
         // ALTER TABLE if column not present
         // SetMeta schema_version = 2
     }
     ```
   - Idempotent: PRAGMA table_info("secrets") 로 preview column 존재 확인 후 ALTER 시도.
4. **`internal/vault/vault.go`** new methods:
   ```go
   func (v *Vault) ListSecretMetadata(env string) ([]domain.VaultKeyMeta, error)
   func (v *Vault) UpdateSecretPreview(name, env, preview string) error
   func (v *Vault) BackfillPreviews(env string, derive func(plaintext, name string) (string, error)) (int, error)
   ```
   - `ListSecretMetadata` SQL: `SELECT name, version, updated_at, preview FROM secrets WHERE environment = ?`
   - 절대 `encrypted_value` 포함 금지.
5. **`pkg/domain/vault_key_metadata.go`** additive:
   ```go
   type VaultKeyMeta struct {
       Name      string    `json:"name"`
       Version   int       `json:"version"`
       UpdatedAt time.Time `json:"updated_at"`
       Preview   string    `json:"preview"`           // NEW (Q2) — always emit. Empty string when preview disabled. Q2 "always-string" contract.
   }
   ```
   - **JSON tag: `omitempty` 사용 금지** (Q2 "always-string" 계약). Empty 상태에서도 `"preview": ""` 가 무조건 emit 되어야 함. design §2.3 + design §8.5 D-1 + design §10.2 와 정합.
6. **`internal/cli/migrate.go`** new command:
   ```
   tene migrate                  # show migration status
   tene migrate fill-previews    # unlock once, derive preview for all secrets w/ empty preview
   ```
   - `fill-previews` 는 PermSecretRead tier (unlock 필요)
   - 진행 표시: "Filled previews for N of M secrets in env=default..."
7. **`internal/cli/config.go`** new command:
   ```
   tene config                              # print current effective config
   tene config preview.enabled=false        # turn preview off entirely
   tene config preview.front=4              # opt-in: also show first 4 chars (prefix). default=0
   tene config preview.back=4               # default. tweak if needed (max 8)
   tene config audit.warnAtMB=100
   ```
   - Storage: `vault_meta` table (`config.preview.enabled` 등). 새 테이블 신설 안 함.
   - Validation: preview.front ∈ [0,8], preview.back ∈ [0,8], front+back ≤ 8.
   - Defaults: **preview.enabled=true, preview.front=0, preview.back=4** (2026-05-20 결정 — prefix 노출 차단이 기본).
   - **명시적 confirm 필요**: `preview.front` 값을 0 → N(>0) 으로 바꿀 때는 다음 프롬프트:
     ```
     WARNING: setting preview.front > 0 will expose API key prefixes (sk-, ghp_, AKIA...) in vault.db.
     This makes service identification possible if vault.db leaks. Continue? [y/N]
     ```
     `--force` 플래그로 스킵 가능 (스크립트용 / F8 `audit prune --force` 와 동일 컨벤션).
   - PermMetaRead tier.
8. **`internal/cli/set.go`** modify `runSet`:
   - 기존 `crypto.Encrypt` 직후, `cfg.PreviewEnabled` 라면 `crypto.DerivePreview(value, cfg.PreviewFront, cfg.PreviewBack)` → `v.UpdateSecretPreview(name, env, preview)` 한 번 더 호출. 또는 SetSecret 시그니처에 preview 인자 추가 (선호).
   - 트랜잭션: SetSecret 내부에서 ciphertext + preview 를 같은 INSERT/UPDATE 에 묶기.
9. **`internal/cli/import_cmd.go`** modify `importDotEnv`:
   - 동일 — derive + store preview 한 번에.
10. **Tests** (`internal/vault/migration_test.go`):
    - **TestMigrate_V1_to_V2_AddsPreviewColumn** — v1 vault 열기 → schema_version meta = "2" + secrets 테이블에 preview 컬럼 존재
    - **TestMigrate_V2_Idempotent** — 두 번 열어도 안전, alter 두 번 안 함
    - **TestMigrate_PreservesData** — v1 vault 에 secret 3개 → migrate 후 3개 모두 존재, encrypted_value 무손상
    - **TestBackfillPreviews_Populates** — fill-previews 실행 후 모든 secret preview 채워짐
    - **TestBackfillPreviews_Resumes** — 절반 채운 상태에서 중단 → 다시 실행 시 빈 것만 채움
    - **TestMigrate_ConcurrentLock** — 두 process 동시 진입 → 한쪽만 migrate, 다른쪽 대기

#### DoD

- `go test ./internal/vault/... ./pkg/domain/... ./pkg/crypto/...` 100% pass
- v1 vault.db (기존 사용자 시뮬레이션) → 새 바이너리로 열기 → preview 컬럼 자동 추가, 데이터 무손상
- `cd ../tene-cloud && go build ./...` exit 0
- `grep -A 5 "ListSecretMetadata" internal/vault/vault.go | grep -c "encrypted_value"` = 0
- DerivePreview hard cap 검증 (front+back > 8 → error)

#### Regression risk

- **NEW (Q2)**: schema migration 실패 시 사용자 vault 손상 — `BEGIN IMMEDIATE` + idempotent ALTER + 백업 권고 (`SECURITY.md` 에 "before upgrade, copy your vault.db" 명시) — but tene 는 single-file SQLite 이므로 사용자가 백업 쉬움
- **NEW (Q2)**: `pkg/domain/VaultKeyMeta` 변경 시 tene-cloud sync push payload 가 preview 받음. cloud server 가 이를 그대로 저장하면 dashboard 에 노출 — cloud 의 책임. 이 sprint 에서는 tene-cloud 가 preview 필드를 무시하는지 확인만 (`grep -rn "Preview" tene-cloud/`).
- `pkg/domain` 변경 시 tene-cloud 빌드 — pre-flight 로 차단
- SQL injection 자동 차단 (parameterized) — 기존 패턴 그대로

---

### F2 — Permission Tier Model (declarative)

**Goal**: `PermLevel` enum + command→tier map. root.go 의 PreRunE 가 tier 를 읽어 적절한 unlock 정책을 결정.

#### Files

| Action | Path | LOC est. |
|---|---|---:|
| Add | `internal/auth/permissions.go` | +120 |
| Add | `internal/auth/permissions_test.go` | +90 |
| Modify | `internal/cli/root.go` | +45 (PreRunE hook + helper) |

#### Steps

1. Create `internal/auth/permissions.go`:
   ```go
   package auth

   type PermLevel int

   const (
       PermMetaRead     PermLevel = iota // list, env list, permissions
       PermSecretWrite                   // set, import, delete, env create/delete
       PermSecretRead                    // get, export, run, passwd
   )

   func (p PermLevel) String() string { ... }
   func (p PermLevel) RequiresUnlock() bool {
       return p != PermMetaRead
   }
   ```
2. Add command tier table:
   ```go
   var CommandTier = map[string]PermLevel{
       // PermMetaRead — metadata only, no plaintext value exposure
       "list":         PermMetaRead, // F3: reads preview column directly
       "env":          PermMetaRead, // catch-all root for `env *`
       "env list":     PermMetaRead,
       "env create":   PermMetaRead,
       "env delete":   PermMetaRead, // 환경 자체 삭제는 metadata 작업
       "permissions":  PermMetaRead, // F5 (NEW)
       "whoami":       PermMetaRead,
       "version":      PermMetaRead,
       "update":       PermMetaRead,
       "completion":   PermMetaRead,
       "logout":       PermMetaRead, // cloud session logout, no vault unlock needed
       "audit":        PermMetaRead, // F8 catch-all root for `audit *`
       "audit tail":   PermMetaRead, // F8 (NEW)
       "audit show":   PermMetaRead, // F8 (NEW)
       "audit prune":  PermSecretWrite, // F8 (NEW) — DELETE from audit_log requires write tier
       "config":       PermMetaRead, // F1 (NEW) — set/get config keys; preview.front>0 has its own confirm
       "migrate":      PermMetaRead, // F1 (NEW) — schema/preview migration, no plaintext exposure

       // PermSecretWrite — encrypts new plaintext into vault. encKey required.
       "set":          PermSecretWrite,
       "import":       PermSecretWrite,
       "delete":       PermSecretWrite,
       "init":         PermSecretWrite, // 새 vault 생성 시 password 설정 필요

       // PermSecretRead — decrypts and returns plaintext. STDOUT_SECRET_BLOCKED applies.
       "get":          PermSecretRead,
       "export":       PermSecretRead,
       "run":          PermSecretRead,
       "passwd":       PermSecretRead, // master password rotation needs to decrypt current vault
       "recover":      PermSecretRead, // recovery flow re-derives + re-encrypts
   }
   // Total: 26 entries (16 PermMetaRead + 5 PermSecretWrite + 5 PermSecretRead). Source of truth — design.md §1.1 CommandTier diagram must mirror byte-by-byte.

   func TierFor(cmdPath string) (PermLevel, bool)
   ```
3. `internal/cli/root.go`:
   - rootCmd 에 `PersistentPreRunE` 추가 — 호출되는 command 의 tier 를 `CommandTier` 에서 lookup, 미선언 시 panic (G4 강제).
   - tier 가 PermMetaRead 이면 `loadOrPromptMasterKey` 미호출. tier 가 SecretWrite/SecretRead 이면 현재 흐름 유지.
4. `internal/auth/permissions_test.go`:
   - 모든 rootCmd subcommand 가 CommandTier 에 entry 보유 — `TestAllCommandsHaveTier` (reflection 으로 rootCmd.Commands() 순회)
   - `TierFor("list")` = PermMetaRead
   - `TierFor("nonexistent")` = false

#### DoD

- `go test ./internal/auth/...` 100% pass
- G4 (Permission Tier Coverage) 통과
- panic-on-missing 작동 — 새 명령 추가 시 등록 강제

#### Regression risk

- PersistentPreRunE 가 cobra 의 기존 hook 와 충돌 가능 — 현재 `helpCmd` / `versionCmd` 는 별도 흐름이므로 영향 없음. 다만 `tene --version` 같은 flag 처리는 cobra 자동, hook 안 거침.

---

### F3 — `list` Reads Preview Column Directly (Q2 reshape)

**Goal**: `tene list` 가 unlock 없이 vault DB 의 preview 컬럼을 그대로 출력. unlock 분기 자체가 사라짐 — 단순화.

#### Files

| Action | Path | LOC est. |
|---|---|---:|
| Modify | `internal/cli/list.go` | +40 / -60 (net -20, 단순화) |
| Modify | `internal/cli/cli_test.go` | +30 (new cases) |
| Add | `internal/cli/list_test.go` | +130 |

#### Steps

1. `internal/cli/list.go` 재작성 — vault unlock 분기 완전 제거:
   ```go
   func runList(cmd *cobra.Command, args []string) error {
       app, err := loadApp()
       if err != nil { return err }
       defer app.Vault.Close()

       env := resolveEnv(app)

       // NEW: vault.db 의 preview 컬럼을 그대로 읽음
       metadata, err := app.Vault.ListSecretMetadata(env)
       if err != nil { return err }

       // No unlock, no decrypt, no Argon2id. Just present.
       // ... render JSON or table
   }
   ```
2. JSON shape (Q2 결정 — preview 는 항상 string):
   - `name` — string, 항상 present
   - `preview` — string, 항상 present. 빈 secret 또는 `preview.enabled=false` 시 `""`
   - `version` — int
   - `updatedAt` — RFC3339 string
   - Top-level: `ok`, `project`, `environment`, `secrets`, `count`
   - 제거: `unlocked` 필드 (Q2 결정으로 의미 없어짐 — list 는 unlock 개념 자체와 무관)
3. Text mode:
   - 3컬럼: `NAME | PREVIEW | UPDATED`
   - preview 가 `""` 일 때 표시: `-` (정렬 깨짐 방지)
   - preview.enabled=false 시 모든 row 가 `-` 가 됨 → footer 에 hint: "Preview disabled (tene config preview.enabled=true to re-enable)"
4. 빈 secret/legacy 처리:
   - v1 vault 가 migrate 직후, 기존 secret 들의 preview 는 빈 문자열 — `tene migrate fill-previews` 실행 권고 footer
5. Tests (`internal/cli/list_test.go`):
   - **TestList_NoUnlock_ShowsPreview** — flagNoKeychain + 3 secrets (preview 채워짐) → 출력에 preview 3개 모두 표시. Argon2id 호출 0회 (timing 기반 보조 검증)
   - **TestList_PreviewDisabled_ShowsDashes** — config preview.enabled=false → preview 컬럼 = `-`
   - **TestList_LegacyVault_EmptyPreview** — migrate 후 fill-previews 미실행 → preview 빈 문자열, footer hint
   - **TestList_EmptyEnv** — 0 secrets → "No secrets in..." 메시지
   - **TestList_JSON_PreviewAlwaysString** — JSON `preview` 필드 type 검증 (never null/absent)
   - **TestList_AfterFillPreviews_PopulatesAll** — fill-previews 직후 list → 모두 preview 채워짐
6. Bench:
   - `BenchmarkListWithPreview` — 100 entries → p50 < 15ms (Argon2id 부담 0)

#### DoD

- 위 6개 테스트 + 1개 bench 통과
- 기존 `set_get_test.go` 의 list 관련 케이스 회귀 없음 — preview 가 항상 string 이라 type assertion 단순
- 성능: p50 < 15ms

#### Regression risk

- **NEW (Q2)**: legacy v1 vault 의 첫 list 호출 시 preview 가 비어있어 사용자가 혼란 → footer hint + migrate 명령 안내
- JSON shape 변경: `preview` type 이 `string | null` → `string` 으로 단순화. CHANGELOG `### Changed` entry 필수
- 텍스트 출력 컬럼 — 항상 3컬럼이지만 preview = `-` 일 때 시각적 변동. `awk '{print $1}'` 같은 파싱은 안전 (NAME 컬럼 위치 무변동)

---

### F4 — Audit Logging by Permission Tier

**Goal**: 모든 CLI 호출이 1 row 의 audit_log 에 tier 명시로 기록.

#### Files

| Action | Path | LOC est. |
|---|---|---:|
| Modify | `internal/cli/root.go` | +30 (PreRunE 확장) |
| Modify | `internal/vault/vault.go` | (변경 없음, AddAuditLog 그대로 사용) |
| Add | `internal/cli/audit_test.go` | +90 |

#### Steps

1. F2 의 PersistentPreRunE 확장:
   ```go
   // after tier lookup
   defer func() {
       if app != nil && app.Vault != nil {
           action := fmt.Sprintf("cli.%s.%s", tier.String(), cmd.Name())
           _ = app.Vault.AddAuditLog(action, strings.Join(args, " "), "")
       }
   }()
   ```
2. **주의**: 기존 명령별 audit (예: `secret.write`, `vault.init`) 는 그대로 유지. 새 entry 는 **추가**됨. 한 명령 = 2 audit row (cli.* prefix + 기존 action).
3. Tests:
   - **TestAudit_List_MetaReadEntry** — `tene list` 호출 후 audit_log 에 `cli.metaread.list` 1행
   - **TestAudit_Set_BothEntries** — `tene set X Y` 호출 후 `cli.secretwrite.set` + `secret.write` 둘 다 존재
   - **TestAudit_Init_VaultInitPreserved** — 기존 `vault.init` entry 보존

#### DoD

- G7 (Audit Log Completeness) 통과
- 기존 audit 관련 테스트 회귀 없음

#### Regression risk

- audit_log 크기 약 2배 증가. SQLite 성능 영향 미미 (insert 1ns). 사용자가 audit_log query 스크립트 보유 시 새 row 가 노이즈. **PRD §7 Q3 사용자 확인 필요**.

---

### F5 — `tene permissions` Info Command + Help Integration

**Goal**: `tene permissions` 명령으로 tier 표를 출력. `tene --help` 에 요약 포함.

#### Files

| Action | Path | LOC est. |
|---|---|---:|
| Add | `internal/cli/permissions.go` | +80 |
| Modify | `internal/cli/root.go` | +15 (rootCmd.Long 갱신, AddCommand) |
| Modify | `README.md` | +30 (Permission Tiers section) |
| Modify | `CHANGELOG.md` | +25 (Unreleased 섹션 entries) |
| Add | `internal/cli/permissions_test.go` | +60 |

#### Steps

1. `internal/cli/permissions.go`:
   ```go
   var permissionsCmd = &cobra.Command{
       Use:   "permissions",
       Short: "Show which commands require master password",
       RunE:  runPermissions,
   }
   ```
   - Iterate `auth.CommandTier` map → 3개 그룹 (MetaRead, SecretWrite, SecretRead) 으로 정렬
   - Text mode: 3컬럼 표 (`COMMAND | TIER | UNLOCKS?`)
   - JSON mode: `{ ok: true, commands: [{ name, tier, requiresUnlock }] }`
2. `root.go`:
   - `rootCmd.AddCommand(permissionsCmd)` 추가
   - `Long` 텍스트 끝에 "Run `tene permissions` to see which commands require a master password." 한 줄 추가
3. `README.md`:
   - 새 H2 섹션 "Permission Tiers" 추가, ~150 단어
   - 3 tier 설명 + 예시 + AI-safety 노트
4. `CHANGELOG.md` `[Unreleased]`:
   - `### Added`: `tene permissions` command, `Vault.ListSecretMetadata`, audit cli.* events
   - `### Changed`: `tene list` no-decrypt by default, JSON `preview` nullable
   - `### Security`: explicit permission tier model documented
5. Tests:
   - **TestPermissions_Text_ContainsAllCommands** — 출력에 list, set, get 등 모든 명령 포함
   - **TestPermissions_JSON_Shape** — JSON shape valid
   - **TestPermissions_NoUnlock** — 호출 시 keychain.Load 호출 안 함

#### DoD

- `tene permissions` 호출 시 3 tier 표 출력
- README + CHANGELOG entry 머지
- 명령 자체가 PermMetaRead tier — 즉 password 안 묻음

#### Regression risk

- 새 명령 추가 시 conflict 없음 (`permissions` 충돌 가능성 grep 확인 — `grep -rn "\"permissions\"" internal/cli/`)
- README 톤 일관성 — 기존 섹션 스타일 따라 작성

---

### F6 — Keychain Fallback UX Polish

**Goal**: file-keystore fallback 발동 시 일회성 stderr notice.

#### Files

| Action | Path | LOC est. |
|---|---|---:|
| Modify | `internal/cli/root.go` | +20 (notice helper in loadApp) |
| Modify | `internal/keychain/keychain.go` | +5 (NewStore 가 fallback 발생 시 sentinel 정보 반환) |
| Modify | `internal/keychain/fallback.go` | (변경 없음 가능성) |
| Add | `internal/cli/fallback_notice_test.go` | +50 |

#### Steps

1. `NewStore(projectPath)` 시그니처 변경 보류 (additive 방법 선호):
   - 새 함수 `NewStoreWithStatus(projectPath) (KeyStore, FallbackInfo)` 추가
   - `FallbackInfo{ Used bool, Reason string, Path string }`
2. `loadApp()` 에서:
   - file fallback 발동 시 `~/.tene/.fallback-warned-<dir-hash>` sentinel 확인
   - 없으면 stderr 에 "Note: using file-based keystore at ~/.tene/keyfile (OS keychain unavailable). This message shows once." 출력 + sentinel touch
   - `flagQuiet` 시 출력 skip (sentinel 도 안 만듦 — 다음 non-quiet 호출에 보여줘야 함)
3. Tests:
   - **TestFallback_FirstCall_PrintsNotice** — sentinel 없으면 notice
   - **TestFallback_SecondCall_Quiet** — sentinel 있으면 notice 없음
   - **TestFallback_QuietFlag_NoNotice** — `--quiet` 시 notice 없음 + sentinel 안 생성

#### DoD

- 위 3개 테스트 통과
- macOS 정상 환경에서는 sentinel 미생성 (fallback 안 함)

#### Regression risk

- sentinel 파일 위치 — `~/.tene/.fallback-warned` 가 vault 디렉터리와 분리되어 있는지 확인
- CI 환경 (Docker) 에서 매번 새 컨테이너 → 매번 notice 출력. 의도된 동작.

---

### F7 — `tene init` First-Run UX Update

**Goal**: init 성공 출력에 permission model + preview 동작 안내 (총 3줄 추가). Q2 final default `front=0, back=4` 반영 (prefix 노출 차단).

#### Files

| Action | Path | LOC est. |
|---|---|---:|
| Modify | `internal/cli/init.go` | +8 (출력 텍스트 + preview note) |
| Modify | `internal/cli/init_test.go` | +30 |

#### Steps

1. `internal/cli/init.go` Line ~226 ("Tip: No server needed..." 다음 위치) 에 3줄 추가 (Q2 final default `front=0, back=4` 반영):
   ```go
   fmt.Println("       Run `tene permissions` to see which commands need your password.")
   fmt.Println("       `tene list` shows last 4 chars of each value by default (no prefix exposed).")
   fmt.Println("       Disable: `tene config preview.enabled=false`  |  Opt-in to prefix: `tene config preview.front=N`")
   ```
2. Tests:
   - **TestInit_OutputContainsPermissionHint** — init 출력 stdout 에 "tene permissions" 키워드 포함
   - **TestInit_OutputContainsPreviewNote** — init 출력에 "last 4 chars" + "no prefix exposed" 포함
   - **TestInit_DoesNotMentionFirstFourLastFour** — 옛 텍스트 잔재 검출 회귀 가드 (negative test: "first-4" / "first 4 + last 4" 문자열 0건)

#### DoD

- 출력 verbose 증가 ≤ 3 line
- 기존 `init_test.go` 100% pass

#### Regression risk

- JSON output 영향 없음 (출력은 text mode 만 변경)
- 본문 verbose 증가 — first-time user 가 압도되지 않도록 톤 검토

---

### F8 — Audit Log Management (Q3 decision — NEW)

**Goal**: `tene audit tail|show|prune` 서브명령 + 50MB 임계값 stderr 경고 (24h 1회).

#### Files

| Action | Path | LOC est. |
|---|---|---:|
| Add | `internal/audit/manager.go` | +250 (size check, query, prune logic) |
| Add | `internal/audit/manager_test.go` | +180 |
| Add | `internal/cli/audit.go` | +200 (cobra subcmds: tail, show, prune) |
| Add | `internal/cli/audit_test.go` (F4 + F8 통합) | +60 |
| Modify | `internal/cli/root.go` | +30 (threshold warning hook in PersistentPreRunE) |
| Modify | `internal/vault/vault.go` | +25 (`GetAuditLogSize`, `QueryAuditLog`, `PruneAuditLog`) |
| Modify | `internal/config/` | (F1 의 config 패키지에 audit.warnAtMB 키 추가, 별도 LOC 없음) |

Total LOC est. ~745.

#### Steps

1. **`internal/audit/manager.go`** new package:
   ```go
   package audit

   type Manager struct { vault *vault.Vault }

   type LogEntry struct {
       Timestamp time.Time
       Action    string
       Resource  string
       Details   string
   }

   func (m *Manager) Tail(n int) ([]LogEntry, error)
   func (m *Manager) Show(filter Filter) ([]LogEntry, error)
   func (m *Manager) SizeBytes() (int64, error)
   func (m *Manager) Prune(olderThan time.Duration) (deleted int64, err error)

   type Filter struct {
       Since       time.Time
       Until       time.Time
       ActionGlob  string // e.g. "cli.secretread*"
       Resource    string
   }
   ```
2. **`internal/vault/vault.go`** new methods:
   - `GetAuditLogSize() (int64, error)` — `SELECT SUM(length(action)+length(resource_name)+length(details)+8) FROM audit_log` (rough byte estimate; or use SQLite page size approximation)
   - `QueryAuditLog(filter)` — parameterized SELECT with LIMIT/ORDER
   - `PruneAuditLog(before time.Time)` — `DELETE FROM audit_log WHERE timestamp < ?`
3. **`internal/cli/audit.go`** subcommands:
   ```
   tene audit                  # alias to "tail -n 20"
   tene audit tail -n N
   tene audit show --since 1d --filter cli.secretread* --resource KEY
   tene audit prune --older-than 30d [--force]
   ```
   - `prune` 는 dry-run 표시 → 확인 prompt → 삭제. `--force` 로 skip 가능.
   - 모든 audit 서브명령은 PermMetaRead tier (unlock 없음).
4. **`internal/cli/root.go`** threshold hook:
   - PersistentPreRunE 끝에 size check (skip if `--quiet`):
     - `audit.Manager.SizeBytes()` 호출
     - threshold = config.audit.warnAtMB * 1024 * 1024 (default 50)
     - sentinel `~/.tene/.audit-warned-<dir-hash>` mtime 확인 → 24h 이내면 skip
     - 초과면 stderr 출력: "Audit log is 52MB. Run 'tene audit prune --older-than 30d' to clean up." + sentinel touch
5. **Tests**:
   - **TestAudit_Tail_ReturnsLastN** — 100 entry 삽입 → tail(10) = 최근 10개
   - **TestAudit_Show_FilterByActionGlob** — `cli.secretread*` 매칭만 반환
   - **TestAudit_Show_FilterBySince** — 시간 필터 정상
   - **TestAudit_Prune_DeletesOlder** — 30일 이전 entry 삭제, 최근 보존
   - **TestAudit_Prune_RequiresConfirm** — `--force` 없이 호출 → prompt
   - **TestAudit_SizeBytes_Reasonable** — 0 entry → ~0, 1000 entry → > 0
   - **TestAudit_ThresholdHook_FiresOnce** — threshold 초과 + sentinel 없음 → stderr 출력 + sentinel 생성
   - **TestAudit_ThresholdHook_SuppressedFor24h** — sentinel 24h 이내 → 출력 없음
   - **TestAudit_ThresholdHook_QuietFlag** — `--quiet` → 출력 없음, sentinel 안 만듦
   - **TestAudit_NoAutoDelete_EverywherePolicy** — Manager 의 어떤 public API 도 DELETE 자동 호출 안 함 (static check)

#### DoD

- 위 10개 테스트 통과
- `tene audit prune` 은 explicit confirm 없이 절대 삭제 안 함
- 50 MB threshold warning 이 24h 내 중복 출력 안 됨
- audit tier = PermMetaRead (password 안 묻음)

#### Regression risk

- **NEW (Q3)**: audit_log table 에 `SELECT` 부하 — index `idx_audit_timestamp` 가 이미 schema 에 있음, 쿼리 빠름
- **NEW (Q3)**: sentinel 파일 시스템 권한 (homedir 쓰기 불가 환경) → fallback 으로 stderr 출력 매번 (annoying but safe)
- prune 의 transaction — 큰 audit_log (100K row) prune 시 lock 시간. `DELETE ... LIMIT N` 으로 batch prune 고려, 그러나 sprint scope 내 simple full-DELETE 우선

---

## 2. Test Strategy Matrix

### L1 — Unit (per-package)

| Package | New tests | Existing tests preserved |
|---|---|---|
| `internal/vault` | `metadata_test.go` — ListSecretMetadata 격리. `migration_test.go` — v1→v2 (Q2) | `vault_test.go` 모두 |
| `internal/auth` | `permissions_test.go` — tier coverage + lookup | (신규 패키지) |
| `internal/audit` | `manager_test.go` — Tail/Show/Prune/SizeBytes (F8) | (신규 패키지) |
| `pkg/domain` | `Preview` 필드 JSON shape 검증 (additive) | — |
| `pkg/crypto` | `preview_test.go` — DerivePreview edge cases (Q2) | `crypto_test.go`, `rotation_test.go` 모두 |
| `internal/keychain` | `keychain_test.go` 확장 (FallbackInfo) | 모두 |
| `internal/config` | config get/set 테스트 (preview.*, audit.*) | (신규) |

### L2 — Integration (CLI subcommand)

| File | New tests |
|---|---|
| `internal/cli/list_test.go` (new) | TestList_NoKeychain_ShowsMetadata, TestList_WithKeychain_ShowsPreviews, TestList_JSON_PreviewNullable, TestList_EmptyEnv |
| `internal/cli/audit_test.go` (new) | TestAudit_List_MetaReadEntry, TestAudit_Set_BothEntries, TestAudit_Init_VaultInitPreserved |
| `internal/cli/permissions_test.go` (new) | TestPermissions_Text_ContainsAllCommands, TestPermissions_JSON_Shape, TestPermissions_NoUnlock |
| `internal/cli/fallback_notice_test.go` (new) | TestFallback_FirstCall_PrintsNotice, TestFallback_SecondCall_Quiet, TestFallback_QuietFlag_NoNotice |
| `internal/cli/cli_test.go` | + TestPermissionTier_AllCommandsRegistered |
| `internal/cli/get_guard_test.go` | (회귀 보존, 변경 없음 — F4 audit row 추가만 검증) |
| `internal/cli/set_get_test.go` | (회귀 보존) |
| `internal/cli/init_test.go` | + TestInit_OutputContainsPermissionHint |

### L3 — End-to-End (manual)

| Scenario | Steps |
|---|---|
| E2E-1 Fresh install | `tene init demo` → `tene set FOO bar` → `tene list` (no prompt) → `tene get FOO` (no prompt due to keychain) → `tene run -- env | grep FOO` |
| E2E-2 No-keychain | `tene init --no-keychain demo2` → set password via env var → `tene list` (no prompt, no preview) → `tene get FOO` (prompt) |
| E2E-3 Permissions command | `tene permissions` → table 출력 → `tene permissions --json | jq` |
| E2E-4 Fallback notice | Linux without libsecret: `tene list` → stderr notice 1회 → `tene list` → silent |
| E2E-5 AI safety | `tene get FOO > /tmp/out` → STDOUT_SECRET_BLOCKED |
| E2E-6 Audit log | After above, `sqlite3 .tene/vault.db "SELECT action FROM audit_log"` → all `cli.*` prefix + 기존 entries |

### L4 — Cross-Repo

| Step | Command |
|---|---|
| F1 직후 | `cd /Users/popup-kay/Documents/GitHub/agentkay/tene-cloud && go build ./...` |
| 통합 | `cd /Users/popup-kay/Documents/GitHub/agentkay/tene-cloud && go test ./...` |
| Sync 테스트 | (cloud 명령 비활성이므로 skip; 활성화되면 push/pull round-trip 확인) |

### L5 — Performance Bench

| Bench | Target |
|---|---|
| `BenchmarkListNoDecrypt` (new) | p50 < 15ms (100 secrets) |
| `BenchmarkListWithDecrypt` (rename existing) | p50 < 80ms (regression watch) |
| `BenchmarkSetSecret` | 회귀 없음 |

---

## 3. Regression Risk Checklist

매 PR review 시 다음을 확인:

### 보안 invariant (master-plan §8 echo)

- [ ] `secrets.encrypted_value` 평문 write 경로 0건
- [ ] `crypto.Decrypt` 호출 전 unlock 검증 100%
- [ ] STDOUT_SECRET_BLOCKED 정책 무변경 (`get_guard_test.go` 4 케이스 pass)
- [ ] 새 외부 네트워크 호출 0건 (`grep -rn "http.Get\|http.Post\|net.Dial" internal/cli pkg/`)
- [ ] AI default 가 secret value 받는 경로 0건
- [ ] Argon2id 파라미터 무변경
- [ ] recovery flow 무변경
- [ ] keychain service name (`"tene" + project-hash`) 무변경

### 호환성

- [ ] CLI flag 이름 변경/제거 0건
- [ ] JSON 기존 key 제거 0건 (additive only)
- [ ] `go.mod` major bump 없음
- [ ] vault.db schema 변경 없음 (schema_version=1 그대로)

### Cross-repo

- [ ] `pkg/domain/vault_key_metadata.go` 변경은 additive (필드 추가만)
- [ ] tene-cloud `go build ./...` exit 0
- [ ] `grep -rn "VaultKeyMeta" ../tene-cloud/` 사용처와 호환

### 성능

- [ ] BenchmarkListNoDecrypt p50 < 15ms
- [ ] BenchmarkListWithDecrypt 회귀 없음
- [ ] BenchmarkSetSecret 회귀 없음

### UX

- [ ] `tene --help` 에 permissions 안내 한 줄
- [ ] `tene init` 출력에 permission hint 한 줄
- [ ] `tene permissions` 호출 가능
- [ ] README "Permission Tiers" 섹션 머지
- [ ] CHANGELOG Unreleased 섹션 entries 머지

### Audit / Observability

- [ ] `audit_log.action` 에 `cli.<tier>.<verb>` 새 prefix 등장
- [ ] 기존 audit action (secret.read, vault.init 등) 보존
- [ ] 한 명령 호출 = 1 cli.* row + N existing rows

---

## 4. Branch + Commit Strategy

- 단일 branch `feature/cli-ux-permission-model` 유지
- Feature 별 commit chunk (squash 가능):
  - `feat(vault): add ListSecretMetadata API` (F1)
  - `feat(auth): permission tier enum + command map` (F2)
  - `feat(cli): list works without master password by default` (F3)
  - `feat(audit): record permission tier per CLI invocation` (F4)
  - `feat(cli): add tene permissions command + README + CHANGELOG` (F5)
  - `feat(keychain): one-time notice on file-store fallback` (F6)
  - `feat(init): mention permission model in success output` (F7)
- 사용자가 검토 후 merge 권한 — 이 sprint 안에서는 자동 merge 없음 (L3)

---

## 5. Execution Order (week-by-week)

> **Alignment**: 이 표는 `master-plan.md §4 Sprint Phase Roadmap` 및 `state JSON sprints.*` 와 1:1 동기화. 변경 시 세 곳 모두 갱신.

### Week 1 — Foundation (F1 + F6 병렬, 의존성 0)

- Day 1: F1 step 1-3 (Vault Metadata API + schema v2 migration ALTER TABLE) — pre-flight `grep -rn "VaultKeyMeta" ../tene-cloud/` + tene-cloud `go build ./...` G3
- Day 2: F1 step 4-7 (`DerivePreview`, `pkg/domain.VaultKeyMeta.Preview` 필드, `tene migrate fill-previews`, `tene config <k>=<v>`)
- Day 3: F1 step 8-9 (`tene set` / `tene import` 의 preview 자동 derive 통합) + unit tests
- Day 4: F6 (Keychain fallback notice) — 완전 독립이므로 F1 진행 중 병렬 가능. sentinel `~/.tene/.fallback-warned-<hash>` 구현
- Day 5: F1+F6 통합 QA, G1/G2/G3/G6/G8/G9 PASS 확인, intermediate PR review

### Week 2 — Permission Model + List 변경 (F2 → F3)

- Day 1-2: F2 (Permission Tier Model) — `internal/auth/permissions.go`, CommandTier map **26 entries** (16 PermMetaRead + 5 PermSecretWrite + 5 PermSecretRead), `PersistentPreRunE` hook, G4 panic-on-missing
- Day 3-4: F3 (`list` reads preview column directly) — F1+F2 활용. JSON shape: `preview` 필드 항상 string (omitempty 금지). bench p50 < 15ms (G6)
- Day 5: F2+F3 통합 QA, G2/G4/G5/G6 PASS 확인

### Week 3 — Audit + Discoverability + Init (F4 → F8 → F5 → F7)

- Day 1: F4 (Audit Logging by Permission Tier) — `cli.<tier>.<verb>` row 추가, F2 의 PreRunE 확장. G7 PASS
- Day 2: F8 (Audit Log Management) — `tene audit tail/show/prune` 서브명령 + 50MB 임계값 stderr 1회 경고 + `--force` skip. G10 PASS (자동 삭제 0건). depends on F4.
- Day 3: F5 (`tene permissions` 명령 + README + CHANGELOG + SECURITY.md update)
- Day 4: F7 (`tene init` 3줄 hint) — depends on F5. 새 default `front=0, back=4` 안내 텍스트
- Day 5: 전체 QA (G1-G10 all PASS), bench, E2E scenario, PR finalize, archive 사용자 승인 대기

### Cross-week QA / Gates Checklist

| Week | Required gates | When |
|---|---|---|
| 1 | G1 (Security), G2 (Backward compat), G3 (Cross-repo), G6 (List perf preview readiness), G8 (Schema migration safety), G9 (Preview privacy hard cap) | F1 + F6 완료 시 |
| 2 | G2, G4 (Permission tier coverage), G5 (STDOUT_SECRET_BLOCKED regression), G6 (List p50 < 15ms 정식 측정) | F2 + F3 완료 시 |
| 3 | G7 (Audit log completeness), G10 (Audit auto-delete prohibition), 전체 G1-G10 재검 | F4 + F8 + F5 + F7 완료 시 |

---

> **Next phase**: Design (`design.md`) — 기술 디자인, sequence/class diagram, API contract
