import { ComponentFixture, fakeAsync, TestBed, tick, waitForAsync } from '@angular/core/testing'

import { LoginScreenComponent } from './login-screen.component'
import { AuthenticationMethod } from '../backend'
import { provideHttpClientTesting } from '@angular/common/http/testing'
import { MessageService } from 'primeng/api'
import { of } from 'rxjs'
import { By } from '@angular/platform-browser'
import { provideRouter, Router } from '@angular/router'
import { AuthService } from '../auth.service'
import { provideNoopAnimations } from '@angular/platform-browser/animations'
import { provideHttpClient, withInterceptorsFromDi } from '@angular/common/http'

describe('LoginScreenComponent', () => {
    let component: LoginScreenComponent
    let fixture: ComponentFixture<LoginScreenComponent>
    let authServiceStub: Partial<AuthService>
    let router: Router
    let msgSrv: MessageService

    beforeEach(waitForAsync(() => {
        authServiceStub = {
            getAuthenticationMethods: () =>
                of([
                    {
                        id: 'localId',
                        name: 'local',
                        description: 'local description',
                        formLabelIdentifier: 'localLabelId',
                        formLabelSecret: 'localLabelSecret',
                    },
                    {
                        id: 'ldapId',
                        name: 'ldap',
                        description: 'ldap description',
                        formLabelIdentifier: 'ldapLabelId',
                        formLabelSecret: 'ldapLabelSecret',
                    },
                ] as AuthenticationMethod[]),
            login: () => ({
                id: 1,
                authenticationMethodId: 'ldap',
            }),
        }
        TestBed.configureTestingModule({
            providers: [
                MessageService,
                { provide: AuthService, useValue: authServiceStub },
                provideNoopAnimations(),
                provideHttpClient(withInterceptorsFromDi()),
                provideHttpClientTesting(),
                provideRouter([]),
            ],
        }).compileComponents()
    }))

    beforeEach(() => {
        fixture = TestBed.createComponent(LoginScreenComponent)
        component = fixture.componentInstance
        router = fixture.debugElement.injector.get(Router)
        msgSrv = fixture.debugElement.injector.get(MessageService)
    })

    it('should create', () => {
        expect(component).toBeTruthy()
    })

    it('should display welcome message', fakeAsync(() => {
        spyOn(component.http, 'get').and.returnValue(of('This is a welcome message'))
        component.ngOnInit()
        tick()
        fixture.detectChanges()

        const welcomeMessage = fixture.debugElement.query(By.css('.login-screen__welcome'))
        expect(welcomeMessage).toBeTruthy()
        expect(welcomeMessage.nativeElement.innerText).toContain('This is a welcome message')
    }))

    it('should not display bloated welcome message', fakeAsync(() => {
        spyOn(component.http, 'get').and.returnValue(of('a'.repeat(2049)))
        component.ngOnInit()
        tick()
        fixture.detectChanges()

        const welcomeMessage = fixture.debugElement.query(By.css('.login-screen__welcome'))
        expect(welcomeMessage).toBeFalsy()
    }))

    it('should display authentication methods', fakeAsync(() => {
        // Inject AuthService stub.
        fixture.debugElement.injector.get(AuthService)
        component.ngOnInit()
        tick()
        fixture.detectChanges()

        // Check if data was received from AuthService getAuthenticationMethods().
        expect(component.authenticationMethods).toBeTruthy()
        expect(component.authenticationMethods.length).toEqual(2)

        // There should be a dropdown visible.
        const dropdown = fixture.debugElement.query(By.css('.login-screen__authentication-selector .p-select'))
        expect(dropdown).toBeTruthy()

        dropdown.nativeElement.click()
        fixture.detectChanges()

        // Dropdown should display two methods.
        const listItems = dropdown.queryAll(By.css('.p-select-list li'))
        expect(listItems).toBeTruthy()
        expect(listItems.length).toEqual(2)
        expect(listItems[0].nativeElement.innerText).toContain('local')
        expect(listItems[1].nativeElement.innerText).toContain('ldap')
    }))

    it('should try to sign-in user with selected authentication method', fakeAsync(() => {
        const authService = fixture.debugElement.injector.get(AuthService)
        const loginSpy = spyOn(authService, 'login').and.callThrough()
        fixture.detectChanges()
        tick()

        component.loginForm.patchValue({
            authenticationMethod: component.authenticationMethods.find((m) => m.id === 'ldapId'),
            identifier: 'login',
            secret: 'passwd',
        })
        component.signIn()

        expect(loginSpy).toHaveBeenCalledOnceWith('ldapId', 'login', 'passwd', '/')
    }))

    it('should display toast with auth error information', () => {
        spyOnProperty(router, 'url').and.returnValue('/login/auth-err')
        spyOn(msgSrv, 'add')
        component.ngOnInit()
        expect(msgSrv.add).toHaveBeenCalledOnceWith(
            jasmine.objectContaining({
                severity: 'error',
                summary: 'Authentication error',
                detail: 'Error during authentication process. Please contact Stork admin.',
            })
        )
    })
})
