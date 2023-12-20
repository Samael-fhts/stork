import { Meta, StoryObj, applicationConfig, moduleMetadata } from '@storybook/angular'
import { HostsMigrationButtonComponent } from './hosts-migration-button.component'
import { ButtonModule } from 'primeng/button'
import { SplitButtonModule } from 'primeng/splitbutton'
import { MenuModule } from 'primeng/menu'
import { BadgeModule } from 'primeng/badge'
import { BrowserAnimationsModule } from '@angular/platform-browser/animations'
import { ProgressButtonComponent } from '../progress-button/progress-button.component'
import { toastDecorator } from '../utils-stories'
import { HostsMigrationService, Migration } from '../hosts-migration-service/hosts-migration.service'
import { ToastModule } from 'primeng/toast'
import { ConfirmationService, MessageService } from 'primeng/api'
import {
    Observable,
    concatMap,
    delay,
    generate,
    ignoreElements,
    interval,
    map,
    merge,
    of,
    throwError,
    timer,
} from 'rxjs'
import { DialogModule } from 'primeng/dialog'
import { RouterModule } from '@angular/router'

/**
 * A mock of the HostsMigrationService. It replaces the real service in the
 * component's stories.
 * We use the `Partial` type to keep consistency with the real service.
 * Note, the mock class isn't annotated with `@Injectable` decorator. It don't
 * know it is necessary; it works without it.
 */
class MockHostsMigrationService implements Partial<HostsMigrationService> {
    private startCount = 0

    /**
     * There is no migration in progress.
     * Delayed by 5s to simulate a backend call.
     */
    getCurrentMigration(): Observable<Migration> {
        return of(null as Migration).pipe(delay(5000))
    }

    /**
     * Returns a stream of migration updates.
     * The updates are generated every 250ms.
     */
    getMigrationUpdates(): Observable<Migration> {
        const progress$ = generate({
            initialState: 0,
            condition: (i) => i <= 100,
            iterate: (i) => i + 1,
            resultSelector: (i) =>
                ({
                    progress: i / 100,
                    errors: Math.round(i / 10),
                    inProgress: i !== 100,
                    filter: `filter-${i}`,
                }) as Migration,
        })

        const bound$ = timer(250).pipe(ignoreElements())
        return progress$.pipe(concatMap((v) => merge(of(v), bound$)))
    }

    /**
     * Delayed by 2s to simulate a backend call.
     */
    removeMigration(): Observable<void> {
        return of(null).pipe(delay(2000))
    }

    /**
     * Returns an initial migration status after 5s.
     * Fails every 3rd call.
     */
    startMigration(filter: string): Observable<Migration> {
        this.startCount += 1
        if (this.startCount % 3 === 0) {
            return throwError(() => new Error('Could not start the migration')).pipe(delay(2000))
        }
        return of({
            id: this.startCount,
            errors: 0,
            inProgress: true,
            progress: 0,
            filter: filter,
        } as Migration).pipe(delay(5000))
    }
}

/**
 * Describes the component's arguments.
 */
interface Args {
    filter$: Observable<string>
}

/**
 * FYI: This file doesn't use story template to define stories because it's
 * deprecated and will be removed in the future. Instead, it uses the StoryObj
 * type introduced by CSF3 format (previous solution was compliant with CSF2).
 * It is a first component in the project that uses the new format.
 *
 * This Meta object uses also a different approach to mock the service. Instead
 * of mocking HTTP calls by the `storybook-addon-mock` plugin features, it
 * provides a mock service directly to the component. This approach seems to be
 * more simple if the component makes many various API calls.
 */
export default {
    title: 'App/HostsMigrationButton',
    component: HostsMigrationButtonComponent,
    argTypes: {
        // This property has a complex type, so it cannot be defined manually.
        filter$: {
            table: {
                disable: true,
            },
        },
    },
    decorators: [
        applicationConfig({
            providers: [MessageService],
        }),
        moduleMetadata({
            imports: [
                ButtonModule,
                SplitButtonModule,
                MenuModule,
                BadgeModule,
                BrowserAnimationsModule,
                ToastModule,
                DialogModule,
                RouterModule
            ],
            declarations: [ProgressButtonComponent],
            providers: [
                // Provide a mock service instead of the real one.
                {
                    provide: HostsMigrationService,
                    useClass: MockHostsMigrationService,
                },
            ],
        }),
        toastDecorator,
    ],
} as Meta<Args>

/**
 * The primary story. The component starts in the 'initializing' state.
 */
export const Primary: StoryObj<Args> = {
    args: {
        // Generates a new filter every 3s.
        filter$: interval(3000).pipe(map((v) => `filter-${v}`)),
    },
}
