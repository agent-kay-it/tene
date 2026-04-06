# MVP 범위 브레인스토밍 v3.1
## Cloud 제외, Fake Door Test, AI Agent 자동 인식

> v3.1 (2026-04-06) — 암호화 알고리즘 XChaCha20-Poly1305 통일
> 목적: Cloud 제외된 MVP 최소 범위, Fake Door Test, AI Agent 자동 인식 기능 정의

---

## 1. MVP 철학: 서버 비용 $0, AI 에이전트가 자동 인식

### 1.1 핵심 질문 (v3)

> "솔로 바이브코더가 서버 가입 없이, npm install 하나로,
> 3분 안에 '와, AI가 알아서 인식하네!'라고 느낄 수 있는 최소 기능은 무엇인가?"

### 1.2 MVP 원칙 (v3)

| 원칙 | v2 | v3 |
|------|----|----|
| **One Job** | 서버 없이 시크릿 저장+주입 | 서버 없이 시크릿 저장+주입 + **AI 자동 인식** |
| **3 Minute Magic** | 가입 없이 설치→사용 3분 | 가입 없이 설치→AI 인식→사용 3분 |
| **No Server** | 서버 없음 (MVP) | 서버 없음 (MVP), **Cloud = Phase 2** |
| **AI Native** | CLI Bash 호출 | **CLAUDE.md/.cursorrules 자동 생성** |
| **Fake Door** | Cloud 전환 넛지 | **`tene sync` Fake Door (waitlist)** |
| **Honest** | — | **못 하는 것 정직하게 명시** |

### 1.3 MVP = CLI + SQLite + @tene/crypto

```
npm install -g @tene/cli → tene init → tene set → tene run
```

- 서버 없음, 회원가입 없음, 비용 없음
- CLAUDE.md 자동 생성 (AI 자동 인식)
- `tene export --encrypted`로 수동 백업
- `tene sync` = Fake Door (waitlist 안내만)

---

## 2. MVP 기능 범위 (Cloud 완전 제외)

### Phase 1: Local MVP (서버 없음, 무료, $0)

**"서버 없이, 가입 없이, AI가 자동 인식하는 CLI"**

| # | 기능 | 명령어 | 우선순위 | 근거 |
|---|------|--------|:--------:|------|
| L1 | **프로젝트 초기화 + CLAUDE.md** | `tene init` | P0 | 볼트 생성 + **CLAUDE.md 자동 생성** |
| L2 | **Cursor 통합** | `tene init --cursor` | P0 | **.cursorrules에 가이드 추가** |
| L3 | **시크릿 저장** | `tene set KEY VALUE` | P0 | 핵심 가치의 시작점 |
| L4 | **시크릿 조회** | `tene get KEY` | P0 | AI 에이전트 Bash 호출 핵심 |
| L5 | **시크릿 주입** | `tene run -- CMD` | P0 | "한 줄이면 끝" 핵심 가치 |
| L6 | **시크릿 목록** | `tene list` | P0 | 시크릿 현황 파악 |
| L7 | **시크릿 삭제** | `tene delete KEY` | P0 | 기본 CRUD |
| L8 | **.env 가져오기** | `tene import .env` | P1 | 전환 비용 제로 |
| L9 | **.env 내보내기** | `tene export` | P1 | 기존 도구 호환성 |
| L10 | **암호화 백업** | `tene export --encrypted` | P1 | Cloud 없이 수동 백업 |
| L11 | **환경 전환** | `tene env [dev/prod]` | P1 | 개발/프로덕션 분리 |
| L12 | **Fake Door** | `tene sync` | P1 | Cloud 수요 확인 (waitlist) |
| L13 | **Master Password 암호화** | (내부) | P0 | 보안의 전제 조건 |
| L14 | **Recovery Key 생성** | (내부) | P0 | 패스워드 분실 대비 |
| L15 | **.gitignore 자동 추가** | (내부) | P0 | .tene/ 노출 방지 |

### Phase 2: Cloud (수요 검증 후)

**"`tene sync` Fake Door에서 수요가 확인되면 구축"**

| # | 기능 | 설명 | 전제 조건 |
|---|------|------|----------|
| C1 | Cloud 인증 | `tene login` | waitlist 100명+ |
| C2 | 암호화 클라우드 백업 | 볼트 blob → S3 | waitlist 확인 |
| C3 | 멀티 디바이스 동기화 | 자동 동기화 | waitlist 확인 |
| C4 | 웹 대시보드 | 시크릿 현황, 감사 로그 | Cloud 출시 후 |

**Cloud 인프라 (서버리스 사용 안 함):**
- ECS Fargate + NLB + RDS PostgreSQL + S3

### Phase 2+: 팀 기능 (가설)

| # | 기능 | 상태 |
|---|------|:----:|
| T1 | 팀 볼트 | **가설** — Fake Door 후 결정 |
| T2 | RBAC | **가설** |
| T3 | 에이전트 스코핑 | Phase 2+ |
| T4 | MCP 서버 | Phase 2 |
| T5 | 자동 로테이션 | **못 함** (Phase 2+ 가설) |
| T6 | API key 만료 확인 | **못 함** (Phase 2+ 가설) |

---

## 3. YAGNI 분석: MVP에서 제외하는 것

### 3.1 "지금 만들면 안 되는" 기능

**1. 클라우드 서버/API (Phase 2)**
- 이유: **수요 미확인 상태에서 서버 구축은 낭비**. Fake Door로 수요 먼저 확인
- 대신: `tene sync` Fake Door + `tene export --encrypted` 수동 백업

**2. 회원가입/로그인 (Phase 2)**
- 이유: **가입이 필요 없는 것이 핵심 차별점**
- 대신: Phase 2 Cloud 가입 시에만

**3. 웹 대시보드 (Phase 2)**
- 이유: CLI로 핵심 가치 전달 가능
- 대신: `tene list`, `tene export`로 CLI에서 관리

**4. MCP 서버 (Phase 2)**
- 이유: AI 에이전트는 **CLAUDE.md 자동 인식 + CLI Bash 호출**이면 충분
- MCP 설정에 시크릿 저장하지 않아도 됨 → 오히려 보안적 이점

**5. 자동 로테이션 (못 함)**
- 이유: 각 서비스 API 개별 통합 필요, 구현 비용 매우 높음
- 정직하게 "못 하는 것"으로 명시. Phase 2+ 가설

**6. API key 만료 확인 (못 함)**
- 이유: 서비스마다 만료 확인 API가 다르거나 없음
- 정직하게 "못 하는 것"으로 명시

**7. $1/월 Cloud ($1은 수요 확인 후)**
- 이유: v2에서 $1/월 Cloud를 Phase 1.1에 계획했지만, **수요 미확인 상태에서 서버 구축은 리스크**
- 대신: waitlist로 수요 확인 → 가격은 수요에 따라 결정

**8. 서버리스 인프라 (사용 안 함)**
- 이유: 트래픽 늘면 비용 역전. ECS Fargate가 예측 가능
- Phase 2 Cloud 구축 시: ECS Fargate + NLB + RDS PostgreSQL + S3

### 3.2 반드시 있어야 하는 "숨겨진" 기능

| # | 기능 | 이유 |
|---|------|------|
| H1 | **Master Password 세션 캐시** | 매 명령마다 패스워드 입력하면 사용 불가 |
| H2 | **Recovery Key 생성** | Master Password 분실 시 유일한 복구 수단 |
| H3 | **CLAUDE.md 자동 생성** | AI Agent 자동 인식의 핵심 메커니즘 |
| H4 | **에러 메시지 품질** | "tene init으로 먼저 초기화하세요" |
| H5 | **.gitignore 자동 추가** | .tene/ 노출 방지 |
| H6 | **시크릿 값 마스킹** | `tene list`에서 값 마스킹 |
| H7 | **빠른 응답** | < 200ms |
| H8 | **Master Password 변경** | `tene passwd` 명령어 |

---

## 4. AI Agent 자동 인식 기능 상세

### 4.1 `tene init` → CLAUDE.md 자동 생성

```
$ tene init
? Master Password: ********
? Confirm Password: ********

> 로컬 볼트 생성: .tene/vault.db (XChaCha20-Poly1305)
> .gitignore에 .tene/ 추가 완료
> CLAUDE.md 생성 완료 (Claude Code 자동 인식)
> Recovery Key: apple banana cherry ... (안전한 곳에 보관하세요!)
```

**생성되는 CLAUDE.md 내용:**
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

### 4.2 `tene init --cursor` → .cursorrules 자동 추가

```
$ tene init --cursor
? Master Password: ********

> .cursorrules에 tene 가이드 추가 완료
```

### 4.3 기존 CLAUDE.md가 있는 경우

- 기존 CLAUDE.md가 있으면 하단에 tene 섹션을 **추가** (덮어쓰지 않음)
- 이미 tene 섹션이 있으면 스킵

---

## 5. Fake Door Test: `tene sync`

### 5.1 구현

```
$ tene sync

  +----------------------------------------------------+
  | 클라우드 동기화 기능을 준비하고 있습니다!              |
  |                                                    |
  | 관심이 있으시면 waitlist에 등록하세요:               |
  | https://tene.dev/cloud                             |
  |                                                    |
  | 현재는 tene export --encrypted로                   |
  | 수동 암호화 백업을 권장합니다.                       |
  +----------------------------------------------------+
```

### 5.2 수요 검증 프로세스

```
[Step 1] CLI 출시 → 기본 지표 추적
   npm 다운로드 수, GitHub Stars
   성공 기준: npm 주간 500+, Stars 1,000+

[Step 2] tene sync Fake Door → Cloud 수요 확인
   tene sync 실행 → waitlist 페이지
   성공 기준: CLI 사용자의 10%+ waitlist 등록

[Step 3] waitlist 반응 → Cloud 구축 결정
   100명+ → Cloud 구축 시작
   < 50명 → Cloud 보류, CLI 기능 강화
```

---

## 6. Local MVP 핵심 플로우 상세

### 6.1 첫 사용 플로우 (3분 목표)

```
[0:00] $ npm install -g @tene/cli
       > 설치 완료 (15초)

[0:15] $ cd my-project
       $ tene init
       ? Master Password: ********
       ? Confirm Password: ********
       > 로컬 볼트 생성: .tene/vault.db
       > .gitignore에 .tene/ 추가
       > CLAUDE.md 생성 완료 (Claude Code 자동 인식)
       > Recovery Key: apple banana cherry ... (보관하세요!)
       (30초)

[0:45] $ tene import .env
       > 5개 시크릿 가져오기 완료 (암호화)
       (10초)

[0:55] $ tene run -- cursor .
       > 5 secrets injected as environment variables
       > Starting: cursor .
       (10초)

[1:05] 완료! 1분 5초 만에.
       Claude Code가 CLAUDE.md를 읽고 tene을 자동 인식.
       서버 없음. 가입 없음. 비용 $0.
```

### 6.2 AI 에이전트 사용 플로우 (CLAUDE.md 자동 인식)

```
[Claude Code 세션 - CLAUDE.md가 있는 프로젝트]

사용자: "Stripe 결제 기능 만들어줘"

Claude Code:
  (CLAUDE.md 읽음: "This project uses tene for secret management")
  
  "이 프로젝트는 tene으로 시크릿을 관리하고 있습니다.
   STRIPE_KEY가 필요합니다. 환경변수로 참조하겠습니다."
  
  코드 생성:
  import Stripe from 'stripe';
  const stripe = new Stripe(process.env.STRIPE_KEY!);
  
  "tene run -- npm run dev 로 실행하면 STRIPE_KEY가 자동 주입됩니다."
```

### 6.3 수동 백업 플로우

```
$ tene export --encrypted > ~/backup/my-project-$(date +%Y%m%d).enc
> 암호화된 백업 파일 생성 완료

# 복원
$ tene import --encrypted ~/backup/my-project-20260406.enc
? Master Password: ********
> 5개 시크릿 복원 완료
```

---

## 7. 기술적 MVP 범위

### 7.1 CLI 아키텍처

```
packages/cli/
+-- src/
|   +-- commands/          # CLI 명령어
|   |   +-- init.ts        # Master Password + 볼트 + CLAUDE.md 생성
|   |   +-- set.ts         # 시크릿 암호화 저장
|   |   +-- get.ts         # 시크릿 복호화 조회
|   |   +-- run.ts         # 환경변수 주입 실행
|   |   +-- list.ts        # 시크릿 목록 (마스킹)
|   |   +-- delete.ts      # 시크릿 삭제
|   |   +-- env.ts         # 환경 전환
|   |   +-- import.ts      # .env / --encrypted 가져오기
|   |   +-- export.ts      # .env / --encrypted 내보내기
|   |   +-- recover.ts     # Recovery Key로 복구
|   |   +-- passwd.ts      # Master Password 변경
|   |   +-- sync.ts        # Fake Door Test (waitlist 안내)
|   +-- crypto/            # @tene/crypto 암호화 모듈
|   |   +-- encryption.ts  # XChaCha20-Poly1305
|   |   +-- kdf.ts         # Argon2id
|   |   +-- keys.ts        # Master Key, Recovery Key 관리
|   +-- store/             # 로컬 저장소
|   |   +-- vault.ts       # SQLite 볼트 관리
|   |   +-- session.ts     # Master Password 세션 캐시
|   +-- agent/             # AI Agent 통합
|   |   +-- claude.ts      # CLAUDE.md 생성
|   |   +-- cursor.ts      # .cursorrules 생성
|   +-- utils/
|       +-- config.ts      # 설정 관리
|       +-- output.ts      # 출력 포맷팅
+-- package.json
+-- tsconfig.json
```

**v2 대비 핵심 변경**:
- `agent/` 모듈 추가 (CLAUDE.md, .cursorrules 생성)
- `sync.ts` = Fake Door (waitlist 안내만)
- `export.ts`에 `--encrypted` 옵션 추가
- `sync/` 모듈 삭제 (Cloud는 Phase 2)
- `login.ts` 삭제 (Phase 2)

### 7.2 서버 API (Phase 2에서만, 서버리스 X)

> MVP에는 서버 엔드포인트가 **0개**.
> Phase 2에서 ECS Fargate + NLB + RDS PostgreSQL + S3로 구축.

---

## 8. MVP 성공 기준 (v3)

### 8.1 출시 후 30일 목표

| 지표 | 목표 | 측정 |
|------|------|------|
| npm 설치 | 2,000+ | npm stats |
| GitHub Stars | 500+ | GitHub |
| `tene sync` Fake Door 실행 | 100+ | (간접 추정) |

### 8.2 출시 후 90일 목표

| 지표 | 목표 | 측정 |
|------|------|------|
| npm 총 설치 | 10,000+ | npm stats |
| GitHub Stars | 2,000+ | GitHub |
| waitlist 등록 | 100+ | tene.dev/cloud |
| Cloud 구축 결정 | Yes/No | waitlist 분석 |

---

*Tene v3 Brainstorming Document 03/05*
