import type { ThemeConfig } from 'antd'

const theme: ThemeConfig = {
  token: {
    colorPrimary: '#3A84FF',
    colorSuccess: '#2DCB56',
    colorWarning: '#FF9C01',
    colorError: '#EA3636',
    colorTextBase: '#63656E',
    colorText: '#63656E',
    colorTextSecondary: '#979BA5',
    colorTextTertiary: '#C4C6CC',
    colorTextQuaternary: '#DCDEE5',
    colorBgContainer: '#FFFFFF',
    colorBgLayout: '#F5F7FA',
    colorBgElevated: '#FFFFFF',
    colorBorder: '#DCDEE5',
    colorBorderSecondary: '#E7E9EF',
    borderRadius: 2,
    borderRadiusLG: 2,
    borderRadiusSM: 2,
    fontSize: 14,
    fontSizeSM: 12,
    fontSizeLG: 16,
    fontFamily:
      '-apple-system, BlinkMacSystemFont, "PingFang SC", "Microsoft YaHei", "Helvetica Neue", Arial, sans-serif',
    controlHeight: 32,
    controlHeightSM: 26,
    controlHeightLG: 38,
    boxShadow: '0 1px 2px 0 rgba(0,0,0,0.08)',
    boxShadowSecondary: '0 3px 9px 0 rgba(0,0,0,0.1)',
    lineHeight: 1.5,
    wireframe: false,
    motion: true,
  },
  components: {
    Table: {
      headerBg: '#F0F1F5',
      headerColor: '#313238',
      rowHoverBg: '#F5F7FA',
      borderColor: '#DFE0E5',
      headerBorderRadius: 0,
    },
    Button: {
      primaryShadow: 'none',
      defaultShadow: 'none',
      dangerShadow: 'none',
    },
    Card: {
      headerHeight: 50,
      paddingLG: 20,
    },
    Modal: {
      headerBg: '#FFFFFF',
      contentBg: '#FFFFFF',
    },
    Input: {
      activeBorderColor: '#3C96FF',
      activeShadow: '0 0 4px rgba(58,132,255,0.4)',
      hoverBorderColor: '#979BA5',
    },
    Tag: {
      defaultBg: 'transparent',
    },
    Select: {
      optionActiveBg: '#EAF3FF',
      optionSelectedBg: '#F4F6FA',
    },
  },
}

export default theme
