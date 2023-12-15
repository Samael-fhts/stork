import { Meta, StoryObj, applicationConfig, moduleMetadata } from '@storybook/angular'
import { HostsMigrationButtonComponent } from './hosts-migration-button.component'
import { ButtonModule } from 'primeng/button'
import { SplitButtonModule } from 'primeng/splitbutton'
import { MenuModule } from 'primeng/menu'
import { BadgeModule } from 'primeng/badge'
import { NoopAnimationsModule } from '@angular/platform-browser/animations'
import { ProgressButtonComponent } from '../progress-button/progress-button.component'

interface Args {}

export default {
    title: 'App/HostsMigrationButton',
    component: HostsMigrationButtonComponent,
    decorators: [
        applicationConfig({
            providers: [],
        }),
        moduleMetadata({
            imports: [ButtonModule, SplitButtonModule, MenuModule, BadgeModule, NoopAnimationsModule],
            declarations: [ProgressButtonComponent],
        }),
    ],
} as Meta<Args>

export const Primary: StoryObj<Args> = {}
