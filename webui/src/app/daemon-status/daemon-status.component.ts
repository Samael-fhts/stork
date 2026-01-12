import { Component, Input } from '@angular/core'

import { daemonStatusIconClass, daemonStatusIconTooltip } from '../utils'
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

    get daemonStatusIconClass() {
        return daemonStatusIconClass(this.daemon)
    }
}
