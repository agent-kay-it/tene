# Tene Cloud PRD (Product Requirements Document)
## Zero-Knowledge Cloud Sync + Team Sharing + Dashboard

> **Version**: v1.0
> **Date**: 2026-04-07
> **Feature**: tene-cloud (Phase 2: Local-Free CLI -> Cloud-Paid SaaS)
> **Status**: PRD Complete -- Plan Phase Ready
> **PM Agent Team**: pm-discovery, pm-strategy, pm-research, pm-prd (4-agent synthesis)
> **Upstream**: `docs/01-plan/features/tene-cloud.plan.md` (v1.0)

---

## Executive Summary

| Perspective | Description |
|-------------|-------------|
| **Problem** | Tene CLI(v0.9.3) 사용자가 멀티 디바이스 동기화, 팀 시크릿 공유, 중앙 감사 로그 없이 로컬 전용으로만 운영. `tene sync` Fake Door에서 수요 신호 확인 |
| **Solution** | Zero-Knowledge Envelope 이중 암호화 기반 Cloud Sync(Solo $5/mo) + X25519 ECDH 팀 키 공유(Team $10/user/mo) + 대시보드(app.tene.sh) |
| **Functional UX Effect** | CLI `push/pull` 2개 명령어로 Vault 동기화, 대시보드에서 디바이스/감사/팀 관리, 시크릿 값은 서버에서 절대 노출 불가 |
| **Core Value** | 경쟁사 대비 40-76% 저렴한 가격($5-10)으로 완전한 Zero-Knowledge 보안 + AI 에이전트 자동 인식을 유지하며 클라우드 편의성 제공 |

---

## Context Anchor

| Dimension | Content |
|-----------|---------|
| **WHY** | 로컬 전용 CLI의 한계(동기화, 공유, 백업) 해소 + 유료 전환으로 지속 가능한 비즈니스 모델 구축 |
| **WHO** | 1차: 멀티 디바이스 솔로 개발자, 2차: 소규모 개발팀(2-10명), 3차: AI-first 개발 조직 |
| **RISK** | Cloud 수요 부족, X25519 구현 보안 오류, Stripe 한국 결제 불안정, Vault Sync 충돌/데이터 손실 |
| **SUCCESS** | MRR $500+, Solo 유료 전환 50+, Team 10+팀, Push/Pull 성공률 99.5%, API P99 < 500ms |
| **SCOPE** | Phase 2a(Solo Sync 4주) + Phase 2b(Team 3주) + Phase 2c(안정화 1주) = 총 8주 |

---

# Part 1: Discovery Analysis

> PM Agent: pm-discovery | Framework: Teresa Torres Opportunity Solution Tree + 5-Step Discovery Chain

## 1. Brainstorm: Cloud 전환의 Pain Point 발견

### 1.1 전환 배경

Tene CLI v0.9.3는 로컬 전용 시크릿 관리에서 성공적으로 자리잡았다. 그러나 사용자 성장에 따라 새로운 Pain Point가 드러나고 있다:

> **"로컬에서 잘 동작하지만, 회사 맥북과 개인 맥북에서 같은 시크릿을 쓸 수 없다"**
> **"팀원에게 시크릿을 전달하려면 결국 Slack DM으로 보내게 된다"**

### 1.2 Cloud 전환 Pain Points

| # | Pain Point | 심각도 | 빈도 | Cloud 해결 |
|---|-----------|--------|------|-----------|
| CP1 | 멀티 디바이스에서 동일 시크릿 접근 불가 | Critical | 높음 | `tene push/pull` Vault Sync |
| CP2 | 팀원 간 시크릿 공유가 안전하지 않음 (Slack DM, 이메일) | Critical | 높음 | X25519 ECDH 팀 키 공유 |
| CP3 | 로컬 볼트 손실 시 복구 불가 (디스크 장애) | High | 중간 | 클라우드 백업 (S3) |
| CP4 | 누가 언제 어떤 시크릿에 접근했는지 추적 불가 | High | 중간 | 중앙화된 감사 로그 |
| CP5 | 환경별(dev/staging/prod) 접근 제어 없음 | High | 중간 | RBAC + 환경별 권한 |
| CP6 | 기존 도구(Doppler $21, Infisical $18)가 비쌈 | Medium | 높음 | Solo $5, Team $10 |
| CP7 | 기존 Cloud 도구가 Zero-Knowledge가 아님 (Doppler) | High | 높음 | Sync Envelope 이중 암호화 |

### 1.3 핵심 인사이트: "Zero-Knowledge Cloud"는 존재하지 않는다

현재 시장에서 **완전한 Zero-Knowledge + Cloud Sync + 팀 공유**를 모두 제공하는 시크릿 관리 도구는 사실상 없다:

| 도구 | Zero-Knowledge | Cloud Sync | 팀 공유 | 가격 |
|------|:-----------:|:----------:|:------:|------|
| Doppler | **X** (서버측 암호화) | O | O | $21/user |
| Infisical | **부분적** (Workspace Key) | O | O | $18/user |
| 1Password | O (2SKD) | O | O | $7.99/user |
| Bitwarden SM | O (RSA-OAEP) | O | O | $6/user |
| **Tene Cloud** | **O** (Envelope + ECDH) | **O** | **O** | **$5-10** |

Tene Cloud는 **가장 저렴하면서 완전한 Zero-Knowledge를 보장하는 유일한 개발자 시크릿 관리 도구**가 될 수 있다.

---

## 2. Assumptions: 핵심 가정 식별

| # | 가정 (Assumption) | 유형 | 검증 상태 |
|---|------------------|------|----------|
| CA1 | CLI 사용자의 10%+가 Cloud Sync에 유료 전환할 의향이 있다 | Viability | Fake Door 측정 중 |
| CA2 | Zero-Knowledge가 개발자에게 유료 결정의 핵심 요인이다 | Desirability | 미검증 |
| CA3 | $5/mo(Solo), $10/user/mo(Team)가 적정 가격대다 | Viability | 경쟁사 대비 검증 |
| CA4 | X25519 ECDH 기반 팀 키 공유가 안전하게 구현 가능하다 | Feasibility | 기술적 검증 필요 |
| CA5 | 대시보드가 CLI 사용자 유지/전환에 기여한다 | Desirability | 미검증 |
| CA6 | AWS 인프라 비용 ~$58/월로 초기 운영이 가능하다 | Feasibility | 아키텍처 산정 완료 |
| CA7 | OAuth 전용 인증(패스워드 없음)이 개발자에게 수용된다 | Usability | 업계 표준 |
| CA8 | Sync 충돌 해결 UX가 사용자에게 직관적이다 | Usability | 미검증 |

---

## 3. Prioritize: 가정 우선순위 (Impact x Risk Matrix)

```
         높은 리스크 (불확실)
              |
    CA8 *     |     * CA1 (유료 전환율)
    CA5 *     |     * CA2 (ZK 가치)
              |     * CA4 (X25519 안전성)
    ----------+----------
              |     * CA3 (가격)
    CA7 *     |     * CA6 (비용)
              |
         낮은 리스크 (확실)
              
   낮은 임팩트          높은 임팩트
```

### 최우선 검증 대상 (High Impact + High Risk)

| 순위 | 가정 | 검증 방법 |
|------|------|----------|
| 1 | **CA1: 유료 전환율 10%** | Fake Door waitlist 데이터 + 초기 Stripe Checkout 전환율 |
| 2 | **CA2: ZK가 유료 결정 요인** | 랜딩페이지 A/B 테스트 ("ZK" 강조 vs 일반) |
| 3 | **CA4: X25519 구현 안전성** | 보안 감사 + 오픈소스 crypto 라이브러리 검증 |

---

## 4. Experiments: 검증 실험 설계

### Experiment 1: 유료 전환율 검증 (CA1)

| 항목 | 내용 |
|------|------|
| **방법** | `tene sync` Fake Door → waitlist 등록 → Early Access 결제 |
| **대상** | 기존 CLI 사용자 |
| **가설** | CLI 사용자의 10%+가 waitlist 등록, waitlist의 20%+가 Early Access 결제 |
| **메트릭** | waitlist 등록률, Early Access 결제 전환률, MRR |
| **기간** | 4주 (Phase 2a 중) |
| **성공 기준** | waitlist 100+ 등록, 결제 전환 20+ |

### Experiment 2: Zero-Knowledge 가치 검증 (CA2)

| 항목 | 내용 |
|------|------|
| **방법** | 랜딩페이지 A/B 테스트 |
| **대상** | tene.sh 방문자 |
| **가설A** | "클라우드 Sync + 팀 공유. $5/월부터." |
| **가설B** | "Zero-Knowledge 암호화. 서버도 당신의 시크릿을 볼 수 없습니다. $5/월." |
| **메트릭** | CTA 클릭률, waitlist 전환률 비교 |
| **기간** | 2주 |
| **성공 기준** | 가설B가 가설A 대비 15%+ 높은 전환율 |

### Experiment 3: Sync 충돌 UX 검증 (CA8)

| 항목 | 내용 |
|------|------|
| **방법** | 프로토타입 사용자 테스트 (5명) |
| **대상** | Tene CLI 파워 사용자 |
| **가설** | 충돌 시 "Remote 선택 / Local 유지 / 수동 병합" 3옵션 UX가 직관적 |
| **메트릭** | 충돌 해결 성공률, 완료 시간, 만족도 |
| **기간** | 1주 |
| **성공 기준** | 5명 중 4명이 30초 이내 충돌 해결 |

---

## 5. Opportunity Solution Tree (OST)

```
                    +-------------------------------------------+
                    |          Desired Outcome                  |
                    |  "Tene CLI 사용자가 Zero-Knowledge를       |
                    |   유지하며 Cloud Sync + 팀 공유로           |
                    |   유료 전환하여 지속 가능한 비즈니스 구축"   |
                    +-------------------+-----------------------+
                                        |
            +---------------------------+---------------------------+
            |                           |                           |
   +--------v--------+       +---------v--------+       +----------v--------+
   |  Opportunity 1   |       |  Opportunity 2   |       |  Opportunity 3    |
   |  Solo Sync       |       |  Team Sharing    |       |  Dashboard +      |
   |  멀티 디바이스    |       |  안전한 팀       |       |  운영 가시성      |
   |  동기화           |       |  시크릿 공유     |       |                   |
   +--------+---------+       +---------+--------+       +----------+--------+
            |                           |                           |
   +--------+--------+       +---------+--------+       +----------+--------+
   |                  |       |                  |       |                   |
+--v--+          +---v--+ +--v--+          +---v--+ +---v--+          +---v--+
| S1  |          | S2   | | S3  |          | S4   | | S5   |          | S6   |
|Sync |          |Cloud | |X25519|         |RBAC  | |Vault |          |Audit |
|Enve-|          |백업  | |ECDH |         |환경별| |목록  |          |로그  |
|lope |          |+S3   | |팀 키 |         |접근  | |디바  |          |구독  |
|push/|          |버전  | |공유  |         |제어  | |이스  |          |관리  |
|pull |          |관리  | |     |         |     | |관리  |          |     |
+--+--+          +--+--+ +--+--+          +--+--+ +--+--+          +--+--+
   |                |       |                |       |                 |
+--v--------+  +---v----+ +--v--------+ +---v----+ +--v--------+ +---v----+
| E1: push  | | E2:    | | E3: 팀원  | | E4:    | | E5: 대시  | | E6:    |
| /pull 전환 | | 백업   | | 초대 흐름 | | prod   | | 보드 사용 | | 감사   |
| 율 10%+   | | 복원   | | 5명 테스트| | 읽기전용| | 빈도 추적 | | 로그   |
| 검증      | | 테스트 | |           | | 검증   | |           | | 가치   |
+------------+ +--------+ +-----------+ +--------+ +----------+ +--------+

[Phase 2a: Solo]           [Phase 2b: Team]          [Phase 2a+2b]
```

### OST 상세 설명

#### Opportunity 1: Solo Sync -- 멀티 디바이스 동기화 ($5/mo)

> "회사와 집에서 같은 시크릿을, Zero-Knowledge로"

- **S1: Sync Envelope 이중 암호화 push/pull**
  - L1: 시크릿 값 XChaCha20-Poly1305 (개별 레코드, 기존)
  - L2: 메타데이터 + DB 구조 XChaCha20-Poly1305 (Sync Envelope, 신규)
  - L3: TLS 1.2+ (네트워크 전송)
  - L4: S3 SSE-S3 AES-256 (디스크 저장)
  - 서버는 평문 시크릿/마스터키에 접근 불가

- **S2: 클라우드 백업 + S3 버전 관리**
  - S3 Object Versioning으로 이전 버전 복원
  - `tene export --encrypted`의 자동화 버전
  - 디스크 장애 시에도 클라우드에서 복구

#### Opportunity 2: Team Sharing -- 안전한 팀 시크릿 공유 ($10/user/mo)

> "Slack DM으로 API 키를 보내지 마세요. X25519로 암호화하세요."

- **S3: X25519 ECDH 팀 키 공유 프로토콜**
  - Project Key(PK)를 각 멤버의 X25519 공개키로 래핑
  - ECDH shared secret + HKDF + XChaCha20-Poly1305 래핑
  - 서버는 PK 평문을 절대 볼 수 없음
  - 멤버 제거 시 키 회전 + 재암호화

- **S4: RBAC + 환경별 접근 제어**
  - 역할: admin, member
  - 환경별 권한: `{"dev": ["read","write"], "prod": ["read"]}`
  - prod 시크릿은 admin만 수정 가능

#### Opportunity 3: Dashboard + 운영 가시성

> "시크릿 값은 안 보이지만, 누가 언제 접근했는지는 보입니다"

- **S5: Vault 목록, 디바이스 관리**
  - 대시보드에서 Vault 이름/환경/마지막 동기화 확인
  - 디바이스 목록 관리 (등록/해제)
  - 시크릿 값 표시/편집 불가 (Zero-Knowledge 유지)

- **S6: 감사 로그 + 구독 관리**
  - vault.push, vault.pull, team.invite 등 모든 작업 기록
  - Stripe Billing Portal 연동 셀프서비스 구독 관리

---

# Part 2: Strategy Analysis

> PM Agent: pm-strategy | Frameworks: JTBD 6-Part VP, Lean Canvas, SWOT, Porter's 5 Forces

## 1. JTBD (Jobs-to-be-Done) 분석 -- Cloud 전환 관점

### 1.1 Core Job Statement

> **멀티 디바이스로 개발하거나 팀과 협업할 때,**
> **시크릿을 Zero-Knowledge로 안전하게 동기화/공유하고,**
> **서버가 평문 시크릿에 접근할 수 없도록 보장하면서,**
> **기존 CLI 워크플로우를 그대로 유지하고 싶다.**

### 1.2 Job Map (Cloud 전환 관점)

| 단계 | Job Step | 현재 Pain (로컬 전용) | Tene Cloud |
|------|----------|---------------------|-----------|
| 1. 인증 | 클라우드 서비스 시작 | 이메일+패스워드 가입 | OAuth (GitHub/Google) -- 패스워드 없음 |
| 2. 동기화 | 디바이스 간 시크릿 공유 | USB/이메일로 수동 복사 | `tene push/pull` -- 이중 암호화 |
| 3. 백업 | 시크릿 데이터 보호 | `tene export --encrypted` 수동 | 자동 클라우드 백업 (S3 버전 관리) |
| 4. 팀 공유 | 팀원에게 시크릿 전달 | Slack DM, 이메일 (비보안) | X25519 ECDH 키 교환 (ZK) |
| 5. 접근 제어 | 환경별 권한 관리 | 불가 | RBAC + 환경별 권한 |
| 6. 감사 | 접근 기록 추적 | 로컬 로그만 | 중앙화된 감사 로그 |
| 7. 관리 | 시각적 운영 관리 | CLI 전용 | 대시보드 (app.tene.sh) |

### 1.3 JTBD 6-Part Value Proposition

| Part | Tene Cloud 내용 |
|------|----------------|
| **1. Job Performer** | 멀티 디바이스 솔로 개발자, AI-first 소규모 팀 (기존 Tene CLI 사용자) |
| **2. Core Functional Job** | 시크릿을 Zero-Knowledge로 클라우드 동기화 + 팀 공유, CLI 워크플로우 유지 |
| **3. Related Jobs** | 디바이스 관리, 구독 관리, 팀 멤버 온보딩, 접근 권한 감사 |
| **4. Emotional Jobs** | "서버가 내 시크릿을 볼 수 없다는 확신", "합리적 가격으로 프로 도구 사용하는 만족감" |
| **5. Desired Outcomes** | (1) push/pull 2초 이내 (2) 팀원 초대 30초 이내 (3) 서버 ZK 보장 (4) $5-10 합리적 가격 |
| **6. Constraints** | (1) 기존 CLI UX 유지 필수 (2) 마스터키 서버 전송 금지 (3) 오프라인 로컬 기능 유지 |

---

## 2. Lean Canvas (Tene Cloud)

```
+--------------------+--------------------+--------------------+
|  1. PROBLEM         |  4. SOLUTION        |  3. UVP            |
|                     |                     |                     |
| - 멀티 디바이스     | - Sync Envelope    | "Zero-Knowledge    |
|   동기화 불가       |   이중 암호화       |  Cloud Sync.       |
|                     |   push/pull         |  서버도 당신의     |
| - 팀 시크릿 공유가  |                     |  시크릿을 볼 수    |
|   안전하지 않음     | - X25519 ECDH      |  없습니다.         |
|   (Slack DM)        |   팀 키 공유        |  $5/월부터."       |
|                     |                     |                     |
| - 기존 ZK 도구가    | - 대시보드          |  High-Level        |
|   비쌈 ($6-21/user) |   (app.tene.sh)     |  Concept:          |
|                     |                     |  "가장 저렴한      |
|  Existing Alt.:     | - OAuth 인증        |  Zero-Knowledge    |
|  Doppler, Infisical |   (가입 없음)       |  시크릿 Cloud"     |
|  1Password, Bitwarden|                    |                     |
+--------------------+--------------------+--------------------+
|  8. KEY METRICS     |                     |  9. UNFAIR ADV.     |
|                     |                     |                     |
| - MRR               |                     | - 기존 CLI 사용자   |
| - Solo 전환율       |                     |   base (Phase 1)    |
| - Team 팀 수        |                     | - AI 에이전트       |
| - Push/Pull 성공률  |                     |   자동 인식 유지    |
| - Churn rate        |                     | - 완전 ZK + 최저가  |
| - NPS               |                     |   ($5 Solo)         |
|                     |                     | - Go 모노레포       |
+--------------------+--------------------+--------------------+
|  7. COST            |                     |  6. REVENUE         |
|                     |  2. CUSTOMER        |                     |
| AWS 인프라:         |  SEGMENTS           | Solo: $5/mo (flat)  |
| - 초기 ~$58/월      |                     | Team: $10/user/mo   |
| - 100명: ~$58/월    | - 1차: 기존 CLI     |   (per-seat)        |
| - 1K명: ~$149/월    |   사용자 (Solo)     |                     |
| - 10K명: ~$462/월   |                     | 예상:               |
|                     | - 2차: 소규모       | 100명: $500+/월     |
| Stripe 수수료:      |   개발팀 (Team)     | 1K명: $5,000+/월    |
| 2.9% + $0.30/건     |                     | 10K명: $50,000+/월  |
|                     | - 3차: AI-first     |                     |
| 마케팅: $0          |   개발 조직         | 이익률:             |
| (오픈소스+커뮤니티)  |                     | 88% → 97% → 99%    |
|                     |                     |                     |
|  5. CHANNELS        |                     |                     |
| - CLI 내장 업셀     |                     |                     |
| - tene.sh 랜딩      |                     |                     |
| - GitHub 오픈소스   |                     |                     |
| - Dev 커뮤니티      |                     |                     |
+--------------------+--------------------+--------------------+
```

---

## 3. SWOT 분석 (Tene Cloud)

### 3.1 SWOT Matrix

| | **Helpful** | **Harmful** |
|---|---|---|
| **Internal** | **Strengths** | **Weaknesses** |
| | S1: **완전 Zero-Knowledge** (Envelope + ECDH) | W1: 1인 팀 -- 개발 속도 제한 |
| | S2: **최저가** ($5 Solo, $10 Team) | W2: 클라우드 운영 경험 부족 |
| | S3: **기존 CLI 사용자 base** (전환 파이프라인) | W3: 브랜드 인지도 제로 |
| | S4: **AI 에이전트 자동 인식** 유지 | W4: Sync 충돌 해결 UX 미검증 |
| | S5: **Go 모노레포** (CLI+Server 코드 공유) | W5: 대시보드 기능 제한 (ZK 제약) |
| | S6: **$58/월** 초기 인프라 비용 | W6: Waitlist 수요 검증 미완 |
| **External** | **Opportunities** | **Threats** |
| | O1: Secrets Management 시장 $4.2B → $10B (CAGR 13.4%) | T1: 1Password Unified Access (브랜드+자본) |
| | O2: AI 시크릿 누출 81% 급증 -- 보안 수요 증가 | T2: Infisical 오픈소스 + 12.7K Stars |
| | O3: 개발자 "가격 민감" -- $21(Doppler) 대비 76% 저렴 | T3: Bitwarden SM $6/user 가격 경쟁 |
| | O4: 중소 개발팀 시크릿 관리 미성숙 | T4: 대형 경쟁사의 무료 티어 확대 |
| | O5: Local-First + Cloud-Optional 트렌드 | T5: "로컬 전용으로 충분" 관성 |

### 3.2 SO/WT 전략

**SO 전략 (강점 x 기회)**

- **SO1**: 완전 ZK(S1) + 시크릿 누출 위기(O2) = **"서버도 볼 수 없는 시크릿 관리"** 보안 마케팅
- **SO2**: 최저가(S2) + 가격 민감(O3) = **"Doppler의 1/4 가격으로 더 나은 보안"** 가격 포지셔닝
- **SO3**: CLI base(S3) + 중소팀 미성숙(O4) = **Free CLI -> Solo -> Team 자연 업셀** 성장 파이프라인
- **SO4**: Go 모노레포(S5) + $4.2B 시장(O1) = 빠른 개발 속도로 시장 선점

**WT 전략 (약점 보완 x 위협 대응)**

- **WT1**: 1인 팀(W1) + 대형 경쟁사(T1) = Solo 개발자 니치에 집중, 엔터프라이즈 회피
- **WT2**: 인지도 부재(W3) + Infisical Stars(T2) = 오픈소스 CLI의 자연 확산에 의존
- **WT3**: 수요 미검증(W6) + "로컬 충분" 관성(T5) = Fake Door 데이터로 Go/No-Go 판단
- **WT4**: ZK 제약(W5) + Bitwarden 가격 경쟁(T3) = CLI-first, 대시보드는 보조적 역할

---

## 4. Porter's 5 Forces 분석 (Tene Cloud)

### 4.1 5 Forces Summary

```
신규 진입 위협: 중간 (3/5) -- Cloud 운영은 진입 장벽 있음
          |
          v
공급자 교섭력 --> 산업 경쟁 강도 <-- 구매자 교섭력
   낮음 (2/5)       높음 (4/5)       높음 (4/5)
   (AWS 대체 가능)   (Doppler,         (무료 대안 다수,
                     Infisical,        전환 비용 낮음)
                     1Password 등)
          ^
          |
대체재 위협: 높음 (4/5) -- .env, GitHub Secrets가 무료 대체재
```

### 4.2 전략적 시사점

| Force | 대응 전략 |
|-------|----------|
| 높은 경쟁 강도 | **"ZK + AI 자동 인식 + 최저가"** 3중 차별화 니치 |
| 높은 구매자 교섭력 | Free CLI로 Lock-in 후 Solo/Team 자연 업셀 |
| 높은 대체재 위협 | CLI → Cloud 자연 전환 경로 설계 (push/pull UX) |
| 중간 신규 진입 | 빠른 실행 + 커뮤니티 선점 |
| 낮은 공급자 교섭력 | AWS 비용 최적화 ($58/월 시작) |

---

# Part 3: Market Research

> PM Agent: pm-research | Frameworks: Persona, Competitive Analysis, TAM/SAM/SOM, Customer Journey Map

## 1. 사용자 페르소나 (Cloud 전환)

### Persona 1: "재현" -- Solo Sync 핵심 사용자

| 항목 | 상세 |
|------|------|
| **이름** | 박재현 |
| **나이/직업** | 31세 / 인디 해커, 사이드 프로젝트 다수 |
| **기술 수준** | 상급. 프로젝트 10개+ 동시 운영 |
| **디바이스** | MacBook Pro (회사) + MacBook Air (개인) + Ubuntu (서버) |
| **관리 시크릿 수** | 20-50개 |
| **현재 방법** | Tene CLI (Free) + `tene export --encrypted` 수동 백업 |
| **JTBD** | "회사와 집 맥북에서 같은 시크릿을 안전하게 쓰고 싶다" |
| **Pain Points** | (1) 수동 백업 번거로움 (2) 새 디바이스에서 시크릿 재설정 (3) 디스크 장애 불안 |
| **Trigger** | "tene push 하면 자동으로 암호화 백업? $5면 싸네" |
| **WTP** | **$5/월** -- 멀티 디바이스 동기화 + 자동 백업 |

### Persona 2: "민수" -- Team 핵심 사용자

| 항목 | 상세 |
|------|------|
| **이름** | 김민수 |
| **나이/직업** | 34세 / 스타트업 Tech Lead (팀 5명) |
| **디바이스** | 팀원 5명 x 개인 랩탑 |
| **관리 시크릿 수** | 50-100개 (팀 전체, dev/staging/prod) |
| **현재 방법** | .env 파일 Slack DM 공유 + 1Password 부분 사용 |
| **JTBD** | "팀원에게 prod 시크릿을 안전하게 전달하고, 접근 기록을 추적하고 싶다" |
| **Pain Points** | (1) Slack DM 시크릿 공유 보안 위험 (2) 1Password $7.99/user 비용 부담 (3) prod 접근 제어 없음 |
| **Trigger** | "팀원 초대하면 X25519로 키가 래핑된다고? $10이면 1Password보다 싸다" |
| **WTP** | **$10/user/월** -- 팀 5명 = $50/월 (1Password $40 대비 유사하나 ZK 우위) |

### Persona 3: "Emily" -- AI-First 팀 사용자 (잠재)

| 항목 | 상세 |
|------|------|
| **이름** | Emily Park |
| **나이/직업** | 29세 / AI 스타트업 개발자 (팀 3명) |
| **특징** | Claude Code + Cursor로 전체 개발, AI 에이전트 의존도 높음 |
| **JTBD** | "AI 에이전트가 자동으로 인식하면서 팀과 시크릿을 공유할 수 있는 도구" |
| **WTP** | **$10/user/월** -- AI 자동 인식 + 팀 공유 조합 |
| **Status** | 잠재적 -- AI-first 팀 시크릿 공유 수요 검증 필요 |

---

## 2. 경쟁사 분석 (Cloud 관점, 5개사)

### 2.1 경쟁사 비교 매트릭스

| 항목 | **Tene Cloud** | Infisical | Doppler | 1Password | Bitwarden SM |
|------|:------------:|:---------:|:-------:|:---------:|:------------:|
| **Zero-Knowledge** | **완전** (Envelope+ECDH) | **부분적** (Workspace Key) | **X** (서버측) | **O** (2SKD) | **O** (RSA-OAEP) |
| **가격 (Solo)** | **$5/mo** | $6/user | $21/user | $7.99/user | $6/user |
| **가격 (Team)** | **$10/user** | $18/user | $21/user | $7.99/user | $6/user |
| **AI 자동 인식** | **O** (CLAUDE.md) | X | X | X | X |
| **로컬 CLI** | **O** (Go 바이너리) | O (Node.js) | O (Node.js) | O | X |
| **오프라인** | **O** (Free 유지) | X | X | X | X |
| **오픈소스** | **MIT** | MIT | X | X | X (SM) |
| **팀 키 교환** | **X25519 ECDH** | Workspace Key | 서버 관리 | SRP+2SKD | RSA-OAEP |
| **환경 제어** | **O** (RBAC) | O | O | X | X |
| **감사 로그** | **O** | O (유료) | O | O (유료) | X |
| **GitHub Stars** | ~500 (성장 중) | 12,700+ | N/A | N/A | 2,000+ |
| **타겟** | 솔로+소규모팀 | DevOps 팀 | 엔터프라이즈 | 엔터프라이즈 | IT팀 |

### 2.2 경쟁사별 심층 분석

#### (1) Infisical -- $6-18/user/mo

- **강점**: 오픈소스 (MIT), 12.7K Stars, Agent Sentinel (MCP), PKI/PAM 확장
- **약점**: Zero-Knowledge가 부분적 (Workspace Key는 서버에서 관리 가능), 복잡도 증가
- **Tene 차별점**: 완전 ZK, AI 자동 인식, 더 단순한 UX, Solo $5로 더 저렴

#### (2) Doppler -- $21/user/mo

- **강점**: 30+ 네이티브 통합, 빠른 개발자 온보딩, Universal Dashboard
- **약점**: **Zero-Knowledge 아님** (서버측 암호화), 가장 비쌈
- **Tene 차별점**: 완전 ZK (Doppler가 절대 못 하는 것), **76% 저렴** ($5 vs $21)

#### (3) 1Password -- $7.99/user/mo

- **강점**: 강력한 브랜드, 2SKD+SRP, Unified Access (2026), Anthropic 파트너십
- **약점**: 패스워드 관리자에서 시크릿 관리로 확장 -- 전문성 부족, 비공개소스
- **Tene 차별점**: 오픈소스, CLI-first, AI 자동 인식, Solo $5로 더 저렴

#### (4) Bitwarden Secrets Manager -- $6/user/mo

- **강점**: Zero-Knowledge (RSA-OAEP), 가격 경쟁력, 기존 Bitwarden 사용자 base
- **약점**: CLI 도구 부재, AI 에이전트 통합 없음, 시크릿 관리 기능 제한적
- **Tene 차별점**: Go CLI, AI 자동 인식, 로컬 오프라인 유지, Solo $5로 더 저렴

#### (5) dotenv-vault -- Deprecated

- **상태**: 2025년 Deprecated. 단일 키 + 클라우드 종속 모델의 실패 사례
- **교훈**: Cloud 의존 전용 모델은 실패. **Local-First + Cloud-Optional이 정답**

### 2.3 Battlecard: Tene Cloud vs 경쟁사

| 비교 포인트 | Tene Cloud 우위 | 경쟁사 우위 |
|------------|----------------|-----------|
| **가격** | Solo $5 (최저) | Bitwarden SM $6 (유사) |
| **Zero-Knowledge** | 완전 (Envelope + ECDH) | 1Password (2SKD), Bitwarden (RSA) |
| **AI 자동 인식** | **유일** (CLAUDE.md) | 없음 |
| **오프라인** | **O** (Free 유지) | 모든 경쟁사 X |
| **오픈소스** | **MIT** | Infisical (MIT) |
| **설치 간편성** | `brew install` | Doppler (빠른 온보딩) |
| **팀 규모** | 소규모 (2-10) | Doppler/Infisical (대규모) |
| **엔터프라이즈** | X (미지원) | 1Password, Infisical, Doppler |
| **통합 수** | 제한 (CLI 중심) | Doppler 30+, Infisical K8s/Docker |

---

## 3. 시장 규모 (TAM/SAM/SOM) -- Cloud 서비스 관점

### 3.1 TAM (Total Addressable Market)

| 출처 | 연도 | 시장 규모 | CAGR |
|------|------|-----------|------|
| KBV Research | 2032 | $10.09B | 13.4% |
| Mordor Intelligence | 2025 | $4.22B | 13.8% |
| Growth Market Reports | 2033 | $5.6B | 16.5% |
| Business Research Insights | 2025 | $1.85B (Key Mgmt) | 19.77% |

> **TAM = $4.22B** (2025, Secrets Management 전체)

### 3.2 SAM (Serviceable Addressable Market) -- 방법론 2개

**방법 A: 사용자 수 기반**
```
전체 개발자 수 (2026): ~30M
AI 코딩 도구 사용: 30M x 60% = 18M
시크릿 관리 유료 의향: 18M x 15% = 2.7M
$5-10/mo 가격대 수용: 2.7M x 40% = 1.08M
연간 SAM = 1.08M x $60-120/yr = $64.8M - $129.6M
```

**방법 B: 시장 세분화 기반**
```
TAM $4.22B
개발자 시크릿 관리 (전체의 20%): $844M
소규모 팀 + 개인 (30%): $253M
CLI-first 선호 (40%): $101M
연간 SAM = ~$101M
```

> **SAM = ~$100M** (두 방법의 수렴)

### 3.3 SOM (Serviceable Obtainable Market) -- 12개월 목표

```
Phase 2a (Solo Sync, Month 1-4):
  목표: 유료 전환 50명 x $5/mo = MRR $250
  
Phase 2b (Team, Month 5-7):
  목표: 10팀 x 평균 4명 x $10/mo = MRR $400
  
Phase 2c+ (안정화, Month 8-12):
  목표: Solo 100명 + Team 20팀 = MRR $500 + $800 = $1,300

Year 1 SOM = MRR $1,300 = ARR ~$15,600
Year 2 목표 = MRR $5,000 = ARR ~$60,000
```

> **Year 1 SOM = ARR ~$15,600** (보수적), **Year 2 목표 = ARR ~$60,000**

---

## 4. Customer Journey Map (Solo Sync 사용자 -- 재현)

```
[Stage 1: 인식]
재현은 Tene CLI를 6개월째 사용 중. 회사와 집에서 시크릿이 다르다.
"tene export --encrypted 매주 하는데 귀찮다..."
                    |
[Stage 2: 발견]
$ tene sync
→ "클라우드 동기화가 출시되었습니다! tene login으로 시작하세요."
→ "Solo $5/월? 그 정도면 괜찮네"
                    |
[Stage 3: 가입 + 결제]
$ tene login
→ GitHub OAuth → 즉시 인증
$ tene billing
→ Stripe Checkout → Solo $5/mo 결제
→ "1분 만에 끝났다"
                    |
[Stage 4: 첫 동기화]
$ tene push
→ Sync Envelope 암호화 → S3 업로드
→ "Uploading... [████████] 100%. Push 완료 (v1, 1.2 MB)"
                    |
[Stage 5: 멀티 디바이스]
(개인 MacBook에서)
$ tene login
$ tene pull
→ "Pull 완료 — 15 secrets synced"
→ "와, 회사에서 설정한 시크릿이 다 있다!"
                    |
[Stage 6: 습관화]
$ tene push  (하루 1-2회, 작업 후)
$ tene pull  (디바이스 전환 시)
→ 자동 백업, 디바이스 동기화가 습관
→ "이제 수동 export는 안 해도 된다"
                    |
[Stage 7: 업셀 모먼트]
→ 팀원이 생김 → "시크릿 어떻게 공유하지?"
→ $ tene team create myteam
→ Team $10/user/mo 업그레이드
```

---

# Part 4: PRD Synthesis & Execution

> PM Agent: pm-prd | Frameworks: ICP, Beachhead Segment, GTM, PRD 8-Section, Pre-mortem

## 1. ICP (Ideal Customer Profile)

### 1.1 ICP Definition

| Dimension | Solo (Primary) | Team (Secondary) |
|-----------|---------------|-----------------|
| **회사 규모** | 개인 / 1-2인 | 2-10인 스타트업 |
| **기술 스택** | Go, Node.js, Python + Claude Code | 동일 + 팀 협업 도구 |
| **시크릿 수** | 10-50개 | 50-200개 |
| **현재 방법** | Tene CLI Free + 수동 백업 | .env + Slack DM |
| **디바이스 수** | 2-3대 | 5-15대 (팀 전체) |
| **보안 요구** | "서버가 볼 수 없어야" | "접근 제어 + 감사 필요" |
| **가격 민감도** | 높음 ($5 이하 선호) | 중간 ($10/user 수용) |
| **전환 트리거** | 멀티 디바이스 동기화 필요 | 팀원 합류 + 시크릿 공유 |

---

## 2. Beachhead Segment (Geoffrey Moore)

### 2.1 4-Criteria Scoring

| 기준 | Solo 개인 개발자 | 소규모 팀 (2-10명) | AI-First 조직 |
|------|:-------------:|:----------------:|:------------:|
| **접근 가능성** (기존 CLI base) | 9/10 | 6/10 | 5/10 |
| **구매 긴급성** (Pain 강도) | 7/10 | 9/10 | 6/10 |
| **지불 의향** ($5-10) | 8/10 | 7/10 | 7/10 |
| **참조 가치** (입소문 영향) | 6/10 | 8/10 | 9/10 |
| **합계** | **30/40** | **30/40** | **27/40** |

### 2.2 Beachhead 선택: **Solo 개인 개발자** (우선) + **소규모 팀** (즉시 이어서)

**선택 근거**:
- Solo와 Team이 동점(30)이지만, Solo가 먼저인 이유:
  1. **접근 가능성 9/10**: 기존 CLI 사용자 base에서 바로 전환 가능
  2. **구현 순서**: Phase 2a(Solo) → Phase 2b(Team) 자연 순서
  3. **성장 경로**: Solo → Team 자연 업셀 (팀원 합류 시)

---

## 3. GTM (Go-To-Market) 전략

### 3.1 전략 개요

| Phase | 기간 | 목표 | 핵심 액션 |
|-------|------|------|----------|
| **Phase 2a** | 4주 | Solo 50명 유료 전환 | CLI 내장 업셀 + waitlist 전환 |
| **Phase 2b** | 3주 | 10팀 유료 전환 | Team 기능 출시 + 팀 레퍼럴 |
| **Phase 2c** | 1주 | MRR $500+ 달성 | 안정화 + 마케팅 강화 |

### 3.2 채널 전략

| 채널 | 전략 | 예상 효과 |
|------|------|----------|
| **CLI 내장 업셀** | `tene sync` → Cloud 출시 안내, `tene push` 시 로그인 안내 | **최고 전환 채널** (기존 사용자) |
| **tene.sh 랜딩** | Cloud 기능 섹션 추가, 가격 페이지, ZK 설명 | 신규 방문자 전환 |
| **GitHub README** | Cloud 기능 배지, Solo/Team 가격 표시 | 오픈소스 발견 경로 |
| **Dev 커뮤니티** | Hacker News, Reddit r/vibecoding, GeekNews | 입소문 + 인지도 |
| **Twitter/X** | @tene_sh 개발 로그, 보안 인사이트 공유 | 팔로워 확보 |

### 3.3 핵심 메시지

**Solo 메시지**:
> "tene push 한 번이면 모든 디바이스에서 시크릿 동기화.
> Zero-Knowledge -- 서버도 당신의 시크릿을 볼 수 없습니다. $5/월."

**Team 메시지**:
> "Slack DM으로 API 키를 보내지 마세요.
> X25519 암호화로 팀 시크릿을 안전하게 공유하세요. $10/user/월."

### 3.4 Growth Loops

```
[Loop 1: Solo Organic]
CLI 무료 사용자 → tene sync → Cloud 발견 → Solo $5 전환
→ 만족 → 커뮤니티 추천 → 새 CLI 사용자 유입

[Loop 2: Team Referral]
Solo 사용자 → 팀원 합류 → tene team invite → Team $10 전환
→ 팀원도 개인 프로젝트에 Solo 사용 → 확산

[Loop 3: Content Loop]
ZK 보안 블로그 → 개발자 유입 → CLI 설치 → Solo/Team 전환
→ 사례 공유 → 더 많은 콘텐츠 유입
```

### 3.5 가격 전략 근거

| 플랜 | 가격 | 포지셔닝 | 경쟁 대비 |
|------|------|---------|----------|
| **Free** | $0 | 가장 안전한 무료 로컬 시크릿 관리 | 유일 (CLI+오프라인+AI) |
| **Solo** | $5/mo | "커피 한 잔 가격의 ZK Cloud Sync" | Bitwarden $6 대비 17% 저렴 |
| **Team** | $10/user/mo | "Doppler의 절반 가격으로 더 나은 보안" | Doppler $21 대비 52% 저렴 |

---

## 4. PRD 8-Section

### Section 1: Product Overview

**제품명**: Tene Cloud
**버전**: Phase 2 (v1.0)
**목표**: Tene CLI의 Zero-Knowledge 보안을 유지하며 Cloud Sync + Team Sharing + Dashboard 제공

### Section 2: Objectives & Key Results

| Objective | Key Result | 측정 방법 |
|-----------|-----------|----------|
| **O1: Solo 유료 전환** | KR1: Solo 50명 달성 (3개월) | Stripe 구독 수 |
| | KR2: 전환율 10% (CLI → Solo) | CLI 사용자 대비 Solo 비율 |
| **O2: Team 유료 전환** | KR3: 10팀 달성 (3개월) | 팀 생성 수 |
| | KR4: 평균 팀 크기 4명 | 팀 멤버 수 / 팀 수 |
| **O3: 매출** | KR5: MRR $500+ 달성 | Stripe MRR |
| | KR6: Churn Rate < 5% | 월간 해지율 |
| **O4: 기술 품질** | KR7: Push/Pull 성공률 99.5%+ | API 모니터링 |
| | KR8: API P99 < 500ms | CloudWatch 메트릭 |
| **O5: 서비스 안정성** | KR9: 가용성 99.9% | Uptime 모니터링 |

### Section 3: User Stories

#### Epic 1: 인증 (Authentication)

| ID | User Story | Priority | INVEST |
|----|-----------|----------|--------|
| US-101 | 사용자로서, GitHub OAuth로 로그인하여 별도 가입 없이 시작하고 싶다 | P0 | I:독립 N:OAuth S:작음 T:검증가능 E:견적가능 V:필수 |
| US-102 | 사용자로서, Google OAuth로 로그인하여 GitHub이 없어도 시작하고 싶다 | P0 | 동일 |
| US-103 | 사용자로서, CLI에서 `tene login`으로 브라우저 기반 인증을 하고 싶다 | P0 | I:독립 N:CLI-browser S:작음 T:검증가능 |
| US-104 | 사용자로서, `tene logout`으로 토큰을 안전하게 삭제하고 싶다 | P1 | I:독립 S:작음 |

#### Epic 2: Vault Sync (Solo)

| ID | User Story | Priority | INVEST |
|----|-----------|----------|--------|
| US-201 | 사용자로서, `tene push`로 암호화된 vault를 클라우드에 업로드하고 싶다 | P0 | I:독립 N:ZK sync S:중간 T:검증가능 E:견적가능 V:핵심 |
| US-202 | 사용자로서, `tene pull`로 클라우드에서 vault를 다운로드하고 싶다 | P0 | 동일 |
| US-203 | 사용자로서, push/pull 시 Sync Envelope 이중 암호화가 적용되어 서버가 내 시크릿을 볼 수 없음을 확신하고 싶다 | P0 | V:핵심 차별점 |
| US-204 | 사용자로서, Sync 충돌 시 명시적 해결 옵션(Remote/Local/Manual)을 선택하고 싶다 | P1 | N:충돌 해결 S:중간 |
| US-205 | 사용자로서, `tene sync`를 실행하면 push+pull이 자동으로 수행되길 원한다 | P2 | S:작음 |

#### Epic 3: Team Sharing

| ID | User Story | Priority | INVEST |
|----|-----------|----------|--------|
| US-301 | 팀 리더로서, `tene team create`로 팀을 생성하고 싶다 | P0 | I:독립 S:작음 |
| US-302 | 팀 리더로서, `tene team invite`로 멤버를 초대하고 X25519로 키를 공유하고 싶다 | P0 | N:ECDH S:큼 T:보안테스트 V:핵심 |
| US-303 | 팀 관리자로서, 멤버 제거 시 자동 키 회전이 발생하여 이전 멤버가 접근할 수 없음을 확인하고 싶다 | P0 | V:보안 필수 |
| US-304 | 팀 관리자로서, 멤버별 환경 접근 권한(dev: read/write, prod: read)을 설정하고 싶다 | P1 | N:RBAC S:중간 |
| US-305 | 팀 멤버로서, `tene pull`로 팀 시크릿에 접근하고 싶다 | P0 | S:작음 |

#### Epic 4: Dashboard

| ID | User Story | Priority | INVEST |
|----|-----------|----------|--------|
| US-401 | 사용자로서, 대시보드에서 내 Vault 목록과 마지막 동기화 시간을 보고 싶다 | P1 | I:독립 S:작음 |
| US-402 | 사용자로서, 대시보드에서 등록된 디바이스를 관리하고 싶다 | P1 | S:작음 |
| US-403 | 사용자로서, 대시보드에서 감사 로그(누가 언제 push/pull 했는지)를 확인하고 싶다 | P1 | S:중간 |
| US-404 | 팀 관리자로서, 대시보드에서 팀 멤버와 역할을 관리하고 싶다 | P1 | S:중간 |
| US-405 | 사용자로서, 대시보드에서 시크릿 값을 절대 볼 수 없어야 한다 (ZK 원칙) | P0 | V:보안 필수 |

#### Epic 5: Billing

| ID | User Story | Priority | INVEST |
|----|-----------|----------|--------|
| US-501 | 사용자로서, `tene billing`으로 현재 구독 상태를 확인하고 싶다 | P1 | S:작음 |
| US-502 | 사용자로서, Stripe Checkout으로 Solo/Team 구독을 시작하고 싶다 | P0 | S:중간 T:Stripe 테스트 |
| US-503 | 사용자로서, Stripe Portal로 구독을 셀프 관리하고 싶다 | P1 | S:작음 |
| US-504 | 시스템으로서, 결제 실패 시 7일 유예 후 Free로 다운그레이드되어야 한다 | P1 | S:중간 |

### Section 4: Functional Requirements

#### FR-1: Authentication

| ID | 요구사항 | Priority |
|----|---------|----------|
| FR-1.1 | GitHub OAuth 2.0 인증 흐름 (redirect → callback → JWT) | P0 |
| FR-1.2 | Google OAuth 2.0 인증 흐름 | P0 |
| FR-1.3 | JWT Access Token (15분) + Refresh Token (30일) | P0 |
| FR-1.4 | CLI Device Code Flow (브라우저 ↔ CLI 인증 연결) | P0 |
| FR-1.5 | 첫 로그인 시 자동 계정 생성 (별도 가입 없음) | P0 |
| FR-1.6 | Refresh Token DB 저장 + 폐기 기능 | P1 |

#### FR-2: Vault Sync

| ID | 요구사항 | Priority |
|----|---------|----------|
| FR-2.1 | Sync Envelope 암호화: vault.db 전체를 XChaCha20-Poly1305로 추가 암호화 | P0 |
| FR-2.2 | `POST /vaults/:id/push` -- 암호화된 blob S3 업로드 | P0 |
| FR-2.3 | `GET /vaults/:id/pull` -- S3에서 blob 다운로드 | P0 |
| FR-2.4 | vault_version + vault_hash 기반 충돌 감지 | P0 |
| FR-2.5 | 충돌 해결: Remote 선택 / Local 유지 / 수동 병합 | P1 |
| FR-2.6 | S3 Object Versioning으로 이전 버전 복원 | P2 |
| FR-2.7 | Device fingerprint 기반 멀티 디바이스 관리 | P1 |

#### FR-3: Team Management

| ID | 요구사항 | Priority |
|----|---------|----------|
| FR-3.1 | X25519 키 쌍 생성 (디바이스별, UMK로 개인키 암호화) | P0 |
| FR-3.2 | X25519 ECDH 키 교환: Project Key를 멤버 공개키로 래핑 | P0 |
| FR-3.3 | 멤버 제거 시 키 회전: 새 PK 생성 → 재래핑 → 재암호화 | P0 |
| FR-3.4 | RBAC: admin (전체 접근), member (제한 접근) | P1 |
| FR-3.5 | 환경별 접근 권한: env_permissions JSONB 필드 | P1 |
| FR-3.6 | Team 생성/초대/제거/목록 API | P0 |

#### FR-4: Dashboard

| ID | 요구사항 | Priority |
|----|---------|----------|
| FR-4.1 | OAuth 로그인 (GitHub/Google) | P0 |
| FR-4.2 | Vault 목록 (이름, 환경, 시크릿 수, 마지막 동기화) | P1 |
| FR-4.3 | 디바이스 관리 (목록, 등록 해제) | P1 |
| FR-4.4 | 감사 로그 조회 (action, user, timestamp) | P1 |
| FR-4.5 | 팀 관리 (멤버 목록, 역할 변경, 초대) | P1 |
| FR-4.6 | 구독 관리 (현재 플랜, Stripe Portal 링크) | P1 |
| FR-4.7 | 시크릿 값 표시/편집 기능 **없음** (ZK 원칙 강제) | P0 |

#### FR-5: Billing

| ID | 요구사항 | Priority |
|----|---------|----------|
| FR-5.1 | Stripe Checkout Session 생성 (Solo, Team) | P0 |
| FR-5.2 | Stripe Webhook 수신 (subscription.created/updated/deleted) | P0 |
| FR-5.3 | invoice.payment_failed → 7일 유예 → Free 다운그레이드 | P1 |
| FR-5.4 | Team per-seat billing (수량 변경 시 proration) | P0 |
| FR-5.5 | Stripe Billing Portal URL 생성 | P1 |

### Section 5: Non-Functional Requirements

| 카테고리 | 요구사항 | 기준 |
|----------|---------|------|
| **성능** | API P99 응답시간 | < 500ms |
| **성능** | Push/Pull 성공률 | 99.5%+ |
| **가용성** | 서비스 가용성 | 99.9% (월 43분 이하 다운타임) |
| **보안** | Zero-Knowledge 보장 | 서버 측 평문 시크릿/마스터키 접근 불가 |
| **보안** | 전송 암호화 | TLS 1.2+ |
| **보안** | 저장 암호화 | S3 SSE-S3 (AES-256) |
| **보안** | Rate Limiting | Free 100 req/min, Paid 1,000 req/min |
| **확장성** | 10,000 사용자까지 수직 확장 | RDS/ECS 스케일 업 |
| **비용** | 초기 인프라 | ~$58/월 |
| **규정** | OWASP Top 10 준수 | Phase 2c 보안 감사 |

### Section 6: Technical Architecture (Summary)

**상세 아키텍처는 `docs/01-plan/features/tene-cloud.plan.md` Section 2-5 참조**

- API 서버: Go (Echo/Gin), ECS Fargate
- DB: PostgreSQL (RDS), 6 tables
- Storage: S3 (암호화된 Vault blob)
- Auth: JWT + OAuth (GitHub, Google)
- Billing: Stripe Checkout + Webhook
- Dashboard: Next.js + shadcn/ui (Vercel)
- IaC: Terraform (VPC, ECS, ALB, RDS, S3)
- CI/CD: GitHub Actions -> ECR -> ECS

### Section 7: Milestones & Timeline

| Phase | 기간 | 핵심 산출물 |
|-------|------|-----------|
| **Phase 2a: Solo Sync** | Week 1-4 | 인프라 + 인증 + Sync API + Stripe Solo + 대시보드 MVP |
| **Phase 2b: Team** | Week 5-7 | X25519 키 교환 + Team API + RBAC + Stripe Team + 대시보드 Team |
| **Phase 2c: 안정화** | Week 8 | E2E 테스트 + 보안 감사 + 문서 + 랜딩 업데이트 |

**상세 주차별 계획은 `docs/01-plan/features/tene-cloud.plan.md` Section 10 참조**

### Section 8: Success Criteria

| 기준 | 목표 (출시 3개월) | 측정 방법 |
|------|:----------------:|----------|
| 회원가입 수 | 500+ | DB users count |
| Solo 유료 전환 | 50+ (전환율 10%) | Stripe subscription |
| Team 유료 전환 | 10+ 팀 | DB teams count |
| MRR | $500+ | Stripe MRR |
| Push/Pull 성공률 | 99.5%+ | API monitoring |
| API P99 응답시간 | < 500ms | CloudWatch |
| 서비스 가용성 | 99.9% | Uptime monitor |
| NPS | 40+ | 사용자 설문 |
| Churn Rate | < 5%/월 | Stripe churn |

---

## 5. Pre-mortem Analysis

### 5.1 Top 3 Risks

| 순위 | 리스크 | 확률 | 영향 | 사전 시나리오 | 완화 전략 |
|------|-------|:----:|:----:|------------|----------|
| 1 | **Cloud 수요 부족**: waitlist 100명 미달, 유료 전환 10명 미달 | 40% | Critical | "8주 개발했지만 MRR $50도 안 된다" | Fake Door 데이터로 Go/No-Go 판단, 인프라 $58/월이므로 손실 최소 |
| 2 | **X25519 키 교환 보안 결함**: 구현 버그로 키 노출 | 15% | Critical | "팀 키가 노출되어 시크릿 유출 사고 발생" | golang.org/x/crypto 사용, 보안 감사, 오픈소스 리뷰 |
| 3 | **Sync 충돌 데이터 손실**: push/pull 충돌 해결 오류로 시크릿 소실 | 25% | High | "pull 했더니 최근 추가한 시크릿이 사라졌다" | S3 버전 관리, 로컬 백업 유지, 명시적 충돌 UX |

### 5.2 추가 리스크

| 리스크 | 확률 | 영향 | 완화 |
|-------|:----:|:----:|------|
| Stripe 한국 결제 불안정 | 20% | Medium | Lemon Squeezy 대안 준비 |
| 1인 팀 번아웃 (8주 집중 개발) | 30% | High | Phase 분할로 점진적 출시 |
| 경쟁사 가격 인하 (Bitwarden $4 등) | 15% | Medium | ZK+AI 자동 인식 차별화 유지 |
| OAuth 제공자 장애 (GitHub/Google) | 5% | Medium | 2개 OAuth 제공, 로그인 토큰 30일 유지 |

---

## 6. Test Scenarios (핵심)

### TS-1: Authentication

| ID | 시나리오 | 예상 결과 |
|----|---------|----------|
| TS-1.1 | GitHub OAuth로 첫 로그인 | 계정 자동 생성, JWT 발급, CLI에 토큰 저장 |
| TS-1.2 | Refresh Token 만료 후 접근 | 401 → 자동 갱신 → 정상 접근 |
| TS-1.3 | `tene logout` 실행 | 토큰 삭제, Keychain에서 제거 |

### TS-2: Vault Sync

| ID | 시나리오 | 예상 결과 |
|----|---------|----------|
| TS-2.1 | `tene push` (첫 push) | Sync Envelope 암호화 → S3 업로드 → vault_version=1 |
| TS-2.2 | `tene pull` (다른 디바이스) | S3 다운로드 → Envelope 복호화 → 로컬 vault 업데이트 |
| TS-2.3 | Push 충돌 (원격 v3, 로컬 v2) | 충돌 감지 → "Remote/Local/Manual" 선택 프롬프트 |
| TS-2.4 | S3에서 vault blob 직접 열기 시도 | Envelope 암호화로 평문 접근 불가 |
| TS-2.5 | 네트워크 끊김 시 push | 에러 메시지 + 로컬 vault 무결성 유지 |

### TS-3: Team Sharing

| ID | 시나리오 | 예상 결과 |
|----|---------|----------|
| TS-3.1 | `tene team invite alice@example.com` | Alice 공개키 조회 → ECDH → wrapped_pk 업로드 |
| TS-3.2 | Alice가 `tene pull` | wrapped_pk 복호화 → PK 획득 → vault 복호화 |
| TS-3.3 | `tene team remove alice` | PK' 생성 → 나머지 멤버 재래핑 → vault 재암호화 |
| TS-3.4 | Member가 prod 환경 write 시도 (권한: read-only) | 403 Forbidden |
| TS-3.5 | 서버 DB에서 wrapped_pk 직접 조회 | 암호문만 보임, PK 복호화 불가 |

### TS-4: Billing

| ID | 시나리오 | 예상 결과 |
|----|---------|----------|
| TS-4.1 | Solo 구독 시작 | Stripe Checkout → 결제 → plan="solo" 업데이트 |
| TS-4.2 | Team 멤버 추가 | 수량 증가 → proration 적용 |
| TS-4.3 | 결제 실패 | invoice.payment_failed → 이메일 알림 → 7일 유예 |
| TS-4.4 | 구독 해지 | Free 다운그레이드 → Cloud Sync 비활성화 → 로컬 데이터 유지 |

---

## 7. Stakeholder Map

| Stakeholder | 관심사 | 영향력 | 관여 방식 |
|-------------|--------|:------:|----------|
| **Steve (Founder)** | 기술 아키텍처, 보안, 비즈니스 모델 | High | 모든 의사결정 |
| **CLI 사용자** | UX 유지, 가격, 보안 | High | Fake Door, waitlist, 피드백 |
| **잠재 팀 사용자** | 팀 기능, RBAC, 가격 | Medium | waitlist, 인터뷰 |
| **Stripe** | 결제 통합, 수수료 | Low | API 통합 |
| **AWS** | 인프라 비용, 가용성 | Low | 서비스 이용 |

---

## 8. Attribution & References

### PM Agent Team

| Agent | Framework | 분석 범위 |
|-------|-----------|----------|
| **pm-discovery** | Teresa Torres OST + 5-Step Discovery Chain | Cloud Pain Points, 가정, 실험, OST |
| **pm-strategy** | JTBD 6-Part VP, Lean Canvas, SWOT, Porter's 5 Forces | Cloud VP, 비용구조, 전략 |
| **pm-research** | Persona, Competitive Analysis, TAM/SAM/SOM, Customer Journey Map | 3 Personas, 5 Competitors, Market Sizing |
| **pm-prd** | ICP, Beachhead (Geoffrey Moore), GTM, PRD 8-Section, Pre-mortem | ICP, Battlecards, User Stories, Test Scenarios |

### Framework Credits

- Opportunity Solution Tree: Teresa Torres, *Continuous Discovery Habits*
- JTBD: Clayton Christensen, Tony Ulwick
- Lean Canvas: Ash Maurya, *Running Lean*
- Beachhead Market: Geoffrey Moore, *Crossing the Chasm*
- Porter's 5 Forces: Michael Porter
- PM frameworks integrated from [pm-skills](https://github.com/phuryn/pm-skills) by Pawel Huryn (MIT License)

### Market Data Sources

- Secrets Management Market: KBV Research ($10.09B by 2032), Mordor Intelligence ($4.22B 2025)
- Secret Management Software: Verified Market Reports ($5.6B by 2033)
- Encryption Key Management: Business Research Insights ($12.41B by 2034)
- AI Secret Sprawl: GitGuardian State of Secrets Sprawl 2026
- Competitor Pricing: CyberSecTool 2026 Pricing Comparison

### Upstream Documents

| Document | Path | Status |
|----------|------|--------|
| CLI MVP PRD | `docs/00-pm/tene.prd.md` | Complete |
| CLI Discovery | `docs/00-pm/tene-discovery.md` | v4 Complete |
| CLI Strategy | `docs/00-pm/tene-strategy.md` | v4 Complete |
| CLI Research | `docs/00-pm/tene-research.md` | v4 Complete |
| **Cloud Plan** | `docs/01-plan/features/tene-cloud.plan.md` | **v1.0 Complete** |

---

*PRD synthesized by PM Agent Team (pm-lead orchestrator)*
*Date: 2026-04-07 | Feature: tene-cloud | Total Frameworks: 12+*
*Architecture: Go CLI (Free) + Cloud SaaS (Solo $5/mo, Team $10/user/mo)*
*Security Model: Zero-Knowledge Envelope + X25519 ECDH*
