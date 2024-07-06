/**
 * @typedef ExtraTheme
 * @property {string} titleColor
 * @property {string} sidebarColor
 * @property {string} splitColor
 */

/**
 *
 * @type ExtraTheme
 */
export const extraLightTheme = {
    titleColor: '#F2F2F2',
    sidebarColor: '#F2F2F2',
    splitColor: '#DADADA',
}

/**
 *
 * @type ExtraTheme
 */
export const extraDarkTheme = {
    titleColor: '#18181C',
    sidebarColor: '#18181C',
    splitColor: '#474747',
}

/**
 *
 * @param {boolean} dark
 * @return ExtraTheme
 */
export const extraTheme = (dark) => {
    return dark ? extraDarkTheme : extraLightTheme
}
