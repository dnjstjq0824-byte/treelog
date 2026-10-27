# FOLLOW_SUBSCRIPTION.md — 팔로우·이웃·상호이웃·구독 통합 관계 모델 명세 (제8조)

## 1. 4대 관계 개념 비교 및 권한 정의
| 구분 | 관계 코드 | 성립 방식 | 열람 가능 게시글 범위 | 주요 활용 목적 |
| :--- | :--- | :--- | :--- | :--- |
| **1. 팔로우** | `FOLLOW` | 일방 클릭 즉시 성립 (`ACCEPTED`) | `PUBLIC` | 가벼운 관심 표시 및 추천 피드 시그널 반영 |
| **2. 이웃** | `NEIGHBOR` | 일방 추가 즉시 성립 (커스텀 폴더 분류 가능) | `PUBLIC` + **`NEIGHBOR_ONLY` (이웃 공개)** | 네이버 블로그 스타일의 관심 블로그 폴더별 모아보기 |
| **3. 상호이웃 (서로이웃)** | `MUTUAL_NEIGHBOR` | 신청 메시지 전송(`PENDING`) → 상대방 수락(`ACCEPTED`) 시 양방향 성립 | `PUBLIC` + `NEIGHBOR_ONLY` + **`MUTUAL_ONLY` (상호이웃 공개)** | 친밀한 블로거 간 교류 및 프라이빗 일상/깊은 이야기 공유 |
| **4. 구독** | `SUBSCRIPTION` | 알림 채널(인앱/푸시/이메일) 또는 유료 멤버십 결제 계약 | 무료 구독은 `PUBLIC`, 유료 멤버십(`PAID_TIER`)은 **`PAID_MEMBERS_ONLY`** 열람 | 새 글 발행 시 즉각적인 푸시/이메일 뉴스레터 수신 및 유료 후원 |

## 2. 상태 전이 및 차단(`user_blocks`) 연동 규칙
- 사용자 A가 사용자 B를 차단(`POST /api/v1/users/{B}/block`)하는 즉시, A와 B의 블로그 사이에 맺어진 모든 `FOLLOW`, `NEIGHBOR`, `MUTUAL_NEIGHBOR`, 무료 `SUBSCRIPTION` 레코드는 단일 트랜잭션 내에서 `TERMINATED` 상태로 강제 종료되며, B는 A의 `NEIGHBOR_ONLY` / `MUTUAL_ONLY` 게시글에 즉시 접근할 수 없게 된다.
