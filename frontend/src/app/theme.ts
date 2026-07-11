import { ConfigProvider } from 'antd';
import thTH from 'antd/locale/th_TH';

/**
 * Ant Design Theme Configuration
 * ใช้สีตาม design system ที่กำหนด
 * ทุก component ต้องใช้ theme ที่นี่เป็นหลัก ห้ามใส่สี hardcode ใน component
 */
export const antdThemeConfig = {
  token: {
    // ===== Colors =====
    colorPrimary: '#6647F0',
    colorSuccess: '#00BE8F',
    colorWarning: '#D4B106',
    colorError: '#FF3D89',
    colorInfo: '#1890FF',
    colorLink: '#6647F0',
    colorLinkHover: '#4F33CE',
    colorLinkActive: '#3C22AC',

    // ===== Text =====
    colorTextBase: '#000000D9',
    colorTextSecondary: '#00000073',
    colorTextTertiary: '#00000040',
    colorTextPlaceholder: '#00000040',

    // ===== Border & Background =====
    colorBorder: '#d9d9d9',
    colorBorderSecondary: '#f0f0f0',
    colorBgContainer: '#ffffff',
    colorBgLayout: '#f5f5f5',
    colorBgElevated: '#ffffff',

    // ===== Shape =====
    borderRadius: 8,
    borderRadiusLG: 12,
    borderRadiusSM: 6,

    // ===== Font =====
    fontFamily:
      "'Noto Sans Thai', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif",
    fontSize: 14,

    // ===== Spacing =====
    controlHeight: 36,
    controlHeightLG: 40,
    controlHeightSM: 28,

    // ===== Motion =====
    motionDurationSlow: '0.3s',
    motionDurationMid: '0.2s',
    motionDurationFast: '0.1s',
  },
  components: {
    // ===== Layout =====
    Layout: {
      siderBg: '#100446',
      headerBg: '#ffffff',
      bodyBg: '#f5f5f5',
      headerHeight: 64,
      headerPadding: '0 24px',
    },

    // ===== Menu =====
    Menu: {
      darkItemBg: 'transparent',
      darkItemSelectedBg: '#2A158A',
      darkItemHoverBg: '#1C0B68',
      darkSubMenuItemBg: 'transparent',
      darkItemColor: 'rgba(255, 255, 255, 0.75)',
      darkItemSelectedColor: '#ffffff',
      itemBorderRadius: 8,
      subMenuItemBorderRadius: 8,
    },

    // ===== Input (Focus: outline สี secondary primary) =====
    Input: {
      controlHeight: 36,
      borderRadius: 8,
      activeBorderColor: '#6647F0',
      hoverBorderColor: '#AE9CFF',
      activeShadow: '0 0 0 3px rgba(102, 71, 240, 0.12)',
      warningActiveShadow: '0 0 0 3px rgba(212, 177, 6, 0.12)',
      errorActiveShadow: '0 0 0 3px rgba(255, 61, 137, 0.15)',
    },

    // ===== TextArea =====
    // TextArea min-height กำหนดผ่าน CSS (80px) ที่ index.css

    // ===== InputNumber =====
    InputNumber: {
      controlHeight: 36,
      borderRadius: 8,
      activeBorderColor: '#6647F0',
      hoverBorderColor: '#AE9CFF',
      activeShadow: '0 0 0 3px rgba(102, 71, 240, 0.12)',
    },

    // ===== Select =====
    Select: {
      controlHeight: 36,
      borderRadius: 8,
      optionSelectedBg: '#F0EDFF',
      optionActiveBg: '#F0EDFF',
      selectorBg: '#ffffff',
    },

    // ===== DatePicker =====
    DatePicker: {
      controlHeight: 36,
      borderRadius: 8,
      activeBorderColor: '#6647F0',
      hoverBorderColor: '#AE9CFF',
      activeShadow: '0 0 0 3px rgba(102, 71, 240, 0.12)',
    },

    // ===== Button =====
    Button: {
      controlHeight: 36,
      borderRadius: 8,
      primaryShadow: '0 2px 8px rgba(102, 71, 240, 0.3)',
      defaultBorderColor: '#d9d9d9',
      defaultColor: '#000000D9',
      fontWeight: 500,
    },

    // ===== Card =====
    Card: {
      borderRadiusLG: 12,
      paddingLG: 24,
    },

    // ===== Table =====
    Table: {
      headerBg: '#F0EDFF',
      headerColor: '#000000D9',
      headerSortActiveBg: '#E6E0FF',
      rowHoverBg: '#F0EDFF',
      borderColor: '#f0f0f0',
      headerBorderRadius: 8,
    },

    // ===== Tabs =====
    Tabs: {
      inkBarColor: '#6647F0',
      itemActiveColor: '#6647F0',
      itemSelectedColor: '#6647F0',
      itemHoverColor: '#AE9CFF',
    },

    // ===== Tag =====
    Tag: {
      borderRadiusSM: 6,
    },

    // ===== Modal =====
    Modal: {
      borderRadiusLG: 12,
      titleFontSize: 16,
    },

    // ===== Drawer =====
    Drawer: {
      borderRadiusLG: 12,
    },

    // ===== Divider =====
    Divider: {
      colorSplit: '#f0f0f0',
    },

    // ===== Tooltip =====
    Tooltip: {
      borderRadius: 6,
    },

    // ===== Avatar =====
    Avatar: {
      colorTextPlaceholder: '#ffffff',
    },

    // ===== Badge =====
    Badge: {
      dotSize: 8,
    },

    // ===== Message =====
    Message: {
      borderRadiusLG: 8,
    },

    // ===== Notification =====
    Notification: {
      borderRadiusLG: 12,
    },

    // ===== Pagination =====
    Pagination: {
      borderRadius: 8,
      itemActiveBg: '#6647F0',
      itemActiveColor: '#ffffff',
      itemActiveBorderColor: '#6647F0',
      itemBg: '#ffffff',
      itemInputBg: '#ffffff',
    },

    // ===== Switch =====
    Switch: {
      colorPrimary: '#6647F0',
      colorPrimaryHover: '#4F33CE',
    },

    // ===== Form =====
    Form: {
      labelFontSize: 14,
      verticalLabelPadding: '0 0 6px',
    },

    // ===== Breadcrumb =====
    Breadcrumb: {
      linkColor: '#00000073',
      linkHoverColor: '#6647F0',
      lastItemColor: '#000000D9',
      separatorColor: '#00000040',
    },

    // ===== Dropdown =====
    Dropdown: {
      borderRadiusLG: 8,
      controlItemBgHover: '#F0EDFF',
    },

    // ===== Skeleton =====
    Skeleton: {
      borderRadiusSM: 6,
    },

    // ===== Result =====
    Result: {
      titleFontSize: 20,
      subtitleFontSize: 14,
    },
  },
};

export const antdLocale = thTH;

export { ConfigProvider };
