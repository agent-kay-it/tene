# LemonSqueezy 설정 및 결제 연동 가이드

> **대상**: Tene Cloud Pro $5/month 개인 구독  
> **작성일**: 2026-04-07  
> **관련 문서**: `docs/01-plan/features/tene-cloud.plan.md`, `docs/02-design/features/tene-cloud.design.md`

---

## 1. 도메인 구성 (URL 정리)

| 도메인 | 용도 | 호스팅 |
|--------|------|--------|
| `tene.sh` | 랜딩페이지 | CloudFront + S3 (기존) |
| `www.tene.sh` | 랜딩페이지 리다이렉트 | → tene.sh |
| `app.tene.sh` | 웹 대시보드 | Vercel (신규) |
| `api.tene.sh` | Go API 서버 | ALB + ECS Fargate (신규) |
| **`buy.tene.sh`** | **LemonSqueezy 스토어 (결제 페이지)** | **LemonSqueezy 커스텀 도메인** |

### buy.tene.sh 설정 방법

Route 53에서 A 레코드 추가:

```
Type:  A
Name:  buy.tene.sh
Value: 3.33.255.208
TTL:   300
```

LemonSqueezy 대시보드 > Settings > Domains > `buy.tene.sh` 입력 후 "Verify Domain" 클릭.  
SSL은 LemonSqueezy가 자동 프로비저닝합니다.

---

## 2. 환경 분리 (Staging vs Production)

LemonSqueezy는 **같은 계정 내 Test Mode / Live Mode 토글** 방식입니다.  
별도 Sandbox 계정은 없습니다.

| 항목 | Test Mode (Staging) | Live Mode (Production) |
|------|:-------------------:|:---------------------:|
| API Key | `test_xxx...` | `live_xxx...` |
| Webhook URL | `https://api-staging.tene.sh/...` | `https://api.tene.sh/...` |
| Webhook Secret | 별도 생성 | 별도 생성 |
| Product/Variant | 별도 생성 필요 | 별도 생성 필요 |
| 결제 | 테스트 카드 (4242...) | 실제 카드 |
| Customer/Order | 분리됨 | 분리됨 |

**중요**: Test Mode에서 만든 Product/Variant ID는 Live Mode에서 사용 불가. 양쪽 모두에서 동일한 구성으로 Product를 생성해야 합니다.

---

## 3. LemonSqueezy 설정 단계 (Step by Step)

### Step 1: 계정 가입 및 스토어 생성

1. [app.lemonsqueezy.com](https://app.lemonsqueezy.com)에서 가입
2. Store 설정:
   - **Store Name**: `Tene`
   - **Store Slug**: `tene` → `tene.lemonsqueezy.com`
   - **Custom Domain**: `buy.tene.sh` (Step 6에서 설정)

### Step 2: KYC 인증 (수익 정산 전 필수)

1. Settings > Payouts > 본인 확인
2. 필요 정보:
   - 이름 (영문)
   - 주소 (영문)
   - 생년월일
   - 신분증 사진 (여권 또는 주민등록증 영문면)
3. 개인 사업자 / 개인 모두 가능
4. 심사 기간: 1-3일

### Step 3: Payoneer 연동 (정산 수령)

1. Settings > Payouts > Connect Payoneer
2. Payoneer 계정이 없으면 가입 진행
3. Payoneer에서 한국 은행 계좌 등록 (KRW 출금)
4. **최소 정산**: $50
5. **정산 주기**: 매월 (전월 확정 수익)
6. **Payoneer 출금 수수료**: ~1-2%

### Step 4: Product 생성 (Test Mode에서 먼저)

> **반드시 Test Mode 토글을 켠 상태에서 시작**하세요.

1. Products > "+ New Product"
2. 설정:

| 항목 | 값 |
|------|---|
| Name | `Tene Pro` |
| Description | `Cloud Sync + Team features for Tene CLI` |
| Type | **Subscription** |
| Tax Category | `SaaS` (자동 세금 계산에 사용) |

3. 기본 Variant 가격: **$5.00**, Billing Period: **1 Month**
4. **Variant ID** 기록 (API 호출에 사용)
5. Product > Share에서 **Checkout Link** 확인

### Step 5: API Key 생성

1. Settings > API Keys > "+ New API Key"
2. **Test Mode 상태에서 생성** → Test API Key
3. Key를 안전하게 저장 (생성 시 한 번만 표시)
4. **Live Mode로 전환 후 동일하게 생성** → Live API Key

```
# 결과 예시
Test: lmsq_test_abc123def456...
Live: lmsq_live_xyz789ghi012...
```

### Step 6: Webhook 설정

1. Settings > Webhooks > "+ New Webhook"
2. 설정:

| 환경 | URL | 이벤트 |
|------|-----|--------|
| **Test** | `https://api-staging.tene.sh/api/v1/billing/webhook` | 아래 목록 |
| **Live** | `https://api.tene.sh/api/v1/billing/webhook` | 아래 목록 |

3. **필수 이벤트 선택**:
   - `subscription_created`
   - `subscription_updated`
   - `subscription_cancelled`
   - `subscription_expired`
   - `subscription_paused`
   - `subscription_unpaused`
   - `subscription_payment_success`
   - `subscription_payment_failed`
   - `subscription_payment_recovered`

4. **Signing Secret** 복사 → 서버 환경변수로 저장

> Test/Live 모드 각각 별도 Webhook을 생성해야 합니다.

### Step 7: 커스텀 도메인 설정

1. Settings > Domains > Custom Domain
2. `buy.tene.sh` 입력
3. DNS 레코드 안내에 따라 Route 53에 CNAME 추가
4. 검증 완료 후 SSL 자동 발급
5. 이후 결제 URL: `https://buy.tene.sh/checkout/buy/{variant-uuid}`

### Step 8: Live Mode 동일 설정 반복

Test Mode에서 검증 완료 후:
1. Test Mode 토글 OFF → Live Mode
2. Product "Tene Pro" 재생성 (동일 설정)
3. API Key 재생성 (Live용)
4. Webhook 재생성 (Live URL + Live Secret)
5. Live Variant ID 기록

---

## 4. 환경변수 (Tene 서버)

### 4.1 AWS Secrets Manager 구조

```
tene/{env}/lemonsqueezy
```

| Secret Key | Staging 값 | Production 값 | 설명 |
|-----------|-----------|--------------|------|
| `LEMON_API_KEY` | `lmsq_test_xxx...` | `lmsq_live_xxx...` | API 인증 키 |
| `LEMON_WEBHOOK_SECRET` | `whsec_test_xxx...` | `whsec_live_xxx...` | Webhook HMAC 서명 검증 |
| `LEMON_STORE_ID` | `12345` (Test) | `12345` (동일) | 스토어 ID |
| `LEMON_VARIANT_PRO` | Test Variant ID | Live Variant ID | Pro $5/mo Variant ID |

### 4.2 Go 서버 Config 구조

```go
// internal/config/config.go
type LemonSqueezyConfig struct {
    APIKey        string `envconfig:"LEMON_API_KEY"        required:"true"`
    WebhookSecret string `envconfig:"LEMON_WEBHOOK_SECRET" required:"true"`
    StoreID       string `envconfig:"LEMON_STORE_ID"       required:"true"`
    VariantProID string `envconfig:"LEMON_VARIANT_PRO" required:"true"`
}
```

### 4.3 ECS Task Definition 환경변수

```json
{
  "secrets": [
    {
      "name": "LEMON_API_KEY",
      "valueFrom": "arn:aws:secretsmanager:ap-northeast-2:xxx:secret:tene/{env}/lemonsqueezy:api_key::"
    },
    {
      "name": "LEMON_WEBHOOK_SECRET",
      "valueFrom": "arn:aws:secretsmanager:ap-northeast-2:xxx:secret:tene/{env}/lemonsqueezy:webhook_secret::"
    }
  ],
  "environment": [
    { "name": "LEMON_STORE_ID", "value": "12345" },
    { "name": "LEMON_VARIANT_PRO", "value": "..." }
  ]
}
```

> `LEMON_API_KEY`와 `LEMON_WEBHOOK_SECRET`은 민감 정보이므로 **반드시 Secrets Manager**에서 주입. Store ID와 Variant ID는 비밀이 아니므로 환경변수로 직접 전달 가능.

### 4.4 Terraform Secrets Manager 모듈

```hcl
# infra/terraform/modules/secrets/main.tf 에 추가
resource "aws_secretsmanager_secret" "lemonsqueezy" {
  name = "${var.project}/${var.environment}/lemonsqueezy"
}

resource "aws_secretsmanager_secret_version" "lemonsqueezy" {
  secret_id = aws_secretsmanager_secret.lemonsqueezy.id
  secret_string = jsonencode({
    api_key        = var.lemon_api_key
    webhook_secret = var.lemon_webhook_secret
  })
  lifecycle { ignore_changes = [secret_string] }
}
```

---

## 5. Checkout 연동 (API)

### 5.1 Checkout URL 생성 (서버 → LemonSqueezy API)

```go
// internal/billing/checkout.go

// CreateCheckoutURL creates a LemonSqueezy checkout URL for the given user.
func (s *billingService) CreateCheckoutURL(ctx context.Context, userID string) (string, error) {
    payload := map[string]interface{}{
        "data": map[string]interface{}{
            "type": "checkouts",
            "attributes": map[string]interface{}{
                "checkout_data": map[string]interface{}{
                    "custom": map[string]string{
                        "user_id": userID,
                    },
                },
                "product_options": map[string]interface{}{
                    "redirect_url": s.redirectURL, // https://app.tene.sh/billing?success=true
                },
            },
            "relationships": map[string]interface{}{
                "store": map[string]interface{}{
                    "data": map[string]string{
                        "type": "stores",
                        "id":   s.storeID,
                    },
                },
                "variant": map[string]interface{}{
                    "data": map[string]string{
                        "type": "variants",
                        "id":   s.variantProID,
                    },
                },
            },
        },
    }

    body, _ := json.Marshal(payload)
    req, _ := http.NewRequestWithContext(ctx, "POST",
        "https://api.lemonsqueezy.com/v1/checkouts", bytes.NewReader(body))
    req.Header.Set("Authorization", "Bearer "+s.apiKey)
    req.Header.Set("Content-Type", "application/vnd.api+json")
    req.Header.Set("Accept", "application/vnd.api+json")

    resp, err := s.httpClient.Do(req)
    // ... 에러 처리, JSON 파싱 ...
    // 응답에서 data.attributes.url 추출
    return checkoutURL, nil
}
```

### 5.2 Webhook 수신 및 검증

```go
// internal/billing/webhook.go

func (s *billingService) HandleWebhook(ctx context.Context, payload []byte, signature string) error {
    // 1. HMAC SHA-256 서명 검증
    mac := hmac.New(sha256.New, []byte(s.webhookSecret))
    mac.Write(payload)
    expected := hex.EncodeToString(mac.Sum(nil))
    if !hmac.Equal([]byte(expected), []byte(signature)) {
        return ErrWebhookSignature
    }

    // 2. 이벤트 파싱
    var event LemonEvent
    if err := json.Unmarshal(payload, &event); err != nil {
        return fmt.Errorf("webhook unmarshal: %w", err)
    }

    // 3. 멱등성 확인 (event ID 중복 방지)
    if processed, _ := s.isEventProcessed(ctx, event.Meta.EventName+event.Data.ID); processed {
        return nil
    }

    // 4. 이벤트 처리
    userID := event.Meta.CustomData["user_id"]

    switch event.Meta.EventName {
    case "subscription_created", "subscription_payment_success":
        return s.activatePro(ctx, userID, event.Data.ID)
    case "subscription_cancelled", "subscription_expired":
        return s.deactivatePro(ctx, userID)
    case "subscription_payment_failed":
        return s.handlePaymentFailed(ctx, userID)
    case "subscription_paused":
        return s.pauseSubscription(ctx, userID)
    case "subscription_unpaused", "subscription_payment_recovered":
        return s.activatePro(ctx, userID, event.Data.ID)
    }

    return nil
}

// LemonEvent는 LemonSqueezy Webhook 페이로드 구조체다.
type LemonEvent struct {
    Meta struct {
        EventName  string            `json:"event_name"`
        CustomData map[string]string `json:"custom_data"`
    } `json:"meta"`
    Data struct {
        ID         string `json:"id"`
        Type       string `json:"type"`
        Attributes struct {
            Status        string `json:"status"`
            CustomerID    int    `json:"customer_id"`
            VariantID     int    `json:"variant_id"`
            RenewsAt      string `json:"renews_at"`
            EndsAt        string `json:"ends_at"`
            URLs          struct {
                CustomerPortal string `json:"customer_portal"`
            } `json:"urls"`
        } `json:"attributes"`
    } `json:"data"`
}
```

### 5.3 Customer Portal URL

Subscription 조회 시 응답에 포함되는 `urls.customer_portal`을 저장하여 사용:

```go
func (s *billingService) GetPortalURL(ctx context.Context, userID string) (string, error) {
    var portalURL string
    err := s.db.QueryRowContext(ctx,
        "SELECT lemon_portal_url FROM users WHERE id = $1", userID,
    ).Scan(&portalURL)
    return portalURL, err
}
```

> `subscription_created` Webhook 수신 시 `data.attributes.urls.customer_portal`을 DB에 저장합니다.

---

## 6. CLI 연동 (tene billing)

### 6.1 tene billing

```bash
$ tene billing
  Tene Billing

  Plan:    Pro ($5/month)
  Status:  Active
  Renews:  2026-05-07

  Manage: tene billing portal
```

### 6.2 tene billing upgrade

```bash
$ tene billing upgrade
  Opening checkout page in browser...
  # buy.tene.sh Checkout Overlay 또는 Hosted 페이지 오픈
```

### 6.3 tene billing portal

```bash
$ tene billing portal
  Opening billing portal in browser...
  # LemonSqueezy Customer Portal 오픈
  # 결제 수단 변경, 구독 취소, 인보이스 조회 가능
```

---

## 7. 테스트 방법

### 7.1 로컬 개발 시

```bash
# 1. ngrok으로 로컬 서버 노출
ngrok http 8080

# 2. LemonSqueezy Test Mode Webhook URL에 ngrok URL 등록
#    https://xxxx.ngrok.io/api/v1/billing/webhook

# 3. 테스트 카드로 결제
#    카드번호: 4242 4242 4242 4242
#    만료: 아무 미래 날짜
#    CVC: 아무 3자리

# 4. Webhook 로그 확인
#    LemonSqueezy 대시보드 > Webhooks > Recent Deliveries
```

### 7.2 Staging 환경

```bash
# Test Mode API Key + Test Webhook 사용
# api-staging.tene.sh에 Test Mode Webhook URL 연결
# 테스트 카드로 전체 흐름 검증
```

### 7.3 Production 전환 체크리스트

- [ ] Live Mode Product "Tene Pro" 생성 완료
- [ ] Live API Key 생성 → Secrets Manager 저장
- [ ] Live Webhook 생성 (api.tene.sh URL + Live Secret)
- [ ] Live Variant ID → 환경변수 업데이트
- [ ] KYC 인증 완료
- [ ] Payoneer 연동 완료
- [ ] 실제 카드로 결제 테스트 (즉시 환불)
- [ ] buy.tene.sh 커스텀 도메인 SSL 확인

---

## 8. 요약: 설정해야 하는 것

### LemonSqueezy 대시보드

| 순서 | 작업 | 환경 |
|:----:|------|------|
| 1 | 계정 가입 + 스토어 생성 | 공통 |
| 2 | KYC 인증 | 공통 |
| 3 | Payoneer 연동 | 공통 |
| 4 | Test Mode에서 Product "Tene Pro" $5/month 생성 | Test |
| 5 | Test Mode API Key 생성 | Test |
| 6 | Test Mode Webhook 설정 (staging URL) | Test |
| 7 | 커스텀 도메인 buy.tene.sh 연결 | 공통 |
| 8 | Live Mode에서 Product/API Key/Webhook 동일 생성 | Live |

### Tene 인프라

| 순서 | 작업 | 파일/위치 |
|:----:|------|----------|
| 1 | Route 53에 `buy.tene.sh` CNAME 추가 | `infra/terraform/modules/route53/` |
| 2 | Secrets Manager에 LemonSqueezy 시크릿 저장 | `infra/terraform/modules/secrets/` |
| 3 | ECS Task Definition에 환경변수 추가 | `infra/terraform/modules/ecs/` |
| 4 | `internal/billing/` 패키지 구현 | Go 서버 |
| 5 | Webhook 엔드포인트 구현 | `internal/api/handler/billing.go` |
| 6 | `tene billing` CLI 명령어 구현 | `internal/cli/billing.go` |

### 환경변수 요약

```bash
# Staging (.env.staging 또는 Secrets Manager)
LEMON_API_KEY=lmsq_test_xxx...
LEMON_WEBHOOK_SECRET=whsec_test_xxx...
LEMON_STORE_ID=12345
LEMON_VARIANT_PRO=...           # Test Mode Variant ID

# Production (Secrets Manager)
LEMON_API_KEY=lmsq_live_xxx...
LEMON_WEBHOOK_SECRET=whsec_live_xxx...
LEMON_STORE_ID=12345       # 동일
LEMON_VARIANT_PRO=...           # Live Mode Variant ID
```
