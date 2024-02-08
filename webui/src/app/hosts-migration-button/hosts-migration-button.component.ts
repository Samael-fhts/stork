import { Component, EventEmitter, Input, OnDestroy, OnInit, Output } from '@angular/core'
import { MessageService } from 'primeng/api'
import { HostsMigrationService, Migration } from '../hosts-migration-service/hosts-migration.service'
import { Observable, Subscription, lastValueFrom } from 'rxjs'
import { getErrorMessage } from '../utils'
import { Router } from '@angular/router'
import { QueryParamsFilter } from '../hosts-page/query-params-filter'

/**
 * The UI of this component is organized in the shape of a state machine.
 * The state machine has the following states:
 * - initializing: The migration status is unknown and fetching from API.
 *                 All interaction is disabled. The loading indicator is shown.
 * - ready: The migration is not running and the button is ready to start a new
 *          migration.
 * - migrating: The migration is running. The button acts as a progress bar and
 *              a link to the migration details page. More actions are
 *              available in the dropdown menu. The loading indicator is shown.
 *              The migration progress is updated in real time and can be
 *              stopped. Tne migrating state has two phases:
 *              - requesting migration: The new migration is created. It is
 *                only entered from the ready state. The loading indicator is
 *                shown. The button is disabled.
 *              - migration in progress: The migration is in progress. The
 *                button is enabled.
 *              This state has two phases because they should be the same from
 *              the UX point of view (the user expects the migration to start
 *              after clicking the button), but they are handled differently
 *              (creating an entity vs reading the stream of updates).
 * - done: The migration is done. The button acts as a link to the migration
 *         details page. More actions are available in the dropdown menu. The
 *         loading indicator is hidden.
 * - error: Any connectivity or API error occurred. Clicking the button will
 *          transition to the initializing state.
 * 
 * The state machine transitions are as follows:
 * 
 *     +-----------------------------(retry)-----------------  +-------+
 *     |  +---------------------(API error)----------------->  | error |
 *     V  |                                                    +-------+
 *                                                                  
 * +--------------+                              +-------+           ^ 
 * | initializing |  -(migration not running)->  | ready |           | 
 * +--------------+                              +-------+           | 
 *                                                                (API error)
 *  ^          \     \                               |  |            | 
 *  |           \     \                       (start migration)      | 
 *  |            \     \                             |  |            | 
 *  |             \    (migration is running)        v  +------------+
 *  |              \                 \                               |
 *  |               \                 \         +------------+--+    |
 *  |                \                 \        | requesting |  |  --+
 *  |                 \                 \       +----v-------+  |    |
 *  |                  \                 ---->  |   migrating   |  --+
 +  +-------------------*--(stop)-------------  +---------------+    |
 *  |                    \                                           |
 *  |                  (migration is done)           |               |
 *  |                               \            (finished)          |
 *  |                                \               |               |
 *  |                                 \              v               |
 *  |                                  \                             |
 *  |                                   \         +------+           |
 *  |                                    ------>  | done |           |
 *  |                                             +------+           |
 *  |                                                                |
 *  |                                                |  |            |
 *  |                                            (mark as read)      |
 *  +------------------------------------------------+  +------------+
 */
type State = 'initializing' | 'ready' | 'migrating' | 'done' | 'error'

/**
 * The hosts migration button component.
 * It allows to control the host reservation migration and track its progress.
 * It is built on top of the progress button component (mix of standard button,
 * progress bar, and drop down menu).
 *
 * The UI look and behavior depends on the current state of the migration.
 * The state and values of the related members are changed by the dedicated
 * transition methods that ensure the consistency of the UI.
 */
@Component({
    selector: 'app-hosts-migration-button',
    templateUrl: './hosts-migration-button.component.html',
    styleUrls: ['./hosts-migration-button.component.sass'],
})
export class HostsMigrationButtonComponent implements OnInit, OnDestroy {
    // Component members.

    /**
     * The label of the current UI state.
     */
    state: State

    /**
     * The current migration.
     * It is empty when the state is 'initializing', 'ready', or 'error'.
     */
    migration: Migration = null

    /**
     * The subscription to the migration updates.
     * It is defined while the migration is in progress.
     */
    updateSubscription: Subscription = null

    /**
     * Indicates whether the component is fetching data from the API.
     * The loading indicator is shown and the button is disabled when this is
     * true.
     */
    fetchingAPI: boolean = false

    /**
     * The current value of the host reservation filter in the hosts table.
     */
    currentFilter: QueryParamsFilter = null

    /**
     * The subscriptions subscribed by the component.
     * It doesn't include the subscription to the migration updates.
     */
    subscriptions: Subscription = new Subscription()

    /**
     * Indicates whether the confirmation dialog is shown.
     */
    showingConfirmation: boolean = false

    // Component inputs.

    /**
     * The observable that emits the current value of the host reservation
     * filter in the hosts table.
     */
    @Input() filter$: Observable<QueryParamsFilter>

    // Event emitters.

    /**
     * Emits the event to request the hosts table to filter the hosts.
     */
    @Output() filterList = new EventEmitter<QueryParamsFilter>()

    /**
     * Accepts the external services including the host reservation migration
     * service.
     * Binds the MenuItem callbacks to the component instance.
     */
    constructor(
        private router: Router,
        private messageService: MessageService,
        private migrationService: HostsMigrationService
    ) {
        // The MenuItem commands must be bound to the component instance.
        this.onStartMigrationClick = this.onStartMigrationClick.bind(this)
        this.onFilterErroredHostsClick = this.onFilterErroredHostsClick.bind(this)
        this.onFilterAffectedHostsClick = this.onFilterAffectedHostsClick.bind(this)
        this.onStopMigrationClick = this.onStopMigrationClick.bind(this)
        this.onMarkAsReadClick = this.onMarkAsReadClick.bind(this)
    }

    // State transitions.

    /**
     * Sets the current state and resets all related members.
     * Unsubscribes from the migration updates (if subscribed).
     * @param state The state label
     * @param migration The migration object to set (optional).
     */
    private setState(state: State, migration: Migration = null) {
        this.state = state
        this.migration = migration
        this.deregisterFromUpdates()
    }

    /**
     * Transitions to the 'initializing' state.
     * It fetches the current migration status from the API.
     *
     * If the migration is in progress, it transitions to the 'migrating'
     * state.
     * If the migration is done, it transitions to the 'done' state.
     * If there is no migration, it transitions to the 'ready' state.
     * If the fetching status failed, it transitions to the 'error' state.
     */
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

    /**
     * Transitions to the 'ready' state.
     * It waits for the user to start a new migration by click the button.
     */
    private transitionToReadyState() {
        this.setState('ready')
    }

    /**
     * Transitions to the first phase of the 'migrating' state.
     * It creates a new migration. Then it transitions to the second phase of
     * the 'migrating' state.
     * If the migration creation failed, it transitions to the 'error' state.
     */
    private transitionToMigrationRequestedState() {
        const filter = this.currentFilter
        this.setState('migrating', {
            errors: 0,
            inProgress: true,
            progress: 0,
            filter: filter,
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

    /**
     * Transitions to the second phase of the 'migrating' state.
     * It subscribes to the migration updates. Then it waits for the migration
     * to finish and transitions to the 'done' state.
     * @param migration The initial migration object.
     */
    private transitionToMigratingState(migration: Migration) {
        this.setState('migrating', migration)

        // Register for updates
        this.updateSubscription = this.migrationService.getMigrationUpdates().subscribe((m) => {
            this.migration = m
            if (!m.inProgress) {
                this.transitionToDoneState(m)
            }
        })
    }

    /**
     * Transitions to the 'done' state.
     * It waits for the user to click the button to mark the migration as read.
     * @param migration The final migration object.
     */
    private transitionToDoneState(migration: Migration) {
        this.setState('done', migration)
    }

    /**
     * Transitions to the 'error' state.
     * It generates a toast message and waits for the user to click the button
     * to retry.
     * @param err The related error object.
     */
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

    /**
     * Called when the component is initialized by Angular.
     * It subscribes to changes of the host reservation filter in the hosts
     * table (if the observable was provided). Then it transitions to the
     * 'initializing' state.
     */
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

    /**
     * Called when the component is destroyed by Angular.
     * It unsubscribes from all subscriptions.
     */
    ngOnDestroy(): void {
        this.deregisterFromUpdates()
        this.subscriptions.unsubscribe()
    }

    // UI event handlers.

    /**
     * Called when the user requests to start a migration. It shows a
     * confirmation dialog and then it transitions to the first phase of the
     * 'migrating' state.
     */
    onStartMigrationClick() {
        this.showingConfirmation = true
    }

    /**
     * Called when the user confirms the migration start.
     */
    onConfirmStartingMigrationClick() {
        this.showingConfirmation = false
        this.transitionToMigrationRequestedState()
    }

    /**
     * Called when the user cancels the migration start.
     */
    onCancelStartingMigrationClick() {
        this.showingConfirmation = false
    }

    /**
     * Called when the user requests to see the migration details page.
     * It redirects the user to the migration details page.
     */
    onRedirectToMigrationDetailsClick() {
        this.redirectToMigrationDetails()
    }

    /**
     * Called when the user requests to retry after an connectivity error.
     * It transitions to the 'initializing' state.
     */
    onRetryOnErrorClick() {
        this.transitionToInitializingState()
    }

    /**
     * Called when the user requests to see the hosts that failed to migrate.
     * It emits the event to filter the hosts table.
     */
    onFilterErroredHostsClick() {
        this.emitFilterList(this.migration.filter, true)
    }

    /**
     * Called when the user requests to see the hosts that were affected by the
     * migration.
     * It emits the event to filter the hosts table.
     */
    onFilterAffectedHostsClick() {
        this.emitFilterList(this.migration.filter, false)
    }

    /**
     * Called when the user requests to stop the migration.
     * It removes the migration and then it transitions to the 'initializing'
     * state. If the removal failed, it transitions to the 'error' state.
     */
    onStopMigrationClick() {
        this.removeMigration()
    }

    /**
     * Called when the user requests to mark the migration as read.
     * It removes the migration and then it transitions to the 'initializing'
     * state. If the removal failed, it transitions to the 'error' state.
     */
    onMarkAsReadClick() {
        this.removeMigration()
    }

    // Event emitters.

    /**
     * Emits the event to filter the hosts table.
     * @param filter The filter to apply.
     * @param errorsOnly Indicates whether to show only the hosts that failed.
     */
    private emitFilterList(filter: QueryParamsFilter, errorsOnly: boolean) {
        filter = this.buildFilter(filter, errorsOnly)
        this.filterList.emit(filter)
    }

    // Helpers.

    /**
     * Navigates to the migration details page.
     */
    private redirectToMigrationDetails() {
        this.router.navigate(['/hosts/migration'])
    }

    /**
     * Unsubscribes from the migration updates (if subscribed).
     */
    private deregisterFromUpdates() {
        if (this.updateSubscription) {
            this.updateSubscription.unsubscribe()
            this.updateSubscription = null
        }
    }

    /**
     * Removes the current migration.
     * It transitions to the 'initializing' state. If the removal failed, it
     * transitions to the 'error' state.
     */
    private removeMigration() {
        if (!this.migration) {
            // Nothing to do.
            return
        }

        // Unsubscribe from updates.
        this.deregisterFromUpdates()

        // Remove the migration.
        this.fetchingAPI = true
        lastValueFrom(this.migrationService.removeMigration())
            .finally(() => {
                this.fetchingAPI = false
            })
            .then(() => {
                this.transitionToInitializingState()
            })
            .catch((err) => {
                this.transitionToErrorState(err)
            })
    }

    /**
     * Builds the filter to apply to the hosts table.
     * @param base The base filter.
     * @param errorsOnly Indicates whether to show only the hosts that failed.
     * @returns The filter to apply.
     */
    private buildFilter(base: QueryParamsFilter, errorsOnly: boolean): QueryParamsFilter {
        // Copy the filter object.
        base = { ...base }
        if (errorsOnly) {
            throw new Error('Not implemented')
        }
        return base
    }
}
