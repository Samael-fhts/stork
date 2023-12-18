import { Meta, StoryObj, applicationConfig, moduleMetadata } from '@storybook/angular'
import { HostsMigrationButtonComponent } from './hosts-migration-button.component'
import { ButtonModule } from 'primeng/button'
import { SplitButtonModule } from 'primeng/splitbutton'
import { MenuModule } from 'primeng/menu'
import { BadgeModule } from 'primeng/badge'
import { NoopAnimationsModule } from '@angular/platform-browser/animations'
import { ProgressButtonComponent } from '../progress-button/progress-button.component'
import { toastDecorator } from '../utils-stories'
import { HostsMigrationService, Migration } from '../hosts-migration-service/hosts-migration.service'
import { ToastModule } from 'primeng/toast'
import { MessageService } from 'primeng/api'
import { Observable, concatMap, delay, generate, ignoreElements, interval, merge, of, throwError, timer } from 'rxjs'


class MockHostsMigrationService implements Partial<HostsMigrationService> {
    private startCount = 0

    cancelMigration(migrationId: number): Observable<void> {
        return of(null).pipe(delay(2000));
    }

    getCurrentMigration(): Observable<Migration> {
        return of(null as Migration).pipe(delay(5000));
    }

    getMigrationUpdates(migrationId: number): Observable<Migration> {
        const progress$ = generate({
            initialState: 0,
            condition: (i) => i <= 100,
            iterate: (i) => i + 1,
            resultSelector: (i) => ({
                id: migrationId,
                progress: i/100,
                errors: Math.round(i / 10),
                inProgress: i !== 100,
            } as Migration),
        })

        const bound$ = timer(250).pipe(ignoreElements())
        return progress$.pipe(
            concatMap(v => merge(of(v), bound$))
        )
    }

    removeMigration(migrationId: number): Observable<void> {
        return of(null).pipe(delay(2000));
    }

    startMigration(): Observable<Migration> {
        this.startCount += 1
        if (this.startCount % 3 === 0) {
            return throwError(() => new Error('Could not start the migration')).
                pipe(delay(2000))
        }
        return of({
            id: this.startCount,
            errors: 0,
            inProgress: true,
            progress: 0,
        } as Migration).pipe(delay(5000))
    }
}

interface Args {}

export default {
    title: 'App/HostsMigrationButton',
    component: HostsMigrationButtonComponent,
    decorators: [
        applicationConfig({
            providers: [
                MessageService,
            ],
        }),
        moduleMetadata({
            imports: [ButtonModule, SplitButtonModule, MenuModule, BadgeModule, NoopAnimationsModule, ToastModule],
            declarations: [ProgressButtonComponent],
            providers: [
                {
                    provide: HostsMigrationService,
                    useClass: MockHostsMigrationService,
                }
            ]
        }),
        toastDecorator,
    ],
} as Meta<Args>

export const Primary: StoryObj<Args> = {}
