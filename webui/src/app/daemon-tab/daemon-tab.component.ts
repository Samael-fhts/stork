import { Component, computed, EventEmitter, input, Output } from '@angular/core'

import { AnyDaemon } from '../backend'
import { daemonStatusIconClass, daemonStatusIconTooltip } from '../utils'
import { isKeaDaemon } from '../version.service'

@Component({
    selector: 'app-daemon-tab',
    standalone: false,
    templateUrl: './daemon-tab.component.html',
    styleUrl: './daemon-tab.component.sass',
})
export class DaemonTabComponent {
    daemon = input.required<AnyDaemon>(null)
    @Output() refreshDaemon = new EventEmitter<number>()

    /**
     * The CSS class to display the icon to be used to indicate daemon status
     */
    daemonStatusIconClass = computed(() => daemonStatusIconClass(this.daemon()))

    /**
     * Tooltip for the icon presented for the daemon status
     */
    daemonStatusIconTooltip = computed(() => daemonStatusIconTooltip(this.daemon()))

    /**
     * Indicates if the given daemon is a Kea daemon.
     * @param daemon
     * @returns true if the daemon is Kea daemon; otherwise false.
     */
    isKeaDaemon = computed(() => isKeaDaemon(this.daemon()?.name))

    /**
     * Emits the refresh event.
     */
    refresh() {
        const daemon = this.daemon()
        if (daemon?.id !== undefined) {
            this.refreshDaemon.emit(daemon.id)
        }
    }
}
