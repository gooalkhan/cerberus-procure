# 비즈니스 로직 명세서 (Business Logic Specification)

본 문서는 **Cerberus Procurement** 솔루션의 핵심 운영 로직을 정의합니다. 모든 시스템 구현은 본 명세의 로직을 최우선으로 따릅니다.

## 1. 물류 흐름 및 적재 관리 (Logistics Flow)

부킹부터 최종 입고까지의 물리적 상태 변화를 전역 UUID를 기반으로 정밀하게 추적합니다.

*   **N:N:N 아이템 매핑 및 추적**: 하나의 컨테이너 내에 여러 PO와 여러 CI가 혼적되는 상황을 `Container_Item`을 통해 SKU 단위로 관리합니다. 특정 PO의 아이템이 어느 CI로 청구되어 어떤 컨테이너/BL에 실렸는지 전 구간을 추적합니다. `Container_Item`에는 임시 ETA를 기입할 수 있으며, BL과 연결되기 전에도 임시 ETA를 통해 구간별 반입예상수량을 가늠해 볼 수 있습니다.
*   **선적 및 항해 정보 관리 (BL)**: Vessel Name, ETD, ETA 등 항해 관련 정보는 `BL` 테이블에서 통합 관리하며, 하나의 BL에 속한 여러 컨테이너는 동일한 항해 스케줄을 공유합니다.
*   **입고(GR) 및 로트 분할 (Unpacking)**: 컨테이너 입고 시 GR을 생성후, 실재고의 최소 단위인 '로트(Lot)'로 분할 저장합니다. GR을 대상으로 매입채무(AP)를 발행할 수 있으며, 따라서 같은 날에 입고된 수량이라고 해도 AP가 발행되는 GR과 발행되지 않는 GR로 나누어 GR을 작성해야 합니다.(재포장 작업이나 스티커 부착 작업이 필요한 물품이 동일한 날 입고된 경우, 해당 작업을 하는 물품은 GR을 분리하여 작성해야 함) 동일 SKU라도 유통기한, 생산 차수(Lot No) 등에 따라 별도의 `Inventory_Lot`을 생성하여 선입선출(FIFO) 및 로트별 원가 귀속의 기초로 삼습니다.
*   **전역 UUID 기반 추적**: 모든 주요 문서(PO, CI, BL, Container, Container_Item, GR, Lot)에 부여된 UUID를 통해 시스템 전체에서 데이터의 기원과 흐름을 즉시 식별할 수 있습니다.

## 2. 랜딩 코스트 및 원가 산출 (Landing Cost)

물품 대금 외의 모든 부대비용을 5대 기준에 따라 로트 단위로 안분하여 정밀한 원가(Landed Cost)를 산출합니다.

### 5대 비용 배분(Allocation) 기준
| 기준 | 설명 | 적용 예시 |
| :--- | :--- | :--- |
| **금액 (Value)** | 가액 비중(Unit Price * Qty)에 따라 배분 | 관세, 보험료 |
| **수량 (Qty)** | 아이템 개수 비중에 따라 배분 | 검사비, 하역비 |
| **부피 (Volume)** | CBM 비중에 따라 배분 | 해상 운임, 보관료 |
| **중량 (Weight)** | 총중량(Gross Weight) 비중에 따라 배분 | 항공 운임, 국내송료 |
| **건 (Unit)** | 대상 로트수로 균등 배분 (1/N) | 서류비, 신고수수료 |

### 일괄 배분 및 증빙 로직 (Batch Allocation)
*   **배분 증빙(Cost_Allocation)**: 유저가 하나 또는 그 이상의 부대비용(AP)을 선택하여 일괄 배분할 때 하나의 '배분 이벤트' 헤더가 생성됩니다. 여기에는 배분 시점의 환율, 배분 기준, 총액이 기록되어 원가 산출의 명확한 증빙이 됩니다.
*   **상세 귀속(Cost_Allocation_Item)**: 배분 이벤트의 결과로 각 로트에 귀속된 금액은 상세 테이블에 기록되어, 특정 로트의 원가가 어떤 청구서들로부터 구성되었는지 완벽한 Audit Trail을 제공합니다.
*   **비용 상속 및 마감**: 선적 전 발생 비용은 입고 시점에 로트로 자동 상속됩니다. 인코텀즈별 필수 비용이 모두 입력되어야 원가 확정(Lock)이 가능하며, 확정된 단가는 `Inventory_Lot.Landed_Cost_Per_Unit`에 스냅샷으로 기록됩니다.

### 일반 배분 vs 추후 배분 (Normal vs Late Cost Allocation)
*   **일반 배분 (Normal Cost Allocation)**: 한 로트(Lot)는 기본적으로 한 번만 배분 대상이 될 수 있습니다. 배분 이력이 없는 로트만 선택 가능하며, 선택된 로트들에 대해 AP 금액을 일괄 배분합니다.
*   **추후 배분 (Late Cost Allocation)**: 이미 배분 이력이 있는 로트도 추가로 선택할 수 있습니다. 예를 들어 선적 후 발생한 추가 비용(관세 추징금, 창고 보관료 등)을 기존에 배분 완료한 로트에 다시 귀속해야 할 때 사용합니다. 추후 배분에서는 이미 배분된 AP 잔액을 기준으로만 추가 배분이 이루어집니다.

### 배분 금액 산출 로직 (Allocation Amount Calculation)
선택된 로트들에 대한 AP 배분 금액은 다음 공식으로 산출됩니다.

```
remainingAP = AP.Local_Amount - alreadyAllocated
remainingBase = totalRefBase - allocatedBase
Lot 금액 = remainingAP * LotBase / remainingBase
```

*   `AP.Local_Amount`: 배분 대상 AP의 현지 통화 금액
*   `alreadyAllocated`: 해당 AP가 이미 Cost_Allocation_Item에서 사용된 금액 합계
*   `totalRefBase`: AP가 참조하는 문서(PO, CI, BL, Container, Container_Item, GR, Lot)의 전체 기준량(Quantity, Weight, Volume, Value, Unit)
*   `allocatedBase`: 이미 배분된 로트들의 기준량 합계
*   `LotBase`: 분배 대상 로트의 기준량

이 로직은 다음을 보장합니다.

1. **이미 배분된 금액은 고정**: `alreadyAllocated`를 먼저 제외하므로, 기존 배분 금액은 재계산되지 않고 보호됩니다.
2. **미선적/미입고 항목 제외**: `totalRefBase`에는 아직 선택되지 않았거나 미입고(unlanded)인 품목의 기준량도 포함되므로, 해당 품목의 몫만큼 자동으로 제외됩니다.
3. **PO 변경 시 비율 재적용**: PO 수량이나 단가가 변경되면 `totalRefBase`가 달라지므로, 새로운 배분 계산에서는 변경된 비율을 기준으로 남은 금액을 분배합니다. 다만 기존에 배분된 금액 자체는 고정되어 있으므로, 기존 배분과 새 배분의 합계가 AP 금액과 정확히 일치하지 않을 수 있습니다.

### PO 변경 시 경고 (PO Modification Guardrail)
PO의 품목(PO_Item)을 수정하거나 삭제하면 해당 PO에 연결된 로트의 기준량 비율이 변경될 수 있으므로, 기존 Cost Allocation의 비율이 물리적으로 틀어질 위험이 있습니다. 따라서 이미 Cost Allocation에 사용된 PO를 수정할 때는 시스템이 다음 정보를 포함한 경고를 표시합니다.

*   배분 이력이 있는 로트 수
*   해당 PO를 통해 배분된 총 금액
*   "기존 배분 비율에 영향을 줄 수 있으므로 주의해서 진행" 안내

해당 경고는 저장을 차단하지 않고 사용자의 확인(Confirm)을 요구합니다. 사용자는 변경의 영향을 인지한 상태에서 PO를 수정할 수 있습니다.

## 3. 수입 현금흐름 및 채무 관리 (Cash Flow)

다양한 출처의 비용을 단일 채무 창구(`Account_Payable`)로 집결시켜 통합 관리합니다.

*   **다형성 참조 (Polymorphic Reference)**: `Account_Payable`은 `Reference_UUID`와 `Reference_Type`을 통해 PO, CI, BL, Container, Container_Item, GR, Lot 등 다양한 원천 문서와 유연하게 연결됩니다. 단, `Container_Item`이나 `Lot`의 경우, 수량이 변경되면 그때마다 새로 `UUID`를 적용합니다. 이는 수량이 변경되어도 기존 `UUID`를 유지하면 연결된 `AP`의 배분시 수량이 달라져서 논리적인 오류가 발생하기 때문입니다.
*   **AP 타겟 그룹 (AP Target Group)**: 여러 PO, 여러 CI, 여러 BL 등 **동일 종류의 원천 문서**를 하나의 그룹 UUID로 묶어 1회 청구형 AP를 발행할 수 있습니다. 그룹은 `AP_Target_Group`과 `AP_Target_Group_Item`으로 관리되며, `AP_Target_Group.Reference_Type`에 의해 그룹이 대표하는 문서 종류가 결정됩니다(PO, CI, BL, Container, Container_Item, GR, Lot). 발행된 AP는 그룹의 UUID를 `Reference_UUID`로, `Reference_Type`을 `AP_Target_Group`으로 참조합니다. 그룹 내 개별 문서는 `AP_Target_Group_Item.Reference_UUID`에 기록되어 단순 참조 목록(Audit Trail)을 제공합니다.
*   **다국어 및 환율 처리**: 모든 채무는 발생 통화(Currency)와 함께 배분/결제 시점의 환율을 적용한 `Local_Amount`(현지 통화 금액)를 병행 기록하여 정확한 외환 차손익 및 원가 계산을 지원합니다.
*   **확정 데이터 중심 운영**: 예측치가 아닌, 유저가 수동으로 확정한 전표(PO 선금, CI 잔액, 부대비용 청구서)만을 공식 현금흐름에 반영합니다.
*   **마이너스 전표 기반 크레딧 노트**: 부족분(Shortage)이나 파손에 대한 클레임은 마이너스(-) 금액의 AP 전표를 발행하여 PO또는 CI 앞으로 발행된 AP에 대한 채무액을 자동 상쇄하는 방식으로 처리합니다.
