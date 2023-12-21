import { ComponentFixture, TestBed, fakeAsync, flush, tick } from '@angular/core/testing'

import { HostsMigrationButtonComponent } from './hosts-migration-button.component'
import { ProgressButtonComponent } from '../progress-button/progress-button.component'
import { By } from '@angular/platform-browser'
import { ConfirmationService, MessageService } from 'primeng/api'
import { HostsMigrationService, Migration } from '../hosts-migration-service/hosts-migration.service'
import { ButtonModule } from 'primeng/button'
import { SplitButtonModule } from 'primeng/splitbutton'
import { MenuModule } from 'primeng/menu'
import { BadgeModule } from 'primeng/badge'
import { NoopAnimationsModule } from '@angular/platform-browser/animations'
import { ToastModule } from 'primeng/toast'
import { DialogModule } from 'primeng/dialog'
import { RouterTestingModule } from '@angular/router/testing'
import { EMPTY, Observable, Subject, concat, merge, of, throwError } from 'rxjs'

describe('HostsMigrationButtonComponent', () => {
    let component: HostsMigrationButtonComponent
    let fixture: ComponentFixture<HostsMigrationButtonComponent>
    let migrationService: HostsMigrationService

    /**
     * Retrieves the progress button component from the fixture.
     */
    function getProgressButton(): ProgressButtonComponent | null {
        const progressButton = fixture.debugElement.query(By.directive(ProgressButtonComponent))
        return progressButton?.componentInstance
    }

    /**
     * Checks if the loading indicator is displayed.
     */
    function hasLoadingIndicator(): boolean {
        const progressButton = getProgressButton()
        if (!progressButton) {
            return false
        }
        return progressButton.progressing
    }

    /**
     * Retrieves the error count from the progress button.
     */
    function getErrorCount(): number {
        const progressButton = getProgressButton()
        if (!progressButton) {
            return null
        }
        return progressButton.badgeCount
    }

    /**
     * Checks if the button is disabled.
     */
    function isDisabled(): boolean {
        const progressButton = getProgressButton()
        if (!progressButton) {
            return false
        }
        return progressButton.disabled
    }

    /**
     * Retrieves the progress value from the progress button.
     */
    function getProgressValue(): number {
        const progressButton = getProgressButton()
        if (!progressButton) {
            return null
        }
        return progressButton.value
    }

    beforeEach(() => {
        TestBed.configureTestingModule({
            declarations: [HostsMigrationButtonComponent, ProgressButtonComponent],
            providers: [MessageService, HostsMigrationService],
            imports: [
                ButtonModule,
                SplitButtonModule,
                MenuModule,
                BadgeModule,
                NoopAnimationsModule,
                ToastModule,
                DialogModule,
                RouterTestingModule,
            ],
        })
        fixture = TestBed.createComponent(HostsMigrationButtonComponent)
        component = fixture.componentInstance
        migrationService = TestBed.inject(HostsMigrationService)
        fixture.detectChanges()
    })

    it('should create', () => {
        expect(component).toBeTruthy()
    })

    it('should start in the initializing state', () => {
        expect(component.state).toBe('initializing')
    })

    it('should be initially in the initializing state', fakeAsync(() => {
        // Arrange
        spyOn(migrationService, 'getCurrentMigration').and.returnValue(of(null))

        // Act
        component.ngOnInit()

        // Assert
        expect(component.state).toBe('initializing')
        expect(component.migration).toBeNull()
        expect(component.fetchingAPI).toBe(true)
        expect(component.updateSubscription).toBeNull()
        expect(hasLoadingIndicator()).toBe(true)
        expect(isDisabled()).toBe(true)
        expect(getErrorCount()).toBe(0)
        expect(getProgressValue()).toBe(0)
    }))

    it('should transition to the ready state if no migration is in progress', fakeAsync(() => {
        // Arrange
        spyOn(migrationService, 'getCurrentMigration').and.returnValue(of(null))

        // Act
        component.ngOnInit()

        flush()
        fixture.detectChanges()

        // Assert
        expect(component.state).toBe('ready')
        expect(component.migration).toBeNull()
        expect(component.fetchingAPI).toBe(false)
        expect(component.updateSubscription).toBeNull()
        expect(hasLoadingIndicator()).toBe(false)
        expect(isDisabled()).toBe(false)
        expect(getErrorCount()).toBe(0)
        expect(getProgressValue()).toBe(0)
    }))

    it('should transition to the migrating state if a migration is in progress', fakeAsync(() => {
        // Arrange
        spyOn(migrationService, 'getCurrentMigration').and.returnValue(
            of({
                errors: 42,
                filter: 'filter',
                inProgress: true,
                progress: 0.84,
            } as Migration)
        )
        spyOn(migrationService, 'getMigrationUpdates').and.returnValue(EMPTY)

        // Act
        component.ngOnInit()

        flush()
        fixture.detectChanges()

        // Assert
        expect(component.state).toBe('migrating')
        expect(component.migration).toEqual({
            errors: 42,
            filter: 'filter',
            inProgress: true,
            progress: 0.84,
        } as Migration)

        expect(component.fetchingAPI).toBe(false)
        expect(component.updateSubscription).not.toBeNull()
        expect(hasLoadingIndicator()).toBe(true)
        expect(isDisabled()).toBe(false)
        expect(getErrorCount()).toBe(42)
        expect(getProgressValue()).toBe(0.84)
    }))

    it('should transition to the done state if a migration is already done', fakeAsync(() => {
        // Arrange
        spyOn(migrationService, 'getCurrentMigration').and.returnValue(
            of({
                errors: 42,
                filter: 'filter',
                inProgress: false,
                progress: 1,
            } as Migration)
        )

        // Act
        component.ngOnInit()

        flush()
        fixture.detectChanges()

        // Assert
        expect(component.state).toBe('done')
        expect(component.migration).toEqual({
            errors: 42,
            filter: 'filter',
            inProgress: false,
            progress: 1,
        } as Migration)

        expect(component.fetchingAPI).toBe(false)
        expect(component.updateSubscription).toBeNull()
        expect(hasLoadingIndicator()).toBe(false)
        expect(isDisabled()).toBe(false)
        expect(getErrorCount()).toBe(42)
        expect(getProgressValue()).toBe(1)
    }))

    it('should transition to the migration state from the ready state after starting a migration', fakeAsync(() => {
        // Prepare the spies.
        spyOn(migrationService, 'getCurrentMigration').and.returnValue(of(null))
        spyOn(migrationService, 'startMigration').and.callFake((filter) => {
            expect(filter).toBe('filter')
            return of({
                errors: 0,
                filter: filter,
                inProgress: true,
                progress: 0,
            } as Migration)
        })
        spyOn(migrationService, 'getMigrationUpdates').and.returnValue(EMPTY)
        component.currentFilter = 'filter'

        // Go to the ready state.
        component.ngOnInit()
        flush()
        expect(component.state).toBe('ready')

        // Start a migration.
        component.onConfirmMigrationClick()
        flush()
        fixture.detectChanges()

        // Assert.
        expect(component.state).toBe('migrating')
        expect(component.migration).toEqual({
            errors: 0,
            filter: 'filter',
            inProgress: true,
            progress: 0,
        } as Migration)
    }))

    it('should display the confirmation dialog before starting a migration', fakeAsync(() => {
        // Prepare the spies.
        spyOn(migrationService, 'getCurrentMigration').and.returnValue(of(null))
        spyOn(migrationService, 'startMigration').and.returnValue(
            of({
                errors: 0,
                filter: 'filter',
                inProgress: true,
                progress: 0,
            } as Migration)
        )
        spyOn(migrationService, 'getMigrationUpdates').and.returnValue(EMPTY)

        // Go to the ready state.
        component.ngOnInit()
        flush()
        expect(component.state).toBe('ready')

        // Click the start button.
        component.onStartMigrationClick()
        flush()

        // It should display the confirmation dialog and be still in the ready
        // state.
        expect(component.state).toBe('ready')
        expect(component.showingConfirmation).toBeTrue()

        // Click the confirm button.
        component.onConfirmMigrationClick()
        flush()

        // It should transition to the migrating state.
        expect(component.state).toBe('migrating')
        expect(component.showingConfirmation).toBeFalse()
    }))

    it('should transition to the error state from ready state after failing to start a migration', fakeAsync(() => {
        // Prepare the spies.
        spyOn(migrationService, 'getCurrentMigration').and.returnValue(of(null))
        spyOn(migrationService, 'startMigration').and.returnValue(
            throwError(() => new Error('error'))
        )

        // Go to the ready state.
        component.ngOnInit()
        flush()
        expect(component.state).toBe('ready')

        // Start a migration.
        component.onConfirmMigrationClick()
        flush()
        fixture.detectChanges()

        // Assert.
        expect(component.state).toBe('error')
        expect(component.migration).toBeNull()
        expect(component.fetchingAPI).toBe(false)
        expect(component.updateSubscription).toBeNull()
        expect(hasLoadingIndicator()).toBe(false)
        expect(isDisabled()).toBe(false)
        expect(getErrorCount()).toBeNull()
        expect(getProgressValue()).toBeNull()
    }))

    it('should receive migration updates', fakeAsync(() => {
        // Prepare the spies.
        spyOn(migrationService, 'getCurrentMigration').and.returnValue(of(null))
        spyOn(migrationService, 'startMigration').and.returnValue(of({
                errors: 0,
                filter: null,
                inProgress: true,
                progress: 0,
        } as Migration))
        // Multiple updates are received.
        spyOn(migrationService, 'getMigrationUpdates').and.returnValue(of(
            {
                errors: 0,
                filter: null,
                inProgress: true,
                progress: 25,
            } as Migration,
            {
                errors: 0,
                filter: null,
                inProgress: true,
                progress: 50,
            } as Migration,
            {
                errors: 0,
                filter: null,
                inProgress: true,
                progress: 75,
            } as Migration,
            {
                errors: 0,
                filter: null,
                inProgress: false,
                progress: 100,
            } as Migration,
        ))

        // Go to the ready state.
        component.ngOnInit()
        flush()

        // Start a migration.
        component.onConfirmMigrationClick()
        flush()

        // Assert.
        expect(component.state).toBe('done')
    }))

    it('should receive the filter value updates', fakeAsync(() => {
        // Prepare the spies.
        spyOn(migrationService, 'getCurrentMigration').and.returnValue(of(null))
        spyOn(migrationService, 'startMigration').and.returnValue(of({
                errors: 0,
                filter: null,
                inProgress: true,
                progress: 0,
        } as Migration))

        // Filter observable.
        const filterObservable = new Subject<string>()
        component.filter$ = filterObservable.asObservable()

        // Go to the ready state.
        component.ngOnInit()
        flush()
        expect(component.state).toBe('ready')

        // Check the initial filter value.
        expect(component.currentFilter).toBeNull()

        // Update the filter value.
        filterObservable.next('filter')

        // Assert.
        expect(component.currentFilter).toBe('filter')
    }))

    it('should preserve the ready state if starting a migration is canceled', fakeAsync(() => {
        // Prepare the spies.
        spyOn(migrationService, 'getCurrentMigration').and.returnValue(of(null))

        // Go to the ready state.
        component.ngOnInit()
        flush()
        expect(component.state).toBe('ready')

        // Click the start button.
        component.onStartMigrationClick()
        flush()

        // It should display the confirmation dialog and be still in the ready
        // state.
        expect(component.state).toBe('ready')
        expect(component.showingConfirmation).toBeTrue()

        // Click the cancel button.
        component.onCancelMigrationClick()

        // It should be still in the ready state. The confirmation dialog
        // should be hidden.
        expect(component.state).toBe('ready')
        expect(component.showingConfirmation).toBeFalse()
    }))

    it('should transition to the initializing state on retry request', fakeAsync(() => {
        // Prepare the spies.
        spyOn(migrationService, 'getCurrentMigration').and.returnValues(
            throwError(() => new Error('error')),
            of(null)
        )

        // Go to the error state.
        component.ngOnInit()
        flush()
        expect(component.state).toBe('error')

        // Click the retry button.
        component.onRetryOnErrorClick()
        flush()

        // Assert.
        expect(component.state).toBe('ready')
    }))

    it('should emit the filter for affected reservations event', fakeAsync(() => {}))

    it('should emit the filter for errored reservations event', fakeAsync(() => {}))

    it('should stop the migration on demand', fakeAsync(() => {}))

    it('should remove the migration if it is marked as done', fakeAsync(() => {}))

    it('should attach the proper filter to the new migration', fakeAsync(() => {}))
})
