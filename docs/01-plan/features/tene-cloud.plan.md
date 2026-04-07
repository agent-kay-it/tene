# Tene Cloud 서비스 상세 계획서

> **버전**: v1.0  
> **작성일**: 2026-04-07  
> **작성자**: AI Agent Team (경쟁사 조사, 인프라 아키텍처, 보안 설계, 결제 시스템, CLI 설계, API/DB 설계, 프론트엔드 설계 — 7개 전문 에이전트 종합)  
> **상태**: Phase 2 설계 확정 대기  

---

## 1. 프로젝트 개요

### 1.1 배경

Tene CLI(v0.9.3)는 로컬 시크릿 관리 도구로 출시 완료. XChaCha20-Poly1305 + Argon2id 암호화, 5개 AI 에이전트(Claude, Cursor, Windsurf, Gemini, Codex) 지원, `tene sync` Fake Door로 클라우드 수요를 측정 중.

### 1.2 목표

로컬 CLI의 Zero-Knowledge 보안을 유지하면서 클라우드 Sync + 팀 공유 + 대시보드를 제공하여 유료 전환 실현.

### 1.3 요금제

| 플랜 | 가격 | 대상 | 핵심 기능 |
|------|------|------|----------|
| **Free** | $0/forever | 개인 개발자 | 로컬 시크릿 CRUD, 5개 AI 에이전트, 복구키 |
| **Pro** | $5/month | 개인 결제 | Vault Sync, 클라우드 백업, 디바이스 관리, 감사 로그, 팀 시크릿 공유, RBAC, 환경별 접근 제어, 팀 감사 로그 |

> **결제 모델**: 각 사용자가 개인 카드로 결제 (Claude Code, GitHub Copilot 방식). 팀 기능은 Pro 사용자끼리 팀을 구성하여 사용. 관리자 일괄 결제(per-seat) 없음.
> **결제 수단**: LemonSqueezy (MoR — 한국 개인사업자/개인 계정 지원, 글로벌 세금 자동 처리)

### 1.4 경쟁 포지셔닝

| 도구 | Zero-Knowledge | 가격 | 로컬 우선 |
|------|:-----------:|------|:-------:|
| **Tene** | **O** | Pro $5/mo | **O** |
| Bitwarden SM | O | $6/user | X |
| 1Password | O | $7.99/user | X |
| Infisical | 부분적 | $18/user | X |
| Doppler | X | $21/user | X |
| dotenv-vault | - | Deprecated | - |

---

## 2. 아키텍처

### 2.1 시스템 구성도

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
     CloudFront       CloudFront          ALB
     (tene.sh)       (app.tene.sh)   (api.tene.sh)
     랜딩페이지       대시보드          API 서버
          │                │                │
     S3 Bucket        S3 Bucket      ┌──────┴──────┐
     (landing)       (dashboard)     │ ECS Fargate  │
                                     │ Go API 서버  │
                                     └──────┬──────┘
                                            │
                              ┌─────────────┼─────────────┐
                              │                           │
                         RDS PostgreSQL              S3 Bucket
                        (메타데이터 DB)          (암호화된 Vault Blob)
```

### 2.2 기술 스택

| 레이어 | 기술 | 이유 |
|--------|------|------|
| **API 서버** | Go (Echo/Gin) | CLI와 동일 언어, crypto 코드 공유, 메모리 효율 |
| **로드밸런서** | ALB | REST API에 최적, WAF 연동, ACM TLS 종료 |
| **DB** | PostgreSQL (RDS) | 관계형 데이터, RBAC 쿼리, 감사 로그 |
| **스토리지** | S3 | 암호화된 vault blob, 버전 관리, lifecycle |
| **결제** | LemonSqueezy | MoR, 한국 개인 계정 지원, Checkout Overlay, 개인 구독 |
| **대시보드** | Next.js + shadcn/ui | 기존 디자인 시스템 유지, App Router |
| **인증** | JWT + OAuth (GitHub, Google) | 개발자 대상, 패스워드 없는 인증 |
| **IaC** | Terraform | 재현 가능한 인프라, 환경 분리 |
| **CI/CD** | GitHub Actions → ECR → ECS | 기존 CI 확장 |

### 2.3 모노레포 구조

```
tene/
├── cmd/
│   ├── tene/              ← CLI 엔트리포인트 (기존)
│   └── server/            ← Cloud API 엔트리포인트 (신규)
├── internal/
│   ├── crypto/            ← CLI + Server 공유
│   ├── vault/             ← CLI + Server 공유
│   ├── cli/               ← CLI 전용
│   ├── api/               ← Server 전용 (라우터, 핸들러, 미들웨어)
│   ├── auth/              ← 인증 (JWT, OAuth)
│   ├── sync/              ← Push/Pull 동기화 로직
│   └── billing/           ← LemonSqueezy 통합
├── apps/
│   ├── web/               ← 랜딩페이지 (기존)
│   └── dashboard/         ← 대시보드 (신규)
├── infra/
│   └── terraform/         ← AWS IaC
└── docs/
```

---

## 3. 보안 모델: Zero-Knowledge 아키텍처

### 3.1 키 계층 구조

```
┌─────────────────────────────────────────────────────────────┐
│                    사용자(User) 계층                          │
│                                                              │
│  마스터 패스워드 + Salt                                       │
│      └─ Argon2id ──► User Master Key (UMK, 256-bit)         │
│              ├─ HKDF("tene-encryption-key") ──► 개인 암호화키 │
│              ├─ HKDF("tene-sync-envelope") ──► Sync 암호화키  │
│              └─ HKDF("tene-device-key") ──► 디바이스 키       │
│                                                              │
│  X25519 키 쌍 (디바이스별)                                    │
│      ├─ 공개키: 서버에 등록                                   │
│      └─ 개인키: UMK로 암호화하여 로컬 저장                    │
├─────────────────────────────────────────────────────────────┤
│                 프로젝트(Project) 계층 — Team 전용             │
│                                                              │
│  Project Key (PK, 256-bit, CSPRNG)                           │
│      ├─ HKDF("tene-project-encryption") ──► 프로젝트 암호화키│
│      └─ 각 멤버의 X25519 공개키로 래핑 ──► wrapped_pk[]     │
├─────────────────────────────────────────────────────────────┤
│                  시크릿(Secret) 계층                          │
│                                                              │
│  XChaCha20-Poly1305(key, plaintext, AAD="키이름|환경|버전")  │
└─────────────────────────────────────────────────────────────┘
```

### 3.2 Sync Envelope (이중 암호화)

vault.db에는 시크릿 값(암호화됨)뿐 아니라 키 이름, 환경, 감사 로그 등 **평문 메타데이터**가 존재. S3 업로드 시 전체 blob을 추가 암호화:

| 계층 | 보호 대상 | 메커니즘 |
|------|----------|---------|
| L1 | 시크릿 값 | XChaCha20-Poly1305 (개별 레코드) |
| L2 | 메타데이터 + DB 구조 | XChaCha20-Poly1305 (Sync Envelope) |
| L3 | 네트워크 전송 | TLS 1.2+ |
| L4 | 디스크 저장 | S3 SSE-S3 (AES-256) |

### 3.3 팀 키 공유 프로토콜 (X25519 ECDH)

**RSA 대신 X25519 선택 이유:**
- 키 크기 32 bytes vs RSA 256 bytes
- 128-bit 보안 수준 (RSA-2048은 112-bit)
- `golang.org/x/crypto/curve25519`로 기존 의존성과 호환
- 상수 시간 구현으로 사이드채널 공격 방어

**팀원 초대 흐름:**
```
1. Owner: tene team invite alice@example.com
2. 서버에서 Alice의 X25519 공개키 조회
3. Owner 측:
   a. ECDH: shared_secret = X25519(owner_private, alice_public)
   b. wrap_key = HKDF(shared_secret, "tene-key-wrap", salt=projectID)
   c. wrapped_pk = XChaCha20-Poly1305(wrap_key, PK, AAD=aliceUserID)
4. wrapped_pk를 서버에 업로드
5. Alice: 자신의 개인키로 역순 복호화
   → 서버는 PK 평문을 절대 볼 수 없음
```

**키 회전 (멤버 제거 시):**
1. 새 PK' 생성 (CSPRNG)
2. 제거된 멤버를 제외한 모든 멤버에게 PK' 재래핑
3. vault.db의 모든 시크릿을 PK'로 재암호화
4. 재암호화된 vault를 sync

### 3.4 서버가 보는 것 vs 볼 수 없는 것

| 데이터 | 서버 접근 | 이유 |
|--------|:--------:|------|
| 사용자 ID/이메일 | 평문 | 인증 필요 |
| X25519 공개키 | 평문 | 공개키는 비밀 아님 |
| Argon2id salt | 평문 | salt는 비밀 아님 |
| 프로젝트 이름 | 평문 | 대시보드 표시 |
| wrapped_pk | **암호화** | ECDH shared_secret으로 래핑 |
| vault blob | **암호화** | Sync Envelope |
| 시크릿 키 이름 | **불가** | Envelope 내부 |
| 시크릿 값 | **불가** | 이중 암호화 |
| 마스터키/PK | **불가** | 서버 경유 없음 |

---

## 4. REST API 설계

### 4.1 인증 API (OAuth 전용 — 회원가입/패스워드 없음)

| Method | Endpoint | 설명 |
|--------|----------|------|
| GET | `/auth/github` | GitHub OAuth 시작 (redirect) |
| GET | `/auth/github/callback` | GitHub OAuth 콜백 → JWT 발급 |
| GET | `/auth/google` | Google OAuth 시작 (redirect) |
| GET | `/auth/google/callback` | Google OAuth 콜백 → JWT 발급 |
| POST | `/auth/refresh` | Access Token 갱신 |
| DELETE | `/auth/logout` | 로그아웃 (Refresh Token 폐기) |

> **참고**: 이메일+패스워드 회원가입 없음. GitHub/Google OAuth로 첫 로그인 시 자동 계정 생성.

### 4.2 Vault Sync API

| Method | Endpoint | 설명 |
|--------|----------|------|
| GET | `/vaults` | 사용자의 vault 목록 |
| POST | `/vaults` | 새 vault 등록 |
| POST | `/vaults/:id/push` | 암호화된 vault blob 업로드 |
| GET | `/vaults/:id/pull` | 암호화된 vault blob 다운로드 |
| DELETE | `/vaults/:id` | vault 삭제 |

### 4.3 Team API

| Method | Endpoint | 설명 |
|--------|----------|------|
| POST | `/teams` | 팀 생성 |
| GET | `/teams` | 내 팀 목록 |
| POST | `/teams/:id/invite` | 멤버 초대 |
| DELETE | `/teams/:id/members/:uid` | 멤버 제거 |
| PATCH | `/teams/:id/members/:uid` | 역할 변경 |

### 4.4 Billing API

| Method | Endpoint | 설명 |
|--------|----------|------|
| GET | `/billing` | 현재 구독 상태 |
| POST | `/billing/checkout` | LemonSqueezy Checkout URL 생성 |
| POST | `/billing/portal` | LemonSqueezy 고객 포털 URL |
| POST | `/billing/webhook` | LemonSqueezy Webhook 수신 |

### 4.5 Waitlist API

| Method | Endpoint | 설명 |
|--------|----------|------|
| POST | `/waitlist` | 이메일 + 관심 플랜 등록 |

### 4.6 기타 API

| Method | Endpoint | 설명 |
|--------|----------|------|
| POST | `/devices` | 디바이스 등록 |
| GET | `/devices` | 디바이스 목록 |
| DELETE | `/devices/:id` | 디바이스 제거 |
| GET | `/audit-logs` | 감사 로그 조회 |

### 4.7 응답 형식

```json
// 성공
{ "ok": true, "data": { ... }, "meta": { "timestamp": "...", "request_id": "..." } }

// 에러
{ "ok": false, "error": "VAULT_NOT_FOUND", "message": "...", "status": 404 }
```

### 4.8 인증 토큰

- **Access Token**: JWT, 만료 15분, `sub`=user_id, `plan`=free/pro
- **Refresh Token**: JWT, 만료 30일, DB에 저장하여 폐기 가능
- **Rate Limit**: Free 100 req/min, Paid 1,000 req/min

---

## 5. PostgreSQL 스키마

### 5.1 users

```sql
CREATE TABLE users (
  id UUID PRIMARY KEY,
  email VARCHAR(255) UNIQUE NOT NULL,
  name VARCHAR(255),
  auth_provider VARCHAR(50) NOT NULL, -- "github", "google"
  github_id BIGINT UNIQUE,
  google_id VARCHAR(255) UNIQUE,
  avatar_url VARCHAR(512),
  plan VARCHAR(50) DEFAULT 'free',    -- "free", "pro"
  lemon_customer_id VARCHAR(255),
  x25519_public_key BYTEA,           -- 32 bytes, 팀 키 교환용
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);
```

### 5.2 devices

```sql
CREATE TABLE devices (
  id UUID PRIMARY KEY,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  name VARCHAR(255) NOT NULL,
  os VARCHAR(50),                    -- "macos", "linux", "windows"
  fingerprint VARCHAR(255),
  last_sync TIMESTAMPTZ,
  created_at TIMESTAMPTZ DEFAULT NOW()
);
```

### 5.3 vaults

```sql
CREATE TABLE vaults (
  id UUID PRIMARY KEY,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  team_id UUID REFERENCES teams(id) ON DELETE SET NULL,
  project_name VARCHAR(255) NOT NULL,
  s3_key VARCHAR(512),               -- "users/{uid}/vaults/{vid}/vault.enc"
  vault_version INT DEFAULT 0,
  vault_hash VARCHAR(64),            -- SHA-256
  secret_count INT DEFAULT 0,
  last_synced_at TIMESTAMPTZ,
  last_synced_by UUID REFERENCES devices(id),
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW(),
  UNIQUE(user_id, project_name)
);
```

### 5.4 teams

```sql
CREATE TABLE teams (
  id UUID PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  slug VARCHAR(255) UNIQUE NOT NULL,
  owner_id UUID NOT NULL REFERENCES users(id),
  lemon_subscription_id VARCHAR(255),
  created_at TIMESTAMPTZ DEFAULT NOW()
);
```

### 5.5 team_members

```sql
CREATE TABLE team_members (
  team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  role VARCHAR(50) NOT NULL,          -- "admin", "member"
  env_permissions JSONB,              -- {"dev": ["read","write"], "prod": ["read"]}
  wrapped_project_key BYTEA,          -- X25519 ECDH로 래핑된 PK
  joined_at TIMESTAMPTZ DEFAULT NOW(),
  PRIMARY KEY (team_id, user_id)
);
```

### 5.6 audit_logs

```sql
CREATE TABLE audit_logs (
  id UUID PRIMARY KEY,
  user_id UUID REFERENCES users(id),
  team_id UUID REFERENCES teams(id),
  vault_id UUID REFERENCES vaults(id),
  action VARCHAR(100) NOT NULL,       -- "vault.push", "vault.pull", "team.invite"
  resource_name VARCHAR(255),
  ip_address INET,
  user_agent VARCHAR(512),
  created_at TIMESTAMPTZ DEFAULT NOW()
);
```

### 5.7 waitlist

```sql
CREATE TABLE waitlist (
  id UUID PRIMARY KEY,
  email VARCHAR(255) UNIQUE NOT NULL,
  plan VARCHAR(50),                   -- "pro" (관심 플랜 기록용)
  created_at TIMESTAMPTZ DEFAULT NOW()
);
```

### 5.8 S3 구조

```
tene-vault-prod-ap-northeast-2/
  users/{user_id}/vaults/{vault_id}/
    vault.enc              ← Sync Envelope 암호화된 blob
    manifest.json          ← 버전, 해시, 타임스탬프
```

---

## 6. CLI 클라우드 명령어

### 6.1 신규 명령어 목록

| 명령어 | 설명 | 구현 우선순위 |
|--------|------|:-----------:|
| `tene login` | GitHub / Google OAuth 로그인 | 1 |
| `tene logout` | 로그아웃, 토큰 삭제 | 1 |
| `tene push` | vault를 클라우드에 업로드 | 2 |
| `tene pull` | 클라우드에서 vault 다운로드 | 2 |
| `tene sync` | 기존 Fake Door → push/pull 통합 | 3 |
| `tene team create` | 팀 생성 | 4 |
| `tene team invite` | 멤버 초대 | 4 |
| `tene team remove` | 멤버 제거 + 키 회전 | 4 |
| `tene team list` | 팀/멤버 목록 | 4 |
| `tene billing` | 구독 상태, LemonSqueezy 고객 포털 링크 | 5 |

### 6.2 주요 UX 흐름

**tene login:**
```bash
$ tene login
? Sign in with:
  > GitHub
    Google
Opening browser for authentication...
✓ Authenticated as octocat (GitHub)
✓ Token stored in OS Keychain
```

**tene push:**
```bash
$ tene push
Preparing vault...
  Encrypting with Sync Envelope...
  Computing checksum...
Uploading to cloud... [████████████████] 100%
✓ Push completed (v4, 2.3 MB)
```

**tene pull:**
```bash
$ tene pull
Fetching from cloud...
  Remote: v5 (2 hours ago)
  Local:  v4
✓ Pull completed — 2 secrets added, 1 updated
```

### 6.3 기존 코드 변경 영향

| 파일/패키지 | 변경 내용 |
|------------|---------|
| `internal/cli/root.go` | 새 명령어 등록 (login, logout, push, pull, team, billing) |
| `internal/cli/sync_cmd.go` | Fake Door → 실제 push/pull 통합으로 전환 |
| `internal/crypto/keymanager.go` | `DeriveSubKeyWithSalt()` 추가, X25519 관련 함수 |
| `internal/config/config.go` | auth 토큰, 마지막 sync 정보 저장 |
| 신규: `internal/api/` | HTTP 클라이언트 (API 통신) |
| 신규: `internal/sync/` | Push/Pull 동기화 + 충돌 해결 로직 |
| 신규: `internal/auth/` | OAuth (GitHub, Google), JWT 토큰 관리 |

---

## 7. 결제 시스템 (LemonSqueezy)

### 7.1 구조

```
LemonSqueezy Products:
  └── "Tene Pro"
        └── Variant: Pro $5/month

결제 흐름:
  CLI/대시보드 → POST /billing/checkout → LemonSqueezy Checkout Overlay
                                           → 각 사용자가 개인 카드로 결제
                                           → Webhook → DB 업데이트
```

> **결제 모델**: 각 사용자가 개인 카드를 등록하여 결제 (Claude Code, GitHub Copilot 방식). 팀 기능은 Pro 사용자끼리 팀을 구성. 관리자 일괄 결제(per-seat) 없음.

### 7.2 핵심 결정

| 항목 | 선택 | 이유 |
|------|------|------|
| 결제 플랫폼 | **LemonSqueezy** | 한국 개인 계정 지원, MoR로 글로벌 세금 자동 처리 |
| Checkout 방식 | **LemonSqueezy Checkout Overlay** | 가장 빠른 구현, PCI 준수 자동 |
| 과금 방식 | **개인 flat 구독** | per-seat 불필요, 구현 단순화 |
| 셀프서비스 | **LemonSqueezy 고객 포털** | 구독 관리, 인보이스 확인 |
| 정산 | **Payoneer → 한국 개인 계좌** | 법인 없이 개인으로 정산 수령 |

### 7.3 Webhook 이벤트 처리

| 이벤트 | 처리 |
|--------|------|
| `subscription_created` | DB plan = "pro", 서비스 활성화 |
| `subscription_updated` | 플랜 변경 반영 |
| `subscription_cancelled` | Free로 다운그레이드 |
| `subscription_payment_failed` | 이메일 알림, 7일 유예 후 서비스 제한 |

### 7.4 수수료

| 항목 | 비용 |
|------|------|
| 거래 수수료 | 5% + $0.50 per transaction |
| Pro $5/mo 기준 | 수수료 $0.75 (5%+$0.50) → 실수령 $4.25 (85%) |

---

## 8. 대시보드 (app.tene.sh)

### 8.1 프로젝트 구조

`apps/dashboard/` — 별도 Next.js 프로젝트 (랜딩과 분리)

### 8.2 기술 스택

| 항목 | 선택 |
|------|------|
| 프레임워크 | Next.js App Router |
| UI | shadcn/ui + Tailwind CSS v4 |
| 인증 | Better Auth 또는 NextAuth v5 (GitHub + Google OAuth) |
| 결제 | LemonSqueezy Checkout Overlay + 고객 포털 |
| 데이터 페칭 | TanStack Query v5 |
| 폼 | React Hook Form + Zod |
| 전역 상태 | Zustand |
| 배포 | Vercel (app.tene.sh) |

### 8.3 페이지 구조

```
(auth)/
  └── login/               # GitHub / Google OAuth (회원가입 별도 없음)

(dashboard)/
  ├── layout.tsx           # 사이드바 + 헤더
  ├── vaults/              # Vault 목록 + 시크릿 키 이름 (값 마스킹)
  │   └── [id]/            # Vault 상세
  ├── devices/             # 디바이스 관리
  ├── audit/               # 감사 로그
  ├── team/                # 팀 관리 (RBAC, 멤버)
  └── billing/             # 구독 상태, LemonSqueezy 고객 포털 링크
```

### 8.4 대시보드가 절대 하지 않는 것

- 시크릿 값 표시/편집 (Zero-Knowledge)
- 시크릿 생성/수정 (CLI 전용: `tene set`)
- 마스터키/프로젝트키 서버 전송

---

## 9. AWS 인프라

### 9.1 리전

**ap-northeast-2 (서울)** — 기존 Route 53 + AWS 계정(monsa-sandbox) 활용

### 9.2 핵심 구성

| 컴포넌트 | 사양 (초기) | 비용/월 |
|----------|-----------|:------:|
| ECS Fargate | 0.25 vCPU, 512 MiB, 1 Task | ~$9 |
| ALB | HTTPS, ACM 인증서 | ~$22 |
| RDS PostgreSQL | db.t4g.micro, 20 GiB gp3 | ~$17 |
| S3 | vault blob | ~$0.05 |
| NAT (fck-nat) | t4g.nano | ~$3 |
| Route 53 | 2 호스팅 존 | ~$2 |
| Secrets Manager | 5개 시크릿 | ~$2 |
| CloudWatch | 5 GiB 로그 | ~$3 |
| **합계** | | **~$58/월** |

### 9.3 규모별 비용 예측

| 가입자 | Pro 전환 (15%) | 인프라 비용 | 매출 ($5×전환) | 수수료 (LS+PayPal) | 영업이익 | 이익률 |
|:------:|:-------------:|:---------:|:-------------:|:----------------:|:--------:|:-----:|
| 500명 | 75명 | ~$75/월 | $375 | ~$60 | ~$141 | 38% |
| 1,000명 | 150명 | ~$149/월 | $750 | ~$119 | ~$413 | 55% |
| 5,000명 | 750명 | ~$300/월 | $3,750 | ~$594 | ~$2,756 | 74% |
| 10,000명 | 1,500명 | ~$462/월 | $7,500 | ~$1,189 | ~$5,751 | 77% |

> 수익은 Pro $5/user/month 기준, 전환율 15% 가정. 수수료 = LS(5%+$0.50) + PayPal(1%) + 마케팅 $100/월 포함

### 9.4 네트워크 구성

```
VPC: 10.0.0.0/16
  Public Subnets:   10.0.1.0/24, 10.0.2.0/24   (ALB, NAT)
  Private Subnets:  10.0.10.0/24, 10.0.11.0/24  (ECS Fargate)
  Isolated Subnets: 10.0.20.0/24, 10.0.21.0/24  (RDS)
```

### 9.5 보안그룹

```
ALB:  443/tcp from 0.0.0.0/0
ECS:  8080/tcp from ALB only, 443/tcp outbound
RDS:  5432/tcp from ECS only
```

### 9.6 도메인 DNS

```
tene.sh       → CloudFront (랜딩페이지)
api.tene.sh   → ALB (Go API)
app.tene.sh   → CloudFront (대시보드)
```

### 9.7 CI/CD 파이프라인

```
Push to main
  → GitHub Actions: test → build → Docker → ECR push
  → Deploy to Staging (자동)
  → Deploy to Production (수동 승인)
```

### 9.8 Terraform 모듈 구조

```
infra/terraform/
  modules/
    vpc/  ecs/  alb/  rds/  s3/  ecr/  route53/  acm/  iam/
  environments/
    staging/    (terraform.tfvars)
    prod/       (terraform.tfvars)
```

---

## 10. 구현 로드맵

### Phase 2a: Solo Sync (4주)

| 주차 | 작업 |
|:----:|------|
| **W1** | Terraform 인프라 구축 (VPC, ECS, RDS, S3, ALB) |
| | PostgreSQL 스키마 마이그레이션 (users, devices, vaults, waitlist, audit_logs) |
| **W2** | Go API 서버 기본 구조 (라우터, 미들웨어, 에러 핸들링) |
| | 인증 API (signup, login, GitHub OAuth, JWT) |
| | `tene login` / `tene logout` CLI 명령어 |
| **W3** | Vault Sync API (push, pull) |
| | Sync Envelope 암호화 구현 (crypto 패키지 확장) |
| | `tene push` / `tene pull` CLI 명령어 |
| | `tene sync` Fake Door → 실제 기능 전환 |
| **W4** | LemonSqueezy 통합 (Pro 플랜 checkout, webhook) |
| | Waitlist API (POST /waitlist) |
| | 대시보드 MVP (로그인, vault 목록, 디바이스, 감사 로그) |
| | CI/CD 파이프라인 |

### Phase 2b: Team (3주)

| 주차 | 작업 |
|:----:|------|
| **W5** | X25519 키 쌍 생성 + 관리 (crypto 패키지) |
| | Project Key 생성 + ECDH 래핑 프로토콜 |
| | Team API (create, invite, remove) |
| **W6** | `tene team` CLI 서브커맨드 |
| | 키 회전 + 재암호화 로직 |
| | RBAC + 환경별 접근 제어 |
| **W7** | 대시보드 Team 기능 (멤버 관리, RBAC, 팀 감사 로그) |
| | `tene billing` CLI 명령어 |

### Phase 2c: 안정화 (1주)

| 주차 | 작업 |
|:----:|------|
| **W8** | E2E 테스트 (CLI ↔ API ↔ Dashboard) |
| | 보안 감사 (OWASP Top 10) |
| | 문서 업데이트 (README, CLAUDE.md) |
| | 랜딩페이지 업데이트 (Coming Soon → 실제 링크) |

---

## 11. 성공 지표 (KPI)

| 지표 | 목표 (출시 3개월) |
|------|:----------------:|
| 회원가입 수 | 500+ |
| Pro 유료 전환 | 75+ (전환율 15%) |
| MRR | $375+ |
| Push/Pull 성공률 | 99.5%+ |
| API P99 응답시간 | < 500ms |
| 서비스 가용성 | 99.9% |

---

## 12. 리스크 및 완화 전략

| 리스크 | 영향 | 완화 |
|--------|------|------|
| 클라우드 수요 부족 | 수익 없음 | Fake Door 데이터로 Go/No-Go 판단, 인프라 비용 최소화 |
| X25519 키 교환 구현 오류 | 보안 사고 | 오픈소스 crypto 라이브러리 사용, 보안 감사 |
| LemonSqueezy 정산 지연 | 현금 흐름 영향 | PayPal 연동, 잔액 $100+ 모아서 출금, 외화 계좌 활용 |
| vault sync 충돌 | 데이터 손실 | S3 버전 관리, 로컬 백업 유지, 명시적 충돌 해결 UX |
| 서버 비용 증가 | 이익률 감소 | Fargate Spot, Reserved Instance, fck-nat |

---

## 13. 기존 계획 대비 변경사항

| 항목 | 기존 (PRD/브리핑) | 변경 | 이유 |
|------|------------------|------|------|
| 클라우드 가격 | $1/user/mo | **Pro $5/mo** | 경쟁사 대비 최저가, 개인 결제 모델, 전환율 극대화 |
| API 서버 | Hono (Node.js) | **Go** | CLI와 crypto 코드 공유, 메모리 효율 |
| 로드밸런서 | NLB | **ALB** | REST API에 L7 라우팅이 적합 |
| AI 에이전트 | Claude Code 전용 | **5개 에이전트** | v0.9.1에서 이미 구현 완료 |
| 키 교환 | 미정 | **X25519 ECDH** | Bitwarden 참고, RSA 대비 효율적 |
| Sync 모델 | 미정 | **Envelope 이중 암호화** | 메타데이터 보호 |
| 초기 비용 예측 | ~$43/월 | **~$58/월** | NAT, ALB 비용 반영 |
| 대시보드 | apps/web 통합 | **apps/dashboard 분리** | 의존성, 배포 파이프라인 독립 |

---

## 부록: 참고 자료

### 경쟁사
- Infisical: AES-256-GCM E2E, Workspace Key, REST API, Pro $18/user/mo
- Doppler: 서버사이드 암호화 (Zero-knowledge 아님), $21/user/mo
- Bitwarden SM: RSA-OAEP Zero-knowledge, $6/user/mo
- 1Password: 2SKD + SRP, $7.99/user/mo
- dotenv-vault: Deprecated — 단일 키 + 클라우드 의존 실패

### 기술 결정
- LemonSqueezy Checkout Overlay + 개인 구독 (per-seat 불필요)
- JWT Access(15분) + Refresh(30일) 토큰
- GitHub OIDC for CI/CD (장기 크레덴셜 제거)
- fck-nat로 NAT Gateway 비용 절감 ($32→$3/월)
- db.t4g.micro(Graviton2 ARM)로 시작, 수직 확장
