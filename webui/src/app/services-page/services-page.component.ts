import { Component } from '@angular/core'
import { MenuItem } from 'primeng/api'

@Component({
    selector: 'app-services-page',
    templateUrl: './services-page.component.html',
    styleUrl: './services-page.component.sass',
})
export class ServicesPageComponent {
    breadcrumbs: MenuItem[] = [{ label: 'Services' }]
}
