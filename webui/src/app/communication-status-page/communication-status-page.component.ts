import { Component, OnInit } from '@angular/core'
import { App, ServicesService } from '../backend'
import { lastValueFrom, of, throwError, timer } from 'rxjs'
import { getErrorMessage } from '../utils'
import { MessageService } from 'primeng/api'
import { switchMap } from 'rxjs/operators'

type TestT = { id: number; name: string; content: string }

/**
 * A component displaying a page showing the communication issues with
 * the monitored apps.
 *
 * It fetches the list of communication issues from the Stork server and
 * displays them as a tree using the CommunicationStatusTreeComponent.
 *
 * The list can be reloaded on demand by pressing a button.
 */
@Component({
    selector: 'app-communication-status-page',
    standalone: false,
    templateUrl: './communication-status-page.component.html',
    styleUrl: './communication-status-page.component.sass',
})
export class CommunicationStatusPageComponent implements OnInit {
    /**
     * Configures the breadcrumbs for the component.
     */
    breadcrumbs = [{ label: 'Monitoring' }, { label: 'Communication' }]

    /**
     * A list of communication issues fetched from the server.
     */
    apps: Array<App> = []

    /**
     * A boolean flag indicating if the data are being loaded.
     */
    loading = true

    tabOptions: TestT[] = [
        {
            name: 'test 1',
            id: 1,
            content: 'test content 1',
        },
        {
            name: 'test 2',
            id: 2,
            content: 'test content 2',
        },
    ]
    testComm: TestT
    childCallable = (n) => {
        console.log('called callable from child with arg', n)
        return this.onGetEntity(n)
    }

    /**
     * Constructor.
     *
     * @param messageService message service used for displaying communication
     *        errros with the Stork server.
     * @param servicesService a service used for fetching the communication issues
     *        from the Stork server.
     */
    constructor(
        private messageService: MessageService,
        private servicesService: ServicesService
    ) {}

    /**
     * A component lifecycle hook invoked when the component is loaded.
     *
     * It fetches the list of communication issues from the server.
     */
    ngOnInit(): void {
        this.reload()
    }

    /**
     * Reloads the list of the communication issues from the server.
     *
     * It is called during the component initialization and on demand, when
     * the reload button is pressed.
     */
    private reload(): void {
        this.loading = true
        lastValueFrom(this.servicesService.getAppsWithCommunicationIssues())
            .then((data) => {
                this.apps = data.items || []
            })
            .catch((err) => {
                const msg = getErrorMessage(err)
                this.messageService.add({
                    severity: 'error',
                    summary: 'Cannot create new transaction',
                    detail: 'Failed to create transaction for adding new host: ' + msg,
                    life: 10000,
                })
                this.apps = []
            })
            .finally(() => {
                this.loading = false
            })
    }

    /**
     * Reloads the list of communication issues on demand.
     *
     * It contacts the Stork server to fetch the list of communication issues.
     */
    onReload(): void {
        this.reload()
    }

    onGetEntity(event: number) {
        return lastValueFrom(
            timer(300).pipe(
                switchMap(() => {
                    if (event > 4) {
                        return throwError(() => new Error('there is no such entity'))
                    }
                    return of({
                        id: event,
                        name: `Test ent ${event}`,
                        content: `Test content of entity no ${event}`,
                    } as TestT)
                })
            )
        )
    }
}
