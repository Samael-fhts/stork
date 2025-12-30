import { Component, Input } from '@angular/core'

import { daemonStatusIconName, daemonStatusIconColor, daemonStatusIconTooltip } from '../utils'
import { Daemon } from '../backend'

@Component({
    selector: 'daemons-status',
    standalone: false,
    templateUrl: './daemons-status.component.html',
    styleUrls: ['./daemons-status.component.sass'],
})
export class DaemonsStatusComponent {
    /**
     * Daemons to show status for. Sorted by importance.
     */
    private _daemons: Daemon[]

    /**
     * Sets daemons to show status for. The daemons are sorted by importance.
     */
    @Input() set daemons(daemons: Daemon[]) {
        this._daemons = this.sortDaemonsByImportance(daemons)
    }

    /**
     * Gets daemons to show status for.
     */
    get daemons(): Daemon[] {
        return this._daemons
    }
    
    constructor() {}

    /** Returns a list of daemon sorted using custom rules. */
    private sortDaemonsByImportance(daemons: Daemon[]) {
        return daemons.sort((a, b) => {
            const order = ['dhcp4', 'dhcp6', 'd2', 'ca', 'netconf', 'named', 'pdns']
            const indexA = order.indexOf(a.name)
            const indexB = order.indexOf(b.name)
            return indexA - indexB
        })
    }

    /**
     * Returns tooltip for the icon in presenting daemon status
     *
     * @param daemon data structure holding the information about the daemon.
     *
     * @returns Tooltip as text. It includes hints about the communication
     *          problems when such problems occur, e.g. it includes the
     *          hint whether the communication is with the agent or daemon.
     */
    daemonStatusIconTooltip(daemon) {
        return daemonStatusIconTooltip(daemon)
    }

    /**
     * Returns the color of the icon used in presenting daemon status
     *
     * @param daemon data structure holding the information about the daemon.
     *
     * @returns grey color if the daemon is not active, red if the daemon is
     *          active but there are communication issues, green if the
     *          communication with the active daemon is ok.
     */
    daemonStatusIconColor(daemon) {
        return daemonStatusIconColor(daemon)
    }

    /**
     * Returns the name of the icon used in presenting daemon status
     *
     * The icon selected depends on whether the daemon is active or not
     * active and whether there is a communication with the daemon or
     * not.
     *
     * @param daemon data structure holding the information about the daemon.
     *
     * @returns ban icon if the daemon is not active, times icon if the daemon
     *          should be active but the communication with it is broken and
     *          check icon if the communication with the active daemon is ok.
     */
    daemonStatusIconName(daemon) {
        return daemonStatusIconName(daemon)
    }
}
