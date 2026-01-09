import { Component, Input } from '@angular/core'

import { daemonStatusIconName, daemonStatusIconColor, daemonStatusIconTooltip } from '../utils'
import { AnyDaemon } from '../backend'

@Component({
    selector: 'app-daemon-status',
    standalone: false,
    templateUrl: './daemon-status.component.html',
    styleUrls: ['./daemon-status.component.sass'],
})
export class DaemonStatusComponent {
    @Input({ required: true }) daemon: AnyDaemon

    get daemonStatusIconTooltip() {
        return daemonStatusIconTooltip(this.daemon)
    }

    get daemonStatusIconColor() {
        return daemonStatusIconColor(this.daemon)
    }

    get daemonStatusIconName() {
        return daemonStatusIconName(this.daemon)
    }
}
