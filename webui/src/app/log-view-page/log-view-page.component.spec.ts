import { provideHttpClientTesting } from '@angular/common/http/testing'
import { ComponentFixture, TestBed, waitForAsync } from '@angular/core/testing'
import { ActivatedRoute, convertToParamMap, provideRouter, RouterModule } from '@angular/router'
import { By } from '@angular/platform-browser'
import { ServicesService } from '../backend'
import { LogViewPageComponent } from './log-view-page.component'
import { of } from 'rxjs'
import { PanelModule } from 'primeng/panel'
import { NoopAnimationsModule } from '@angular/platform-browser/animations'
import { ButtonModule } from 'primeng/button'
import { ProgressSpinnerModule } from 'primeng/progressspinner'
import { SharedModule } from 'primeng/api'
import { EntityLinkComponent } from '../entity-link/entity-link.component'
import { DaemonNiceNamePipe } from '../pipes/daemon-name.pipe'
import { provideHttpClient, withInterceptorsFromDi } from '@angular/common/http'

describe('LogViewPageComponent', () => {
    let component: LogViewPageComponent
    let fixture: ComponentFixture<LogViewPageComponent>

    beforeEach(waitForAsync(() => {
        TestBed.configureTestingModule({
            declarations: [LogViewPageComponent, EntityLinkComponent, DaemonNiceNamePipe],
            imports: [
                PanelModule,
                NoopAnimationsModule,
                ButtonModule,
                ProgressSpinnerModule,
                SharedModule,
                RouterModule,
            ],
            providers: [
                ServicesService,
                {
                    provide: ActivatedRoute,
                    useValue: {
                        paramMap: of(convertToParamMap({})),
                    },
                },
                provideHttpClient(withInterceptorsFromDi()),
                provideHttpClientTesting(),
                provideRouter([]),
            ],
        }).compileComponents()
    }))

    beforeEach(() => {
        fixture = TestBed.createComponent(LogViewPageComponent)
        component = fixture.componentInstance
        fixture.detectChanges()
    })

    it('should create', () => {
        expect(component).toBeTruthy()
    })

    it('should include daemon link', () => {
        component.loaded = true
        component.data = { logTargetOutput: '/tmp/xyz', machine: { id: 1 } }
        component.daemonName = 'fantastic-daemon'
        component.daemonId = 15
        fixture.detectChanges()
        const daemonLink = fixture.debugElement.query(By.css('#daemon-link'))
        const daemonLinkComponent = daemonLink.componentInstance
        expect(daemonLinkComponent).toBeDefined()
        expect(daemonLinkComponent.attrs.hasOwnProperty('name')).toBeTrue()
        expect(daemonLinkComponent.attrs.name).toEqual('fantastic-daemon')
        expect(daemonLinkComponent.attrs.id).toEqual(15)
    })
})
