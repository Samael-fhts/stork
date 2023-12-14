import { Component, EventEmitter, Input, Output } from '@angular/core'
import { MenuItem } from 'primeng/api'

/**
 * Progress button component.
 * It is an outlined button that background is filled depending on the value.
 * It can also display a badge with a count.
 */
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
     * Indicates if the progress is active.
     * If true, the loading indicator is displayed.
     */
    @Input() progressing: boolean = true

    /**
     * The count displayed in the badge.
     * If 0, the badge is not displayed.
     */
    @Input() badgeCount: number = 0

    /**
     * Button label.
     */
    @Input() label: string = ''

    /**
     * Progress button style class.
     * Accepts the same values as the p-button styleClass. The p-button-outline
     * class is always added.
     */
    @Input() styleClass: string = ''

    /**
     * Menu item model
     */
    @Input() model: MenuItem[] = null

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
