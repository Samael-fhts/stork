import { Component, EventEmitter, Input, Output } from '@angular/core'

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
    @Input() daemon: AnyDaemon
    @Output() refreshDaemon = new EventEmitter<number>()

    /**
     * Returns the CSS class to display the icon to be used to indicate daemon status
     */
    get daemonStatusIconClass() {
        return daemonStatusIconClass(this.daemon)
    }

    /**
     * Returns tooltip for the icon presented for the daemon status
     */
    get daemonStatusIconTooltip() {
        return daemonStatusIconTooltip(this.daemon)
    }

    /**
     * Indicates if the given daemon is a Kea daemon.
     * @param daemon
     * @returns true if the daemon is Kea daemon; otherwise false.
     */
    get isKeaDaemon() {
        return isKeaDaemon(this.daemon?.name)
    }

    /**
     * Emits the refresh event.
     */
    refresh() {
        if (this.daemon?.id !== undefined) {
            this.refreshDaemon.emit(this.daemon.id)
        }
    }
}
