import { Component, EventEmitter, Input, OnDestroy, OnInit, Output } from '@angular/core'
import { ConfirmationService, MessageService } from 'primeng/api'
import { HostsMigrationService, Migration } from '../hosts-migration-service/hosts-migration.service'
import { Observable, Subscription, lastValueFrom } from 'rxjs'
import { getErrorMessage } from '../utils'

type State = 'initializing' | 'ready' | 'migrating' | 'done' | 'error'

@Component({
    selector: 'app-hosts-migration-button',
    templateUrl: './hosts-migration-button.component.html',
    styleUrls: ['./hosts-migration-button.component.sass'],
})
export class HostsMigrationButtonComponent implements OnInit, OnDestroy {
    // Component state.
    state: State
    migration: Migration = null
    updateSubscription: Subscription = null
    fetchingAPI: boolean = false
    currentFilter: string = null
    subscriptions: Subscription = new Subscription()

    // Component inputs.
    @Input() filter$: Observable<string>

    // Event emitters.
    @Output() filterList = new EventEmitter<string>()

    constructor(
        private messageService: MessageService,
        private confirmationService: ConfirmationService,
        private migrationService: HostsMigrationService
    ) {
        // The MenuItem commands must be bound to the component instance.
        this.onStartMigrationClick = this.onStartMigrationClick.bind(this)
        this.onShowErroredHostsClick = this.onShowErroredHostsClick.bind(this)
        this.onShowAffectedHostsClick = this.onShowAffectedHostsClick.bind(this)
        this.onCancelMigrationClick = this.onCancelMigrationClick.bind(this)
        this.onMarkAsReadClick = this.onMarkAsReadClick.bind(this)
    }

    private setState(state: State, migration: Migration = null) {
        this.state = state
        this.migration = migration
        this.deregisterFromUpdates()
    }

    private transitionToInitializingState() {
        this.setState('initializing')
        this.fetchingAPI = true

        // Check the current migration status.
        lastValueFrom(this.migrationService.getCurrentMigration())
            .finally(() => {
                this.fetchingAPI = false
            })
            .then((migration) => {
                if (migration) {
                    if (migration.inProgress) {
                        this.transitionToMigratingState(migration)
                    } else {
                        this.transitionToDoneState(migration)
                    }
                } else {
                    this.transitionToReadyState()
                }
            })
            .catch((err) => {
                this.transitionToErrorState(err)
            })
    }

    private transitionToReadyState() {
        this.setState('ready')
    }

    private transitionToMigrationRequestedState() {
        const filter = this.currentFilter
        this.setState('migrating', {
            errors: 0,
            id: null,
            inProgress: true,
            progress: 0,
            filter: filter
        })
        this.fetchingAPI = true

        // Start a new migration.
        lastValueFrom(this.migrationService.startMigration(filter))
            .finally(() => {
                this.fetchingAPI = false
            })
            .then((migration) => {
                this.transitionToMigratingState(migration)
            })
            .catch((err) => {
                this.transitionToErrorState(err)
            })
    }

    private transitionToMigratingState(migration: Migration) {
        this.setState('migrating', migration)

        // Register for updates
        this.updateSubscription = this.migrationService.getMigrationUpdates(migration.id).subscribe(m => {
            this.migration = m
            if (!m.inProgress) {
                this.transitionToDoneState(m)
            }
        })
    }

    private transitionToDoneState(migration: Migration) {
        this.setState('done', migration)
    }

    private transitionToErrorState(err: Error) {
        this.setState('error')

        // Generate an error message.
        const errorMessage = getErrorMessage(err)
        this.messageService.add({
            severity: 'error',
            summary: 'Migration error',
            detail: errorMessage,
        })
    }

    // Component lifecycle hooks.

    ngOnInit(): void {
        if (this.filter$ != null) {
            this.subscriptions.add(
                this.filter$.subscribe((filter) => {
                    this.currentFilter = filter
                })
            )
        }

        this.transitionToInitializingState()
    }
    ngOnDestroy(): void {
        this.deregisterFromUpdates()
        this.subscriptions.unsubscribe()
    }

    // UI event handlers.
    onStartMigrationClick() {
        this.confirmationService.confirm({
            accept: () => {
                this.transitionToMigrationRequestedState()
            },
        })
    }

    onRedirectToMigrationDetailsClick() {
        this.redirectToMigrationDetails(this.migration.id)
    }

    onRetryOnErrorClick() {
        this.transitionToInitializingState()
    }

    onShowErroredHostsClick() {
        this.emitFilterList('filter', true)
    }

    onShowAffectedHostsClick() {
        this.emitFilterList('filter', false)
    }

    onCancelMigrationClick() {
        this.deregisterFromUpdates()
        this.fetchingAPI = true
        lastValueFrom(this.migrationService.cancelMigration(this.migration.id))
            .finally(() => { this.fetchingAPI = false })
            .then(() => {
                this.transitionToInitializingState()
            })
            .catch((err) => {
                this.transitionToErrorState(err)
            })
    }

    onMarkAsReadClick() {
        this.fetchingAPI = true
        lastValueFrom(this.migrationService.removeMigration(this.migration.id))
            .finally(() => { this.fetchingAPI = false })
            .then(() => {
                this.transitionToInitializingState()
            })
            .catch((err) => {
                this.transitionToErrorState(err)
            })
    }

    // Event emitters.
    private emitFilterList(filter: string, errorsOnly: boolean) {
        filter = this.buildFilter(filter, errorsOnly)
        this.filterList.emit(filter)
    }

    // Helpers.
    private redirectToMigrationDetails(migrationId: number) {
        // ToDo
    }

    private deregisterFromUpdates() {
        if (this.updateSubscription) {
            this.updateSubscription.unsubscribe()
            this.updateSubscription = null
        }
    }

    private buildFilter(base: string, errorsOnly: boolean): string {
        throw new Error('Method not implemented.')
    }
}
