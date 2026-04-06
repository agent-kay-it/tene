# Tene MVP 기술 Plan v3.1 -- Local-Only MVP + AI Agent 자동 인식

> v3.1 (2026-04-06) — Recovery Key 12단어 니모닉 통일, npm 패키지 @tene/cli 통일, tene passwd/recover 추가, 암호화 상세 명세 보강
>
> **Summary**: Tene Agentic Secret Runtime Platform의 Local-Only MVP 기술 Plan. 서버 비용 $0, AI Agent 자동 인식이 핵심 차별점
>
> **Project**: Tene
> **Version**: 3.1.0
> **Author**: CTO Lead (Steve)
> **Date**: 2026-04-06
> **Status**: Draft (v3 -- Local-Only MVP, Cloud → Phase 2)

---

## Executive Summary

| 관점 | 내용 |
|------|------|
| **Problem** | AI 에이전트와 바이브코더의 75%가 시크릿을 .env/하드코딩으로 관리. 기존 도구(Vault, Doppler, Infisical)는 서버 가입을 강제하고, $6-21/유저/월로 비싸며, AI 에이전트를 1등 시민으로 지원하지 않음. 2025년 GitHub 시크릿 노출 2,865만 건(+34%), AI 서비스 시크릿 누출 81% 급증 |
| **Solution** | **서버 없는** CLI 시크릿 관리 + **AI Agent 자동 인식**. Master Password + Argon2id KDF + XChaCha20-Poly1305 + SQLite 로컬 볼트. `npm install -g @tene/cli` → `tene init` 하면 CLAUDE.md 자동 생성으로 AI Agent가 즉시 시크릿 관리법을 인식. 서버 비용 $0 |
| **Function/UX Effect** | CLI 명령어로 로컬 오프라인 시크릿 관리. `tene init --claude`(기본값)로 CLAUDE.md 자동 생성, `--cursor`로 .cursorrules 자동 생성. `tene sync` 실행 시 Cloud waitlist 안내 (Fake Door Test). `tene export --encrypted`로 암호화 백업 |
| **Core Value** | "서버가 없다 = 해킹 대상이 없다." 로컬 암호화 + AI Agent 자동 인식 + 오픈소스 + 서버 비용 $0 = 시크릿 관리의 Git. AI Agent가 tene 사용법을 자동으로 학습하는 유일한 시크릿 매니저 |

---

## Context Anchor

| Key | Value |
|-----|-------|
| **WHY** | 바이브코더/AI 에이전트의 시크릿 하드코딩 및 .env 노출 문제 해결. 기존 도구의 서버 강제 + 고가격 + AI 에이전트 미지원 해소. "서버가 없으면 해킹 대상도 없다" |
| **WHO** | 솔로 바이브코더 (Claude Code, Cursor 사용자). 5-15개 시크릿을 관리하는 개인 개발자. 서버 가입/결제에 거부감이 있는 개발자 |
| **RISK** | 암호화 구현 결함 시 제품 신뢰도 치명적 타격. Master Password 분실 시 복구 불가 (Recovery Key로 완화). ".env로 충분" 인식 극복 필요 |
| **SUCCESS** | CLI 명령어 로컬 동작, 오프라인 100%, XChaCha20-Poly1305 + Argon2id 암호화, 설치→첫 시크릿 3분 이내, AI Agent 자동 인식(CLAUDE.md), Fake Door Test로 Cloud 수요 검증 |
| **SCOPE** | Phase 1 (MVP, 2주): 로컬 CLI + AI Agent 통합 + Fake Door. Phase 2 (수요 검증 후): $1 Cloud (ECS + RDS + S3 + 대시보드). Phase 3: 팀 기능 (가설) |

---

## 1. 개요 및 배경

### 1.1 목적

Tene MVP의 Local-Only 기술 아키텍처를 정의한다. "서버 없이, 가입 없이, 무료로, 서버 비용 $0로" 시크릿을 관리하는 CLI를 코어로 하고, AI Agent가 자동으로 시크릿 관리법을 인식하도록 하는 것이 핵심 차별점이다. Cloud 기능은 MVP에서 완전 제외하고, Fake Door Test(tene sync → waitlist)로 수요를 검증한 후 Phase 2에서 구축한다.

### 1.2 배경

- **바이브코딩 폭발적 성장**: AI 생성 코드의 45%가 보안 취약점 포함, 시크릿 하드코딩이 주요 문제
- **NHI(비인간 ID) 폭증**: 인간 ID 대비 100:1 비율에 도달한 기업 존재
- **기존 도구의 빈틈**: Vault(과잉 + $1,152+/월), Doppler($21/유저/월), Infisical(셀프호스팅 서버 필요)
- **Local-First 트렌드**: 개발자 클라우드 피로감 증가, "내 데이터는 내 디바이스에" 선호 급증
- **경쟁 공백**: 로컬 전용 + AI 에이전트 네이티브 + $0 가격대의 시크릿 관리 도구 부재
- **AI Agent 시대**: Claude Code, Cursor, Windsurf 등 AI 에이전트가 코드 작성의 주류가 됨. 이들에게 시크릿 관리법을 자동으로 알려주는 도구가 없음

### 1.3 v2 → v3 핵심 변화

| 항목 | v2 Plan | v3 Plan |
|------|---------|---------|
| **MVP 범위** | Phase 1a(로컬 CLI 2주) + Phase 1b($1 Cloud 3주) | **Phase 1(로컬 CLI 2주)만 MVP**. Cloud는 Phase 2 |
| **서버 비용** | Phase 1a: $0, Phase 1b: $46+/월 | **MVP 전체: $0** |
| **AI Agent 통합** | --json 플래그만 | **CLAUDE.md / .cursorrules 자동 생성** (핵심 차별점) |
| **Cloud 인프라** | Lambda + Aurora Serverless v2 | **ECS Fargate + NLB + RDS PostgreSQL + S3** (Phase 2) |
| **tene sync** | Cloud 동기화 구현 | **Fake Door Test** (waitlist 안내만) |
| **tene export** | .env 형식만 | .env 형식 + **--encrypted 암호화 백업** |
| **Phase 구조** | 1a(로컬) → 1b(Cloud) → 2(팀) | **1(MVP 로컬) → 2(Cloud, 수요 검증 후) → 3(팀)** |

### 1.4 관련 문서

- PRD v2: `docs/00-pm/tene.prd.md`
- Strategy v2: `docs/00-pm/tene-strategy.md`
- Discovery: `docs/00-pm/tene-discovery.md`
- Research: `docs/00-pm/tene-research.md`

---

## 2. 범위

### 2.1 In Scope — Phase 1: MVP (로컬 CLI, 2주)

#### 2.1.1 CLI 코어

- [ ] CLI 코어: 10개 명령어 (init, set, get, run, list, delete, import, export, passwd, recover)
- [ ] 환경 전환: `tene env` (dev/staging/prod)
- [ ] Master Password 설정 + Recovery Key 생성 (12단어 니모닉, BIP-39)
- [ ] Master Password 변경: `tene passwd` (볼트 재암호화 + 새 Recovery Key)
- [ ] Master Password 복구: `tene recover` (Recovery Key로 재설정)
- [ ] 로컬 암호화: Argon2id KDF + XChaCha20-Poly1305 + SQLite
- [ ] OS Keychain 연동 (Master Key 저장)
- [ ] 오프라인 100% 동작 (네트워크 통신 없음)
- [ ] .env 마이그레이션 (import/export)
- [ ] 플랫폼: macOS, Linux, Windows(WSL)
- [ ] npm 패키지 배포 (`npm install -g @tene/cli`)
- [ ] --json 플래그 (AI 에이전트 파싱용)

#### 2.1.2 AI Agent 자동 인식 (핵심 차별점)

- [ ] `tene init` 시 CLAUDE.md 자동 생성 (기본값, `--claude` 플래그)
- [ ] `tene init --cursor` 시 .cursorrules에 tene 가이드 추가
- [ ] `tene init --windsurf` 등 확장 가능한 구조
- [ ] AI Agent가 tene 명령어를 즉시 인식하도록 컨텍스트 파일 자동 관리

#### 2.1.3 Fake Door Test

- [ ] `tene sync` 명령어 포함 (실제 동기화 미구현)
- [ ] 실행 시 Cloud waitlist 안내 화면 표시
- [ ] waitlist 등록 수 수집 (Analytics)

#### 2.1.4 암호화 백업

- [ ] `tene export --encrypted` 암호화된 볼트 파일 수동 백업

### 2.2 Out of Scope — Phase 2: $1 Cloud (수요 검증 후)

Cloud 기능은 MVP에서 완전 제외한다. Fake Door Test(tene sync → waitlist)에서 수요가 확인되면 Phase 2에서 구축한다.

- GitHub OAuth 인증 (Cloud 계정)
- 암호화된 볼트 백업: 로컬 SQLite → 암호화 blob → AWS S3
- 멀티 디바이스 동기화 프로토콜
- API 서버: ECS Fargate + NLB (Hono 프레임워크)
- Cloud DB: RDS PostgreSQL (사용자 + 메타데이터)
- 웹 대시보드: Next.js static export (S3 + CloudFront)
- 감사 로그 (Cloud 측)
- Stripe 결제 ($1/월 구독)

### 2.3 Out of Scope — Phase 3: 팀 기능 (가설)

- 팀 볼트 / RBAC
- 에이전트 스코핑 (에이전트별 시크릿 접근 제어)
- 프로젝트 간 글로벌 키 공유
- 자동 시크릿 로테이션
- MCP 서버 (AI 에이전트 네이티브 통합)
- SSO / SCIM
- Docker / 셀프호스팅
- Homebrew / curl 설치 (Phase 1은 npm만)
- 모바일 앱

---

## 3. Tene가 해결하는 것 / 못 하는 것

### 3.1 해결하는 것 (MVP)

| 문제 | Tene의 해결 방식 |
|------|-----------------|
| **시크릿 안전 저장** | XChaCha20-Poly1305 + SQLite 로컬 볼트. .env 평문 저장 대비 암호화 |
| **시크릿 암호화** | Argon2id KDF → Master Key → XChaCha20-Poly1305. 디스크에 평문 없음 |
| **환경변수 주입** | `tene run -- <command>`로 암호화된 시크릿을 환경변수로 자동 주입 |
| **AI Agent 자동 인식** | `tene init`시 CLAUDE.md/.cursorrules 자동 생성. AI Agent가 시크릿 관리법을 즉시 학습 |
| **.env 마이그레이션** | `tene import .env` 한 줄로 기존 .env에서 전환 |
| **환경별 시크릿 분리** | `tene env dev/staging/prod`로 환경별 시크릿 관리 |
| **암호화 백업** | `tene export --encrypted`로 수동 백업 가능 |

### 3.2 못 하는 것 (MVP 한계)

| 한계 | 설명 | 해결 시기 |
|------|------|----------|
| **시크릿 만료 확인** | 시크릿에 만료일을 설정하거나 만료 알림을 받을 수 없음 | 미정 |
| **자동 갱신** | 시크릿 자동 로테이션/갱신 기능 없음 | Phase 3+ |
| **프로젝트 간 글로벌 키 공유** | 같은 키를 여러 프로젝트에서 공유할 수 없음 (프로젝트별 독립) | Phase 3 (팀 기능) |
| **Cloud 동기화** | 멀티 디바이스 동기화 미지원 (로컬 전용) | Phase 2 (수요 검증 후) |
| **웹 대시보드** | 브라우저에서 시크릿 현황 조회 불가 | Phase 2 |
| **팀 협업** | 팀원 간 시크릿 공유/RBAC 불가 | Phase 3 |

---

## 4. AI Agent 자동 인식 — 핵심 차별점

### 4.1 개요

Tene의 핵심 차별점은 AI Agent가 시크릿 관리법을 **자동으로 인식**하는 것이다. `tene init` 실행 시 AI Agent 컨텍스트 파일(CLAUDE.md, .cursorrules 등)을 자동 생성하여, AI Agent가 별도 학습 없이 즉시 tene을 통해 시크릿을 관리한다.

### 4.2 지원 AI Agent 및 컨텍스트 파일

| AI Agent | 컨텍스트 파일 | 생성 명령어 | 비고 |
|----------|-------------|-----------|------|
| **Claude Code** | `CLAUDE.md` | `tene init` (기본값) 또는 `tene init --claude` | 기본 동작 |
| **Cursor** | `.cursorrules` | `tene init --cursor` | .cursorrules에 tene 섹션 추가 |
| **Windsurf** | `.windsurfrules` | `tene init --windsurf` | 확장 가능 구조 |
| **복수 Agent** | 복수 파일 동시 생성 | `tene init --claude --cursor` | 조합 가능 |

### 4.3 자동 생성되는 CLAUDE.md 내용 (영어)

```markdown
# Secrets Management

This project uses [tene](https://github.com/agentkay/tene) for secret management.

## Usage
- Get a secret: `tene get <KEY>`
- List secrets: `tene list`
- Run with secrets injected: `tene run -- <command>`
- Set a secret: `tene set <KEY> <VALUE>`

## Rules
- Never hardcode secret values in source code
- Access secrets via `process.env.KEY_NAME`
- Do not create .env files — use `tene run` instead
- Use `tene list` to see available secrets
```

### 4.4 .cursorrules에 추가되는 내용

```
# Secrets Management (tene)
This project uses tene for secrets management.
- Get a secret: `tene get <KEY>`
- List secrets: `tene list`
- Run with secrets: `tene run -- <command>`
- NEVER hardcode secret values in code
- Access secrets via process.env.KEY_NAME
```

### 4.5 구현 설계

```typescript
// packages/cli/src/commands/init.ts

interface InitOptions {
  claude?: boolean;   // --claude (기본값 true)
  cursor?: boolean;   // --cursor
  windsurf?: boolean; // --windsurf
}

async function initCommand(name?: string, options: InitOptions = {}) {
  // 1. 볼트 생성 (기존 로직)
  await createVault(name);
  
  // 2. AI Agent 컨텍스트 파일 생성
  const agents = getSelectedAgents(options); // 기본값: ['claude']
  
  for (const agent of agents) {
    await generateAgentContext(agent, projectDir);
  }
  
  // 3. .gitignore에 .tene/ 추가
  await updateGitignore(projectDir);
}

// Agent 컨텍스트 생성기 (확장 가능 구조)
const agentGenerators: Record<string, AgentContextGenerator> = {
  claude: {
    file: 'CLAUDE.md',
    generate: (projectDir) => generateClaudeMd(projectDir),
  },
  cursor: {
    file: '.cursorrules',
    generate: (projectDir) => appendToCursorRules(projectDir),
  },
  windsurf: {
    file: '.windsurfrules',
    generate: (projectDir) => appendToWindsurfRules(projectDir),
  },
};
```

### 4.6 기존 파일 병합 정책

| 상황 | 동작 |
|------|------|
| CLAUDE.md가 없을 때 | 새로 생성 |
| CLAUDE.md가 이미 있을 때 | "# Secrets Management" 섹션만 추가/업데이트 |
| .cursorrules가 이미 있을 때 | 파일 끝에 tene 섹션 추가 (기존 내용 보존) |
| tene 섹션이 이미 있을 때 | 스킵 (중복 방지) |

### 4.7 AI Agent 사용 시나리오

```bash
# 1. 프로젝트 초기화 (CLAUDE.md 자동 생성)
$ tene init my-project
  ✓ Vault created (.tene/vault.db)
  ✓ CLAUDE.md generated (AI agent will auto-detect tene)
  
# 2. 시크릿 저장
$ tene set STRIPE_KEY sk_test_xxxxx

# 3. Claude Code가 코드 작성 시 자동으로 tene 사용:
#    (CLAUDE.md를 읽고 tene 명령어를 인식)
#    "이 프로젝트는 tene으로 시크릿을 관리하므로,
#     STRIPE_KEY=$(tene get STRIPE_KEY) 를 사용합니다."

# 4. Cursor 사용자의 경우:
$ tene init my-project --cursor
  ✓ Vault created (.tene/vault.db)
  ✓ CLAUDE.md generated
  ✓ .cursorrules updated with tene guide
```

---

## 5. Fake Door Test — Cloud 수요 검증

### 5.1 개요

Cloud 동기화 기능(`tene sync`)을 MVP에 명령어로 포함하되, 실제 동기화는 구현하지 않는다. 사용자가 `tene sync`를 실행하면 Cloud waitlist 안내 화면을 표시하고, waitlist 등록 수를 수집하여 Cloud 수요를 검증한다.

### 5.2 tene sync 실행 화면

```
$ tene sync

  ☁️  Tene Cloud Sync — Coming Soon!

  Cloud sync will enable:
  • Multi-device secret synchronization
  • Encrypted cloud backup (zero-knowledge)
  • Web dashboard for secret overview
  • All for just $1/month

  Join the waitlist to get early access:
  → https://tene.dev/waitlist

  In the meantime, use `tene export --encrypted` for local backup.

  [Open waitlist page? (Y/n)]
```

### 5.3 수요 검증 기준

| 지표 | 목표 | 행동 |
|------|------|------|
| `tene sync` 실행 횟수 / DAU | 15%+ | Cloud 구축 시작 (Phase 2) |
| waitlist 등록 수 | 100명+ | Cloud 구축 시작 (Phase 2) |
| `tene sync` 실행 횟수 / DAU | < 5% | Cloud 보류, 로컬 기능 강화 |

### 5.4 수동 백업 대안: tene export --encrypted

Cloud 동기화가 없는 MVP 기간에는 `tene export --encrypted`로 암호화된 볼트 파일을 수동 백업할 수 있다.

```
$ tene export --encrypted
  Encrypted vault exported to: ./my-project.tene.enc
  
  This file is encrypted with your Master Password.
  To restore: tene import --encrypted my-project.tene.enc
  
  Store this file in a safe place (USB, cloud drive, etc.)
```

```
$ tene import --encrypted my-project.tene.enc
  Enter Master Password: ********
  
  Restoring vault from encrypted backup...
  5 secrets restored to "my-project" vault.
```

---

## 6. 시스템 아키텍처

### 6.1 전체 아키텍처 — Phase 1 MVP (로컬 전용)

```
╔═══════════════════════════════════════════════════════════════════════════╗
║                     PHASE 1 MVP (로컬 전용, 서버 없음, 비용 $0)            ║
║                                                                         ║
║   사용자 / AI 에이전트                                                    ║
║         |                                                               ║
║         v                                                               ║
║   ┌─────────────────────────────────────────────────────┐               ║
║   │              CLI (tene)                              │               ║
║   │              npm global package                      │               ║
║   │                                                      │               ║
║   │   ┌──────────┐  ┌──────────┐  ┌───────────────┐    │               ║
║   │   │ Commands │  │  Crypto  │  │   Keychain    │    │               ║
║   │   │ init/set │  │ Argon2id │  │  OS Native    │    │               ║
║   │   │ get/run  │  │ XChaCha  │  │  Master Key   │    │               ║
║   │   │ list/del │  │ 20-Poly  │  │  Storage      │    │               ║
║   │   │ imp/exp  │  │ 1305     │  │               │    │               ║
║   │   │ sync(FD) │  │          │  │               │    │               ║
║   │   └────┬─────┘  └────┬─────┘  └───────────────┘    │               ║
║   │        │              │                              │               ║
║   │        v              v                              │               ║
║   │   ┌────────────────────────────────────────────┐    │               ║
║   │   │         SQLite Local Vault                  │    │               ║
║   │   │         (.tene/vault.db)                    │    │               ║
║   │   │                                             │    │               ║
║   │   │  secrets: name + encrypted_value + env     │    │               ║
║   │   │  metadata: project info, environments      │    │               ║
║   │   │  audit: local access log                   │    │               ║
║   │   └────────────────────────────────────────────┘    │               ║
║   │                                                      │               ║
║   │   ┌────────────────────────────────────────────┐    │               ║
║   │   │         AI Agent 자동 인식                   │    │               ║
║   │   │                                             │    │               ║
║   │   │  CLAUDE.md      ← tene init (기본)          │    │               ║
║   │   │  .cursorrules   ← tene init --cursor        │    │               ║
║   │   │  .windsurfrules ← tene init --windsurf      │    │               ║
║   │   └────────────────────────────────────────────┘    │               ║
║   └─────────────────────────────────────────────────────┘               ║
║                                                                         ║
║   ** 네트워크 통신: 없음 **                                               ║
║   ** 서버: 없음 **                                                       ║
║   ** 해킹 대상: 없음 **                                                   ║
║   ** 비용: $0 **                                                         ║
║                                                                         ║
║   ┌─────────────────────────────────────────────────────┐               ║
║   │  Fake Door: tene sync → waitlist 안내               │               ║
║   │  대안: tene export --encrypted → 수동 암호화 백업     │               ║
║   └─────────────────────────────────────────────────────┘               ║
╚═══════════════════════════════════════════════════════════════════════════╝
```

### 6.2 핵심 아키텍처 원칙

| 원칙 | 설명 |
|------|------|
| **Local-Only (MVP)** | MVP는 100% 로컬. Cloud는 Phase 2 (수요 검증 후) |
| **Server Cost $0** | MVP 기간 서버 인프라 비용 완전 제로 |
| **AI Agent First** | AI Agent 자동 인식이 핵심 차별점. CLAUDE.md/.cursorrules 자동 생성 |
| **Zero-Knowledge (설계)** | Phase 2 Cloud 구축 시에도 서버는 시크릿 평문을 절대 알 수 없는 구조 |
| **Offline 100%** | 모든 기능이 인터넷 없이 완전히 동작 |
| **Encrypted at Rest** | 모든 시크릿은 XChaCha20-Poly1305로 암호화 저장 |

---

## 7. 모노레포 디렉토리 구조

```
tene/
├── packages/
│   ├── cli/                              # CLI 패키지 (tene) — npm global
│   │   ├── src/
│   │   │   ├── commands/
│   │   │   │   ├── init.ts               # tene init (볼트 + Master PW + AI Agent 컨텍스트)
│   │   │   │   ├── set.ts                # tene set KEY VALUE
│   │   │   │   ├── get.ts                # tene get KEY (stdout)
│   │   │   │   ├── run.ts                # tene run -- COMMAND
│   │   │   │   ├── list.ts               # tene list
│   │   │   │   ├── delete.ts             # tene delete KEY
│   │   │   │   ├── import.ts             # tene import .env / --encrypted
│   │   │   │   ├── export.ts             # tene export / --encrypted
│   │   │   │   ├── env.ts                # tene env [name]
│   │   │   │   ├── passwd.ts              # tene passwd (Master Password 변경)
│   │   │   │   ├── recover.ts            # tene recover (Recovery Key로 복구)
│   │   │   │   ├── sync.ts               # tene sync (Fake Door → waitlist)
│   │   │   │   └── whoami.ts             # tene whoami
│   │   │   ├── lib/
│   │   │   │   ├── vault.ts              # SQLite 볼트 매니저
│   │   │   │   ├── crypto.ts             # 암호화 래퍼 (@tene/crypto 사용)
│   │   │   │   ├── keychain.ts           # OS Keychain 연동
│   │   │   │   ├── config.ts             # CLI 설정 관리
│   │   │   │   ├── agent-context.ts      # AI Agent 컨텍스트 파일 생성기
│   │   │   │   ├── fake-door.ts          # Fake Door Test 로직
│   │   │   │   └── output.ts             # CLI 출력 포매터
│   │   │   └── index.ts                  # CLI 엔트리포인트
│   │   ├── bin/
│   │   │   └── tene.ts                   # 실행 바이너리
│   │   ├── templates/                    # AI Agent 컨텍스트 템플릿
│   │   │   ├── claude.md.hbs             # CLAUDE.md 템플릿
│   │   │   ├── cursorrules.hbs           # .cursorrules 템플릿
│   │   │   └── windsurfrules.hbs         # .windsurfrules 템플릿
│   │   ├── tsconfig.json
│   │   └── package.json
│   │
│   ├── crypto/                           # 공유 암호화 모듈 (핵심!)
│   │   ├── src/
│   │   │   ├── kdf.ts                    # Argon2id 키 유도
│   │   │   ├── encrypt.ts                # XChaCha20-Poly1305 암호화
│   │   │   ├── decrypt.ts                # XChaCha20-Poly1305 복호화
│   │   │   ├── key-manager.ts            # 마스터키 / 파생키 관리
│   │   │   ├── recovery.ts               # Recovery Key 생성/검증
│   │   │   ├── vault-export.ts           # 암호화 볼트 내보내기/가져오기
│   │   │   └── index.ts                  # 공개 API
│   │   ├── __tests__/                    # 95%+ 커버리지 목표
│   │   │   ├── kdf.test.ts
│   │   │   ├── encrypt.test.ts
│   │   │   ├── decrypt.test.ts
│   │   │   ├── recovery.test.ts
│   │   │   └── vault-export.test.ts
│   │   ├── tsconfig.json
│   │   └── package.json
│   │
│   └── types/                            # 공유 타입 패키지
│       ├── src/
│       │   ├── secret.ts
│       │   ├── vault.ts
│       │   ├── agent.ts                  # AI Agent 관련 타입
│       │   └── index.ts
│       ├── tsconfig.json
│       └── package.json
│
├── docs/                                 # PDCA 문서
│   ├── 00-pm/
│   ├── 01-plan/
│   │   └── features/
│   ├── 02-design/
│   ├── 03-analysis/
│   └── 04-report/
│
├── .github/
│   └── workflows/
│       ├── ci.yml                        # CI: lint, test, build
│       └── release.yml                   # npm 배포
│
├── pnpm-workspace.yaml                   # pnpm workspace
├── biome.json                            # Biome 린트/포맷
├── tsconfig.json                         # Root tsconfig
├── package.json
└── CLAUDE.md
```

### 7.1 패키지 간 의존성

```
                 ┌──────────┐
                 │ @tene/   │
                 │ types    │
                 └────┬─────┘
                      │ (모든 패키지가 참조)
                      │
                 ┌────▼───┐
                 │ @tene/ │
                 │ crypto │
                 └───┬────┘
                     │
                ┌────▼──────────────┐
                │   packages/cli    │
                │   (tene CLI)      │
                └───────────────────┘

핵심 원칙:
- @tene/crypto: CLI에서 사용하는 암호화 코어 (향후 Cloud에서도 동일 로직 재사용)
- @tene/types: 모든 패키지의 공통 타입 정의
- CLI: 로컬 모드에서 crypto + SQLite만 사용 (Cloud 의존 없음)
- apps/api, apps/web: Phase 2에서 추가 (MVP에는 없음)
```

---

## 8. 기술 스택 선정 및 근거

### 8.1 기술 스택 총괄 — Phase 1 MVP

| 영역 | 기술 | 선정 근거 |
|------|------|----------|
| **언어** | TypeScript 5.x | CLI-API-Web 전체 통일. 바이브코더 생태계(npm). 타입 안전성 |
| **런타임** | Node.js 22 LTS | 장기 지원. CLI 런타임 |
| **모노레포** | pnpm workspace | 의존성 효율, npm 호환, 업계 표준 |
| **CLI 프레임워크** | Commander.js | 경량(~30KB), 성숙도, 낮은 학습 곡선 |
| **로컬 DB** | better-sqlite3 | Node.js 최고 SQLite 바인딩, 동기 API, 빠른 검색 |
| **암호화 라이브러리** | libsodium (libsodium-wrappers) | 검증된 암호화, Node.js/브라우저 크로스플랫폼 |
| **KDF** | Argon2id (via libsodium) | 최신 키 유도 함수, 메모리 하드, OWASP 권장 |
| **대칭 암호화** | XChaCha20-Poly1305 | 192-bit nonce(충돌 안전), libsodium 네이티브, AES-256-GCM과 동등한 256-bit 보안 |
| **키체인** | keytar | OS 네이티브 키체인 (macOS Keychain, Linux Secret Service, Win Credential Vault) |
| **유효성 검사** | Zod | TypeScript-first 스키마 검증, CLI 인자 검증 |
| **테스트** | Vitest | 빠름, ESM 네이티브 |
| **린트/포맷** | Biome | ESLint + Prettier 대체, 빠름, 설정 최소화 |
| **CI/CD** | GitHub Actions | GitHub 통합, npm 자동 배포 |

### 8.2 핵심 기술 선정 상세 근거

#### 로컬 DB: better-sqlite3 (vs. 파일 기반 .env 암호화)

| 기준 | better-sqlite3 | 암호화된 .env |
|------|---------------|--------------|
| 구조화 검색 | 인덱스 기반 O(log n) | 파일 전체 파싱 O(n) |
| 환경별 분리 | SQL WHERE 조건 | 파일 분리 필요 |
| 동시 접근 | WAL 모드 안전 | 파일 락 필요 |
| 부분 업데이트 | 행 단위 UPDATE | 전체 파일 재작성 |
| 감사 로그 | 같은 DB에 기록 | 별도 파일 필요 |
| 결론 | **선택** — 구조화 장점이 압도적 | Dotenvx 방식 |

---

## 9. 보안 아키텍처

### 9.1 보안 모델 (로컬 전용)

| 층위 | 전략 | 적용 |
|------|------|------|
| **Layer 1: Server-Free** | 서버가 없다 (MVP) | 해킹 대상 자체가 없음. 공격 표면 = 0 |
| **Layer 2: Encrypted at Rest** | 모든 시크릿이 암호화 저장 | 디바이스 분실/도난 시에도 시크릿 안전 |
| **Layer 3: OS Keychain** | Master Key를 OS 보안 저장소에 보관 | 프로세스 간 격리, 하드웨어 암호화 |

### 9.2 보안 플로우

```
[사용자 디바이스 — 오직 여기에만 존재]
+──────────────────────────────────────────+
│                                          │
│  Master Password (사용자만 알고 있음)     │
│       |                                  │
│       v                                  │
│  Argon2id KDF                            │
│  (memory: 64MB, iterations: 3,           │
│   parallelism: 1, outputLen: 32)         │
│       |                                  │
│       v                                  │
│  Master Key (256-bit)                    │
│       |                                  │
│       +---> OS Keychain 저장              │
│       |     (macOS Keychain /            │
│       |      Linux Secret Service)       │
│       |                                  │
│       v                                  │
│  XChaCha20-Poly1305 Encrypt              │
│  (192-bit random nonce, AAD=key name)    │
│       |                                  │
│       v                                  │
│  SQLite DB (.tene/vault.db)              │
│  encrypted secrets stored locally        │
│                                          │
│  ** 네트워크 통신: 없음 **                │
│  ** 서버: 없음 **                        │
│  ** 해킹 대상: 없음 **                   │
+──────────────────────────────────────────+
```

### 9.3 보안 원칙 요약

| 원칙 | MVP (Phase 1) |
|------|:-------------:|
| 시크릿 평문은 로컬에만 존재 | O |
| 모든 시크릿은 암호화 저장 | O (SQLite) |
| Master Key는 OS Keychain에 저장 | O |
| 오프라인 100% 동작 | O |
| 네트워크 통신 없음 | O |
| 오픈소스 검증 가능 | O |

---

## 10. 키 유도 및 암호화/복호화 플로우

### 10.1 키 유도 (Key Derivation)

```typescript
// packages/crypto/src/kdf.ts

// 1단계: Master Password에서 Master Key 유도
const salt = crypto.randomBytes(16); // 128-bit random salt
const masterKey = argon2id(password, salt, {
  memoryLimit: 64 * 1024 * 1024,   // 64MB (로컬 환경 고려)
  opsLimit: 3,                      // 3 iterations
  parallelism: 1,                   // 단일 스레드
  outputLength: 32                  // 256-bit
});

// 2단계: Master Key에서 용도별 키 파생 (HKDF 방식)
const encKey   = deriveKey(masterKey, "tene-encryption-key", 32);  // 시크릿 암호화
const authHash = deriveKey(masterKey, "tene-auth-hash",     32);  // (Phase 2: Cloud 인증용)

// deriveKey는 libsodium의 crypto_kdf_derive_from_key 사용
```

### 10.2 Recovery Key 생성 (12단어 니모닉, BIP-39 스타일)

```typescript
// packages/crypto/src/recovery.ts

// tene init 시 Recovery Key 생성
// 128-bit entropy → BIP-39 워드리스트 기반 12단어 니모닉
const entropy = crypto.randomBytes(16); // 128-bit
const mnemonic = entropyToMnemonic(entropy, BIP39_WORDLIST);
// 예: "apple banana cherry dolphin eagle frost grape harbor island jungle kite lemon"

// 니모닉에서 Recovery Key 유도
const recoveryKey = mnemonicToKey(mnemonic); // 니모닉 → 256-bit key (PBKDF2 또는 HKDF)

// Recovery Key로 Master Key 암호화하여 볼트에 저장
const encryptedMasterKey = XChaCha20Poly1305.encrypt(
  key: deriveKey(recoveryKey, "tene-recovery", 32),
  plaintext: masterKey,
  nonce: randomNonce()
);

// vault.db의 vault_meta 테이블에 recovery_blob (base64)으로 저장
// 니모닉 자체는 사용자에게만 표시 (볼트에 저장하지 않음)

// 니모닉 검증 함수
function validateMnemonic(mnemonic: string): boolean {
  const words = mnemonic.trim().split(/\s+/);
  if (words.length !== 12) return false;
  return words.every(word => BIP39_WORDLIST.includes(word));
  // BIP-39 워드리스트: 2048개 영어 단어
}
```

### 10.3 시크릿 암호화 플로우 (tene set KEY VALUE)

```
[CLI] tene set STRIPE_KEY sk_test_xxxxx

1. OS Keychain에서 Master Key 로드
   (없으면 Master Password 입력 요청 → Argon2id → Master Key)

2. Master Key에서 Encryption Key 파생
   encKey = deriveKey(masterKey, "tene-encryption-key", 32)

3. 시크릿 값 암호화:
   nonce = random_192bit()
   encrypted = XChaCha20-Poly1305.encrypt(
     key: encKey,
     plaintext: "sk_test_xxxxx",
     nonce: nonce,
     additionalData: "STRIPE_KEY"      // 키 이름을 AAD로 (변조 방지)
   )

4. SQLite 볼트에 저장:
   INSERT INTO secrets (name, encrypted_value, environment, version)
   VALUES ('STRIPE_KEY', base64(nonce + encrypted), 'default', 1)

5. 감사 로그 기록 (로컬):
   INSERT INTO audit_log (action, resource_name, timestamp)
   VALUES ('secret.write', 'STRIPE_KEY', NOW())
```

### 10.4 시크릿 복호화 플로우 (tene get KEY)

```
[CLI] tene get STRIPE_KEY

1. SQLite 볼트에서 암호화된 blob 조회:
   SELECT encrypted_value FROM secrets
   WHERE name='STRIPE_KEY' AND environment='default'

2. OS Keychain에서 Master Key 로드

3. Encryption Key 파생:
   encKey = deriveKey(masterKey, "tene-encryption-key", 32)

4. 복호화:
   blob = base64Decode(encrypted_value)
   nonce = blob[0:24]                    // 앞 24바이트
   ciphertext = blob[24:]
   plaintext = XChaCha20-Poly1305.decrypt(
     key: encKey,
     ciphertext: ciphertext,
     nonce: nonce,
     additionalData: "STRIPE_KEY"        // AAD 검증
   )

5. stdout 출력: "sk_test_xxxxx"
   (AI 에이전트가 Bash에서 파싱 가능)
```

### 10.5 시크릿 주입 플로우 (tene run -- COMMAND)

```
[CLI] tene run -- cursor .

1. 현재 환경의 모든 시크릿을 SQLite에서 조회
2. 각 시크릿을 로컬에서 복호화
3. 환경변수로 설정한 자식 프로세스 생성:
   child_process.spawn("cursor", ["."], {
     env: {
       ...process.env,
       STRIPE_KEY: "sk_test_xxxxx",
       DATABASE_URL: "postgresql://...",
     },
     stdio: "inherit"
   })
4. 자식 프로세스 종료 시 환경변수 자동 정리
   (디스크에 시크릿이 평문으로 저장되지 않음)
```

### 10.6 암호화 백업 플로우 (tene export --encrypted)

```
[CLI] tene export --encrypted

1. SQLite 볼트 전체를 읽기 (이미 암호화된 시크릿 포함)
2. 볼트 전체를 XChaCha20-Poly1305로 2차 암호화:
   - Master Key로 전체 볼트 데이터를 암호화
   - 결과물에 KDF 파라미터, salt 포함
3. 단일 .tene.enc 파일로 저장:
   - 파일 헤더: magic bytes + version + KDF params
   - 페이로드: 암호화된 볼트 데이터

[CLI] tene import --encrypted my-project.tene.enc

1. 파일 헤더에서 KDF 파라미터 추출
2. Master Password 입력 요청
3. Argon2id로 Master Key 유도
4. 2차 암호화 복호화 → 볼트 데이터 복원
5. 로컬 SQLite 볼트에 머지
```

---

## 11. 데이터 모델

### 11.1 Local SQLite 스키마 (.tene/vault.db)

```sql
-- vault_meta: 볼트 메타데이터
CREATE TABLE vault_meta (
  key             TEXT PRIMARY KEY,
  value           TEXT NOT NULL
);
-- 저장 항목: vault_version, created_at, kdf_salt (base64),
--           kdf_params (JSON), recovery_blob (base64)

-- environments: 환경 관리
CREATE TABLE environments (
  id              INTEGER PRIMARY KEY AUTOINCREMENT,
  name            TEXT NOT NULL UNIQUE,              -- default, dev, staging, prod
  is_default      INTEGER NOT NULL DEFAULT 0,
  created_at      TEXT NOT NULL DEFAULT (datetime('now'))
);

-- secrets: 암호화된 시크릿
CREATE TABLE secrets (
  id              INTEGER PRIMARY KEY AUTOINCREMENT,
  environment_id  INTEGER NOT NULL REFERENCES environments(id) ON DELETE CASCADE,
  name            TEXT NOT NULL,                     -- 시크릿 키 이름 (평문)
  encrypted_value TEXT NOT NULL,                     -- nonce + 암호문 (base64)
  version         INTEGER NOT NULL DEFAULT 1,
  created_at      TEXT NOT NULL DEFAULT (datetime('now')),
  updated_at      TEXT NOT NULL DEFAULT (datetime('now')),
  UNIQUE(environment_id, name)
);

-- audit_log: 로컬 감사 로그
CREATE TABLE audit_log (
  id              INTEGER PRIMARY KEY AUTOINCREMENT,
  action          TEXT NOT NULL,                     -- secret.read, secret.write, etc.
  resource_name   TEXT,                              -- 시크릿 이름
  environment     TEXT,                              -- 환경 이름
  source          TEXT NOT NULL DEFAULT 'cli',       -- cli | agent | export
  timestamp       TEXT NOT NULL DEFAULT (datetime('now'))
);

-- 인덱스
CREATE INDEX idx_secrets_env_name ON secrets(environment_id, name);
CREATE INDEX idx_audit_timestamp ON audit_log(timestamp DESC);
```

### 11.2 로컬 저장소 구조

```
~/.tene/                              # 글로벌 CLI 설정 디렉토리
└── config.json                       # CLI 전역 설정 (파일 퍼미션 0600)
    {
      "defaultEnvironment": "default",
      "analytics": {
        "syncAttempts": 0,            // Fake Door 측정용
        "lastSyncAttempt": null
      }
    }
    # 참고: salt는 글로벌이 아닌 각 프로젝트의 vault.db vault_meta 테이블에 kdf_salt로 저장
    # 프로젝트마다 다른 salt = 프로젝트 간 암호화 독립성 보장

project/.tene/                        # 프로젝트 로컬 볼트 (tene init으로 생성)
├── vault.db                          # SQLite 암호화 볼트
├── vault.json                        # 볼트 메타데이터
│   {
│     "projectName": "my-project",
│     "createdAt": "2026-04-06T12:00:00Z",
│     "vaultVersion": 1,
│     "agents": ["claude"]            // 활성화된 AI Agent 목록
│   }
└── .gitignore                        # .tene/ 전체를 Git에서 제외

project/CLAUDE.md                     # AI Agent 컨텍스트 (tene init이 생성)
project/.cursorrules                  # Cursor용 (tene init --cursor)

OS Keychain:
  - Service: "tene"
  - Account: "{project-path}" 또는 "global"
  - Password: Master Key (32 bytes, base64 encoded)
```

---

## 12. CLI 명령어 설계

### 12.1 Phase 1 MVP 명령어

| 명령어 | 우선순위 | 설명 | 인자 |
|--------|:--------:|------|------|
| `tene init` | P0 | 프로젝트 초기화 (볼트 + Master PW + AI Agent 컨텍스트) | `[name]` `--claude` `--cursor` `--windsurf` |
| `tene set <key> <value>` | P0 | 시크릿 저장 (로컬 암호화) | `--env <name>` `--stdin` |
| `tene get <key>` | P0 | 시크릿 조회 (stdout) | `--env <name>` |
| `tene run -- <command>` | P0 | 시크릿 주입 후 명령 실행 | `--env <name>` |
| `tene list` | P0 | 시크릿 목록 (값 마스킹) | `--env <name>` |
| `tene delete <key>` | P0 | 시크릿 삭제 | `--env <name>` `--force` |
| `tene import <file>` | P1 | .env에서 일괄 가져오기 | `--env <name>` `--overwrite` `--encrypted` |
| `tene export` | P1 | .env 형식/암호화 백업 내보내기 | `--env <name>` `--file <path>` `--encrypted` |
| `tene env [name]` | P1 | 환경 전환/목록/생성 | `list`, `create <name>` |
| `tene passwd` | P0 | Master Password 변경 + 볼트 재암호화 + 새 Recovery Key 발급 | -- (대화형) |
| `tene recover` | P0 | Recovery Key로 Master Password 재설정 | -- (대화형) |
| `tene sync` | P1 | Fake Door: waitlist 안내 표시 | -- |
| `tene whoami` | P2 | 현재 상태 표시 | -- |

### 12.2 명령어 상세 동작

#### tene init (핵심 — 서버 없음, AI Agent 컨텍스트 자동 생성)

```
$ cd my-project
$ tene init

  Welcome to Tene! Let's set up your local secret vault.

  Project name (my-project):

  Set your Master Password (used to encrypt all secrets):
  Master Password: ********
  Confirm: ********

  Generating encryption keys...

  Recovery Key (write this down and keep it safe!):
  +--------------------------------------------------+
  |   apple banana cherry dolphin eagle frost          |
  |   grape harbor island jungle kite lemon            |
  |                                                    |
  |   If you forget your Master Password,              |
  |   this is the ONLY way to recover.                 |
  +--------------------------------------------------+

  ✓ Created .tene/vault.db (local encrypted vault)
  ✓ Added .tene/ to .gitignore
  ✓ Master Key saved to OS Keychain
  ✓ Generated CLAUDE.md (AI agents will auto-detect tene)

  Project "my-project" initialized.
  Default environment "default" created.

  Next: tene set KEY VALUE to add your first secret.

  Tip: No server needed. Your secrets stay on this device.
       AI agents (Claude Code) will automatically use tene.
```

```
# Cursor 사용자:
$ tene init --cursor

  ... (위와 동일)
  ✓ Generated CLAUDE.md (AI agents will auto-detect tene)
  ✓ Updated .cursorrules with tene guide

# 복수 Agent:
$ tene init --claude --cursor --windsurf

  ... (위와 동일)
  ✓ Generated CLAUDE.md
  ✓ Updated .cursorrules with tene guide
  ✓ Updated .windsurfrules with tene guide
```

#### tene set (로컬 암호화 저장)

```
$ tene set STRIPE_KEY sk_test_xxxxx
  STRIPE_KEY saved (encrypted, default)

# stdin 지원 (shell history 방지)
$ echo "sk_test_xxxxx" | tene set STRIPE_KEY --stdin
  STRIPE_KEY saved (encrypted, default)

# 환경 지정
$ tene set DATABASE_URL postgresql://user:pass@host/db --env prod
  DATABASE_URL saved (encrypted, prod)
```

#### tene get (stdout 출력 — AI 에이전트 호출용)

```
$ tene get STRIPE_KEY
sk_test_xxxxx

# AI 에이전트 활용:
$ STRIPE_KEY=$(tene get STRIPE_KEY)

# JSON 출력:
$ tene get STRIPE_KEY --json
{"name":"STRIPE_KEY","value":"sk_test_xxxxx","environment":"default"}
```

#### tene run (시크릿 주입 실행)

```
$ tene run -- cursor .
  Injecting 5 secrets into environment...
  Starting: cursor .

$ tene run --env prod -- node server.js
  Injecting 8 secrets (prod) into environment...
  Starting: node server.js
```

#### tene list (목록 + 마스킹)

```
$ tene list
  Project: my-project (default)

  NAME              VALUE           UPDATED
  STRIPE_KEY        sk_te*****      2 minutes ago
  DATABASE_URL      postg*****      5 minutes ago
  API_SECRET        eyJhb*****      1 hour ago

  3 secrets in "default" environment

# JSON 출력:
$ tene list --json
[{"name":"STRIPE_KEY","preview":"sk_te*****","updatedAt":"..."}]
```

#### tene import / export

```
$ tene import .env
  Found 5 secrets in .env:
    STRIPE_KEY, DATABASE_URL, API_SECRET, SENDGRID_KEY, JWT_SECRET

  Import 5 secrets to "my-project" (default)? (y/N) y
  5 secrets imported (encrypted).

  Tip: You can now delete .env and use tene run instead.

$ tene export
  STRIPE_KEY=sk_test_xxxxx
  DATABASE_URL=postgresql://user:pass@host/db

$ tene export --file .env.local
  5 secrets exported to .env.local
  Warning: This file contains plain-text secrets. Do not commit it.

$ tene export --encrypted
  Encrypted vault exported to: ./my-project.tene.enc
  
  This file is encrypted with your Master Password.
  To restore: tene import --encrypted my-project.tene.enc

$ tene import --encrypted my-project.tene.enc
  Enter Master Password: ********
  5 secrets restored to "my-project" vault.
```

#### tene sync (Fake Door Test)

```
$ tene sync

  ☁️  Tene Cloud Sync — Coming Soon!

  Cloud sync will enable:
  • Multi-device secret synchronization
  • Encrypted cloud backup (zero-knowledge)
  • Web dashboard for secret overview
  • All for just $1/month

  Join the waitlist to get early access:
  → https://tene.dev/waitlist

  In the meantime, use `tene export --encrypted` for local backup.

  [Open waitlist page? (Y/n)]
```

### 12.3 CLI 에러 처리

| 상황 | 에러 메시지 | 종료 코드 |
|------|-----------|:---------:|
| 볼트 미초기화 | `Not in a Tene project. Run \`tene init\` first.` | 1 |
| 시크릿 없음 | `Secret "KEY" not found in "env-name" environment.` | 1 |
| Master PW 오류 | `Invalid Master Password. Try again or use Recovery Key.` | 2 |
| Recovery Key 오류 | `Invalid Recovery Key.` | 2 |
| 이미 존재 | `Secret "KEY" already exists. Use --overwrite to replace.` | 1 |
| Keychain 실패 | `Cannot access OS Keychain. Enter Master Password manually.` | 0 (폴백) |

### 12.4 글로벌 플래그

| 플래그 | 설명 |
|--------|------|
| `--version, -v` | 버전 출력 |
| `--help, -h` | 도움말 |
| `--json` | JSON 형식 출력 (AI 에이전트 파싱용) |
| `--quiet, -q` | 최소 출력 (에러만) |
| `--env <name>` | 환경 지정 (기본: 현재 환경) |
| `--no-color` | 색상 출력 비활성화 |

---

## 13. Phase 정리

### 13.1 Phase 1: MVP (2주) — 로컬 CLI + AI Agent 통합 + Fake Door

| 구성 | 내용 |
|------|------|
| **핵심** | CLI + SQLite + crypto + AI Agent 자동 인식 + Fake Door |
| **비용** | $0 (서버 없음) |
| **기간** | 2주 |
| **목표** | 시크릿 관리 + AI Agent 자동 인식. Cloud 수요 검증 |

### 13.2 Phase 2: $1 Cloud (수요 검증 후)

**진입 조건**: Fake Door Test에서 `tene sync` 실행률 15%+ 또는 waitlist 100명+

| 구성 | 내용 |
|------|------|
| **인프라** | ECS Fargate + NLB + RDS PostgreSQL + S3 |
| **기능** | Cloud 동기화, 웹 대시보드, 감사 로그, Stripe 결제 |
| **비용** | 사용자에게 $1/월. 인프라 비용은 ECS+RDS 기준 |
| **기간** | 3-4주 (수요 확인 후) |

**Cloud 인프라: Lambda/Aurora Serverless를 사용하지 않는 이유**

| 항목 | Lambda + Aurora Serverless | ECS Fargate + RDS PostgreSQL |
|------|--------------------------|------------------------------|
| 비용 예측성 | 요청당 과금 (변동) | 고정 비용 (예측 가능) |
| Cold Start | 100-500ms | 없음 |
| DB 최소 비용 | Aurora 0.5 ACU ~$43/월 | RDS db.t4g.micro ~$15/월 |
| 운영 복잡도 | API Gateway + Lambda 설정 복잡 | 단순한 컨테이너 |
| 확장성 | 자동 (0→무한) | ECS Auto Scaling |

**Phase 2 비용 추정 (ECS + RDS)**

| 항목 | 100 유료 사용자 | 1,000 유료 사용자 |
|------|:--------------:|:----------------:|
| **ECS Fargate** (0.25 vCPU + 0.5GB) | ~$10/월 | ~$20/월 |
| **NLB** | ~$16/월 | ~$16/월 |
| **RDS PostgreSQL** (db.t4g.micro) | ~$15/월 | ~$30/월 |
| **S3 (볼트 저장)** | ~$0.50 | ~$5 |
| **CloudFront** | ~$1 | ~$5 |
| **Route 53** | ~$0.50 | ~$0.50 |
| **월간 총 비용** | **~$43** | **~$77** |
| **월간 총 수익** | **$100** | **$1,000** |
| **이익률** | **57%** | **92.3%** |

### 13.3 Phase 3: 팀 기능 (가설)

| 구성 | 내용 |
|------|------|
| **기능** | 팀 볼트 공유, RBAC, 에이전트 스코핑, MCP 서버 |
| **진입 조건** | Phase 2에서 유료 사용자 확보 + 팀 기능 수요 확인 |
| **가격** | 미정 (Fake Door Test로 수요 확인) |

---

## 14. 요구사항

### 14.1 기능 요구사항

| ID | 요구사항 | 우선순위 | Phase | 상태 |
|----|---------|:--------:|:-----:|:----:|
| FR-01 | `tene init`: Master Password + 볼트 + Recovery Key + AI Agent 컨텍스트 | P0 | 1 | Pending |
| FR-02 | `tene set KEY VALUE`: XChaCha20-Poly1305 로컬 암호화 저장 | P0 | 1 | Pending |
| FR-03 | `tene get KEY`: 복호화 후 stdout 출력 (AI 에이전트 Bash 호출) | P0 | 1 | Pending |
| FR-04 | `tene run -- CMD`: 모든 시크릿을 환경변수로 주입 후 명령 실행 | P0 | 1 | Pending |
| FR-05 | `tene list`: 시크릿 목록 표시 (값 마스킹) | P0 | 1 | Pending |
| FR-06 | `tene delete KEY`: 시크릿 삭제 (확인 프롬프트) | P0 | 1 | Pending |
| FR-07 | `tene import .env`: .env 파일에서 시크릿 일괄 가져오기 | P1 | 1 | Pending |
| FR-08 | `tene export`: .env 형식으로 내보내기 | P1 | 1 | Pending |
| FR-09 | `tene export --encrypted`: 암호화 볼트 백업 | P1 | 1 | Pending |
| FR-10 | `tene import --encrypted`: 암호화 볼트 복원 | P1 | 1 | Pending |
| FR-11 | `tene env [name]`: 환경 전환/목록/생성 | P1 | 1 | Pending |
| FR-12 | Master Password + Argon2id KDF + XChaCha20-Poly1305 암호화 | P0 | 1 | Pending |
| FR-13 | OS Keychain 연동: Master Key를 OS 키체인에 저장 + 파일 폴백 | P0 | 1 | Pending |
| FR-14 | Recovery Key: 생성, 표시, Master Password 복구 | P0 | 1 | Pending |
| FR-15 | --json 플래그: JSON 형식 출력 (AI 에이전트 파싱용) | P1 | 1 | Pending |
| FR-16 | --stdin 플래그: tene set에서 stdin 입력 (shell history 방지) | P1 | 1 | Pending |
| FR-17 | 로컬 감사 로그: 모든 시크릿 접근/수정 기록 (SQLite) | P2 | 1 | Pending |
| FR-18 | AI Agent 자동 인식: `tene init`시 CLAUDE.md 자동 생성 | P0 | 1 | Pending |
| FR-19 | AI Agent 확장: `--cursor`, `--windsurf` 플래그 | P1 | 1 | Pending |
| FR-20 | Fake Door: `tene sync` → waitlist 안내 화면 | P1 | 1 | Pending |
| FR-24 | `tene passwd`: Master Password 변경 → 볼트 재암호화 → 새 Recovery Key 발급 | P0 | 1 | Pending |
| FR-25 | `tene recover`: Recovery Key 입력 → Master Password 재설정 → 새 Recovery Key 발급 | P0 | 1 | Pending |
| FR-21 | Cloud 동기화: GitHub OAuth + ECS API + S3 백업 | P0 | 2 | Pending |
| FR-22 | 웹 대시보드: 시크릿 목록 조회 + 접근 로그 | P1 | 2 | Pending |
| FR-23 | Stripe 결제: $1/월 구독 | P0 | 2 | Pending |

### 14.2 비기능 요구사항

| 범주 | 기준 | Phase | 측정 방법 |
|------|------|:-----:|----------|
| **성능** | CLI 명령 응답 < 200ms (로컬, P95) | 1 | time 명령 벤치마크 |
| **성능** | CLI 시작 시간 < 300ms (cold start) | 1 | time 명령 |
| **오프라인** | 모든 기능 100% 오프라인 동작 | 1 | 네트워크 차단 테스트 |
| **보안** | 암호화: XChaCha20-Poly1305 (256-bit) | 1 | 코드 리뷰 |
| **보안** | KDF: Argon2id (64MB, 3 iterations) | 1 | 코드 리뷰 |
| **확장성** | 사용자당 시크릿 1,000개 지원 | 1 | SQLite 성능 테스트 |
| **호환성** | macOS 12+, Ubuntu 20.04+, Windows 10+ (WSL) | 1 | CI 테스트 |
| **크기** | CLI npm 패키지 < 10MB | 1 | npm pack |
| **코드** | 테스트 커버리지 > 80% (crypto > 95%) | 1 | Vitest coverage |
| **AI Agent** | CLAUDE.md 생성 시간 < 100ms | 1 | 벤치마크 |
| **보안** | Zero-Knowledge: Cloud에서 시크릿 복호화 불가 | 2 | 보안 감사 |
| **성능** | API 응답 < 500ms (P95) | 2 | CloudWatch 메트릭 |
| **동기화** | Cloud 동기화 < 3초 | 2 | E2E 테스트 |

---

## 15. 구현 우선순위

### 15.1 Phase 1: MVP 로컬 CLI (2주)

```
Week 1: 코어 인프라 + 암호화 + AI Agent 통합
──────────────────────────────────────────────
Day 1-2:
  1. 모노레포 초기화 (pnpm workspace)
  2. 공유 패키지 설정 (@tene/types)
  3. @tene/crypto 패키지 구현:
     - Argon2id KDF
     - XChaCha20-Poly1305 encrypt/decrypt
     - Key derivation (Master Key → Enc Key)
     - Recovery Key 생성/검증
     - 암호화 볼트 내보내기/가져오기
  4. @tene/crypto 테스트 (95%+ 커버리지)

Day 3-4:
  5. CLI 기본 구조 (Commander.js)
  6. SQLite 볼트 매니저 (better-sqlite3)
  7. OS Keychain 연동 (keytar) + 파일 폴백
  8. tene init (볼트 생성 + Master PW + Recovery Key)
  9. AI Agent 컨텍스트 생성기:
     - CLAUDE.md 자동 생성 (기본값)
     - .cursorrules 추가 (--cursor)
     - .windsurfrules 추가 (--windsurf)
     - 기존 파일 병합 로직

Day 5:
  10. tene set / get (암호화/복호화 + SQLite)
  11. tene run (환경변수 주입)

Week 2: CLI 확장 + Fake Door + 배포
──────────────────────────────────────────────
Day 6-7:
  12. tene list / delete
  13. tene import / export (.env 형식)
  14. tene export --encrypted / import --encrypted (암호화 백업)
  15. tene env (환경 전환)
  16. --json / --stdin / --quiet 플래그

Day 8-9:
  17. tene sync (Fake Door → waitlist 안내)
  18. tene whoami
  19. 로컬 감사 로그
  20. 에러 처리 + 사용자 경험 개선
  21. 통합 테스트

Day 10:
  22. npm 패키지 빌드 + 배포
  23. CI 파이프라인 (GitHub Actions)
  24. README + 퀵스타트 문서
```

### 15.2 Critical Path

```
Phase 1:
  @tene/crypto → SQLite 볼트 → tene init (+ AI Agent 컨텍스트) → tene set/get → tene run
       |
       └──> 이 경로의 지연이 Phase 1 전체에 영향

공통:
  @tene/crypto 품질 = 전체 제품 보안 신뢰도
  AI Agent 컨텍스트 생성 = 핵심 차별점 (경쟁사 대비)
```

---

## 16. 리스크 및 완화

| 리스크 | 영향 | 가능성 | 완화 방안 |
|--------|:----:|:------:|----------|
| **암호화 구현 결함**: 암호화 로직 버그로 시크릿 노출 | 치명적 | 중간 | libsodium 사용(검증된 라이브러리), @tene/crypto 95%+ 테스트 커버리지, 오픈소스 커뮤니티 보안 리뷰 |
| **Master Password 분실**: 로컬 전용이라 서버에서 복구 불가 | 치명적 | 높음 | Recovery Key 생성 + 안전 보관 가이드 + `tene export --encrypted` 백업 유도 |
| **".env로 충분"**: 사용자가 전환 동기 부족 | 높음 | 높음 | `tene import .env` 한 줄 마이그레이션 + AI Agent 자동 인식 차별점 + "Git 커밋해도 안전" |
| **AI Agent 컨텍스트 파일 충돌**: 기존 CLAUDE.md/.cursorrules와 충돌 | 중간 | 중간 | 섹션 단위 추가/업데이트, 기존 내용 보존, 중복 감지 |
| **keytar 호환성**: OS Keychain이 특정 환경에서 동작하지 않음 | 중간 | 중간 | 폴백: 암호화된 파일 저장 (~/.tene/keyfile, 퍼미션 0600). Master Password 재입력 |
| **Shell History 노출**: `tene set KEY VALUE`가 shell history에 남음 | 높음 | 높음 | `--stdin` 플래그 지원. 문서에서 `echo VALUE \| tene set KEY --stdin` 권장 |
| **npm 패키지 크기**: libsodium + better-sqlite3로 10MB 초과 가능 | 낮음 | 중간 | libsodium-wrappers 경량 빌드 사용, better-sqlite3 prebuild 최적화 |
| **Dotenvx 직접 경쟁**: 유사 포지셔닝 | 높음 | 중간 | AI Agent 자동 인식(CLAUDE.md) + SQLite 볼트(구조화) + 환경 전환으로 차별화 |
| **Cloud 수요 부재**: Fake Door Test 결과 수요 없음 | 중간 | 중간 | 로컬 기능 강화에 집중. Cloud 비용 $0 유지 |

---

## 17. 성공 기준

### 17.1 Phase 1 MVP Definition of Done

- [ ] CLI 코어 명령어 전체 작동 (init, set, get, run, list, delete, import, export, env, passwd, recover, whoami, sync)
- [ ] Master Password + Argon2id KDF + XChaCha20-Poly1305 암호화 작동
- [ ] Recovery Key 생성 및 복구 작동
- [ ] OS Keychain 연동 + 파일 폴백
- [ ] 오프라인 100% 동작 확인
- [ ] **AI Agent 자동 인식**: `tene init` 시 CLAUDE.md 자동 생성
- [ ] **AI Agent 확장**: `--cursor`, `--windsurf` 플래그 동작
- [ ] **Fake Door**: `tene sync` 실행 시 waitlist 안내 표시
- [ ] **암호화 백업**: `tene export --encrypted` / `tene import --encrypted` 동작
- [ ] macOS + Linux 테스트 통과
- [ ] npm 패키지 배포 (`npm install -g @tene/cli`)
- [ ] 설치 → 첫 시크릿 저장 3분 이내 달성
- [ ] --json 플래그 동작 (AI 에이전트 파싱)
- [ ] --stdin 플래그 동작 (shell history 방지)
- [ ] 전체 테스트 커버리지 > 80%
- [ ] @tene/crypto 테스트 커버리지 > 95%
- [ ] npm 패키지 < 10MB
- [ ] CLI 응답 시간 < 200ms (로컬, P95)
- [ ] GitHub 오픈소스 공개 (MIT License)
- [ ] 퀵스타트 README 작성

### 17.2 품질 기준

- [ ] 린트 에러 0개 (Biome)
- [ ] 빌드 성공 (모든 패키지)
- [ ] 시크릿이 에러 메시지에 노출되지 않음

---

## 18. 컨벤션

### 18.1 코딩 컨벤션

| 범주 | 규칙 |
|------|------|
| **네이밍** | camelCase (변수/함수), PascalCase (타입/클래스), SCREAMING_SNAKE_CASE (상수) |
| **파일 이름** | kebab-case.ts (모듈), PascalCase.tsx (React 컴포넌트) |
| **import 순서** | 1) node: 내장 2) 외부 패키지 3) @tene/* 내부 패키지 4) 상대 경로 |
| **에러 처리** | 커스텀 Error 클래스 (TeneError, CryptoError, VaultError) |
| **비동기** | async/await 사용, Promise 체이닝 금지 |
| **주석** | JSDoc (공개 API), 인라인 주석은 WHY에만 사용 |

### 18.2 Git 컨벤션

```
feat(cli): add tene init with AI agent context generation
feat(crypto): implement XChaCha20-Poly1305 encryption
feat(cli): add tene sync fake door test
fix(cli): fix keychain fallback on Linux
chore(deps): update libsodium to latest
test(crypto): add KDF edge case tests
docs: add quickstart guide
```

---

## 19. Phase 2 Cloud 상세 (수요 검증 후)

> 이 섹션은 Phase 2에서 구현할 Cloud 기능의 참고 자료이다. MVP에는 포함되지 않는다.

### 19.1 Cloud 아키텍처 (ECS Fargate + RDS PostgreSQL)

```
╔═══════════════════════════════════════════════════════════════════════════╗
║                     $1/월 CLOUD TIER (Phase 2, 수요 검증 후)               ║
║                                                                         ║
║   ┌──────────────┐          ┌──────────────────────────┐                ║
║   │  CLI         │          │  Web Dashboard           │                ║
║   │  (sync cmd)  │          │  (Next.js static export) │                ║
║   │              │          │  S3 + CloudFront 배포     │                ║
║   └──────┬───────┘          └──────────┬───────────────┘                ║
║          │                             │                                 ║
║          │  HTTPS (E2E Encrypted)      │  HTTPS                         ║
║          │                             │                                 ║
║          ▼                             ▼                                 ║
║   ┌─────────────────────────────────────────────────────────────┐       ║
║   │            NLB + ECS Fargate (Hono API)                      │       ║
║   │                                                              │       ║
║   │   ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐  │       ║
║   │   │   Auth   │  │   Sync   │  │  Audit   │  │ Billing  │  │       ║
║   │   │ GitHub   │  │  Upload  │  │  Logs    │  │ Stripe   │  │       ║
║   │   │ OAuth    │  │  Download│  │          │  │          │  │       ║
║   │   └────┬─────┘  └────┬─────┘  └────┬─────┘  └────┬─────┘  │       ║
║   └────────┼──────────────┼─────────────┼──────────────┼────────┘       ║
║            │              │             │              │                 ║
║            ▼              ▼             ▼              ▼                 ║
║   ┌────────────────┐  ┌────────────────────────────────────────┐       ║
║   │ RDS            │  │ S3 Bucket                               │       ║
║   │ PostgreSQL     │  │ (Encrypted Vault Blobs)                 │       ║
║   │                │  │                                         │       ║
║   │ users          │  │ {userId}/{deviceId}/vault.enc           │       ║
║   │ devices        │  │ {userId}/sync-manifest.json             │       ║
║   │ subscriptions  │  │                                         │       ║
║   │ audit_logs     │  │ 서버는 암호화된 blob만 저장              │       ║
║   └────────────────┘  │ 복호화 키는 클라이언트에만 존재           │       ║
║                       └────────────────────────────────────────┘       ║
║                                                                         ║
║   ┌───────────────────────────────────────┐                             ║
║   │ CloudFront + Route 53                  │                             ║
║   │ - tene.dev (웹 대시보드)                │                             ║
║   │ - api.tene.dev (API)                   │                             ║
║   └───────────────────────────────────────┘                             ║
╚═══════════════════════════════════════════════════════════════════════════╝
```

### 19.2 Cloud API 엔드포인트 (Phase 2 참고)

| Method | Path | 설명 | 인증 |
|--------|------|------|:----:|
| **인증** | | | |
| POST | `/api/auth/github` | GitHub OAuth 코드 교환 | No |
| POST | `/api/auth/refresh` | JWT 토큰 갱신 | Refresh |
| DELETE | `/api/auth/session` | 로그아웃 | Yes |
| GET | `/api/auth/me` | 현재 사용자 정보 | Yes |
| POST | `/api/auth/register-key` | Auth Hash 등록 | Yes |
| **동기화** | | | |
| POST | `/api/sync/upload` | 암호화 볼트 blob 업로드 | Yes |
| GET | `/api/sync/download` | 암호화 볼트 blob 다운로드 | Yes |
| GET | `/api/sync/manifest` | 동기화 매니페스트 조회 | Yes |
| POST | `/api/sync/register-device` | 디바이스 등록 | Yes |
| **감사** | | | |
| GET | `/api/audit/logs` | 감사 로그 조회 | Yes |
| **결제** | | | |
| POST | `/api/billing/subscribe` | $1/월 구독 시작 | Yes |
| POST | `/api/billing/cancel` | 구독 취소 | Yes |
| GET | `/api/billing/status` | 구독 상태 조회 | Yes |
| POST | `/api/billing/webhook` | Stripe 웹훅 수신 | Stripe Sig |

### 19.3 Cloud DB 스키마 (Phase 2 참고 — RDS PostgreSQL)

```sql
-- users: Cloud 사용자 (GitHub OAuth)
CREATE TABLE users (
  id              TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
  github_id       BIGINT UNIQUE NOT NULL,
  email           TEXT UNIQUE NOT NULL,
  name            TEXT,
  avatar_url      TEXT,
  auth_hash       TEXT NOT NULL,
  plan            TEXT NOT NULL DEFAULT 'free',
  stripe_customer_id TEXT,
  stripe_subscription_id TEXT,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- devices: 사용자 디바이스 등록
CREATE TABLE devices (
  id              TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
  user_id         TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  name            TEXT NOT NULL,
  device_hash     TEXT NOT NULL,
  last_sync_at    TIMESTAMPTZ,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(user_id, device_hash)
);

-- sync_manifests: 동기화 매니페스트
CREATE TABLE sync_manifests (
  id              TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
  user_id         TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  device_id       TEXT NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
  vault_hash      TEXT NOT NULL,
  s3_key          TEXT NOT NULL,
  version         INTEGER NOT NULL DEFAULT 1,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- audit_logs: Cloud 감사 로그
CREATE TABLE audit_logs (
  id              TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
  user_id         TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  action          TEXT NOT NULL,
  device_id       TEXT,
  ip_address      TEXT,
  user_agent      TEXT,
  metadata        JSONB,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 인덱스
CREATE INDEX idx_devices_user ON devices(user_id);
CREATE INDEX idx_manifests_user ON sync_manifests(user_id, created_at DESC);
CREATE INDEX idx_audit_user ON audit_logs(user_id, created_at DESC);
```

### 19.4 동기화 프로토콜 (Phase 2 참고)

#### 동기화 모델: Last-Write-Wins with Conflict Detection

```
디바이스 A                          Cloud (S3)                         디바이스 B
    |                                  |                                    |
    |-- tene sync push --------------->|                                    |
    |   1. 로컬 볼트 암호화             |                                    |
    |   2. blob + vault_hash 업로드    |                                    |
    |   3. sync_manifest 갱신          |                                    |
    |                                  |                                    |
    |                                  |<------------- tene sync pull ------|
    |                                  |   1. 매니페스트 확인                 |
    |                                  |   2. blob 다운로드                  |
    |                                  |   3. 로컬 복호화 + 볼트 갱신         |
    |                                  |                                    |
```

#### Push (업로드)

```
1. 로컬 볼트의 모든 시크릿 + 환경 데이터를 JSON 직렬화
2. XChaCha20-Poly1305로 전체 JSON 암호화 (Master Key 사용)
3. 암호화된 blob을 S3에 업로드:
   PUT s3://tene-vaults/{userId}/{deviceId}/vault.enc
4. 볼트 해시(SHA-256)와 타임스탬프를 sync_manifests에 기록
5. 감사 로그: sync.upload
```

#### Pull (다운로드)

```
1. sync_manifests에서 최신 버전 확인
2. 로컬 vault_hash와 비교
   - 동일하면: 스킵 (이미 최신)
   - 다르면: 다운로드 진행
3. S3에서 암호화된 blob 다운로드
4. 로컬에서 복호화 (Master Key 사용)
5. 로컬 볼트에 머지 (Last-Write-Wins)
6. 충돌 발생 시: 양쪽 버전 보존 + 사용자에게 알림
```

#### 충돌 해결 정책

| 시나리오 | 정책 |
|---------|------|
| 같은 키에 다른 값 | Last-Write-Wins (timestamp 기준) |
| 한쪽에서 삭제, 다른 쪽에서 수정 | 수정 우선 (삭제 취소) |
| 새 키 추가 (양쪽) | 양쪽 모두 유지 |
| 충돌 감지 시 | CLI에 경고 메시지 + `tene sync conflicts` 명령어 |

### 19.5 웹 대시보드 (Phase 2 참고)

#### 배포 방식: Next.js Static Export → S3 + CloudFront

| 페이지 | 경로 | 설명 |
|--------|------|------|
| 로그인 | `/login` | GitHub OAuth 로그인 |
| OAuth Callback | `/auth/callback` | GitHub callback 처리 |
| 대시보드 홈 | `/` | 프로젝트 요약 + 시크릿 수 |
| 시크릿 목록 | `/secrets` | 시크릿 이름 + 환경 + 마지막 접근 |
| 접근 로그 | `/logs` | 동기화/접근 기록 |
| 구독 설정 | `/settings` | 구독 상태, 결제 정보 |

**대시보드에서 할 수 없는 것 (보안):**
- 시크릿 값 조회/복호화 (Master Key가 브라우저에 없으므로)
- 시크릿 수정/삭제 (CLI에서만 가능)
- 볼트 다운로드

### 19.6 보안 요약: 누가 뭘 알고 있는가 (Phase 2 Cloud 포함)

| 구성요소 | 시크릿 평문 | Master Key | 암호화된 blob | 메타데이터 |
|---------|:----------:|:----------:|:------------:|:----------:|
| **CLI** | O | O (Keychain) | O | O |
| **SQLite** | X (암호화됨) | X | O | O |
| **ECS API** | X | X | 중계만 | O |
| **RDS DB** | X | X | X | O |
| **S3** | X | X | O (저장) | X |
| **CloudFront/대시보드** | X | X | X | O (표시) |

**결론: 시크릿 평문은 CLI(사용자 디바이스)에서만 존재한다.**

---

## 20. Phase 확장 고려

MVP 설계 시 향후 확장을 위해 미리 고려한 사항:

| 향후 기능 | MVP에서의 준비 |
|----------|---------------|
| Cloud 동기화 (Phase 2) | `tene sync` 명령어 이미 존재 (Fake Door). 구현만 추가하면 됨 |
| 팀 볼트 / RBAC (Phase 3) | SQLite 스키마에 확장 가능한 구조. Phase 3에서 team_id + role 추가 |
| 에이전트 스코핑 | audit_log에 source 필드 존재. Phase 3에서 agent_id + scope 정책 추가 |
| MCP 서버 | CLI가 stdout 기반 인터페이스. Phase 3에서 MCP 프로토콜 레이어 추가 |
| 자동 로테이션 | secrets 테이블에 version 필드 존재. Phase 3+에서 rotation 로직 추가 |
| 새로운 AI Agent | agent-context.ts의 확장 가능 구조. 새 Agent 추가 시 템플릿만 추가 |
| Homebrew/curl 설치 | CLI는 Node.js 독립. Phase 2+에서 pkg/바이너리 빌드 + brew formula |

---

## 21. 다음 단계

1. [ ] Design 문서 작성 (`tene-mvp.design.md`)
2. [ ] @tene/crypto 패키지 상세 설계
3. [ ] AI Agent 컨텍스트 템플릿 상세 설계
4. [ ] Fake Door Test 분석 대시보드 설계
5. [ ] CI/CD 파이프라인 설계

---

## 22. `tene passwd` 명령어 상세

```
$ tene passwd
  Current Master Password: ********
  New Master Password: ********
  Confirm: ********
  
  Re-encrypting vault...
  Master Password updated.
  
  New Recovery Key (write this down!):
  > orbit piano queen river stone tiger umbrella violet whale xray yellow zebra
  
  Previous Recovery Key is now invalid.
```

- 인자: 없음 (대화형)
- 동작: Master Password 변경 → 볼트의 모든 시크릿을 새 키로 재암호화 → 새 Recovery Key 자동 발급 (기존 것 무효화)
- 종료 코드: 0(성공), 1(실패)
- 에러: "Invalid Master Password" / "Vault not found"

---

## 23. `tene recover` 명령어 상세

```
$ tene recover
  Enter your 12-word Recovery Key:
  > apple banana cherry dolphin eagle frost grape harbor island jungle kite lemon
  
  Recovery Key verified.
  Set new Master Password: ********
  Confirm: ********
  
  Master Password updated.
  All secrets re-encrypted with new key.
  
  New Recovery Key (write this down!):
  > orbit piano queen river stone tiger umbrella violet whale xray yellow zebra
  
  Master Key saved to OS Keychain.
```

- 인자: 없음 (대화형)
- 종료 코드: 0(성공), 1(실패)
- 에러: "Invalid Recovery Key" / "Vault not found"

---

## 24. `tene export --encrypted` 파일 포맷 정의

```
파일 형식: .tene.enc
구조:
  [4 bytes]  Magic: "TENE"
  [2 bytes]  Version: 0x0001
  [16 bytes] Salt
  [4 bytes]  KDF params (memory_mb:u16 + iterations:u8 + parallelism:u8)
  [24 bytes] Nonce
  [N bytes]  XChaCha20-Poly1305 encrypted payload (JSON 직렬화된 볼트)
  [16 bytes] Poly1305 auth tag (암호문에 포함)
```

---

## 25. `--json` 출력 스키마 정의

```typescript
// tene list --json
[{"name":"STRIPE_KEY","environment":"default","updatedAt":"2026-04-06T12:00:00Z","version":1}]

// tene env list --json
[{"name":"default","isDefault":true,"secretCount":5},{"name":"prod","isDefault":false,"secretCount":3}]

// tene whoami --json
{"mode":"local","project":"my-project","environment":"default","secretCount":5,"cloud":null}
// cloud 연결 시: {"mode":"cloud","project":"my-project","environment":"default","secretCount":5,"cloud":{"email":"steve@example.com","plan":"cloud"}}

// 에러 --json
{"error":{"code":"SECRET_NOT_FOUND","message":"Secret \"KEY\" not found in \"default\" environment"},"exitCode":1}
```

---

## 26. `--stdin` 처리 규칙

- 첫 번째 줄만 사용 (trailing newline 제거)
- 멀티라인: 첫 줄만 value, 나머지 무시
- stdin과 positional VALUE 동시 제공: `--stdin` 우선, VALUE 무시 + 경고
- EOF 처리: stdin이 비면 에러 (exit code 1)

---

## 27. `tene env` 서브명령어 상세

```
tene env              → 현재 활성 환경 표시
tene env list         → 모든 환경 목록
tene env create NAME  → 새 환경 생성 (이미 존재 시 에러)
tene env delete NAME  → 환경 삭제 (default 삭제 불가, --force 필요 시에도 불가)
tene env NAME         → 활성 환경 전환
```

---

## 28. npm 패키지 설정

```json
// packages/cli/package.json
{
  "name": "@tene/cli",
  "version": "0.1.0",
  "bin": { "tene": "./dist/index.js" },
  "engines": { "node": ">=20.0.0" },
  "os": ["darwin", "linux", "win32"]
}
```

- native dependency 전략: better-sqlite3는 prebuild-install로 prebuilt binary 제공. keytar도 동일. 빌드 실패 시 npm 설치 에러와 대처법 문서화.

---

## 29. 에러 코드 체계

```
종료 코드:
  0 = 성공
  1 = 일반 에러 (시크릿 미존재, 볼트 미초기화 등)
  2 = 인증 에러 (Master Password 오류, Recovery Key 오류)

--json 에러 코드:
  VAULT_NOT_FOUND    = tene init 안 됨
  SECRET_NOT_FOUND   = 시크릿 없음
  SECRET_EXISTS      = 이미 존재 (--overwrite 없이)
  AUTH_FAILED        = Master Password/Recovery Key 틀림
  KEYCHAIN_ERROR     = OS Keychain 접근 실패
  ENV_NOT_FOUND      = 환경 없음
  INVALID_INPUT      = 잘못된 인자
  CLOUD_NOT_CONNECTED = Cloud 미연결 (sync 시)
```

---

## 30. keytar 폴백 상세

```
keytar 없는 환경 (CI, Docker, WSL 일부):
1. 환경변수 TENE_MASTER_PASSWORD 확인 → 있으면 사용
2. 없으면 stdin에서 Master Password 입력 요청
3. 세션 내 캐시: 첫 입력 후 프로세스 종료까지 메모리에 유지
4. 파일 폴백 없음 (보안상 Master Key를 파일에 저장하지 않음)
```

---

## 31. 테스트 전략

```
@tene/crypto 단위 테스트:
  - KDF: 동일 입력 → 동일 출력 (결정론적)
  - Encrypt/Decrypt: roundtrip 검증
  - 다른 키로 복호화 → 실패
  - AAD 변조 → 실패
  - Recovery Key: 생성 → 니모닉 변환 → 복구 roundtrip
  - 커버리지 목표: 95%+

CLI 통합 테스트:
  - init → set → get → list → delete 전체 플로우
  - import .env → list 검증
  - run -- echo $KEY 검증
  - env create → env switch → set → get (환경별 분리)
  - passwd → 기존 시크릿 접근 가능 여부
  - recover → 새 Master Password로 접근 가능 여부
  - --json 출력 파싱 가능 여부
  - 커버리지 목표: 80%+

테스트 프레임워크: Vitest
테스트 러너: pnpm test (Turborepo 병렬)
```

---

## 32. 프로젝트 구성 파일

```yaml
# pnpm-workspace.yaml
packages:
  - "packages/*"
  - "apps/*"
```

```json
// root tsconfig.json
{
  "compilerOptions": {
    "target": "ES2022",
    "module": "Node16",
    "moduleResolution": "Node16",
    "strict": true,
    "esModuleInterop": true,
    "skipLibCheck": true,
    "outDir": "dist",
    "declaration": true
  }
}
```

```json
// biome.json
{
  "formatter": { "indentStyle": "space", "indentWidth": 2 },
  "linter": { "rules": { "recommended": true } }
}
```

---

## Version History

| Version | Date | Changes | Author |
|---------|------|---------|--------|
| 1.0 | 2026-04-06 | Initial draft (서버 중심 아키텍처) | CTO Lead |
| 2.0 | 2026-04-06 | **전면 재설계**: Local-Free + Cloud-Paid 하이브리드. SQLite 로컬 볼트, AWS Lambda, $1 Cloud 모델 | CTO Lead |
| 3.0 | 2026-04-06 | **v3 보완**: Local-Only MVP (Cloud → Phase 2). AI Agent 자동 인식 (CLAUDE.md/.cursorrules). Fake Door Test (tene sync → waitlist). tene export --encrypted. Cloud 인프라 ECS+RDS로 변경. "해결하는 것/못 하는 것" 추가. Phase 재정리 (1=MVP, 2=Cloud, 3=팀) | CTO Lead |
| 3.1 | 2026-04-06 | **v3.1 보완**: Recovery Key 12단어 니모닉(BIP-39) 통일. npm 패키지 @tene/cli 통일. tene passwd/recover 명령어 추가. export --encrypted 파일 포맷 정의. --json 출력 스키마, --stdin 처리 규칙, tene env 상세, 에러 코드 체계, keytar 폴백, 테스트 전략, 프로젝트 구성 파일 추가. salt 저장 위치 vault_meta로 확정 | CTO Lead |
