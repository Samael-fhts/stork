import { Component, EventEmitter, Input, Output } from '@angular/core'

import { AnyDaemon } from '../backend'
import { daemonStatusIconColor, daemonStatusIconName, daemonStatusIconTooltip } from '../utils'
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

    get daemonStatusIconName() {
        return daemonStatusIconName(this.daemon)
    }

    get daemonStatusIconColor() {
        return daemonStatusIconColor(this.daemon)
    }

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
