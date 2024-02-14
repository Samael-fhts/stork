import { Meta, StoryObj, applicationConfig, moduleMetadata } from '@storybook/angular'
import { HostsMigrationButtonComponent } from './hosts-migration-button.component'
import { ButtonModule } from 'primeng/button'
import { SplitButtonModule } from 'primeng/splitbutton'
import { MenuModule } from 'primeng/menu'
import { BadgeModule } from 'primeng/badge'
import { BrowserAnimationsModule } from '@angular/platform-browser/animations'
import { ProgressButtonComponent } from '../progress-button/progress-button.component'
import { toastDecorator } from '../utils-stories'
import { HostsMigrationService } from '../hosts-migration-service/hosts-migration.service'
import { ToastModule } from 'primeng/toast'
import { DialogModule } from 'primeng/dialog'
import { RouterModule } from '@angular/router'
import { MessageService } from 'primeng/api'
import { MockHostsMigrationService } from '../hosts-migration-service/hosts-migration-mock.service'
import { PlaceholderPipe } from '../pipes/placeholder.pipe'
import { QueryParamsFilter } from '../hosts-page/query-params-filter'

/**
 * Describes the component's arguments.
 */
interface Args {
    filter: QueryParamsFilter
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
                RouterModule,
            ],
            declarations: [ProgressButtonComponent, PlaceholderPipe],
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
        filter: {
            appId: 42,
            conflict: true,
            global: false,
            keaSubnetId: 24,
            subnetId: 4224,
            text: 'foobar',
        },
    },
}
