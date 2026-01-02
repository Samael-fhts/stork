import { Component, Input } from '@angular/core'

import { daemonStatusIconName, daemonStatusIconColor, daemonStatusIconTooltip } from '../utils'
import { Daemon } from '../backend'

@Component({
    selector: 'app-daemon-status',
    standalone: false,
    templateUrl: './daemon-status.component.html',
    styleUrls: ['./daemon-status.component.sass'],
})
export class DaemonStatusComponent {
    @Input() daemon: Daemon

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
     * @returns Tooltip as text. It includes hints about the communication
     *          problems when such problems occur, e.g. it includes the
     *          hint whether the communication is with the agent or daemon.
     */
    get daemonStatusIconTooltip() {
        return daemonStatusIconTooltip(this.daemon)
    }

    /**
     * Returns the color of the icon used in presenting daemon status
     *
     * @returns grey color if the daemon is not active, red if the daemon is
     *          active but there are communication issues, green if the
     *          communication with the active daemon is ok.
     */
    get daemonStatusIconColor() {
        return daemonStatusIconColor(this.daemon)
    }

    /**
     * Returns the name of the icon used in presenting daemon status
     *
     * The icon selected depends on whether the daemon is active or not
     * active and whether there is a communication with the daemon or
     * not.
     *
     * @returns an icon if the daemon is not active, times icon if the daemon
     *          should be active but the communication with it is broken and
     *          check icon if the communication with the active daemon is ok.
     */
    get daemonStatusIconName() {
        return daemonStatusIconName(this.daemon)
    }
}
