/**
 * Specifies the filter parameters for fetching hosts that may be specified
 * in the URL query parameters.
 */
export interface QueryParamsFilter {
    text: string
    appId: number
    subnetId: number
    keaSubnetId: number
    global: boolean
    conflict: boolean
}
