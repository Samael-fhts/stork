import { Component, computed, input } from '@angular/core'

import { daemonStatusIconClass, daemonStatusIconTooltip } from '../utils'
import { AnyDaemon } from '../backend'

@Component({
    selector: 'app-daemon-status',
    standalone: false,
    templateUrl: './daemon-status.component.html',
    styleUrls: ['./daemon-status.component.sass'],
})
export class DaemonStatusComponent {
    daemon = input<AnyDaemon>(null)

    /**
     * Tooltip for the icon presented for the daemon status
     */
    daemonStatusIconTooltip = computed(() => daemonStatusIconTooltip(this.daemon()))

    /**
     * The CSS class to display the icon to be used to indicate daemon status
     */
    daemonStatusIconClass = computed(() => daemonStatusIconClass(this.daemon()))
}
