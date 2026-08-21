# Omnilogs: Ant Design and CSS Reference

เอกสารนี้เป็นจุดอ้างอิงสำหรับปรับสี ขนาด ระยะห่าง และพฤติกรรมของ UI ฝั่ง frontend ของ Omnilogs โดยบันทึกจากโค้ดที่ใช้งานอยู่ใน `src/`

## ลำดับความสำคัญของการปรับ style

```text
1. Ant Design theme token          src/app/theme.ts
2. Ant Design component token      src/app/theme.ts -> components
3. CSS variables                   src/index.css -> :root
4. Global CSS override             src/index.css -> .ant-...
5. Page/component CSS เฉพาะจุด     src/styles/** หรือ className ของ component
```

ให้เริ่มจาก token ก่อนเสมอ เพราะมีผลกับ component ทั้งระบบ ใช้ selector `.ant-*` เมื่อ token ทำไม่ได้จริง และ scope selector ไว้กับหน้า/feature เพื่อไม่ให้กระทบทั้งระบบโดยไม่ตั้งใจ

## 1. Theme token กลางของ Ant Design

แก้ไขที่ `src/app/theme.ts` ภายใน `antdThemeConfig.token` แล้ว `ConfigProvider` ใน `src/App.tsx` จะส่งผลให้ทุก Ant Design component

| Token | ค่าปัจจุบัน | ใช้ควบคุม |
| --- | --- | --- |
| `colorPrimary` | `#6647F0` | สีหลักของปุ่ม primary, control ที่ active และองค์ประกอบที่เลือก |
| `colorSuccess` | `#00BE8F` | สถานะสำเร็จ |
| `colorWarning` | `#D4B106` | สถานะเตือน |
| `colorError` | `#FF3D89` | สถานะผิดพลาด/ปุ่มอันตราย |
| `colorInfo` | `#1890FF` | สถานะข้อมูล |
| `colorLink` | `#6647F0` | สี link ปกติ |
| `colorLinkHover` | `#4F33CE` | สี link เมื่อ hover |
| `colorLinkActive` | `#3C22AC` | สี link เมื่อ active |
| `colorTextBase` | `#000000D9` | สีข้อความหลัก |
| `colorTextSecondary` | `#00000073` | สีข้อความรอง/คำอธิบาย |
| `colorTextTertiary` | `#00000040` | สีข้อความระดับสาม |
| `colorTextPlaceholder` | `#00000040` | สี placeholder ใน form |
| `colorBorder` | `#d9d9d9` | เส้นขอบ component ทั่วไป |
| `colorBorderSecondary` | `#f0f0f0` | เส้นแบ่ง/เส้นขอบรอง |
| `colorBgContainer` | `#ffffff` | พื้นหลังของ card, input และ container |
| `colorBgLayout` | `#f5f5f5` | พื้นหลัง layout/page |
| `colorBgElevated` | `#ffffff` | พื้นหลังของ popup, dropdown และ layer ยกสูง |
| `borderRadius` | `8` | มุมโค้งมาตรฐาน |
| `borderRadiusLG` | `12` | มุมโค้ง card/modal/drawer ขนาดใหญ่ |
| `borderRadiusSM` | `6` | มุมโค้ง element ขนาดเล็ก |
| `fontFamily` | Noto Sans Thai Looped และ system fonts | ฟอนต์ทั้งระบบ |
| `fontSize` | `14` | ขนาดฟอนต์ฐาน |
| `controlHeight` | `36` | ความสูง input/select/button ปกติ |
| `controlHeightLG` | `40` | ความสูง control ขนาดใหญ่ |
| `controlHeightSM` | `28` | ความสูง control ขนาดเล็ก |
| `motionDurationSlow/Mid/Fast` | `0.3s/0.2s/0.1s` | ความเร็ว motion ของ Ant Design |

ตัวอย่าง: เปลี่ยนสีหลักของระบบ

```ts
// src/app/theme.ts
token: {
  colorPrimary: '#0F766E',
  colorLink: '#0F766E',
}
```

## 2. Component token ของ Ant Design

แก้ไขที่ `src/app/theme.ts` ใน `antdThemeConfig.components`

| Component | Token/ค่า | จุดที่ควบคุม |
| --- | --- | --- |
| `Layout` | `siderBg` | สีพื้นหลัง sidebar |
| `Layout` | `headerBg` | สีพื้นหลัง header |
| `Layout` | `bodyBg` | สีพื้นหลัง body layout |
| `Layout` | `headerHeight`, `headerPadding` | ขนาดและ padding ของ header |
| `Menu` | `darkItemBg` | พื้นหลัง menu dark ปกติ |
| `Menu` | `darkItemSelectedBg` | สีพื้นหลังรายการ sidebar ที่เลือก |
| `Menu` | `darkItemHoverBg` | สีพื้นหลังรายการ sidebar เมื่อ hover |
| `Menu` | `darkItemColor` | สีข้อความ menu dark ปกติ |
| `Menu` | `darkItemSelectedColor` | สีข้อความ menu dark ที่เลือก |
| `Input` | `activeBorderColor` | สีขอบ input เมื่อ focus |
| `Input` | `hoverBorderColor` | สีขอบ input เมื่อ hover |
| `Input` | `activeShadow` | เงา/outline input เมื่อ focus |
| `Input` | `errorActiveShadow` | outline input เมื่อมี validation error |
| `InputNumber` | `controlHeight`, `borderRadius` | ขนาดและมุมโค้ง input number |
| `Select` | `optionSelectedBg` | พื้นหลัง option ที่เลือก |
| `Select` | `optionActiveBg` | พื้นหลัง option เมื่อ hover |
| `Select` | `selectorBg` | พื้นหลัง select |
| `DatePicker` | `activeBorderColor`, `hoverBorderColor`, `activeShadow` | สีและ focus state ของ date picker |
| `Button` | `controlHeight` | ความสูงปุ่ม |
| `Button` | `borderRadius` | มุมโค้งปุ่ม |
| `Button` | `primaryShadow` | เงาของปุ่ม primary |
| `Button` | `defaultBorderColor`, `defaultColor` | ขอบและข้อความของปุ่ม default |
| `Button` | `fontWeight` | น้ำหนักตัวอักษรปุ่ม |
| `Card` | `borderRadiusLG`, `paddingLG` | มุมโค้งและ padding card |
| `Table` | `headerBg`, `headerColor` | พื้นหลัง/ข้อความ table header |
| `Table` | `headerSortActiveBg` | สีหัวตารางที่กำลัง sort |
| `Table` | `rowHoverBg` | สี row เมื่อ hover |
| `Table` | `borderColor`, `headerBorderRadius` | เส้นขอบและมุมหัวตาราง |
| `Tabs` | `inkBarColor` | เส้นใต้ tab ที่เลือก |
| `Tabs` | `itemActiveColor`, `itemSelectedColor`, `itemHoverColor` | สีข้อความ tab ตาม state |
| `Tag` | `borderRadiusSM` | มุมโค้ง tag |
| `Modal` | `borderRadiusLG`, `titleFontSize` | มุมโค้งและขนาด title modal |
| `Drawer` | `borderRadiusLG` | มุมโค้ง drawer |
| `Divider` | `colorSplit` | สีเส้นแบ่ง |
| `Tooltip` | `borderRadius` | มุมโค้ง tooltip |
| `Avatar` | `colorTextPlaceholder` | สีข้อความ placeholder ของ avatar |
| `Badge` | `dotSize` | ขนาด dot badge |
| `Message` | `borderRadiusLG` | มุมโค้ง message |
| `Notification` | `borderRadiusLG` | มุมโค้ง notification |
| `Pagination` | `itemActiveBg`, `itemActiveColor`, `itemActiveBorderColor` | สีเลขหน้าที่เลือก |
| `Switch` | `colorPrimary`, `colorPrimaryHover` | สี switch เมื่อเปิดและ hover |
| `Form` | `labelFontSize`, `verticalLabelPadding` | ขนาดและระยะ label ของ form |
| `Breadcrumb` | `linkColor`, `linkHoverColor`, `lastItemColor`, `separatorColor` | สี breadcrumb ทุกสถานะ |
| `Dropdown` | `borderRadiusLG`, `controlItemBgHover` | มุมและพื้นหลัง dropdown item เมื่อ hover |
| `Skeleton` | `borderRadiusSM` | มุมโค้ง skeleton |
| `Result` | `titleFontSize`, `subtitleFontSize` | ขนาดข้อความหน้า result/error |

## 3. CSS variables ของโปรเจกต์

ประกาศที่ `src/index.css` ใน `:root` ใช้ใน CSS และ inline style ผ่าน `var(--...)`

| Variable | ค่าปัจจุบัน | ใช้สำหรับ |
| --- | --- | --- |
| `--color-primary` | `#6647f0` | สีแบรนด์หลัก, avatar และ loading bar |
| `--color-sidebar-bg` | `#100446` | พื้นหลัง sidebar/login gradient |
| `--color-sidebar-hover` | `#1c0b68` | sidebar hover |
| `--color-sidebar-active` | `#2a158a` | sidebar item ที่ active |
| `--color-purple-500` ถึง `--color-purple-50` | purple scale | accent, table header, hover และ background เบา |
| `--color-error` | `#ff3d89` | ข้อความ/สถานะ error |
| `--color-success` | `#00be8f` | ข้อความ/สถานะ success |
| `--color-warning` | `#d4b106` | ข้อความ/สถานะ warning |
| `--color-info` | `#1890ff` | ข้อความ/สถานะ info |
| `--color-text-primary` | `#000000d9` | ข้อความหลัก |
| `--color-text-secondary` | `#00000073` | ข้อความรอง |
| `--sidebar-width` | `260px` | ความกว้าง sidebar มาตรฐาน |
| `--sidebar-collapsed-width` | `80px` | ความกว้าง sidebar ตอนย่อ |
| `--header-height` | `64px` | ความสูง header/logo/sidebar content offset |

เมื่อเปลี่ยนสีแบรนด์ ให้เปลี่ยนทั้ง `colorPrimary` ใน theme และ `--color-primary` ใน CSS ให้ตรงกัน เพื่อป้องกัน UI คนละเฉด

## 4. Global overrides ที่มีผลกับ Ant Design

อยู่ใน `src/index.css`

| Selector | ผลที่เกิดขึ้น | จุดที่แก้ |
| --- | --- | --- |
| `.ant-layout-sider` | บังคับสีพื้นหลัง sidebar ด้วย `--color-sidebar-bg` | เปลี่ยน variable หรือ selector นี้ |
| `.ant-menu-dark` | ทำพื้นหลัง menu dark ให้โปร่งใส | เปลี่ยน `background` |
| `.ant-menu-dark .ant-menu-item-selected` | สีรายการ sidebar ที่เลือก | `--color-sidebar-active` |
| `.ant-menu-dark .ant-menu-item:hover` | สี hover item ใน sidebar | `--color-sidebar-hover` |
| `.ant-menu-dark .ant-menu-submenu-title:hover` | สี hover ของ submenu | `--color-sidebar-hover` |
| `.ant-btn-primary` | **บังคับปุ่ม primary เป็นดำ** | แก้/ลบ override นี้หากต้องการใช้ `colorPrimary` สีม่วงจาก theme |
| `.ant-btn-primary:hover` | ปุ่ม primary hover เป็นพื้นขาว ตัวอักษร/ขอบดำ | แก้/ลบ override นี้เพื่อใช้ hover จาก Ant Design token |
| `.ant-input-textarea textarea` | กำหนดความสูงต่ำสุด TextArea เป็น `80px` | ปรับ `min-height` |
| `.admin-table .ant-table-thead > tr > th` | หัวตารางสีม่วงอ่อนและเส้นล่าง | `--color-purple-50`, `--color-purple-100` |
| `.admin-table .ant-table-tbody > tr:hover > td` | row ตาราง hover สีม่วงอ่อน | `--color-purple-50` |

ข้อควรระวัง: `.ant-btn-primary` ใช้ `!important` จึงชนะ theme token `Button` และ `colorPrimary` ปัจจุบันปุ่ม `<Button type="primary">` จะเป็นสีดำ ไม่ใช่สีม่วง หากต้องการให้ปุ่ม primary ใช้สีแบรนด์ ให้ลบสอง selector นี้หรือเปลี่ยนเป็น `var(--color-primary)` ตาม state ที่ต้องการ

## 5. Layout และ navigation CSS

### Application shell — `src/styles/layout/app-shell.css`

| Selector | ใช้ควบคุม |
| --- | --- |
| `.app-shell` | padding รอบหน้าหลักและพื้นหลัง canvas ผ่าน `--app-canvas-background` (fallback `#f1f5f9`) |
| `.page-panel` | padding ของพื้นที่ content ภายใน shell |
| media query `max-width: 760px` | ลด padding สำหรับ mobile |

### Floating dock — `src/styles/layout/floating-dock.css`

| กลุ่ม selector | ใช้ควบคุม |
| --- | --- |
| `.nav-dock`, `.nav-dock--open` | พื้นหลัง glass, border, shadow และรูปทรง navigation dock |
| `.nav-dock__brand`, `.nav-dock__quick-link` | สี/พื้นหลัง interaction ของ brand และ quick link |
| `:focus-visible` ของ dock controls | สี focus ring (`rgb(102 71 240 / 24%)`) |
| `.nav-dock__avatar` | สี avatar ใช้ `--color-primary` |
| `.nav-dock__toggle` | ปุ่มเปิด/ปิด dock สีดำ (`#020617`) และ hover (`#1e293b`) |
| `.nav-dock__item`, `--active` | สีข้อความ, hover และ item ที่เลือก |
| `.nav-dock__item-icon` | สีและพื้นหลัง icon ของ navigation |
| media queries `1024px`, `700px`, `380px` | จำนวนคอลัมน์ ขนาด touch target และ text บนจอเล็ก |

สีใน floating dock หลายจุดเขียนเป็น hex/slate ตรง ๆ หากต้องการเปลี่ยนทั้งระบบ แนะนำย้ายค่าซ้ำไปเป็น CSS variables ก่อน แล้วอ้างอิง variable แทน

## 6. Page CSS ที่มีอยู่

| ไฟล์ | Selector สำคัญ | ใช้สำหรับปรับ |
| --- | --- | --- |
| `src/styles/pages/login.css` | `.login-page .login-container` | gradient ของหน้า login (`--color-sidebar-bg` → `--color-primary`) |
| `src/styles/pages/login.css` | `.login-page .login-card` | ความกว้าง มุม และ shadow ของ login card |
| `src/styles/pages/dashboard.css` | `.dashboard-page .stat-card` | มุม, shadow และ hover ของ dashboard cards |
| `src/styles/pages/users.css` | `.users-page .admin-table` | สีหัวตารางและ row hover ของ Users |
| `src/styles/pages/settings.css` | `.settings-section-title` | สี/น้ำหนักหัวข้อ settings (`--color-text-primary`) |
| `src/styles/pages/error.css` | `.error-actions` | layout ของปุ่มในหน้า 403/404 |
| `src/styles/pages/audit-logs.css` | `.audit-title-icon`, `.audit-activity-icon` | icon สี primary บนพื้นม่วงอ่อน |
| `src/styles/pages/audit-logs.css` | `.audit-filter-card`, `.audit-table-card` | border, radius และ shadow ของ audit cards |
| `src/styles/pages/audit-logs.css` | `.audit-table ...` | สีหัวตาราง, row hover, padding และ pagination ของ Audit Logs |
| `src/styles/pages/audit-logs.css` | `.audit-json` | สีพื้นหลัง/ข้อความ/เส้นขอบของ JSON viewer |
| `src/styles/pages/logs-explorer.css` | `#summarry`, `#raw`, `#Expand` | hover ของ action ใน Logs Explorer; ชื่อ `summarry` สะกดต่างจาก `summary` และ CSS นี้ยังไม่ถูก import จาก `src/styles/index.css` |

`src/styles/index.css` ปัจจุบัน import layout, login, dashboard, users, settings และ error แต่ยังไม่ได้ import `audit-logs.css` หรือ `logs-explorer.css` หากต้องการให้ rule ของสองไฟล์นี้มีผลทั่วแอป ต้องเพิ่ม import เช่น:

```css
@import './pages/audit-logs.css';
@import './pages/logs-explorer.css';
```

## 7. จุดที่ใช้ปรับสีตามความต้องการ

### เปลี่ยนสีปุ่ม primary ทั้งระบบ

1. แก้ `colorPrimary` ใน `src/app/theme.ts`
2. แก้ `--color-primary` ใน `src/index.css`
3. ตรวจ `.ant-btn-primary` ใน `src/index.css`; ลบหรือแก้ override สีดำ เพราะมี `!important`

### เปลี่ยนสีข้อความ

- ข้อความหลัก: `colorTextBase` และ `--color-text-primary`
- ข้อความรอง: `colorTextSecondary` และ `--color-text-secondary`
- Placeholder: `colorTextPlaceholder`
- ข้อความ status: `colorSuccess`, `colorWarning`, `colorError`, `colorInfo` หรือ CSS variables สถานะคู่กัน

### เปลี่ยนสีปุ่มสถานะ

- ปุ่ม primary: `colorPrimary` + global `.ant-btn-primary`
- ปุ่ม danger: `colorError`
- ปุ่ม link: `colorLink`, `colorLinkHover`, `colorLinkActive`
- Switch: `components.Switch.colorPrimary`, `colorPrimaryHover`

### เปลี่ยนสี form และ focus

- Input: `components.Input.activeBorderColor`, `hoverBorderColor`, `activeShadow`, `errorActiveShadow`
- InputNumber: `components.InputNumber.*`
- DatePicker: `components.DatePicker.*`
- Select options: `components.Select.optionSelectedBg`, `optionActiveBg`
- TextArea height: `.ant-input-textarea textarea` ใน `src/index.css`

### เปลี่ยนสี table

- ค่าเริ่มต้น: `components.Table.headerBg`, `headerColor`, `rowHoverBg`, `borderColor`
- Users/admin table: `.admin-table` หรือ `.users-page .admin-table`
- Audit table: `.audit-table` ใน `audit-logs.css`

### เปลี่ยนสี sidebar และ menu

- ใช้ `components.Layout.siderBg` และ `components.Menu.*` เป็นหลัก
- CSS global ยัง override sidebar/menu dark ด้วย variables: `.ant-layout-sider`, `.ant-menu-dark ...`
- เปลี่ยน `--color-sidebar-bg`, `--color-sidebar-hover`, `--color-sidebar-active` ให้ครบ

## 8. แนวทางเพิ่ม style ใหม่

1. ใช้ Ant Design props/token ก่อน เช่น `type="primary"`, `status="error"`, `color="warning"`
2. หากต้องใช้ค่าแบรนด์ซ้ำ ให้เพิ่ม CSS variable ใน `src/index.css` พร้อมชื่อสื่อความหมาย
3. หากเป็น token ของ Ant Design component ให้เพิ่มใน `src/app/theme.ts`
4. หากเป็น style เฉพาะหน้า ให้สร้าง/แก้ `src/styles/pages/<page>.css` และ import ผ่าน `src/styles/index.css`
5. scope selector ด้วย page wrapper เช่น `.audit-page .audit-table` ไม่ใช้ `.ant-table` เปล่า ๆ
6. ตรวจ desktop, mobile, focus-visible, hover และ disabled state

## 9. Checklist ก่อนแก้สีหรือ CSS

- สีเปลี่ยนอยู่ใน token หรือ CSS variable ที่ถูกต้องหรือไม่
- มี global override ที่ใช้ `!important` ทับ token อยู่หรือไม่
- สีข้อความมี contrast เพียงพอบนพื้นหลังหรือไม่
- มี hover, focus-visible, disabled, error และ loading state หรือไม่
- CSS เฉพาะหน้า import ผ่าน `src/styles/index.css` แล้วหรือไม่
- selector ไม่กระทบ component/หน้าจออื่นโดยไม่ตั้งใจหรือไม่
- ตรวจบน mobile และ keyboard navigation แล้วหรือไม่
