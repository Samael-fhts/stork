/**
 * Specifies the filter parameters for fetching hosts that may be specified
 * in the URL query parameters.
 * 
 * The undefined values of the properties are used to indicate that the
 * particular filter is not provided.
 * The null values of the properties are used to indicate that the particular
 * filter has an incorrect value.
 */
export interface QueryParamsFilter {
    text: string
    appId: number
    subnetId: number
    keaSubnetId: number
    global: boolean
    conflict: boolean
    migrationError: boolean
}

/**
 * Creates an empty query param filter.
 */
export function getEmptyQueryParamsFilter(): QueryParamsFilter {
    return {
        text: undefined,
        appId: undefined,
        subnetId: undefined,
        keaSubnetId: undefined,
        global: undefined,
        conflict: undefined,
        migrationError: undefined
    }

}

/**
 * Returns the keys of the boolean properties of the QueryParamsFilter.
 * @returns List of keys.
 */
export function getBooleanQueryParamsFilterKeys(): (keyof QueryParamsFilter)[] {
    return ['global', 'conflict', 'migrationError']
}

/**
 * Returns the keys of the numeric properties of the QueryParamsFilter.
 * @returns List of keys.
 */
export function getNumericQueryParamsFilterKeys(): (keyof QueryParamsFilter)[] {
    return ['appId', 'subnetId', 'keaSubnetId']
}
