import { Component, Input, Output, EventEmitter } from '@angular/core'

import { forkJoin, lastValueFrom } from 'rxjs'

import { MessageService } from 'primeng/api'

import { ServicesService } from '../backend/api/api'
import { ServerDataService } from '../server-data.service'

import {
    daemonNameToFriendlyName,
    daemonStatusErred,
    daemonStatusIconName,
    daemonStatusIconTooltip,
    getErrorMessage,
} from '../utils'
import { DaemonTab } from '../daemons'
import { Bind9Daemon, Daemon } from '../backend'

@Component({
    selector: 'app-daemon-tab',
    standalone: false,
    templateUrl: './daemon-tab.component.html',
    styleUrls: ['./daemon-tab.component.sass'],
})
export class DaemonTabComponent {
    /**
     * Event emitter sending an event to the parent component when the daemon is
     * refreshed.
     */
    @Output() refreshDaemon = new EventEmitter<number>()

    /**
     * Information about the daemons.
     */
    @Input() daemon: Daemon

    constructor(
        private servicesApi: ServicesService,
        private serverData: ServerDataService,
        private msgService: MessageService
    ) {}

    /**
     * An action triggered when refresh button is pressed.
     */
    refreshDaemonState() {
        this.refreshDaemon.emit(this.daemon.id)
    }

    /**
     * Returns boolean value indicating if there is an issue with communication
     * with the given daemon
     *
     * @return true if there is a communication problem with the daemon,
     *         false otherwise.
     */
    get daemonStatusErred(): boolean {
        return this.daemon.active && daemonStatusErred(this.daemon)
    }

    /**
     * Returns the name of the icon to be used when presenting daemon status
     *
     * The icon selected depends on whether the daemon is active or not
     * active and whether there is a communication with the daemon or
     * not.
     *
     * @returns ban icon if the daemon is not active, times icon if the daemon
     *          should be active but the communication with it is broken and
     *          check icon if the communication with the active daemon is ok.
     */
    get daemonStatusIconName() {
        return daemonStatusIconName(this.daemon)
    }

    /**
     * Returns error text to be displayed when there is a communication issue
     * with a given daemon
     *
     * @returns Error text. It includes hints about the communication
     *          problems when such problems occur, e.g. it includes the
     *          hint whether the communication is with the agent or daemon.
     */
    get daemonStatusErrorText() {
        return daemonStatusIconTooltip(this.daemon)
    }
}
