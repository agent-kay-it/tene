# Tene Cloud 서비스 상세설계서

> **버전**: v1.0  
> **작성일**: 2026-04-07  
> **작성자**: AI Agent Team (10명 전문 에이전트 종합)  
> **상태**: Design Phase 완료  
> **선행 문서**: `docs/00-pm/tene-cloud.prd.md`, `docs/01-plan/features/tene-cloud.plan.md`

---

## Context Anchor

| Dimension | Content |
|-----------|---------|
| **WHY** | 로컬 전용 CLI의 한계(동기화, 공유, 백업) 해소 + 유료 전환으로 지속 가능한 비즈니스 모델 구축 |
| **WHO** | 1차: 멀티 디바이스 솔로 개발자, 2차: 소규모 개발팀(2-10명), 3차: AI-first 개발 조직 |
| **RISK** | Cloud 수요 부족, X25519 구현 보안 오류, LemonSqueezy 정산 지연, Vault Sync 충돌/데이터 손실 |
| **SUCCESS** | MRR $375+, Pro 유료 전환 75+, Push/Pull 성공률 99.5%, API P99 < 500ms |
| **SCOPE** | Phase 2a(Solo Sync 4주) + Phase 2b(Team 3주) + Phase 2c(안정화 1주) = 총 8주 |

---

## Executive Summary

| 관점 | 설명 |
|------|------|
| **Problem** | Tene CLI(v0.9.3) 사용자가 멀티 디바이스 동기화, 팀 시크릿 공유, 중앙 감사 로그 없이 로컬 전용으로만 운영 |
| **Solution** | Zero-Knowledge Envelope 이중 암호화 기반 Cloud Sync + X25519 ECDH 팀 키 공유 (Pro $5/mo, 개인 결제) + 대시보드(app.tene.sh) |
| **Functional UX Effect** | CLI `push/pull` 2개 명령어로 Vault 동기화, 대시보드에서 디바이스/감사/팀 관리, 시크릿 값은 서버에서 절대 노출 불가 |
| **Core Value** | 경쟁사 대비 최저가($5/mo)로 완전한 Zero-Knowledge 보안 + 팀 기능 포함 + AI 에이전트 자동 인식을 유지하며 클라우드 편의성 제공 |

---

## 목차

1. [시스템 아키텍처 개요](#1-시스템-아키텍처-개요)
2. [보안 모델: Zero-Knowledge 아키텍처](#2-보안-모델-zero-knowledge-아키텍처)
3. [Go API 서버 설계](#3-go-api-서버-설계)
4. [CLI 클라우드 명령어 설계](#4-cli-클라우드-명령어-설계)
5. [Vault Sync 엔진 설계](#5-vault-sync-엔진-설계)
6. [PostgreSQL 데이터베이스 설계](#6-postgresql-데이터베이스-설계)
7. [AWS 인프라 설계](#7-aws-인프라-설계)
8. [LemonSqueezy 결제 시스템 설계](#8-lemonsqueezy-결제-시스템-설계)
9. [프론트엔드 대시보드 설계](#9-프론트엔드-대시보드-설계)
10. [QA/테스트 전략](#10-qa테스트-전략)
11. [구현 로드맵](#11-구현-로드맵)

---

## 1. 시스템 아키텍처 개요

### 1.1 전체 구성도

```
                        Internet
                           │
                    ┌──────┴──────┐
                    │  Route 53   │
                    │  tene.sh    │
                    └──────┬──────┘
                           │
          ┌────────────────┼────────────────┐
          │                │                │
     CloudFront       CloudFront          ALB + WAF
     (tene.sh)       (app.tene.sh)   (api.tene.sh)
     랜딩페이지       대시보드          API 서버
          │                │                │
     S3 Bucket        Vercel         ┌──────┴──────┐
     (landing)                       │ ECS Fargate  │
                                     │ Go API 서버  │
                                     └──────┬──────┘
                                            │
                              ┌─────────────┼─────────────┐
                              │             │             │
                         RDS PostgreSQL  S3 Bucket   Secrets Manager
                        (메타데이터 DB) (Vault Blob)  (API 시크릿)
```

### 1.2 기술 스택

| 레이어 | 기술 | 이유 |
|--------|------|------|
| **API 서버** | Go (Echo v4) | CLI와 동일 언어, crypto 코드 공유, 메모리 효율 |
| **DB** | PostgreSQL 16 (RDS db.t4g.micro) | 관계형 데이터, RBAC 쿼리, 감사 로그 |
| **스토리지** | S3 (SSE-S3 AES-256) | 암호화된 vault blob, 버전 관리, lifecycle |
| **결제** | LemonSqueezy (MoR) | Checkout Overlay, 개인 구독, 한국 개인 계정 지원 |
| **대시보드** | Next.js 15 + shadcn/ui + TanStack Query v5 | App Router, 기존 디자인 시스템 유지 |
| **인증** | ES256 JWT + OAuth 2.0 (GitHub, Google) + PKCE | 비대칭 서명, 개발자 대상 패스워드 없는 인증 |
| **IaC** | Terraform (12 모듈) | 재현 가능한 인프라, 환경 분리 |
| **CI/CD** | GitHub Actions + OIDC → ECR → ECS | 장기 크레덴셜 제거 |

### 1.3 모노레포 구조 (확장)

```
tene/
├── cmd/
│   ├── tene/              ← CLI 엔트리포인트 (기존)
│   └── server/            ← Cloud API 엔트리포인트 (신규)
├── internal/
│   ├── crypto/            ← CLI + Server 공유 (확장: X25519, SyncEnvelope)
│   ├── vault/             ← CLI + Server 공유
│   ├── cli/               ← CLI 전용 (확장: login, push, pull, team, billing)
│   ├── api/               ← Server 전용 (신규)
│   │   ├── server.go      ← Echo 인스턴스, 미들웨어, 라우팅
│   │   ├── handler/       ← 핸들러 (auth, vault, team, billing, device, audit)
│   │   ├── middleware/     ← JWT Auth, Rate Limit, CORS, Security Headers
│   │   ├── storage/       ← S3 클라이언트
│   │   └── response/      ← 표준 응답 포맷
│   ├── auth/              ← 인증 (신규: OAuth, JWT, Device Key)
│   ├── sync/              ← Push/Pull 동기화 (신규: Envelope, Conflict, Merge)
│   ├── billing/           ← LemonSqueezy 통합 (신규)
│   ├── domain/            ← 도메인 모델 + 센티넬 에러 (신규)
│   ├── config/            ← 설정 (확장: CloudConfig)
│   ├── keychain/          ← OS Keychain (확장: JWT 토큰, X25519 키)
│   └── recovery/          ← BIP-39 복구 (기존)
├── apps/
│   ├── web/               ← 랜딩페이지 (기존)
│   └── dashboard/         ← 대시보드 (신규: Next.js)
├── infra/
│   └── terraform/         ← AWS IaC (신규: 12 모듈)
│       ├── modules/       ← vpc, ecs, alb, rds, s3, ecr, route53, acm, iam, nat, waf, secrets
│       └── environments/  ← staging/, prod/
├── migrations/            ← PostgreSQL 마이그레이션 (신규)
└── tests/
    ├── e2e/               ← E2E 테스트 (신규)
    └── integration/       ← 통합 테스트 (신규)
```

---

## 2. 보안 모델: Zero-Knowledge 아키텍처

### 2.1 키 계층 구조

```
Master Password (사용자 입력)
    │
    ▼ Argon2id (64MB, 3 iter, 128-bit salt)
    │
  UMK (User Master Key, 256-bit)   ← 기존 internal/crypto/kdf.go
    │
    ├─ HKDF("tene-encryption-key") → EncKey      ← 기존 (개인 시크릿 암호화)
    ├─ HKDF("tene-sync-envelope")  → SyncKey     ← 신규 (Sync Envelope L2 암호화)
    ├─ HKDF("tene-device-key")     → DeviceKey   ← 신규 (디바이스 식별)
    └─ HKDF("tene-auth-hash")      → AuthHash    ← 기존 (서버 인증용)

  X25519 Key Pair (디바이스별)
    ├─ PrivateKey (32 bytes) → OS Keychain
    └─ PublicKey  (32 bytes) → 서버 등록

  Project Key (PK, 256-bit, CSPRNG) — Team 전용
    └─ X25519 ECDH + HKDF로 각 멤버에게 래핑 전달
```

### 2.2 4계층 암호화 (Sync Envelope)

| 계층 | 보호 대상 | 메커니즘 | 키 |
|------|----------|---------|---|
| **L1** | 시크릿 값 | XChaCha20-Poly1305 (개별 레코드) | EncKey (HKDF) |
| **L2** | 메타데이터 + DB 구조 | XChaCha20-Poly1305 (Sync Envelope) | SyncKey (HKDF) |
| **L3** | 네트워크 전송 | TLS 1.3 | ACM 인증서 |
| **L4** | 디스크 저장 | S3 SSE-S3 (AES-256) | AWS 관리 |

**Envelope 바이너리 포맷**:
```
┌────────┬──────────┬─────────────────────────┬──────────────────┐
│ Header │  Nonce   │       Ciphertext        │      Tag         │
│ 8 bytes│ 24 bytes │     variable length     │    16 bytes      │
└────────┴──────────┴─────────────────────────┴──────────────────┘
AAD: projectID + ":" + environment (버전/vault 위조 방지)
```

### 2.3 X25519 ECDH 팀 키 공유

```go
// internal/crypto/teamkey.go

// WrapProjectKey: Owner가 멤버의 공개키로 PK 래핑
func WrapProjectKey(senderPrivate, recipientPublic []byte,
    projectID, recipientUserID string, projectKey []byte) ([]byte, error)

// UnwrapProjectKey: 멤버가 자신의 개인키로 PK 복원
func UnwrapProjectKey(recipientPrivate, senderPublic []byte,
    projectID, recipientUserID string, wrappedKey []byte) ([]byte, error)
```

**키 회전 (멤버 제거 시)**:
1. 새 PK' 생성 (CSPRNG)
2. 제거된 멤버를 제외한 모든 멤버에게 PK' 재래핑
3. vault.db의 모든 시크릿을 PK'로 재암호화
4. 재암호화된 vault를 sync

### 2.4 서버 가시성 매트릭스

| 데이터 | 서버 접근 | 이유 |
|--------|:--------:|------|
| 사용자 이메일 | **평문** | OAuth 인증 필요 |
| X25519 공개키 | **평문** | 공개키는 비밀 아님 |
| Sync Envelope | **암호문만** | 복호화 불가 |
| 시크릿 키 이름 | **불가** | Envelope 내부 |
| 시크릿 값 | **불가** | L1+L2 이중 암호화 |
| 마스터키/PK | **불가** | 서버 경유 없음 |
| Wrapped PK | **암호문만** | ECDH private key 없이 복호화 불가 |

### 2.5 JWT 토큰 보안

| 항목 | 설계 |
|------|------|
| Access Token | ES256 (ECDSA P-256), 15분 만료, claims: sub, plan(free/pro), did, scope |
| Refresh Token | 256-bit opaque, SHA-256 해시 DB 저장, 30일 만료, 1회용 (rotation) |
| 블랙리스트 | Redis SET, TTL = Access Token 남은 만료시간 |
| PKCE | S256 code_challenge, 로컬 HTTP 서버 콜백 |

### 2.6 OWASP Top 10 대응

| # | 위협 | 대응 |
|---|------|------|
| A01 | Broken Access Control | JWT scope + 서버 측 RBAC 이중 검증 |
| A02 | Cryptographic Failures | XChaCha20+Argon2id, TLS 1.3, 하드코딩 금지 |
| A03 | Injection | Parameterized queries, input regex validation |
| A04 | Insecure Design | Zero-Knowledge 아키텍처 |
| A07 | Auth Failures | ES256 JWT, PKCE OAuth, RT Rotation |

---

## 3. Go API 서버 설계

### 3.1 프로젝트 구조

```
cmd/server/main.go          # 엔트리포인트, 그레이스풀 셧다운
internal/
├── config/config.go         # envconfig 기반 설정 로딩
├── api/
│   ├── server.go            # Echo 인스턴스, 미들웨어 체인, 라우터
│   ├── handler/
│   │   ├── auth.go          # OAuth 콜백, JWT 발급/갱신, 로그아웃
│   │   ├── vault.go         # Vault CRUD, Push(S3 업로드), Pull(Presigned URL)
│   │   ├── team.go          # 팀 CRUD, 멤버 초대/제거, 역할 변경
│   │   ├── billing.go       # Checkout, Portal, Webhook
│   │   ├── device.go        # 디바이스 등록/목록/제거
│   │   ├── audit.go         # 감사 로그 조회
│   │   ├── waitlist.go      # 이메일 등록
│   │   └── health.go        # Liveness + Readiness
│   ├── middleware/
│   │   ├── auth.go          # JWT 검증, Claims 추출
│   │   ├── ratelimit.go     # Token Bucket (Free 100/min, Pro 1000/min)
│   │   ├── cors.go          # 허용 오리진, 메서드, 헤더
│   │   └── security.go      # HSTS, X-Frame-Options, CSP
│   └── response/
│       └── response.go      # 표준 응답: {ok, data, meta} / {ok, error, message, status}
├── service/                 # 비즈니스 로직 (인터페이스 기반)
│   ├── auth_service.go
│   ├── vault_service.go
│   ├── team_service.go
│   └── billing_service.go
├── repository/              # PostgreSQL 쿼리 (직접 SQL)
│   ├── user_repo.go
│   ├── vault_repo.go
│   ├── team_repo.go
│   └── audit_repo.go
└── domain/
    ├── user.go, vault.go, team.go, device.go
    └── errors.go            # 센티넬 에러 → HTTP 에러 매핑
```

### 3.2 API 라우팅

```
# Public
GET    /health                      → Liveness
GET    /health/ready                → Readiness (DB 연결 확인)
GET    /api/v1/auth/:provider/authorize  → OAuth 시작
GET    /api/v1/auth/:provider/callback   → OAuth 콜백 → JWT 발급
POST   /api/v1/auth/refresh         → Access Token 갱신
POST   /api/v1/waitlist             → 이메일 등록
POST   /api/v1/billing/webhook      → LemonSqueezy Webhook (HMAC SHA-256 서명 검증)

# Authenticated (JWT 필수)
POST   /api/v1/auth/signout         → Refresh Token 폐기
GET    /api/v1/auth/me              → 현재 사용자 정보

GET    /api/v1/vaults               → Vault 목록
POST   /api/v1/vaults               → Vault 생성
POST   /api/v1/vaults/:id/push      → 암호화된 vault blob S3 업로드
GET    /api/v1/vaults/:id/pull      → Presigned URL 반환
DELETE /api/v1/vaults/:id           → Vault 삭제

POST   /api/v1/teams               → 팀 생성
GET    /api/v1/teams               → 내 팀 목록
POST   /api/v1/teams/:id/invite    → 멤버 초대 (wrapped_pk 포함)
DELETE /api/v1/teams/:id/members/:uid → 멤버 제거 (키 회전 트리거)
PATCH  /api/v1/teams/:id/members/:uid/role → 역할/권한 변경

GET    /api/v1/billing/subscription → 구독 상태
POST   /api/v1/billing/checkout     → LemonSqueezy Checkout URL
POST   /api/v1/billing/portal       → LemonSqueezy 고객 포털 URL

POST   /api/v1/devices             → 디바이스 등록
GET    /api/v1/devices             → 디바이스 목록
DELETE /api/v1/devices/:id         → 디바이스 제거
GET    /api/v1/audit               → 감사 로그 조회
```

### 3.3 미들웨어 체인

```
Request → RequestID → Logger → Recovery → CORS → SecurityHeaders
  → RateLimit(plan-aware) → [JWT Auth] → Handler → Response
```

### 3.4 응답 포맷

```json
// 성공
{ "ok": true, "data": { ... }, "meta": { "timestamp": "...", "request_id": "..." } }

// 에러
{ "ok": false, "error": "VAULT_NOT_FOUND", "message": "...", "status": 404 }
```

### 3.5 에러 처리

도메인 센티넬 에러 → HTTP 에러 자동 매핑:

| 도메인 에러 | HTTP Status |
|------------|-------------|
| `ErrNotFound`, `ErrVaultNotFound` | 404 |
| `ErrEmailAlreadyExists` | 409 |
| `ErrUnauthorized`, `ErrTokenExpired` | 401 |
| `ErrForbidden`, `ErrNotTeamMember` | 403 |
| `ErrProPlanRequired` | 402 (Free 사용자가 Sync/Team 시도 시) |

### 3.6 그레이스풀 셧다운

SIGTERM 수신 → Echo 셧다운 (30초 타임아웃, 기존 요청 완료 대기) → DB 풀 종료

---

## 4. CLI 클라우드 명령어 설계

### 4.1 신규 명령어 일람

| 명령어 | 설명 | 핵심 동작 |
|--------|------|----------|
| `tene login` | OAuth 로그인 | 로컬 HTTP 서버 → 브라우저 → JWT 저장 → X25519 키 등록 |
| `tene logout` | 로그아웃 | 서버 RT 폐기 → Keychain 삭제 |
| `tene push` | Vault 업로드 | Sync Envelope 암호화 → SHA-256 → API → S3 |
| `tene pull` | Vault 다운로드 | S3 → 체크섬 검증 → 복호화 → vault.db 교체 |
| `tene sync` | Push+Pull 통합 | Fake Door → 실제 기능 전환 |
| `tene team create` | 팀 생성 | Project Key 생성 → X25519 래핑 |
| `tene team invite` | 멤버 초대 | 대상 공개키 → ECDH → PK 래핑 → 서버 전송 |
| `tene team remove` | 멤버 제거 | 멤버 삭제 → PK 회전 → 재래핑 → 재암호화 |
| `tene team list` | 팀/멤버 목록 | 테이블 출력 |
| `tene billing` | 구독 관리 | 상태 조회, LemonSqueezy 고객 포털 브라우저 오픈 |

### 4.2 신규 내부 패키지

#### `internal/auth/` — 인증 관리

```go
type AuthClient interface {
    IsAuthenticated() bool
    StartOAuthFlow(ctx context.Context, cfg OAuthConfig) (*OAuthResult, error)
    StoreTokens(accessToken, refreshToken string) error
    GetAccessToken() (string, error)  // 만료 시 자동 갱신
    GenerateDeviceKey() (*DeviceKey, error)
    RegisterDeviceKey(ctx context.Context, pubKey []byte) error
}
```

OAuth 흐름: PKCE + State → 로컬 127.0.0.1 HTTP 서버 → 브라우저 콜백 → JWT 수신 → Keychain 저장

#### `internal/sync/` — Vault 동기화

```go
type SyncEngine interface {
    Push(ctx context.Context, opts PushOptions) (*PushResult, error)
    Pull(ctx context.Context, opts PullOptions) (*PullResult, error)
    Status(ctx context.Context) (*SyncStatus, error)
    Resolve(ctx context.Context, strategy ConflictStrategy) (*ResolveResult, error)
}
```

#### `internal/api/` (CLI 측) — HTTP 클라이언트

```go
type APIClient interface {
    ExchangeCode(ctx, code, verifier, redirectURI string) (*TokenResponse, error)
    UploadVaultBlob(ctx, vaultID string, blob []byte, version int64) (*UploadResult, error)
    DownloadVaultBlob(ctx, vaultID string) (io.ReadCloser, error)
    CreateTeam(ctx, name string) (*Team, error)
    // ... 전체 API 커버
}
```

### 4.3 기존 파일 변경 영향

| 파일 | 변경 |
|------|------|
| `internal/cli/root.go` | 7개 신규 명령어 등록 |
| `internal/cli/sync_cmd.go` | Fake Door → 실제 sync |
| `internal/config/config.go` | `CloudConfig`, `SyncInfo` 필드 추가 |
| `internal/crypto/keymanager.go` | `DeriveSubKeyWithSalt()`, X25519 함수 추가 |

### 4.4 신규 파일 목록

| 파일 | 설명 |
|------|------|
| `internal/cli/login.go` | OAuth 로그인 |
| `internal/cli/logout.go` | 로그아웃 |
| `internal/cli/push.go` | Vault Push |
| `internal/cli/pull.go` | Vault Pull |
| `internal/cli/team.go` | Team 서브커맨드 |
| `internal/cli/billing.go` | 구독 관리 |
| `internal/crypto/x25519.go` | X25519 키 쌍, ECDH |
| `internal/crypto/teamkey.go` | PK 래핑/언래핑 |
| `internal/sync/engine.go` | Sync 엔진 |
| `internal/sync/envelope.go` | Envelope 암복호화 |
| `internal/sync/conflict.go` | 충돌 감지/해결 |
| `internal/sync/merge.go` | 3-way merge |

---

## 5. Vault Sync 엔진 설계

### 5.1 Push 흐름

```
1. vault.db 읽기 → 2. SyncKey 파생 (HKDF)
→ 3. XChaCha20-Poly1305 Seal (L2) → 4. SHA-256 체크섬
→ 5. API POST /push (If-Match: version) → 6. Presigned URL 수신
→ 7. S3 업로드 (프로그레스 바) → 8. SyncState 업데이트
```

### 5.2 Pull 흐름

```
1. API GET /pull → 2. manifest 수신 (version, hash)
→ 3. 버전 비교 → 4. S3 다운로드 → 5. SHA-256 검증
→ 6. XChaCha20-Poly1305 Open (L2) → 7. vault.db 교체
→ 8. SyncState 업데이트 + base snapshot 저장
```

### 5.3 충돌 감지 및 해결

**충돌 발생 조건**: Push 시 서버 버전 > 로컬 버전 → 409 Conflict

**해결 전략**:

| 전략 | 플래그 | 동작 |
|------|--------|------|
| Server-Wins (기본) | (없음) | pull → merge → push |
| Force Push | `--force` | 서버 덮어쓰기 |
| Force Pull | `--force-pull` | 로컬 덮어쓰기 |
| Interactive | `--interactive` | 키 단위 사용자 선택 |

### 5.4 3-Way Merge 규칙

| Base | Local | Remote | 결과 |
|------|-------|--------|------|
| A=1 | A=2 | A=1 | A=2 (로컬만 수정, 자동) |
| A=1 | A=1 | A=2 | A=2 (리모트만 수정, 자동) |
| A=1 | A=2 | A=3 | **충돌** (양쪽 다르게 수정) |
| — | A=1 | — | A=1 (로컬 추가, 자동) |
| A=1 | 삭제 | A=2 | **충돌** (로컬 삭제+리모트 수정) |

### 5.5 네트워크 복원력

- **재시도**: Exponential backoff (최대 3회, 초기 1초, 최대 30초, 10% jitter)
- **타임아웃**: 소형 vault 30초, 대형 vault(>1MB) 5분
- **오프라인 큐잉**: `.tene/sync_queue.json`에 push 대기, 재연결 시 자동 실행

---

## 6. PostgreSQL 데이터베이스 설계

### 6.1 ERD

```
users ──┬──► devices
        ├──► vaults ──► audit_logs
        ├──► teams ──► team_members
        ├──► refresh_tokens
        └──► waitlist (독립)
```

### 6.2 핵심 테이블

#### users
```sql
CREATE TABLE users (
  id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email             TEXT NOT NULL,
  auth_provider     TEXT NOT NULL CHECK (auth_provider IN ('github', 'google')),
  github_id         TEXT, google_id TEXT,
  plan              TEXT NOT NULL DEFAULT 'free' CHECK (plan IN ('free', 'pro')),
  lemon_customer_id TEXT,
  x25519_public_key BYTEA CHECK (x25519_public_key IS NULL OR octet_length(x25519_public_key) = 32),
  created_at TIMESTAMPTZ DEFAULT now(), updated_at TIMESTAMPTZ DEFAULT now()
);
CREATE UNIQUE INDEX idx_users_email ON users(email);
```

#### vaults
```sql
CREATE TABLE vaults (
  id UUID PRIMARY KEY, user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  team_id UUID REFERENCES teams(id) ON DELETE SET NULL,
  project_name TEXT NOT NULL, s3_key TEXT NOT NULL,
  vault_version INTEGER NOT NULL DEFAULT 1,
  vault_hash BYTEA NOT NULL CHECK (octet_length(vault_hash) = 32),
  secret_count INTEGER DEFAULT 0,
  UNIQUE(user_id, project_name)
);
```

**낙관적 잠금**: `UPDATE vaults SET vault_version = vault_version + 1 WHERE id = $1 AND vault_version = $2` — 0 rows affected → 409 Conflict

#### audit_logs (월별 파티셔닝)
```sql
CREATE TABLE audit_logs (
  id UUID NOT NULL, user_id UUID NOT NULL,
  action TEXT NOT NULL, created_at TIMESTAMPTZ NOT NULL,
  PRIMARY KEY (id, created_at)
) PARTITION BY RANGE (created_at);
```

#### team_members
```sql
CREATE TABLE team_members (
  team_id UUID REFERENCES teams(id) ON DELETE CASCADE,
  user_id UUID REFERENCES users(id) ON DELETE CASCADE,
  role TEXT NOT NULL CHECK (role IN ('admin', 'member')),
  env_permissions JSONB DEFAULT '["dev"]',
  wrapped_project_key BYTEA,
  PRIMARY KEY (team_id, user_id)
);
```

### 6.3 마이그레이션 전략

- **도구**: `golang-migrate`
- **파일**: `migrations/000001_create_users.up.sql` ~ `000008_create_waitlist.up.sql`
- **순서**: users → teams → team_members → devices → vaults → audit_logs → refresh_tokens → waitlist
- **실행**: 앱 시작 시 자동 (Phase 1), CI/CD 분리 (Phase 2)

### 6.4 성능

- **pgxpool**: MaxConns=20, MinConns=5, MaxConnLifetime=30min
- **audit_logs**: 월별 파티셔닝, `pg_cron`으로 자동 생성
- **페이지네이션**: Cursor 기반 (Keyset), OFFSET 지양

---

## 7. AWS 인프라 설계

### 7.1 네트워크

```
VPC: 10.0.0.0/16 (ap-northeast-2)
  Public:   10.0.1.0/24, 10.0.2.0/24   (ALB, fck-nat)
  Private:  10.0.10.0/24, 10.0.11.0/24 (ECS Fargate)
  Isolated: 10.0.20.0/24, 10.0.21.0/24 (RDS)
```

### 7.2 주요 리소스

| 서비스 | 사양 | 월 비용 |
|--------|------|:------:|
| ECS Fargate | 0.25 vCPU, 512 MiB, 1 Task | ~$9 |
| ALB + WAF | HTTPS, ACM, OWASP 규칙 | ~$28 |
| RDS PostgreSQL | db.t4g.micro, 20 GiB gp3 | ~$12 |
| S3 | vault blob + 버전 관리 | ~$0.05 |
| fck-nat | t4g.nano (NAT GW 대체, $29 절감) | ~$3 |
| Route 53 + ACM | 2 호스팅 존, 자동 갱신 인증서 | ~$2 |
| **합계** | | **~$54/월** |

### 7.3 Terraform 모듈 구조

```
infra/terraform/
├── modules/
│   ├── vpc/     (VPC, 서브넷, IGW, 라우팅)
│   ├── nat/     (fck-nat t4g.nano, Auto Recovery)
│   ├── ecs/     (Cluster, Task Def, Service, Auto Scaling)
│   ├── alb/     (ALB, HTTPS 리스너, Target Group)
│   ├── rds/     (PostgreSQL, 파라미터 그룹, 보안 그룹)
│   ├── s3/      (vault 버킷, ALB 로그 버킷, lifecycle)
│   ├── ecr/     (이미지 레지스트리, lifecycle policy)
│   ├── route53/ (A 레코드 Alias)
│   ├── acm/     (*.tene.sh 인증서, DNS 검증)
│   ├── iam/     (ECS Task/Execution Role, GitHub OIDC)
│   ├── waf/     (OWASP Common + Rate Limit)
│   └── secrets/ (DB 비밀번호, JWT, LemonSqueezy, OAuth)
├── environments/
│   ├── staging/ (main.tf, terraform.tfvars, backend.tf)
│   └── prod/    (main.tf, terraform.tfvars, backend.tf)
└── global/      (ECR, OIDC Provider — 환경 공유)
```

### 7.4 보안 그룹

```
ALB:  443/tcp from 0.0.0.0/0
ECS:  8080/tcp from ALB only
RDS:  5432/tcp from ECS only
```

### 7.5 CI/CD 파이프라인

```
Push → Test → Lint → Build (Docker multi-stage) → ECR Push (OIDC)
→ Deploy Staging (자동) → E2E Test → Deploy Production (수동 승인)
```

### 7.6 규모별 비용 예측

| 사용자 | 인프라 비용 | 예상 수익 | 이익률 |
|:------:|:---------:|:--------:|:-----:|
| 가입자 | Pro 전환 (15%) | 인프라 | 매출 ($5×전환) | 수수료 (LS+PayPal) | 영업이익 | 이익률 |
|:------:|:-------------:|:-----:|:-------------:|:----------------:|:--------:|:-----:|
| 500명 | 75명 | ~$75 | $375 | ~$60 | ~$141 | 38% |
| 1,000명 | 150명 | ~$118 | $750 | ~$119 | ~$413 | 55% |
| 5,000명 | 750명 | ~$300 | $3,750 | ~$594 | ~$2,756 | 74% |
| 10,000명 | 1,500명 | ~$460 | $7,500 | ~$1,189 | ~$5,751 | 77% |

---

## 8. LemonSqueezy 결제 시스템 설계

### 8.1 요금제 구조

| 플랜 | 가격 | 과금 방식 | 결제 주체 |
|------|------|----------|----------|
| **Free** | $0/forever | — | — |
| **Pro** | $5/month | 개인 flat 구독 | 각 사용자가 개인 카드 등록 |

> **설계 원칙**: Claude Code, GitHub Copilot과 동일한 개인 결제 모델. 팀 기능은 Pro 사용자끼리 팀을 구성하여 사용. 관리자 일괄 결제(per-seat) 없음 → 구현 극적 단순화.

### 8.2 LemonSqueezy 선택 이유

| 항목 | LemonSqueezy | 비고 |
|------|-------------|------|
| 한국 개인 계정 | **지원** | Stripe는 한국 계정 생성 불가 |
| MoR (Merchant of Record) | **O** | 글로벌 세금 (EU VAT, US Sales Tax) 자동 처리 |
| 정산 | Payoneer → 한국 개인 계좌 | 법인 없이 수령 가능 |
| 수수료 | 5% + $0.50 per transaction | Pro $5/mo 기준: 실수령 $4.25 (85%) |
| Go SDK | 비공식 (`NdoleStudio/lemonsqueezy-go`) | REST API 직접 호출도 가능 |
| Webhook 서명 | HMAC SHA-256 | 검증 지원 |

### 8.3 결제 흐름

```
[사용자]
  │
  ├─ CLI: tene billing upgrade
  │   └─ POST /api/v1/billing/checkout → LemonSqueezy Checkout URL 반환
  │   └─ 브라우저 자동 오픈
  │
  └─ 대시보드: "Upgrade to Pro" 버튼
      └─ LemonSqueezy Checkout Overlay 표시
          │
          ▼
[LemonSqueezy Checkout 페이지]
  - 사용자가 개인 카드 입력 → 결제 완료
          │
          ▼
[LemonSqueezy Webhook → POST /api/v1/billing/webhook]
  - subscription_created → users.plan = "pro"
  - 감사 로그 기록
```

### 8.4 BillingService 인터페이스 (단순화)

```go
// internal/billing/billing.go
type BillingService interface {
    CreateCheckoutURL(ctx context.Context, userID string) (string, error)
    GetPortalURL(ctx context.Context, userID string) (string, error)
    HandleWebhook(ctx context.Context, payload []byte, signature string) error
    GetSubscription(ctx context.Context, userID string) (*Subscription, error)
}

type Subscription struct {
    Plan      string     // "free" or "pro"
    Status    string     // "active", "cancelled", "past_due"
    ExpiresAt *time.Time // 현재 결제 기간 종료일
}
```

> **기존 대비 제거된 것**: `UpdateSeatCount`, `HandleSeatProration`, `CreateTeamCheckoutSession` — per-seat 관련 코드 전부 불필요

### 8.5 Webhook 이벤트 처리

| 이벤트 | 처리 |
|--------|------|
| `subscription_created` | `users.plan = "pro"`, 감사 로그 |
| `subscription_updated` | 플랜 상태 변경 반영 |
| `subscription_cancelled` | `users.plan = "free"`, 데이터 90일 보존 |
| `subscription_payment_failed` | 7일 유예, 이메일 알림 |

**보안**: `X-Signature` 헤더 HMAC SHA-256 검증 + event_id 멱등성

### 8.6 환경변수

```bash
LEMONSQUEEZY_API_KEY=...           # API 키
LEMONSQUEEZY_WEBHOOK_SECRET=...    # Webhook 서명 시크릿
LEMONSQUEEZY_STORE_ID=...          # 스토어 ID
LEMONSQUEEZY_VARIANT_PRO=...           # Pro $5/month Variant ID
```

### 8.7 다운그레이드 정책

| 전환 | 동작 |
|------|------|
| Pro → Free | Sync 비활성화, 팀 참여 불가, 기존 데이터 90일 보존 |
| 결제 실패 | 7일 Grace Period → 읽기 전용 → Free |

### 8.8 팀 참여 조건

```
팀 생성: Pro 사용자만 가능
팀 참여: Pro 사용자만 초대 수락 가능
Free 사용자 초대 시: "Pro로 업그레이드하세요" 안내
Pro 만료 시: 팀에서 자동 비활성화 (데이터 유지, 접근 차단)
```

---

## 9. 프론트엔드 대시보드 설계

### 9.1 기술 스택

| 항목 | 선택 |
|------|------|
| 프레임워크 | Next.js 15 App Router |
| UI | shadcn/ui + Tailwind CSS v4 |
| 데이터 페칭 | TanStack Query v5 |
| 폼 | React Hook Form + Zod |
| 전역 상태 | Zustand (persist) |
| 배포 | Vercel (app.tene.sh) |

### 9.2 페이지 구조

```
(auth)/login/           # GitHub / Google OAuth
(dashboard)/
  ├── page.tsx           # Overview (통계 카드 4개 + 최근 활동)
  ├── vaults/            # Vault 목록 (검색, 환경 필터)
  │   └── [id]/          # Vault 상세 (키 이름만, 값 마스킹 ••••••••)
  ├── devices/           # 디바이스 카드 그리드 (온/오프라인, 해제)
  ├── audit/             # 감사 로그 테이블 (액션/날짜 필터)
  ├── team/              # 팀 관리 (멤버 테이블, 역할 배지, 초대 모달)
  └── billing/           # 플랜 카드 (Free/Pro), LemonSqueezy 고객 포털
```

### 9.3 인증 흐름

- OAuth → API → JWT 수신 → access_token: Zustand 메모리, refresh_token: httpOnly Cookie
- 401 → 자동 갱신 (refresh queue로 중복 방지)
- middleware.ts로 보호 라우트 리다이렉트

### 9.4 Zero-Knowledge 원칙

**대시보드가 절대 하지 않는 것**:
- 시크릿 값 표시/편집 (마스킹 ••••••••)
- 시크릿 생성/수정 (CLI 전용)
- 마스터키/PK 서버 전송
- "값 보기" 버튼 (비활성 + 툴팁: "CLI에서만 확인 가능")

### 9.5 주요 컴포넌트

| 컴포넌트 | 위치 | 역할 |
|---------|------|------|
| `AppSidebar` | layout | 네비게이션 (6 메뉴), 팀 전환, 유저 메뉴 |
| `StatCard` | overview | 통계 표시 (vaults, 키 수, 디바이스, 감사 이벤트) |
| `VaultTable` | vaults | 검색, 환경 필터, 환경별 배지 색상 |
| `SecretKeyRow` | vault detail | 키 이름 + ••••••••, Eye 버튼 disabled |
| `DeviceCard` | devices | 플랫폼 아이콘, 온라인 dot, 해제 버튼 |
| `AuditRow` | audit | 시간, 액션 배지, 액터, 대상 |
| `InviteModal` | team | 이메일 + 역할 선택 (Zod validation) |
| `PlanCard` | billing | 현재 플랜 하이라이트, 업그레이드 버튼 |
| `UsageBar` | billing | 사용량 프로그레스 (80% 이상 경고 색) |

### 9.6 반응형 & 접근성

- lg+: 사이드바 항상 표시 (접기 가능)
- sm 이하: 바텀 탭 네비게이션
- 다크모드: CSS 변수 기반
- WCAG 2.1 AA: 포커스 링, aria-label, 색상 대비 4.5:1+

---

## 10. QA/테스트 전략

### 10.1 테스트 피라미드

```
        E2E (CLI↔API↔Dashboard)         — 주요 흐름 100%
       ╱                            ╲
     Integration (API↔DB, CLI↔API)    — testcontainers + LocalStack
    ╱                                  ╲
  Unit (crypto, sync, auth, billing)    — Go testing + testify
```

### 10.2 커버리지 목표

| 패키지 | 목표 | 미달 시 |
|--------|:----:|---------|
| `internal/crypto/` | **95%+** | 배포 블로킹 |
| `internal/sync/` | **90%+** | 배포 블로킹 |
| `internal/auth/` | **85%+** | PR 블로킹 |
| `internal/billing/` | **85%+** | PR 블로킹 |
| handlers/middleware | **80%+** | 경고 |
| `apps/dashboard` | **70%+** | 경고 |

### 10.3 핵심 테스트 시나리오

**Zero-Knowledge 검증 (Critical)**:
- DB/S3에 평문 시크릿 미존재 확인
- 핸들러 로그에 시크릿 미노출 확인
- X25519 래핑/언래핑 왕복 + 잘못된 키 실패 확인

**Sync E2E**:
- 디바이스A push → 디바이스B pull → 시크릿 일치
- 양쪽 동시 수정 → 충돌 감지 → 해결

**보안 테스트**:
- JWT alg:none 공격, plan 클레임 위변조 → 차단 확인
- SQL Injection 페이로드 → 정상 에러/무결과 확인
- Rate Limit 초과 → 429 반환 + Retry-After 헤더

### 10.4 테스트 환경

- **로컬**: Docker Compose (PostgreSQL + LocalStack)
- **CI**: GitHub Actions (testcontainers, LemonSqueezy test mode)
- **Staging**: AWS 실제 환경 E2E
- **성능**: k6 (P99 < 500ms, 에러율 < 1%)

---

## 11. 구현 로드맵

### 11.1 주차별 계획

| 주차 | Phase | 핵심 작업 | 마일스톤 |
|:----:|:-----:|----------|:--------:|
| **W1** | 2a | Terraform 인프라 + PostgreSQL 스키마 + ECS 배포 | M1: healthcheck 200 OK |
| **W2** | 2a | Go API 부트스트랩 + OAuth + JWT + `tene login/logout` | M2: OAuth 로그인 성공 |
| **W3** | 2a | Sync Envelope + Push/Pull API + CLI + 충돌 감지 | M3: 2대 디바이스 sync |
| **W4** | 2a | LemonSqueezy Pro + Waitlist + 대시보드 MVP + CI/CD | M4: LemonSqueezy 결제 완료 |
| **W5** | 2b | X25519 키 쌍 + ECDH 래핑 + Team API + DB 마이그레이션 | M5: 팀원 복호화 성공 |
| **W6** | 2b | `tene team` CLI + 키 회전 + RBAC 환경별 접근 제어 | M6: RBAC 동작 |
| **W7** | 2b | 대시보드 Team 기능 + `tene billing` CLI | M7: Team 대시보드 완료 |
| **W8** | 2c | E2E 테스트 + 보안 감사 (OWASP) + 문서 + 랜딩페이지 갱신 | M8: 출시 준비 완료 |

### 11.2 세션 가이드

각 주차는 2-3 세션으로 분리:

| 세션 | 목표 | 검증 기준 |
|------|------|----------|
| 1-1 | VPC + ACM + IAM | `terraform plan` 오류 없음 |
| 1-2 | RDS + S3 + ECR | `psql` 접속 성공 |
| 1-3 | ECS + ALB | `curl /health` → 200 |
| 2-1 | Echo 서버 + JWT | 토큰 생성/검증 테스트 통과 |
| 2-2 | OAuth + CLI login | `tene login` 성공 |
| 3-1 | Sync Envelope 암호화 | Seal→Open 왕복 테스트 |
| 3-2 | Push/Pull API + S3 | API 통합테스트 통과 |
| 3-3 | CLI push/pull | 2대 sync 성공 |
| 4-1 | LemonSqueezy + Waitlist | test 결제 → DB plan="pro" |
| 4-2 | 대시보드 MVP + CI/CD | app.tene.sh 로그인 동작 |
| ... | ... | ... |

### 11.3 의존성 그래프

```
Terraform ──► PostgreSQL 스키마 ──► Go API 부트스트랩
  ├──► Auth API ──► Vault Sync API ──► Billing API
  │                      │                │
  └──► Dashboard ────────┘────────────────┘
                                          │
                         Team Crypto ─────┘
```

### 11.4 리스크 및 완화

| 리스크 | 확률 | 영향 | 완화 |
|--------|:----:|:----:|------|
| X25519 구현 오류 | 중 | 상 | golang.org/x/crypto 공식 라이브러리만 사용, 테스트벡터 검증 |
| Sync 충돌 데이터 손실 | 중 | 상 | S3 버전 관리, 로컬 .bak 백업, 항상 사용자 명시 선택 |
| LemonSqueezy 정산 지연 | 저 | 중 | PayPal 연동, 잔액 $100+ 모아서 출금, 외화 계좌 활용 |
| 키 회전 중 서비스 중단 | 중 | 중 | 트랜잭션 처리, 2-phase rotation, 실패 시 롤백 |

### 11.5 KPI

| 지표 | 목표 (3개월) | 측정 |
|------|:-----------:|------|
| 회원가입 | 500+ | DB 쿼리 |
| Pro 전환 | 75+ (15%) | LemonSqueezy 대시보드 |
| MRR | $375+ | LemonSqueezy 대시보드 |
| Push/Pull 성공률 | 99.5%+ | CloudWatch |
| API P99 | < 500ms | CloudWatch |
| 가용성 | 99.9% | CloudWatch |

---

## 부록 A: 신규 Go 의존성

```
github.com/labstack/echo/v4           # HTTP 프레임워크
github.com/golang-jwt/jwt/v5          # ES256 JWT
golang.org/x/oauth2                   # OAuth 2.0
github.com/golang-migrate/migrate/v4  # DB 마이그레이션
github.com/jackc/pgx/v5              # PostgreSQL 드라이버 + 풀
github.com/aws/aws-sdk-go-v2          # S3, Secrets Manager
# LemonSqueezy: REST API 직접 호출 (net/http) 또는 비공식 Go SDK
github.com/go-playground/validator/v10 # 입력 검증
golang.org/x/time/rate                # Rate Limiter
golang.org/x/crypto                   # 기존 (curve25519 추가 사용)
```

## 부록 B: 상세 설계 참조 파일

각 에이전트의 전체 설계 결과는 이 문서의 해당 섹션에 요약되어 있습니다.
10명의 전문 에이전트가 작성한 구성 요소:

1. **Go API 서버** — 핸들러, 서비스, 리포지토리, 에러 처리 (전체 Go 코드 시그니처)
2. **보안/암호화** — 키 계층, Sync Envelope, X25519 ECDH, JWT, STRIDE 위협 모델
3. **AWS 인프라** — Terraform 12 모듈 전체 HCL, CI/CD, 비용 예측
4. **프론트엔드** — Next.js 컴포넌트 Props/이벤트, Zustand 스토어, TanStack Query 키
5. **CLI 명령어** — cobra.Command 구조, Run 함수, 플래그, UX 출력 예시
6. **DB 스키마** — 8 테이블 CREATE TABLE DDL, 인덱스, 마이그레이션 파일
7. **LemonSqueezy 결제** — BillingService 인터페이스 (단순화), Webhook 핸들러, Free/Pro 개인 결제
8. **Sync 엔진** — SyncEngine/ConflictResolver 인터페이스, 3-way merge, 재시도
9. **QA 전략** — 테스트 피라미드, 커버리지 목표, 보안/성능 테스트 코드
10. **구현 로드맵** — 8주 세션 가이드, 마일스톤, 의존성 그래프, 리스크 매트릭스
