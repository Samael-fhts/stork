import { Component, EventEmitter, Input, Output } from '@angular/core'

import { AnyDaemon } from '../backend'
import {
    daemonNameToFriendlyName,
    daemonStatusIconColor,
    daemonStatusIconName,
    daemonStatusIconTooltip,
} from '../utils'

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

    isKeaDaemon(daemon: AnyDaemon) {
        const keaDaemons = ['dhcp4', 'dhcp6', 'd2', 'ca', 'netconf']
        return keaDaemons.includes(daemon?.name)
    }

    appTypeForEvents(daemon: AnyDaemon) {
        if (this.isKeaDaemon(daemon)) {
            return 'kea'
        }

        if (daemon?.name === 'bind9') {
            return 'bind9'
        }

        return null
    }

    refresh() {
        if (this.daemon?.id !== undefined) {
            this.refreshDaemon.emit(this.daemon.id)
        }
    }
}
