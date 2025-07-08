import { Component, OnDestroy, OnInit } from '@angular/core'
import { Tab, TabList, TabPanel, TabPanels, Tabs } from 'primeng/tabs'
import { ActivatedRoute, Router, RouterLink } from '@angular/router'
import { inject } from '@angular/core'
import { Subscription } from 'rxjs'
import { MessageService } from 'primeng/api'
import { TimesIcon } from 'primeng/icons'

export type StorkTab = { title: string; value: number; content: string; route: string }

@Component({
    selector: 'app-stork-tab-view',
    standalone: true,
    imports: [Tabs, TabList, Tab, TabPanels, TabPanel, RouterLink, TimesIcon],
    templateUrl: './stork-tab-view.component.html',
    styleUrl: './stork-tab-view.component.sass',
})
export class StorkTabViewComponent implements OnInit, OnDestroy {
    /**
     * TODO: input
     */
    openTabs: StorkTab[] = []

    tabOptions: StorkTab[] = [
        { title: 'Tab 2', value: 2, content: 'Tab 2 Content', route: '/communication/2' },
        { title: 'Tab 1', value: 1, content: 'Tab 1 Content', route: '/communication/1' },
    ]

    activeTabEntityID: number = 0

    private readonly route = inject(ActivatedRoute)

    private readonly messageService = inject(MessageService)

    private readonly router = inject(Router)

    private subscriptions: Subscription

    /**
     *
     * @param entityID
     */
    openTab(entityID: number) {
        // console.log('openTab', entityID)
        if (entityID === this.activeTabEntityID) {
            console.log('openTab', entityID, 'this tab is already active')
            return
        }
        const indexOfOpenTab = this.openTabs.findIndex((tab) => tab.value === entityID)
        if (indexOfOpenTab > -1) {
            console.log('openTab', entityID, 'this tab is already open, switch active tab')
            // this.router.navigate(['/communication', indexOfOpenTab])
            this.activeTabEntityID = entityID
            return
        }

        console.log('openTab', entityID, 'need to fetch data and create new tab')
        const entityToOpen = this.tabOptions.find((tab) => tab.value === entityID)
        if (!entityToOpen) {
            this.messageService.add({
                detail: `Couldn't find tab to open with id ${entityID}!`,
                severity: 'error',
                summary: `Error opening tab`,
            })
            return
        }

        this.openTabs = [...this.openTabs, entityToOpen]
        this.activeTabEntityID = entityID
    }

    /**
     *
     * @param entityID
     */
    closeTab(entityID: number) {
        const activeTabIndex = this.openTabs.findIndex((tab) => tab.value === this.activeTabEntityID)
        const tabToCloseIndex = this.openTabs.findIndex((tab) => tab.value === entityID)
        if (tabToCloseIndex > -1) {
            this.openTabs.splice(tabToCloseIndex, 1)
            if (tabToCloseIndex <= activeTabIndex) {
                // activate first tab
                this.router.navigate(['/communication'])
            }
        }
    }

    /**
     *
     */
    ngOnInit(): void {
        console.log('storkTabViewComponent onInit')

        this.openTabs = [
            // { title: 'Tab All', value: 0, content: 'Tab All Content - table' },
        ]

        this.subscriptions = this.route.paramMap.subscribe({
            next: (params) => {
                console.log('ActivatedRoute paramMap emits next', params)
                const id = params.get('id')
                if (!id || id === 'all') {
                    // console.log('open first tab')
                    this.activeTabEntityID = 0
                    return
                }
                const numericId = parseInt(id, 10)
                if (!Number.isNaN(numericId)) {
                    this.openTab(numericId)
                    return
                } else {
                    this.messageService.add({
                        detail: `Couldn't parse provided id ${id} to numeric value!`,
                        severity: 'error',
                        summary: `Error opening tab`,
                    })
                    this.activeTabEntityID = 0
                    return
                }
            },
            error: (err) => {
                console.log('error emitted by ActivatedRoute paramMap', err)
            },
            complete: () => {
                console.log('ActivatedRoute paramMap complete')
            },
        })
    }

    ngOnDestroy(): void {
        console.log('storkTabViewComponent onDestroy')
        this.subscriptions.unsubscribe()
    }
}
