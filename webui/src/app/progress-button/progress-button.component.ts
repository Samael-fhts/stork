import { Component, EventEmitter, Input, Output } from '@angular/core'

@Component({
    selector: 'app-progress-button',
    templateUrl: './progress-button.component.html',
    styleUrls: ['./progress-button.component.sass'],
})
export class ProgressButtonComponent {
    /**
     * Progress bar value from 0 to 1.
     */
    @Input() value: number = 0

    /**
     * Indicates if the progress is enabled.
     */
    @Input() enabled: boolean = true

    /**
     * The count displayed in the badge.
     */
    @Input() badgeCount: number = 0

    /**
     * Button label.
     */
    @Input() label: string = ''

    /**
     * Progress button style class.
     */
    @Input() styleClass: string = ''

    /**
     * Click handler.
     */
    @Output() click = new EventEmitter<void>()

    /**
     * Click handler.
     */
    onClick() {
        this.click.emit()
    }
}
