import { Component, Input } from '@angular/core'
import { AccessPoint, AnyDaemon } from '../backend'

/**
 * A component that displays daemon access points.
 */
@Component({
    selector: 'app-access-points',
    standalone: false,
    templateUrl: './access-points.component.html',
    styleUrls: ['./access-points.component.sass'],
})
export class AccessPointsComponent {
    /** Pointer to the structure holding the daemon information. */
    @Input() daemon: AnyDaemon

    /**
     * Conditionally formats an IP address for display.
     *
     * @param addr IPv4 or IPv6 address string.
     * @returns unchanged value if it is an IPv4 address or an IPv6 address
     *          surrounded by [ ].
     */
    formatAddress(addr: string): string {
        if (addr.length === 0 || !addr.includes(':') || (addr.startsWith('[') && addr.endsWith(']'))) return addr

        return `[${addr}]`
    }
}
