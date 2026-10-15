# 비즈니스 로직 명세서 (Business Logic Specification)

본 문서는 **Cerberus Procurement** 솔루션의 핵심 운영 로직을 정의합니다. 모든 시스템 구현은 본 명세의 로직을 최우선으로 따릅니다.

## 1. 물류 흐름 및 적재 관리 (Logistics Flow)

부킹부터 최종 입고까지의 물리적 상태 변화를 전역 UUID를 기반으로 정밀하게 추적합니다.

*   **N:N:N 아이템 매핑 및 추적**: 하나의 컨테이너 내에 여러 PO와 여러 CI가 혼적되는 상황을 `Container_Item`을 통해 SKU 단위로 관리합니다. 특정 PO의 아이템이 어느 CI로 청구되어 어떤 컨테이너/BL에 실렸는지 전 구간을 추적합니다. `Container_Item`에는 임시 ETA를 기입할 수 있으며, BL과 연결되기 전에도 임시 ETA를 통해 구간별 반입예상수량을 가늠해 볼 수 있습니다.
*   **선적 및 항해 정보 관리 (BL)**: Vessel Name, ETD, ETA 등 항해 관련 정보는 `BL` 테이블에서 통합 관리하며, 하나의 BL에 속한 여러 컨테이너는 동일한 항해 스케줄을 공유합니다.
*   **입고(GR) 및 로트 분할 (Unpacking)**: 컨테이너 입고 시 GR을 생성후, 실재고의 최소 단위인 '로트(Lot)'로 분할 저장합니다. GR을 대상으로 매입채무(AP)를 발행할 수 있으며, 따라서 같은 날에 입고된 수량이라고 해도 AP가 발행되는 GR과 발행되지 않는 GR로 나누어 GR을 작성해야 합니다.(재포장 작업이나 스티커 부착 작업이 필요한 물품이 동일한 날 입고된 경우, 해당 작업을 하는 물품은 GR을 분리하여 작성해야 함) 동일 SKU라도 유통기한, 생산 차수(Lot No) 등에 따라 별도의 `Inventory_Lot`을 생성하여 선입선출(FIFO) 및 로트별 원가 귀속의 기초로 삼습니다.
*   **Container_Item CBM 스냅샷 정책**: `Container_Item.Cbm`은 선적(생성) 시점에 `Item_Master.CBM × Load_Qty`로 스냅샷됩니다. 이후 `Item_Master`의 CBM이 변경되어도 이미 생성된 `Container_Item`에는 영향이 없습니다. 기존 `Container_Item`을 수정할 때 새 CBM(`Item_Master.CBM × Load_Qty`)이 현재 저장값과 다르면, 시스템이 **변경 전/후 CBM을 보여주고 사용자 확인 시에만 업데이트**합니다(거부 시 기존 스냅샷 유지).
*   **전역 UUID 기반 추적**: 모든 주요 문서(PO, CI, BL, Container, Container_Item, GR, Lot)에 부여된 UUID를 통해 시스템 전체에서 데이터의 기원과 흐름을 즉시 식별할 수 있습니다.
*   **PO 품목 = (SKU, 단가) 버킷 (A1 정책)**: 하나의 `PO_Item`은 **동일 SKU + 동일 단가**의 버킷 1개이며, `Unit_Price`는 **불변**입니다.
    *   동일 SKU·동일 단가 추가 → 기존 버킷에 **수량 Delta** 병합
    *   동일 SKU·다른 단가 추가 → **새 PO_Item(새 버킷)** 생성
    *   단가 변경 → 기존 버킷 수량 조정 + 새 버킷 추가로 표현 (in-place 단가 변경 없음)
    *   DB 제약: `UNIQUE(PO_ID, Item_ID, Unit_Price)`
*   **PO 수량 변경 정책 (하이브리드)**: PO와 Container_Item은 분리되어 있어 부분/과·부족 선적이 가능합니다. PO 수량 변경은 참조 유무에 따라 다르게 관리합니다.
    *   **참조 없음(AP 미연결, Container_Item 미존재)**: 입력 실수 수정을 위해 **직접 수정 허용**.
    *   **참조 있음(AP 연결 또는 Container_Item 존재)**: 기존 품목의 **수량/단가 직접 수정·삭제를 차단**하고, 수량은 `PO_Item_Adjustment`의 **증분(Delta)으로만** 변경합니다(사유·작성자·시각 기록). **신규 품목 추가는 허용**됩니다(동일 SKU·단가면 기존 버킷에 병합). 참조를 제거하면 다시 직접 수정이 허용됩니다.
    *   **기준 수량(Base_Qty)**: 최초 등록 수량은 불변으로 보존하고, 현행 수량 = 기준 수량 + Σ수량증분 으로 관리합니다.
    *   **취소(Cancelled) 상태 제거**: 취소는 별도 상태 대신 **음수 수량 증분(수량 0까지 감소)**으로 표현합니다.

### PO_Item 버킷 & 수량 원장 구조 (A1)
PO(주문)는 **여러 번 개정되는 "계획"** 이고, Container_Item은 **단 한 번의 "물리적 선적"** 입니다. 이 둘을 정합성 있게 관리하기 위해 다음 구조를 사용합니다.

#### 1) PO_Item = (SKU, 단가) 버킷 (불변)
- 하나의 `PO_Item`은 **동일 SKU + 동일 단가**의 버킷 1개입니다.
- `Unit_Price`는 **불변**이며, DB 제약 `UNIQUE(PO_ID, Item_ID, Unit_Price)`로 1버킷=1단가를 강제합니다.
- 결과: **같은 SKU라도 단가가 다르면 서로 다른 버킷** → "다중 단가 발주"(예: 공짜 / 50% 할인 / 정가)를 그대로 표현합니다.

#### 2) 발주 수량 = 원장(Delta) + 현재값(캐시)
| 위치 | 역할 |
| :--- | :--- |
| `PO_Item.Base_Qty` | **최초 등록 기준 수량** (불변) |
| `PO_Item_Adjustment` | **수량 증분(Delta) 원장** — `Qty_Delta`, 사유, 작성자, 시각 |
| `PO_Item.PO_Qty` | **현행 수량(캐시/denormalized)** = `Base_Qty + Σ Qty_Delta` |

- **원장(Delta)이 감사·이력의 근거**이고, `PO_Qty`는 조회/배분/상태 판정을 빠르게 하기 위한 **materialized 현재값**입니다.
- 수량 변경은 항상 **증분으로 기록**되며, 이력으로부터 임의 시점의 수량을 재구성할 수 있습니다.

#### 3) 동작 규칙
- **동일 SKU·동일 단가 추가** → 기존 버킷에 **수량 Delta 병합** (새 행 생성 안 함)
- **동일 SKU·다른 단가 추가** → **새 PO_Item(새 버킷)** 생성
- **단가 변경** → in-place 변경 금지. **기존 버킷 수량 조정 + 새 버킷 추가**로 표현
- **취소** → 음수 증분(수량 0까지)

#### 4) 직접수정 vs 증분 (참조 유무에 따른 잠금)
| 참조 상태 | 동작 |
| :--- | :--- |
| **참조 없음** (AP 미연결, Container_Item 미존재) | 입력 실수 수정을 위해 **직접 수정 허용** |
| **참조 있음** (AP 연결 또는 Container_Item 존재) | 기존 품목의 **수량/단가 직접수정·삭제 차단**, 수량은 **증분(Delta)로만**. **신규 품목 추가는 허용**(동일 SKU·단가면 병합). 참조 제거 시 다시 직접수정 허용 |

#### 5) Container_Item 연결
- `Container_Item`은 **단가 버킷(`PO_Item_ID`)** 에 연결됩니다.
- 따라서 선적 단위가 **어떤 (SKU, 단가)인지 명확** → 시점에 따른 소급 왜곡이 없고, 별도 단가 스냅샷이 불필요합니다.
- 선적 수량은 컨테이너 단위로 누적(`SUM(Load_Qty)`)되어 버킷의 주문 수량과 비교됩니다. 상태는 `Not Shipped` / `Partially Shipped` / `Shipped` / **`Over Shipped`** 로 산출됩니다.
  - `Over Shipped`: 누적 선적량이 주문 수량을 **초과**(`SUM(Load_Qty) > PO_Qty`)한 경우입니다. 과다 선적은 차단하지 않으며, 초과 상태를 표기해 가시화합니다. (취소로 `PO_Qty`가 0이 되었는데 이미 선적된 버킷도 `Over Shipped`로 표기됩니다.)
  - 완납 판정은 `Shipped` + `Over Shipped` 모두 포함하므로, 과다 선적 품목이 있어도 PO는 정상적으로 `Closed` 처리됩니다.
- **파생값은 애플리케이션(유스케이스) 레이어에서 계산합니다.** `PO_Item.Status`, `Container_Item`의 중량/부피, `Container` 총계(Total_*)는 `Container_Item`/`PO_Item` 쓰기 시 유스케이스가 재계산합니다(SQL 트리거 미사용). 서버(SQLite)와 WASM(메모리) 두 런타임이 동일한 로직을 공유합니다.

#### 6) Cost Allocation과의 관계 (Value 기준)
- **Value = (SKU, 단가) 버킷의 금액(수량 × 단가) 기준**으로 배분합니다.
- 같은 SKU라도 단가가 다르면 **별개 버킷**이므로 **단가별 차등 배분**이 정확히 동작합니다.
- 예) 같은 SKU를 0원 / 50% / 100%로 발주·선적 시, 보험료(Value 배분)는 **버킷 금액 비중대로 차등** 부담합니다. **동일 취급(균등)** 을 원하면 `Value`가 아니라 `Quantity/Weight/Volume` 기준을 사용합니다.
- `Value`는 "금액 비례(차등)", `Quantity/Weight/Volume`은 "물리량 비례(동일 SKU 동일 취급)"라는 구분을 명확히 합니다.

#### 요약 다이어그램
```
[PO_Item]  (SKU, 단가) 버킷  — Unit_Price 불변
   │  Base_Qty (최초)          UNIQUE(PO_ID, Item_ID, Unit_Price)
   │  PO_Qty   (현행 = Base + ΣDelta)  ← 캐시
   │
   ├── [PO_Item_Adjustment]  수량 Delta 원장 (+/- , 사유/작성자/시각)
   │
   └── [Container_Item]  PO_Item_ID 로 버킷 참조 (선적; (SKU, 단가) 명확)
```

## 2. 랜딩 코스트 및 원가 산출 (Landing Cost)

물품 대금 외의 모든 부대비용을 5대 기준에 따라 로트 단위로 안분하여 정밀한 원가(Landed Cost)를 산출합니다.

### 5대 비용 배분(Allocation) 기준
| 기준 | 설명 | 적용 예시 |
| :--- | :--- | :--- |
| **금액 (Value)** | 가액 비중(Unit Price * Qty)에 따라 배분 | 관세, 보험료 |
| **수량 (Qty)** | 아이템 개수 비중에 따라 배분 | 검사비, 하역비 |
| **부피 (Volume)** | CBM 비중에 따라 배분 | 해상 운임, 보관료 |
| **중량 (Weight)** | 총중량(Gross Weight) 비중에 따라 배분 | 항공 운임, 국내송료 |
| **행 (Row)** | 대상 로트수로 균등 배분 (1/N) | 서류비, 신고수수료 |

### 일괄 배분 및 증빙 로직 (Batch Allocation)
*   **배분 증빙(Cost_Allocation)**: 유저가 하나 또는 그 이상의 부대비용(AP)을 선택하여 일괄 배분할 때 하나의 '배분 이벤트' 헤더가 생성됩니다. 여기에는 배분 시점의 환율, 배분 기준, 총액이 기록되어 원가 산출의 명확한 증빙이 됩니다.
*   **상세 귀속(Cost_Allocation_Item)**: 배분 이벤트의 결과로 각 로트에 귀속된 금액은 상세 테이블에 기록되어, 특정 로트의 원가가 어떤 청구서들로부터 구성되었는지 완벽한 Audit Trail을 제공합니다.
*   **비용 상속 및 마감**: 선적 전 발생 비용은 입고 시점에 로트로 자동 상속됩니다. 인코텀즈별 필수 비용이 모두 입력되어야 원가 확정(Lock)이 가능하며, 확정된 단가는 `Inventory_Lot.Landed_Cost_Per_Unit`에 스냅샷으로 기록됩니다.

### 일반 배분 vs 추후 배분 (Normal vs Late Cost Allocation)
*   **일반 배분 (Normal Cost Allocation)**: 한 로트(Lot)는 기본적으로 한 번만 배분 대상이 될 수 있습니다. 배분 이력이 없는 로트만 선택 가능하며, 선택된 로트들에 대해 AP 금액을 일괄 배분합니다.
*   **추후 배분 (Late Cost Allocation)**: 이미 배분 이력이 있는 로트도 추가로 선택할 수 있습니다. 예를 들어 선적 후 발생한 추가 비용(관세 추징금, 창고 보관료 등)을 기존에 배분 완료한 로트에 다시 귀속해야 할 때 사용합니다. 추후 배분에서는 이미 배분된 AP 잔액을 기준으로만 추가 배분이 이루어집니다.

### 배분 금액 산출 로직 (Allocation Amount Calculation)
선택된 로트들에 대한 AP 배분 금액은 **로트별 공정 몫(Fair Share)** 방식으로 산출됩니다.

```
fair_i         = AP.Local_Amount × LotBase_i / totalRefBase     (각 로트의 공정 몫)
need_i         = fair_i - alreadyAllocated_i                    (기배분 금액 제외, 음수면 0)
scale          = min(1, remainingAP / Σ need_i)                 (선택분 총 몫이 잔액보다 크면 축소)
Lot 배분 금액_i = need_i × scale
```

*   `AP.Local_Amount`: 배분 대상 AP의 현지 통화 금액
*   `totalRefBase`: AP가 참조하는 문서(PO, CI, BL, Container, Container_Item, GR, Lot)의 전체 기준량(Quantity, Weight, Volume, Value, Row)
*   `LotBase_i`: 분배 대상 로트의 기준량
*   `alreadyAllocated_i`: 해당 로트에 이미 배분된(수동 포함) 금액
*   `remainingAP`: `AP.Local_Amount - Σ alreadyAllocated_i` (기배분 총액 제외 후 남은 금액)

이 로직은 다음을 보장합니다.

1. **이미 배분된 금액은 고정**: 각 로트의 기배분 금액(`alreadyAllocated_i`)을 공정 몫에서 제외하므로, 기존 배분 금액은 재계산되지 않고 보호됩니다. 수동으로 부분 배분(공정 몫보다 적게)한 경우에도 해당 금액만 차감하고 나머지 로트로 몰지 않으므로, **AP의 분배 비율이 어그러지지 않습니다**.
2. **미선적/미입고 항목 제외**: `totalRefBase`에는 아직 선택되지 않았거나 미입고(unlanded)인 품목의 기준량도 포함되므로, 해당 품목의 몫만큼 자동 제외됩니다.
3. **취소(Cancelled) 품목 제외**: 기준량 집계 시 `Status = 'Cancelled'`인 PO 품목은 제외합니다. PO 표시 금액(Total Amount)이 취소 품목을 제외하고 계산되는 것과 일관성을 유지하기 위함입니다.
4. **PO 변경 시 비율 재적용**: PO 수량이나 단가가 변경되면 `totalRefBase`가 달라지므로, 각 로트의 공정 몫이 변경된 비율로 재계산됩니다. 기배분 금액은 고정된 채 남은 몫만 새 비율로 재분배됩니다.

### AP Target Group 배분 (Group AP Allocation)
여러 개의 PO(또는 CI, BL, Container 등 동일 종류 문서)를 하나로 묶은 **AP Target Group**에 부과된 AP도 동일한 로직으로 배분됩니다.

*   **매칭**: 그룹 AP는 `AP_Target_Group_Item`의 멤버 UUID 목록을 통해 대상 로트와 연결됩니다. 선택된 로트가 그룹 내 어떤 멤버 문서에 속해 있으면 배분 대상이 됩니다.
*   **기준량 집계**: 배분 시 그룹에 속한 **모든 멤버의 기준량(totalRefBase)을 합산**합니다. 멤버 PO에 미선적/미입고 품목이 포함되어 있다면 그 몫만큼 자동 제외됩니다.
*   **2단계 분배와의 동치성**: "그룹 금액 → 기저 PO별 배분 → PO에 입고(GR_Lot)된 부분은 다시 PO_Item/Lot별 배분" 하는 2단계 방식은, 최종적으로 **각 Lot 기준량 / (그룹 전체 기준량 합계)** 비율로 배분하는 것과 수학적으로 동일합니다. 따라서 구현은 단일 공식으로 처리하되, 기저 PO(또는 그룹 기준문서)별로 AP 금액을 안분한 뒤 입고된 Lot에 귀속되는 것과 같은 결과를 산출합니다.
*   **PO 변경 시**: 그룹 내 일부 PO가 변경되면 해당 PO의 기준량이 변하고 그룹 전체 `totalRefBase`가 변하므로, 남은 금액은 변경된 비율로 재분배됩니다. 이미 배분된 금액은 고정됩니다.

### PO 변경 시 경고 (PO Modification Guardrail)
PO의 품목(PO_Item)을 수정하거나 삭제하면 해당 PO에 연결된 로트의 기준량 비율이 변경될 수 있으므로, 기존 Cost Allocation의 비율이 물리적으로 틀어질 위험이 있습니다. 따라서 이미 Cost Allocation에 사용된 PO를 수정할 때는 시스템이 다음 정보를 포함한 경고를 표시합니다.

*   배분 이력이 있는 로트 수
*   해당 PO를 통해 배분된 총 금액
*   "기존 배분 비율에 영향을 줄 수 있으므로 주의해서 진행" 안내

해당 경고는 저장을 차단하지 않고 사용자의 확인(Confirm)을 요구합니다. 사용자는 변경의 영향을 인지한 상태에서 PO를 수정할 수 있습니다.

### 과다 배분 상계: 마이너스 AP(크레딧 노트) (Credit Note Offset)
수동 배분 과정에서 한 AP에 대해 공정 몫보다 **과다한 금액이 배분**된 경우, 과다분을 되돌리기 위해 **마이너스(-) 금액의 AP(크레딧 노트)**를 발행하는 방식으로 상계합니다.

1. **발행**: 과다 배분된 AP와 **동일한 참조**(GR/PO/CI 등)에 `Local_Amount < 0`인 AP를 발행합니다 (예: -20,000).
2. **상계 배분**: 해당 참조의 LOT에 크레딧 AP를 배분합니다.
   *   자동 제안: 음수 AP에 대해 시스템은 각 로트에 음수 금액(`AP.Local_Amount × LotBase / totalRefBase`)을 비례 제안합니다.
   *   수동 상계: 과다 배분된 특정 LOT에 크레딧 전액(예: -20,000)을 직접 입력하여 정확히 상계할 수 있습니다.
3. **결과**: LOT의 순배분 = 기존 배분 + 크레딧. 예: 120,000 과다 배분 + (-20,000) 크레딧 = 100,000 (정상 복원).

크레딧 노트는 일반 배분과 동일한 공정 몫(음수) 로직을 사용하므로, 상계 후에도 해당 AP 분배의 비율이 유지되고 기배분 금액 보호 규칙이 그대로 적용됩니다.

## 3. 수입 현금흐름 및 채무 관리 (Cash Flow)

다양한 출처의 비용을 단일 채무 창구(`Account_Payable`)로 집결시켜 통합 관리합니다.

*   **다형성 참조 (Polymorphic Reference)**: `Account_Payable`은 `Reference_UUID`와 `Reference_Type`을 통해 PO, CI, BL, Container, Container_Item, GR, Lot 등 다양한 원천 문서와 유연하게 연결됩니다. 참조 문서의 UUID는 **수량이 변경되어도 유지**됩니다. Cost Allocation은 `Cost_Allocation_Item`이 로트(Lot_ID) 기준으로 금액을 스냅샷하고, 이후 배분은 현재 기준량으로 공정 몫을 재계산하는 구조이므로, 수량 변경으로 인해 기존 배분이 깨지지 않습니다.
*   **삭제 시 AP 고아 방지**: AP가 참조하고 있는 문서(PO, CI, BL, Container, Container_Item, GR, Lot, AP Target Group)는 **이를 참조하는 AP가 하나라도 있으면 삭제할 수 없습니다.** 시스템이 삭제 시 연결된 AP 목록을 표시하고 삭제를 차단합니다.
*   **AP 타겟 그룹 (AP Target Group)**: 여러 PO, 여러 CI, 여러 BL 등 **동일 종류의 원천 문서**를 하나의 그룹 UUID로 묶어 1회 청구형 AP를 발행할 수 있습니다. 그룹은 `AP_Target_Group`과 `AP_Target_Group_Item`으로 관리되며, `AP_Target_Group.Reference_Type`에 의해 그룹이 대표하는 문서 종류가 결정됩니다(PO, CI, BL, Container, Container_Item, GR, Lot). 발행된 AP는 그룹의 UUID를 `Reference_UUID`로, `Reference_Type`을 `AP_Target_Group`으로 참조합니다. 그룹 내 개별 문서는 `AP_Target_Group_Item.Reference_UUID`에 기록되어 단순 참조 목록(Audit Trail)을 제공합니다.
*   **다국어 및 환율 처리**: 모든 채무는 발생 통화(Currency)와 함께 배분/결제 시점의 환율을 적용한 `Local_Amount`(현지 통화 금액)를 병행 기록하여 정확한 외환 차손익 및 원가 계산을 지원합니다.
*   **배분 완료 AP 금액 잠금**: Cost Allocation에 이미 사용된(배분 이력이 있는) AP는 `Amount`/`Local_Amount`를 **수정할 수 없습니다.** 배분 후 금액이 바뀌면 배분 비율과 원가 증빙이 깨지기 때문입니다. 기타 필드(상태, 납기일, 비고 등)는 수정 가능하며, 금액 수정이 필요하면 크레딧 노트(마이너스 AP)로 상계합니다.
*   **확정 데이터 중심 운영**: 예측치가 아닌, 유저가 수동으로 확정한 전표(PO 선금, CI 잔액, 부대비용 청구서)만을 공식 현금흐름에 반영합니다.
*   **마이너스 전표 기반 크레딧 노트**: 부족분(Shortage)이나 파손에 대한 클레임은 마이너스(-) 금액의 AP 전표를 발행하여 PO또는 CI 앞으로 발행된 AP에 대한 채무액을 자동 상쇄하는 방식으로 처리합니다.

## 4. 행정 절차 제약(Constraint) 체인과 원가배분 가능일 (Cost Allocation Readiness)

입고(GR) 직후에는 곧바로 **Cost Allocation(원가 확정)이 가능하지 않습니다.** 요건승인, 검역, 통관 등 **행정적 절차**를 거쳐야 하며, 이 절차들은 서로 **연쇄(체인)** 되어 앞 절차가 끝나야 다음 절차가 시작됩니다. 본 절은 이 행정 절차를 **제약(Constraint)** 으로 모델링하고, 체이닝을 통해 **최종 Cost Allocation 가능일(= 재고 처분 시점)** 을 자동 산출하는 로직을 정의합니다.

### 4.1 개념

*   **제약(Constraint)**: 하나의 행정 절차(예: 요건승인, 검역, 통관). 각 제약은 **시작일(Start)**, **리드타임(Lead Time)**, **완료일(End)** 을 가집니다.
*   **리드타임(Lead Time)**: 해당 절차에 통상 소요되는 기간. 완료일이 아직 확정되지 않았을 때 **완료일 = 시작일 + 리드타임** 으로 **예상 완료일**을 산출합니다.
*   **체이닝(Chaining)**: 제약을 순서대로 연결(전 제약 → 후 제약). 관계는 **Finish-to-Start** 입니다(선행이 끝나야 후속이 시작).
*   **앵커(Anchor)**: 체인의 첫 제약 시작일을 확정해 주는 물류 이벤트(예: BL ETA, 입고일(GR), 도착예정일).
*   **Cost Allocation 가능일(= 재고 처분 시점)**: 체인의 **마지막 제약의 완료일**. 이 시점에 도달해야 해당 로트의 원가 확정(원가배분)과 처분이 가능합니다.

### 4.2 제약(Constraint) 데이터 모델 (제안)

| 필드 | 설명 |
| :--- | :--- |
| `Constraint_ID` | PK |
| `Constraint_Type` / `Name` | 절차 유형(요건승인 / 검역 / 통관 / …) |
| `Scope_Type`, `Scope_UUID` | 이 제약이 걸린 대상(GR, Lot, Container, BL 등) |
| `Predecessor_ID` | 선행 제약(체인 연결). 없으면 체인 시작 |
| `Sequence` | 체인 내 순서(표시/정렬용) |
| `Lead_Time_Days` | 소요기간(기본 역일, 옵션: 영업일) |
| `Anchor_Type`, `Anchor_UUID` | (체인 시작 제약) 시작일을 확정하는 물류 이벤트 |
| `Planned_Start`, `Actual_Start` | 계획 / 실제 시작일 |
| `Planned_End`, `Actual_End` | 계획 / 실제 완료일 |
| `Status` | `Planned` / `In Progress` / `Done` / `Waived` |

> 규칙: **실제값(`Actual_*`)이 있으면 우선**, 없으면 **파생값**으로 계산합니다.

### 4.3 체이닝 규칙 (Finish-to-Start)

각 제약 `i`에 대해:

```
start_i = Actual_Start_i            (확정된 실제 시작일이 있으면 우선)
        = anchor_i                  (체인 시작 제약이고 앵커가 있으면 그 일자)
        = end_{i-1}                 (그 외: 선행 제약의 완료일)

end_i   = Actual_End_i              (확정된 실제 완료일이 있으면 우선)
        = start_i + Lead_Time_i     (미확정: 시작일 + 리드타임 = 예상 완료일)

Cost_Allocation_가능일 = end_last   (체인의 마지막 제약 완료일)
```

핵심 동작:

1.  **1차 제약의 시작일이 확정되면**, 2차 제약의 시작일은 **1차 리드타임을 더한 1차의 예상 완료일**로 자동 확정됩니다(`start_2 = start_1 + Lead_1`). 3차·4차도 같은 방식으로 연쇄됩니다.
2.  **1차 제약의 완료일이 실제로 확정되면**(`Actual_End`), 후속 제약의 시작일이 그 확정일로 **재설정**되고, 이후 체인 전체가 **자동 재계산(cascade)** 됩니다.
3.  예상 완료일과 실제가 달라지면(빠름/늦음) 그 차이가 **후속 제약 전체와 최종 가능일에 전파**됩니다.

### 4.4 병렬 제약 / 분기

*   한 제약이 **여러 선행 제약**을 가질 수 있습니다(합류). 이때 시작일은 선행 완료일의 **최댓값**입니다(모두 완료되어야 시작):
    `start_i = max(end_p) for p in predecessors`
*   한 제약이 **여러 후속 제약**으로 분기할 수 있습니다(병렬 진행). 이 경우 최종 가능일은 모든 분기 종료일의 **최댓값**입니다:
    `Cost_Allocation_가능일 = max(end of all terminal constraints)`

### 4.5 Cost Allocation 가능일과 재고 처분

*   **재고 처분 시점 = `Cost_Allocation_가능일`**. 이 시점에 도달하기 전에는 해당 로트를 **"행정 절차 진행 중"** 으로 간주하여 Cost Allocation(원가 확정 / Lock)을 **보류**합니다.
*   도달 후에는 §2의 "비용 상속 및 마감" 규칙에 따라 로트 원가를 확정(`Inventory_Lot.Landed_Cost_Per_Unit` 스냅샷)하고 재고를 처분할 수 있습니다.
*   동일한 GR/체인을 공유하는 여러 로트는 **동일한 가능일**을 가집니다.
*   **게이트(구현)**: Cost Allocation **로트 선택 목록에서 가능일 미도달 로트를 제외(Hard)** 하고, 배분 계산(`calculate`) 시에도 미도달 로트가 포함되면 **차단**합니다. (제약 없음 / 전 단계 실제완료 / 가능일 ≤ 오늘 이면 통과)

### 4.6 예시

앵커(입고일) = `2026-02-01`:

| 순서 | 제약 | 리드타임 | 시작일 | 완료일(예상) |
| :--- | :--- | :--- | :--- | :--- |
| 1 | 요건승인 | 3일 | 02-01 | 02-04 |
| 2 | 검역 | 2일 | 02-04 | 02-06 |
| 3 | 통관 | 4일 | 02-06 | 02-10 |

→ **Cost Allocation 가능일 = `2026-02-10`**

*   검역 완료가 예상보다 **1일 빨리 확정**(02-05)되면: 통관 시작 02-05 → 완료 02-09 → **가능일 02-09**.
*   통관 완료가 **2일 지연 확정**(02-12)되면: **가능일 02-12**.

### 4.7 엣지 케이스

*   **절차 면제(Waived)**: 리드타임 0으로 두거나 체인에서 제외하고 후속 시작을 선행 완료일에 연결합니다.
*   **리드타임 단위**: 기본 **역일(calendar day)**, 필요 시 **영업일(business day)** 옵션.
*   **제약 추가 / 삭제 / 순서 변경**: 체인을 재계산합니다.
*   **실제 시작일만 확정, 완료 미확정**: `end = Actual_Start + Lead_Time`.
*   **앵커 미확정**: 체인 시작이 미정이면 전체 가능일도 **미정(pending)** 으로 표기합니다.

### 4.8 기존 로직과의 관계

*   §2 "비용 상속 및 마감(원가 확정 Lock)"의 **선행 조건**으로 작동합니다. 즉, 인코텀즈 필수 비용이 모두 입력되었더라도 **행정 제약 체인이 완료(Cost Allocation 가능일 도달)** 되어야 원가 확정이 가능합니다.
*   §3의 AP/비용 구조와는 독립적으로, **로트의 처분 가능 시점**을 결정하는 축입니다.

### 4.9 요약 다이어그램

```
[물류 이벤트: 입고(GR) / BL ETA]  ← 앵커
        │
        ▼
[Constraint 1: 요건승인]  start₁ ──(lead₁)──▶ end₁
        │  start₂ = end₁
        ▼
[Constraint 2: 검역]      start₂ ──(lead₂)──▶ end₂
        │  start₃ = end₂
        ▼
[Constraint 3: 통관]      start₃ ──(lead₃)──▶ end₃
        │
        ▼
Cost Allocation 가능일 = end₃  (재고 처분 시점)
   → 이후 로트 원가 확정(Landed_Cost_Per_Unit) · 처분 가능
```

## 5. 행정 제약 프로파일과 인스턴스 체인 (Constraint Profile & Instance Chain)

§4가 "제약 체인과 가능일 계산 규칙"을 정의했다면, §5는 이를 **운영 가능한 형태**로 구체화합니다 — **룰(프로파일)**, **선적 시 선택**, **GR 시 인스턴스화**, **cascade**, **체인 편집(삽입/삭제/재정렬)**.

> **스코프 경계**: 이 절은 **예상 가능일 산출 + 제약 진행 추적**에 한정합니다. **Blanket/릴리즈·실측 이벤트 로그·SKU별 리드타임 통계**는 **범위에서 제외**합니다(§5.11).

### 5.1 앵커와 시간축

*   입고(GR) 이후에만 행정 절차가 시작되므로, **체인의 앵커는 "도착/입고 시점"** 입니다.
*   앵커는 문서 진행에 따라 **진화**하며, 동일 공식으로 재계산됩니다:
    `Temporary_ETA → BL.ETA → GR.Receive_Date`
*   즉, 입고 前에는 `Temporary_ETA`(Container_Item)로 **투영**, 입고 後에는 `GR.Receive_Date`로 **실측 기준**을 삼습니다.

### 5.2 룰 계층: Item_Master → Constraint_Profile

적용할 행정 절차 조합은 **SKU 속성**이므로 Item_Master에 둡니다. 단, SKU마다 직접 나열하지 않고 **프로파일(템플릿) 참조**로 재사용합니다.

| 테이블 | 필드(요지) | 역할 |
| :--- | :--- | :--- |
| `Constraint_Profile` | Profile_ID, Name, Remark | 제약 조합 정의(템플릿). 예: 식품 표준 / 일반 / 검역만 |
| `Constraint_Profile_Step` | Profile_ID, Sequence, Constraint_Type, Lead_Time_Days, Is_Optional | 프로파일의 단계 목록. `Is_Optional`은 필요 시 enable |
| `Item_Master` | `Default_Constraint_Profile_ID` (FK) | **SKU별 기본 제약 프로파일** |

*   **Item Master 편집 UI에 "행정 제약 프로파일" 설정 기능**을 추가합니다(기본값).
*   프로파일은 카테고리 단위로 재사용 → SKU가 많아도 중복 정의가 없습니다.

### 5.3 Container_Item: 기본값 + 오버라이드

| 필드 | 역할 |
| :--- | :--- |
| `Container_Item.Constraint_Profile_ID` (FK) | 입력 시 **SKU의 기본 프로파일로 자동 채움**, 유저가 **다른 템플릿으로 변경 가능** |

*   선적 시점에 SKU를 아니까 **SKU별 체인 문제가 기본값으로 자동 해결**됩니다.
*   예외(특정 선적 한정)는 **템플릿 선택**으로 처리합니다.
*   `Container_Item`은 **프로파일(계획)만 소유**하며 **제약 인스턴스는 갖지 않습니다.** 인스턴스는 GR/Lot 생성 시 **Lot으로 클론**됩니다(§5.5).

### 5.4 예상 Cost Allocation 가능일 (투영)

입력된 `Temporary_ETA`를 앵커로, 선택된 프로파일의 단계를 **정방향 합산**합니다(저장 불필요, 조회 시 계산).

```
예상 가능일 = Temporary_ETA
            + 도착→입고 리드
            + Σ (선택된 프로파일 단계의 Lead_Time_Days)
```

*   SKU 기본이면 유저 개입 없이 즉시 표시됩니다.
*   템플릿 변경 시 즉시 재계산합니다.
*   `BL.ETA`/`GR일`이 확정되면 **앵커만 교체**해 재계산합니다.

### 5.5 GR: Lot별 인스턴스 클론

GR(입고)/Lot 생성 시점에, 해당 Lot이 속한 `Container_Item`의 **프로파일을 Lot별로 복제(클론)** 하여 인스턴스화하고, 앵커를 `GR.Receive_Date`로 설정합니다.

*   **소유**: `Container_Item`은 **프로파일만**(계획) 보유하고 **인스턴스는 갖지 않습니다.** 인스턴스는 **`Lot`이 소유**합니다.
*   **클론**: 한 Container_Item이 여러 Lot으로 분할되면 **Lot마다 독립적인 사본**을 가집니다 → BL 분할·부분 통관·Lot별 반출 등 **Lot별 독립 절차**를 자연스럽게 표현합니다.
*   **출처 기록**: 클론 인스턴스에 `Source_Container_Item_ID`, `Source_Profile_ID`를 남겨 추적성을 확보합니다.
*   **스냅샷**: GR 시점 프로파일을 반영 → 이후 템플릿이 바뀌어도 기존 Lot 인스턴스는 불변(안정·감사).
*   **Optional 단계**: 프로파일의 `Is_Optional` 단계는 필요 시 enable 합니다.
*   **주의(선적 前 단계)**: 이 모델은 앵커를 GR 이후로 봅니다. 선적 前에 시작되는 절차(사전 요건승인, 선적 전 검사 등)를 추적하려면, 그 단계만 상위 레벨(Container_Item 또는 그 이전)에 인스턴스화하는 **혼합 모델**로 확장합니다.

### 5.6 인스턴스 모델 (`Predecessor_ID` 기반)

체인은 정수 나열이 아니라 **연결 리스트(선행 참조)** 로 표현 → 삽입/삭제/재정렬이 용이합니다.

| 필드 | 설명 |
| :--- | :--- |
| `Instance_ID` (PK) | |
| `Lot_ID` (FK) | **소유 단위**(GR/Lot 생성 시 클론) |
| `Source_Container_Item_ID` (FK) | 클론 출처(선적) — 일괄 편집·추적용 |
| `Profile_ID` | 클론 출처 프로파일(스냅샷) |
| `Predecessor_ID` (FK, nullable) | **선행 제약**(체인 연결). null이면 체인 시작 |
| `Sequence` | 표시/정렬용(여유 간격 10·20·30 또는 재부여) |
| `Constraint_Type` / `Name` | 요건승인 / 검역 / 통관 / 불시검문검사 / … |
| `Lead_Time_Days` | 소요기간 |
| `Planned_Start`, `Actual_Start` | 계획 / 실제 시작일 |
| `Planned_End`, `Actual_End` | 계획 / 실제 완료일 |
| `Status` | `Planned` / `InProgress` / `Done` / `Waived` |
| audit | Created_By/At, Updated_By/At |

### 5.7 cascade 규칙 (§4.3과 동일)

```
start_i = Actual_Start_i ?? (선행 없음 ? GR.Receive_Date(또는 앵커) : end_predecessor)
end_i   = Actual_End_i   ?? start_i + Lead_Time_i
가능일   = end of terminal   (병렬이면 max)
```

*   **시작일만 입력** → `end = start + lead` → 후속 시작 확정 → 연쇄.
*   **완료일 입력** → 후속 시작 = 그 완료일 → 연쇄.
*   **실제값이 들어온 지점부터** 하위만 재계산합니다(상위 구간은 고정).

### 5.8 체인 편집: 삽입 / 삭제 / 재정렬

체인 편집은 **`Predecessor_ID` 재연결(relink)** 로 수행합니다.

#### 5.8.1 중간 삽입 (불시검문검사 등)

```
기존: A ─▶ B ─▶ C
삽입: A ─▶ B ─▶ [D] ─▶ C
   D.Predecessor = B
   C.Predecessor = D      (B의 후속을 D로 교체)
→ 삽입 지점부터 cascade: D.start = B.end, C.start = D.end → 가능일 갱신
```

*   **서브체인 삽입**: `A ─▶ [S1→S2→S3] ─▶ C` (헤드 S1을 A에, C를 꼬리 S3에 연결).
*   삽입 지점 **이후만** 재계산합니다(앞 구간 불변).

#### 5.8.2 삭제

```
A ─▶ [B] ─▶ C   →   A ─▶ C
   C.Predecessor = A      (B를 건너뛰고 후속을 A에 연결)
```

*   삭제 지점부터 cascade 합니다. (이력이 필요하면 물리 삭제 대신 `Status=Waived`/`Void` 로 보존.)

#### 5.8.3 재정렬

*   순서 교환: 해당 두 노드의 `Predecessor_ID`만 교체 후 **변경 지점부터** cascade.
*   `Sequence`는 표시용으로 재부여해 정렬 일관성을 유지합니다.

#### 5.8.4 병렬 / 합류

*   한 노드가 **여러 후속**(동일 `Predecessor_ID`)을 가질 수 있음 → 병렬 진행.
*   합류 노드의 시작 = 선행 완료들의 **max** (§4.4).

#### 5.8.5 선적 일괄 편집 (Bulk Edit)

Lot별로 클론된 인스턴스는 독립적이므로, 같은 선적의 여러 Lot에 **동일한 변경**을 적용할 때는 **"선적(Container_Item) 전체에 적용"** 일괄 편집을 제공합니다.

*   예: 같은 Container_Item에서 분할된 모든 Lot에 새 단계(불시검문검사)를 추가하거나 리드타임을 조정.
*   균일한 다수 Lot의 **중복 편집 부담을 줄이되, Lot별 독립성은 유지**(일괄 적용 후 개별 Lot에서 다시 수정 가능).
*   대상 Lot은 출처(`Source_Container_Item_ID`)를 기준으로 선택합니다.

### 5.9 돌발 단계 처리

*   **템플릿 optional enable**: 프로파일에 `Is_Optional` 단계(예: 불시검문검사)를 미리 열어두고 필요 시 enable 합니다.
*   **애드혹 삽입**: 카탈로그(타입 + 기본 리드타임)에서 고르거나, 이름/리드타임을 직접 입력해 **중간 삽입**합니다.
*   두 방식 모두 §5.8.1의 relink로 처리합니다.

### 5.10 감사와 안정성

*   인스턴스는 **audit 컬럼**으로 "누가/언제 단계를 추가·삭제·재정렬"했는지 기록합니다.
*   GR 시 `Profile_ID`를 **스냅샷** → 템플릿 변경이 기존 선적에 소급되지 않습니다.
*   (선택) 삭제는 물리 삭제 대신 `Waived`/`Void` 로 이력을 보존합니다.

### 5.11 범위 제외 (Out of Scope)

아래는 **명시적으로 범위에서 제외**합니다(도입하지 않음).

*   **Blanket Order / 릴리즈 앵커**: 포괄 주문·콜오프(`PO_Release`)·생산 마일스톤(`Production_Start`/`CRD`)은 **고려하지 않습니다.** (Standard PO만 대상)
*   **실측 이벤트 로그**: 모든 마일스톤의 발생/변경을 append-only로 누적하는 로그는 **구현하지 않습니다.**
*   **SKU별 리드타임 통계 / 프로파일 보정**: 이벤트/실적 기반 리드타임 통계와 자동 보정은 **구현하지 않습니다.**

### 5.12 요약 다이어그램

```
[Item_Master] ──Default_Constraint_Profile──▶ [Constraint_Profile(+Step)]   ← 룰(템플릿)
        │
        ▼ (SKU 기본값)
[Container_Item]  Constraint_Profile_ID (유저 오버라이드)
        │  Temporary_ETA ──▶ 예상 가능일 (투영, 계산)
        ▼ (GR)
[Lot별 Constraint_Instance]  (GR 시 프로파일 클론, Lot 소유, 앵커 = GR일)
        │   A ─▶ B ─▶ C        ← Predecessor_ID 연결
        │   중간 삽입/삭제/재정렬 = relink
        │   Actual_Start/End 입력 → cascade
        ▼
Cost Allocation 가능일 (재고 처분 시점)
```
