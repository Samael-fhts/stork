import { Component, OnDestroy, OnInit, ViewChild } from '@angular/core'
import { Router, ActivatedRoute, EventType } from '@angular/router'

import { Table, TableLazyLoadEvent } from 'primeng/table'

import { DHCPService } from '../backend/api/api'
import { getErrorMessage } from '../utils'
import {
    getTotalAddresses,
    getAssignedAddresses,
    parseSubnetsStatisticValues,
    parseSubnetStatisticValues,
    extractUniqueSharedNetworkPools,
    SharedNetworkWithUniquePools,
} from '../subnets'
import { Subscription, lastValueFrom, EMPTY } from 'rxjs'
import { catchError, filter, map } from 'rxjs/operators'
import { SharedNetwork } from '../backend'
import { MenuItem, MessageService } from 'primeng/api'
import { PrefilteredTable } from '../table'
import { Location } from '@angular/common'

/**
 * Specifies the filter parameters for fetching shared networks that may be
 * specified in the URL query parameters.
 */
interface QueryParamsFilter {
    text?: string
    dhcpVersion?: 4 | 6
    appId?: string
}

/**
 * Component for presenting shared networks in a table.
 */
@Component({
    selector: 'app-shared-networks-page',
    templateUrl: './shared-networks-page.component.html',
    styleUrls: ['./shared-networks-page.component.sass'],
})
export class SharedNetworksPageComponent extends PrefilteredTable<QueryParamsFilter> implements OnInit, OnDestroy {
    subscriptions = new Subscription()
    breadcrumbs = [{ label: 'DHCP' }, { label: 'Shared Networks' }]

    @ViewChild('networksTable') table: Table
    prefilterKey: keyof QueryParamsFilter = 'appId'
    queryParamBooleanKeys: (keyof QueryParamsFilter)[] = []
    queryParamNumericKeys: (keyof QueryParamsFilter)[] = ['dhcpVersion']
    stateKey: string = 'networks-table-session-all'
    filterBooleanKeys: (keyof QueryParamsFilter)[] = []
    filterNumericKeys: (keyof QueryParamsFilter)[] = ['appId', 'dhcpVersion']

    // networks
    networks: SharedNetworkWithUniquePools[] = []

    // Tab menu

    /**
     * Array of tab menu items with shared network information.
     *
     * The first tab is always present and displays the shared networks list.
     *
     * Note: we cannot use the URL with no segment for the list tab. It causes
     * the first tab to be always marked active. The Tab Menu has a built-in
     * feature to highlight items based on the current route. It seems that it
     * matches by a prefix instead of an exact value (the "/foo/bar" URL
     * matches the menu item with the "/foo" URL).
     */
    tabs: MenuItem[] = [{ label: 'Shared Networks', routerLink: '/dhcp/shared-networks/all' }]

    /**
     * Selected tab menu index.
     *
     * The first tab has an index of 0.
     */
    activeTabIndex = 0

    /**
     * Holds the information about specific shared networks presented in the tabs.
     *
     * The entry corresponding to shared networks list is not related to any specific
     * shared network. Its ID is 0.
     */
    openedSharedNetworks: SharedNetworkWithUniquePools[] = [{ id: 0 }]

    /**
     * Constructor.
     *
     * @param route activated route.
     * @param messageService message service.
     * @param router router.
     * @param dhcpApi a service for communication with the server.
     * @param location location service used to update queryParams
     */
    constructor(
        private route: ActivatedRoute,
        private messageService: MessageService,
        private router: Router,
        private dhcpApi: DHCPService,
        private location: Location
    ) {
        super(router, location)
    }

    /**
     * A component lifecycle hook invoked when the component instance is destroyed.
     *
     * It unsubscribes from all subscriptions.
     */
    ngOnDestroy(): void {
        this.filter$.complete()
        this.subscriptions.unsubscribe()
    }

    /**
     * A component lifecycle hook invoked when the component is initialized.
     */
    ngOnInit() {
        this.dataLoading = true

        const paramMap = this.route.snapshot.paramMap
        const queryParamMap = this.route.snapshot.queryParamMap

        // Get host id and appId.
        const id = paramMap.get('id')
        if (!id || id === 'all') {
            this.parseIdFromQueryParam(queryParamMap)
            if (this.hasPrefilter()) {
                this.stateKey = `networks-table-session-${this.prefilterValue}`
            }
        }

        this.subscribeFilterValidation()
        this.subscribeFilterHandler()

        // subscribe to subsequent changes to query params
        this.subscriptions.add(
            this.router.events
                .pipe(
                    filter((event, idx) => idx === 0 || event.type === EventType.NavigationEnd),
                    catchError((err) => {
                        const msg = getErrorMessage(err)
                        this.messageService.add({
                            severity: 'error',
                            summary: 'Cannot process the URL query',
                            detail: msg,
                            life: 10000,
                        })
                        return EMPTY
                    })
                )
                .subscribe(() => {
                    const paramMap = this.route.snapshot.paramMap
                    const queryParamMap = this.route.snapshot.queryParamMap

                    // Apply to the changes of the host id, e.g. from /dhcp/shared-networks/all to
                    // /dhcp/shared-networks/1. Those changes are triggered by switching between the
                    // tabs.

                    // Get shared-network id.
                    const id = paramMap.get('id')
                    if (!id || id === 'all') {
                        // Update the filter only if the target is shared-networks list.
                        this.updateFilterFromQueryParameters(queryParamMap)
                        this.switchToTab(0)
                        return
                    }
                    // if (id === 'new') {
                    //     this.openNewNetwork()
                    //     return
                    // }
                    const numericId = parseInt(id, 10)
                    if (!Number.isNaN(numericId)) {
                        // The path has a numeric id indicating that we should
                        // open a tab with selected shared-network information or switch
                        // to this tab if it has been already opened.
                        this.openTabBySharedNetworkId(numericId)
                    } else {
                        // In case of failed Id parsing, open list tab.
                        this.switchToTab(0)
                        this.filter$.next({ filter: {} })
                    }
                })
        )
    }

    /**
     * Loads shared networks from the database into the component.
     *
     * @param event Event object containing index of the first row, maximum number
     *              of rows to be returned, dhcp version and text for networks filtering.
     */
    loadData(event: TableLazyLoadEvent) {
        this.dataLoading = true

        lastValueFrom(
            this.dhcpApi
                .getSharedNetworks(
                    event.first,
                    event.rows,
                    this.prefilterValue ?? this.getTableFilterValue('appId', event.filters),
                    this.getTableFilterValue('dhcpVersion', event.filters),
                    this.getTableFilterValue('text', event.filters)
                )
                .pipe(
                    map((sharedNetworks) => {
                        parseSubnetsStatisticValues(sharedNetworks.items)
                        return sharedNetworks
                    })
                )
        )
            .then((data) => {
                this.networks = data.items ?? []
                this.totalRecords = data.total ?? 0
            })
            .catch((error) => {
                // ToDo: Silent error catching. We should display a message to the user.
                console.log(error)
            })
            .finally(() => {
                this.dataLoading = false
            })
    }

    /**
     * Get the total number of addresses in the network.
     */
    getTotalAddresses(network: SharedNetwork) {
        return getTotalAddresses(network)
    }

    /**
     * Get the number of assigned addresses in the network.
     */
    getAssignedAddresses(network: SharedNetwork) {
        return getAssignedAddresses(network)
    }

    /**
     * Get the total number of delegated prefixes in the network.
     */
    getTotalDelegatedPrefixes(network: SharedNetwork) {
        return network.stats?.['total-pds']
    }

    /**
     * Get the number of delegated prefixes in the network.
     */
    getAssignedDelegatedPrefixes(network: SharedNetwork) {
        return network.stats?.['assigned-pds']
    }

    /**
     * Returns a list of applications maintaining a given shared network.
     * The list doesn't contain duplicates.
     *
     * @param net Shared network
     * @returns List of the applications (only ID and app name)
     */
    getApps(net: SharedNetwork) {
        const apps = []
        const appIds = {}

        for (const sn of net.subnets) {
            for (const lsn of sn.localSubnets) {
                if (!appIds.hasOwnProperty(lsn.appId)) {
                    apps.push({ id: lsn.appId, name: lsn.appName })
                    appIds[lsn.appId] = true
                }
            }
        }

        return apps
    }

    /**
     * Returns true if at least one of the shared networks contains at least
     * one IPv6 subnet
     */
    get isAnyIPv6SubnetVisible(): boolean {
        return !!this.networks?.some((n) => n.subnets.some((s) => s.subnet.includes(':')))
    }

    /**
     * Open a shared network tab.
     *
     * If the tab already exists, switch to it without fetching the data.
     * Otherwise, fetch the shared network information from the API and
     * create a new tab.
     *
     * @param sharedNetworkId Shared network ID or a NaN for subnet list.
     */
    openTabBySharedNetworkId(sharedNetworkId: number) {
        const tabIndex = this.openedSharedNetworks.map((t) => t.id).indexOf(sharedNetworkId)
        if (tabIndex < 0) {
            this.createTab(sharedNetworkId).then(() => {
                this.switchToTab(this.openedSharedNetworks.length - 1)
            })
        } else {
            this.switchToTab(tabIndex)
        }
    }

    /**
     * Close a menu tab by index.
     *
     * @param index Tab index.
     * @param event Event triggered upon tab closing.
     */
    closeTabByIndex(index: number, event?: Event) {
        if (index == 0) {
            return
        }

        this.openedSharedNetworks.splice(index, 1)
        this.tabs = [...this.tabs.slice(0, index), ...this.tabs.slice(index + 1)]

        if (this.activeTabIndex === index) {
            // Closing currently selected tab. Switch to previous tab.
            this.switchToTab(index - 1)
            this.router.navigate([this.tabs[index - 1].routerLink])
        } else if (this.activeTabIndex > index) {
            // Sitting on the later tab then the one closed. We don't need
            // to switch, but we have to adjust the active tab index.
            this.activeTabIndex--
        }

        if (event) {
            event.preventDefault()
        }
    }

    /**
     * Create a new tab for a given shared network ID.
     *
     * It fetches the shared network information from the API.
     *
     * @param sharedNetworkId Shared network ID.
     */
    private createTab(sharedNetworkId: number): Promise<void> {
        this.dataLoading = true
        return (
            lastValueFrom(
                // Fetch data from API.
                this.dhcpApi.getSharedNetwork(sharedNetworkId).pipe(
                    map((sharedNetwork) => {
                        if (sharedNetwork) {
                            parseSubnetStatisticValues(sharedNetwork)
                        }
                        return sharedNetwork
                    })
                )
            )
                // Execute and use.
                .then((data) => {
                    if (data) {
                        const networks = extractUniqueSharedNetworkPools([data])
                        this.appendTab(networks[0])
                    }
                })
                .catch((error) => {
                    const msg = getErrorMessage(error)
                    this.messageService.add({
                        severity: 'error',
                        summary: 'Cannot get shared network',
                        detail: `Error getting shared network with ID ${sharedNetworkId}: ${msg}`,
                        life: 10000,
                    })
                })
                .finally(() => {
                    this.dataLoading = false
                })
        )
    }

    /**
     * Append a new tab to the list of tabs.
     *
     * @param sharedNetwork Shared network data.
     */
    private appendTab(sharedNetwork: SharedNetwork) {
        this.openedSharedNetworks.push(sharedNetwork)
        this.tabs = [
            ...this.tabs,
            {
                label: sharedNetwork.name,
                routerLink: `/dhcp/shared-networks/${sharedNetwork.id}`,
            },
        ]
    }

    /**
     * Switch to tab identified by an index.
     *
     * @param index Tab index.
     */
    private switchToTab(index: number) {
        if (this.activeTabIndex === index) {
            return
        }
        this.activeTabIndex = index
    }
}
