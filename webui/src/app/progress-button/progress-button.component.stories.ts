import { Meta, StoryObj, applicationConfig, moduleMetadata } from '@storybook/angular'
import { ProgressButtonComponent } from './progress-button.component'
import { ButtonModule } from 'primeng/button'
import { NoopAnimationsModule } from '@angular/platform-browser/animations'
import { SplitButtonModule } from 'primeng/splitbutton'
import { MenuItem } from 'primeng/api'
import { MenuModule } from 'primeng/menu'
import { BadgeModule } from 'primeng/badge'
import { action } from '@storybook/addon-actions'

interface Args {
    value: number
    badgeCount: number
    label: string
    styleClass: string
    progressing: boolean
    disabled: boolean
    model: MenuItem[]
}

export default {
    title: 'App/ProgressButton',
    component: ProgressButtonComponent,
    decorators: [
        applicationConfig({
            providers: [],
        }),
        moduleMetadata({
            imports: [ButtonModule, SplitButtonModule, MenuModule, BadgeModule, NoopAnimationsModule],
        }),
    ],
    argTypes: {
        badgeCount: { control: 'number' },
        value: { control: { type: 'number', min: 0, max: 1, step: 0.1 } },
        label: { control: 'text' },
        progressing: { control: 'boolean' },
        disabled: { control: 'boolean' },
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
                'unknown',
            ],
        },
        onClick: {
            action: 'onClick',
        },
    },
} as Meta<Args>

export const Primary: StoryObj<Args> = {
    args: {
        value: 0.25,
        badgeCount: 7,
        label: 'Progress button',
        styleClass: 'p-button-success',
    },
}

export const SplitButton: StoryObj<Args> = {
    args: {
        value: 0.25,
        badgeCount: 7,
        label: 'Progress button',
        styleClass: 'p-button-success',
        model: [
            {
                label: 'Save',
                icon: 'pi pi-check',
                command: () => {
                    action('Save')()
                },
            },
            {
                label: 'Update',
                icon: 'pi pi-refresh',
                command: () => {
                    action('Update')()
                },
            },
            {
                label: 'Delete',
                icon: 'pi pi-trash',
                command: () => {
                    action('Delete')()
                },
            },
        ],
    },
}
