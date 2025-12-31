import { HttpErrorResponse } from '@angular/common/http'
import { Injectable } from '@angular/core'
import { Observable, Subject, merge, timer, EMPTY, of } from 'rxjs'
import { switchMap, shareReplay, catchError, filter, map } from 'rxjs/operators'

import { AuthService } from './auth.service'
import { ServicesService, UsersService } from './backend/api/api'
import { Groups } from './backend/model/groups'
import { DaemonsStats } from './backend'

/**
 * Service for providing and caching data from the server.
 */
@Injectable({
    providedIn: 'root',
})
export class ServerDataService {
    private daemonsStats: Observable<DaemonsStats>
    private groups: Observable<Groups>
    private reloadDaemonsStats = new Subject<void>()
    private reloadDaemonConfiguration: { [daemonId: number]: Subject<number> } = {}
    private _daemonConfigurations: { [daemonId: number]: Observable<any> } = {}

    constructor(
        private auth: AuthService,
        public servicesApi: ServicesService,
        private usersApi: UsersService
    ) {}

    /**
     * Get daemons stats from the server and cache it for other subscribers.
     * Cache is refreshed after 30 minutes.
     */
    getDaemonsStats() {
        if (!this.daemonsStats) {
            const refreshInterval = 1000 * 60 * 30 // 30 mins
            const refreshTimer = timer(0, refreshInterval)

            // For each timer tick and and for each reload
            // make an http request to fetch new data
            this.daemonsStats = merge(refreshTimer, this.reloadDaemonsStats, this.auth.currentUser$).pipe(
                filter((x) => x !== null), // filter out trigger which is logout ie user changed to null
                switchMap(() => {
                    return this.servicesApi.getDaemonsStats().pipe(
                        // use subpipe to not complete source due to error
                        catchError(() => EMPTY) // in case of error drop the response (it should not be cached)
                    )
                }),
                shareReplay(1) // cache the response for all subscribers
            )
        }

        return this.daemonsStats
    }

    /**
     * Force reloading cache for daemons stats.
     */
    forceReloadDaemonsStats() {
        this.reloadDaemonsStats.next()
    }

    /**
     * Get system groups from the server and cache it for other subscribers.
     *
     * Cache is refreshed upon user login.
     */
    getGroups() {
        if (!this.groups) {
            this.groups = this.auth.currentUser$.pipe(
                filter((x) => x !== null), // filter out trigger which is logout ie user changed to null
                switchMap(() =>
                    this.usersApi.getGroups().pipe(
                        // use subpipe to not complete source due to error
                        catchError(() => EMPTY) // in case of error drop the response (it should not be cached)
                    )
                ),
                shareReplay(1) // cache the response for all subscribers
            )
        }

        return this.groups
    }

    /**
     * Get name of the system group fetched from the database indicated by group ID.
     *
     * @param groupId Identifier of the group in the database, counted
     *                from 1.
     * @param groupItems List of all groups returned by the server.
     * @returns Group name or unknown string if the group is not found.
     */
    public getGroupName(groupId: number, groupItems: any[]): string {
        // The superadmin group is well known and doesn't require
        // iterating over the list of groups fetched from the server.
        // Especially, if the server didn't respond properly for
        // some reason, we still want to be able to handle the
        // superadmin group.
        if (groupId === 1) {
            return 'superadmin'
        }
        for (const grp of groupItems) {
            if (grp.id === groupId) {
                return grp.name
            }
        }
        return 'unknown'
    }

    /**
     * Get (Kea) daemon configuration from the server and cache it for other subscribers.
     * Cache is refreshed manually or when the user is logged in.
     * @param daemonId Daemon ID
     * @returns Observable of daemon configuration
     */
    public getDaemonConfiguration(daemonId: number): Observable<any | HttpErrorResponse> {
        if (!(daemonId in this._daemonConfigurations)) {
            this.reloadDaemonConfiguration[daemonId] = new Subject<number>()
            this._daemonConfigurations[daemonId] = merge(
                this.reloadDaemonConfiguration[daemonId],
                this.auth.currentUser$
            ).pipe(
                filter((x) => x !== null), // filter out trigger which is logout ie user changed to null
                switchMap(() => {
                    return this.servicesApi.getDaemonConfig(daemonId).pipe(
                        // use subpipe to not complete source due to error
                        catchError((err) => of(err)) // in case of error continue with it to prevent broken pipe
                    )
                }),
                shareReplay(1) // cache the response for all subscribers
            )
        }
        return this._daemonConfigurations[daemonId]
    }

    /**
     * Force reloading cache for daemon configuration.
     * @param daemonId Daemon ID
     */
    forceReloadDaemonConfiguration(daemonId: number) {
        if (daemonId in this.reloadDaemonConfiguration) {
            this.reloadDaemonConfiguration[daemonId].next(daemonId)
        }
    }
}
