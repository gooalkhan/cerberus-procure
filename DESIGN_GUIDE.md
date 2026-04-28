# Cerberus Procure Design Guide

이 문서는 Cerberus Procure 프로젝트의 핵심 UI/UX 디자인 원칙과 가이드라인을 설명합니다. 본 프로젝트는 현대적이고 전문적인 물류 관리 시스템을 지향하며, 프리미엄 다크 모드 테마를 기반으로 구축되었습니다.

---

## 1. Typography (폰트)

가독성과 현대적인 감각을 위해 Google Fonts의 두 가지 서체를 혼용합니다.

-   **Primary Font (Inter)**: 본문, 입력 폼, 데이터 테이블 등에 사용됩니다. (`font-family: 'Inter', sans-serif;`)
-   **Heading Font (Outfit)**: 페이지 제목, 모달 헤더, 주요 섹션 타이틀에 사용되어 전문적인 인상을 줍니다. (`font-family: 'Outfit', sans-serif;`)

| 사용처 | 폰트 | Weight | 특이사항 |
| :--- | :--- | :--- | :--- |
| 시스템 전반 | Inter | 300, 400, 500, 600 | 가독성 중심 |
| 제목 (h1~h3) | Outfit | 600, 700 | 포인트 서체 |

---

## 2. Color Palette (색상 정의)

Slate 및 Navy 톤의 어두운 배경에 밝은 하늘색(Sky Blue) 포인트를 사용하는 **Premium Dark Theme**를 적용했습니다.

| 명칭 | 변수명 | 색상값 | 용도 |
| :--- | :--- | :--- | :--- |
| **Background** | `--bg-color` | `#0f172a` | 메인 배경색 |
| **Panel BG** | `--panel-bg` | `#1e293b` | 카드, 모달, 사이드바 배경 |
| **Accent** | `--accent-color` | `#38bdf8` | 브랜드 컬러, 강조 텍스트, 활성화 아이콘 |
| **Text Primary** | `--text-primary` | `#f8fafc` | 기본 텍스트 (흰색 계열) |
| **Text Secondary** | `--text-secondary` | `#94a3b8` | 라벨, 설명글, 비활성 텍스트 |
| **Border** | `--border-color` | `#334155` | 구분선, 입력창 테두리 |
| **Success** | `--success` | `#10b981` | 긍정적 상태, 저장 버튼 |
| **Danger** | `--danger` | `#ef4444` | 부정적 상태, 삭제/로그아웃 버튼 |

---

## 3. Layout Components

### 3.1 Sidebar (사이드바)
-   **Width**: 260px (`--sidebar-width`)
-   **Structure**: 상단 로고, 중간 메뉴 리스트, 하단 여백 구성.
-   **Interaction**: 활성화된 메뉴는 좌측에 3px Accent Border가 생기며, 배경색이 미세하게 밝아집니다.
-   **Hierarchy**: 주요 메뉴 아래에 들여쓰기(`padding-left: 2.5rem`)와 낮은 불투명도를 적용한 서브 메뉴를 배치하여 계층 구조를 표현합니다.

### 3.2 Filter Container (필터 컨테이너)
-   페이지 상단에 위치하며, `Panel BG` 배경을 가집니다.
-   **Flex Layout**: 필터 그룹들이 유연하게 배치되며, 라벨은 입력창 상단에 작게 배치됩니다.
-   **Glassmorphism**: 헤더 영역에는 `backdrop-filter: blur(8px)`를 적용하여 스크롤 시 세련된 시각 효과를 제공합니다.

### 3.3 Data Table (데이터 테이블)
-   **Row Hover**: 행 위에 마우스를 올리면 부드러운 배경색 변화와 함께 `pointer` 커서가 표시되어 클릭 가능함을 알립니다.
-   **Responsive**: 가로 스크롤을 지원하여 좁은 화면에서도 데이터가 깨지지 않도록 설계되었습니다.

---

## 4. Modal (모달) & Master-Detail

본 프로젝트의 가장 핵심적인 UX 패턴입니다.

-   **Overlay**: `rgba(0, 0, 0, 0.7)` 배경과 블러 처리를 통해 콘텐츠에 집중하게 합니다.
-   **Form Grid**: 기본적으로 2열 그리드(`1fr 1fr`) 구성을 가지며, 특수한 경우(`Remark` 등) Full Width를 사용합니다.
-   **Sub-Table (Detail)**: 마스터 정보 아래에 연관된 상세 내역(예: 컨테이너 내 품목, B/L 내 컨테이너)을 표시하는 서브 테이블이 배치됩니다. 
    -   서브 테이블은 메인 테이블보다 작고 어두운 톤을 사용하여 시각적 위계를 구분합니다.

---

## 5. UI Elements

-   **Inputs**: 포커스 시 `--accent-color` 테두리가 활성화됩니다.
-   **Buttons**: Hover 시 미세하게 투명도가 변하고, Active(클릭) 시 `scale(0.98)`로 작아지는 마이크로 애니메이션이 적용되어 물리적인 피드백을 줍니다.
-   **Scrollbar**: 커스텀 스크롤바 디자인을 통해 다크 모드 테마와 조화를 이룹니다.
