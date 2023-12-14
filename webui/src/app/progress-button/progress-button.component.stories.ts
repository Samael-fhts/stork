import { Meta, StoryFn, StoryObj, applicationConfig, moduleMetadata } from '@storybook/angular'
import { ProgressButtonComponent } from './progress-button.component'
import { ButtonModule } from 'primeng/button'
import { NoopAnimationsModule } from '@angular/platform-browser/animations'
import { action } from '@storybook/addon-actions'

interface Args {
    value: number
    badgeCount: number
    label: string
    styleClass: string
    enabled: boolean
}

export default {
    title: 'App/ProgressButton',
    component: ProgressButtonComponent,
    decorators: [
        applicationConfig({
            providers: [],
        }),
        moduleMetadata({
            imports: [ButtonModule, NoopAnimationsModule],
        }),
    ],
    argTypes: {
        errorCount: { control: 'number' },
        value: { control: { type: 'number', min: 0, max: 1, step: 0.1 } },
        label: { control: 'text' },
        enabled: { control: 'boolean' },
        styleClass: {
            control: 'select',
            options: [
                '',
                'p-button-primary',
                'p-button-success',
                'p-button-info',
                'p-button-warning',
                'p-button-danger',
                'p-button-secondary',
                'unknown'
            ],
        },
        onClick: { action: 'clicked' },
    },
} as Meta<Args>

export const Primary: StoryObj<Args> = {
    args: {
        value: 0.25,
        badgeCount: 7,
        label: 'Progress button',
        styleClass: 'p-button-success'
    },
}
