import { ComponentFixture, TestBed, waitForAsync } from '@angular/core/testing'

import { GlobalSearchComponent } from './global-search.component'
import { SearchService } from '../backend/api/api'
import { provideHttpClientTesting } from '@angular/common/http/testing'
import { By } from '@angular/platform-browser'
import { PopoverModule } from 'primeng/popover'
import { NoopAnimationsModule } from '@angular/platform-browser/animations'
import { FormsModule } from '@angular/forms'
import { provideHttpClient, withInterceptorsFromDi } from '@angular/common/http'
import { provideRouter, RouterModule } from '@angular/router'
import { DaemonNiceNamePipe } from '../pipes/daemon-name.pipe'
import { EntityLinkComponent } from '../entity-link/entity-link.component'

describe('GlobalSearchComponent', () => {
    let component: GlobalSearchComponent
    let fixture: ComponentFixture<GlobalSearchComponent>

    beforeEach(waitForAsync(() => {
        TestBed.configureTestingModule({
            declarations: [GlobalSearchComponent, DaemonNiceNamePipe, EntityLinkComponent, DaemonNiceNamePipe],
            imports: [PopoverModule, NoopAnimationsModule, FormsModule, RouterModule],
            providers: [
                SearchService,
                provideHttpClient(withInterceptorsFromDi()),
                provideHttpClientTesting(),
                provideRouter([]),
            ],
        }).compileComponents()
    }))

    beforeEach(() => {
        fixture = TestBed.createComponent(GlobalSearchComponent)
        component = fixture.componentInstance
        fixture.detectChanges()
    })

    it('should create', () => {
        expect(component).toBeTruthy()
    })

    it('should display app name and proper link to an app in the results', async () => {
        component.searchResults = {
            subnets: { items: [] },
            sharedNetworks: { items: [] },
            hosts: { items: [] },
            machines: { items: [] },
            daemons: { items: [{ id: 1, name: 'dhcp-server' }] },
            users: { items: [] },
            groups: { items: [] },
        }

        // Show search result box, by default it is hidden
        component.searchResultsBox.show(new Event('click'), fixture.nativeElement)
        await fixture.whenRenderingDone()
        fixture.detectChanges()

        const daemonsDiv = fixture.debugElement.query(By.css('#daemons-div'))
        expect(daemonsDiv).toBeDefined()
        expect(daemonsDiv.children.length).toBe(2)
        const daemonDiv = daemonsDiv.children[1]
        expect(daemonDiv.children.length).toBe(1)
        // Entity link component wraps the daemon display
        const daemonLink = daemonDiv.query(By.css('#daemon-link-1'))
        expect(daemonLink).toBeTruthy()
        expect(daemonLink.nativeElement.innerText).toBe('[1] Dhcp-server')
        expect(daemonLink.attributes.href).toBe('/daemons/1')
    })
})
