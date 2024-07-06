import { merge } from 'lodash'

/**
 *
 * @type import('naive-ui').GlobalThemeOverrides
 */
export const themeOverrides = {
    common: {
        primaryColor: '#1f276c',
        primaryColorHover: '#5887be',
        primaryColorPressed: '#1f277d',
        primaryColorSuppl: '#5887be',
        borderRadius: '4px',
        borderRadiusSmall: '3px',
        heightMedium: '32px',
        lineHeight: 1.5,
        scrollbarWidth: '8px',
        tabColor: '#FFFFFF',
    },
    Button: {
        heightMedium: '32px',
        paddingSmall: '0 8px',
        paddingMedium: '0 12px',
    },
    Tag: {
        borderRadius: '4px',
        heightLarge: '32px',
    },
    Input: {
        heightMedium: '32px',
    },
    Tabs: {
        tabGapSmallCard: '2px',
        tabGapMediumCard: '2px',
        tabGapLargeCard: '2px',
        tabFontWeightActive: 450,
    },
    Card: {
        colorEmbedded: '#FAFAFA',
    },
    Form: {
        labelFontSizeTopSmall: '12px',
        labelFontSizeTopMedium: '13px',
        labelFontSizeTopLarge: '13px',
        labelHeightSmall: '18px',
        labelHeightMedium: '18px',
        labelHeightLarge: '18px',
        labelPaddingVertical: '0 0 5px 2px',
        feedbackHeightSmall: '18px',
        feedbackHeightMedium: '18px',
        feedbackHeightLarge: '20px',
        feedbackFontSizeSmall: '11px',
        feedbackFontSizeMedium: '12px',
        feedbackFontSizeLarge: '12px',
        labelTextColor: 'rgb(113,120,128)',
        labelFontWeight: '450',
    },
    Radio: {
        buttonColorActive: '#1f276f',
        buttonTextColorActive: '#FFF',
    },
    DataTable: {
        thPaddingSmall: '6px 8px',
        tdPaddingSmall: '6px 8px',
    },
    Dropdown: {
        borderRadius: '5px',
        optionIconSizeMedium: '18px',
        padding: '6px 2px',
        optionColorHover: '#1f276c',
        optionTextColorHover: '#FFF',
        optionHeightMedium: '28px',
    },
    Divider: {
        color: '#AAAAAB',
    },
}

/**
 *
 * @type import('naive-ui').GlobalThemeOverrides
 */
const _darkThemeOverrides = {
    common: {
        primaryColor: '#eeb866',
        primaryColorHover: '#ce9b4e',
        primaryColorPressed: '#eeb869',
        primaryColorSuppl: '#ce9b4e',
        bodyColor: '#101014',
        tabColor: '#101014', //#18181C
        // borderColor: '#262629',
    },
    Tree: {
        nodeTextColor: '#CECED0',
    },
    Card: {
        // colorEmbedded: '#212121',
    },
    Radio: {
        buttonColorActive: '#eeb867',
        buttonTextColorActive: '#FFF',
    },
    Dropdown: {
        // color: '#272727',
        optionColorHover: '#eeb866',
    },
    Popover: {
        // color: '#2C2C32',
    },
}

export const darkThemeOverrides = merge({}, themeOverrides, _darkThemeOverrides)
