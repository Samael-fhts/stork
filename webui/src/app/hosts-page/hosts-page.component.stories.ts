import { Meta, StoryObj, applicationConfig, moduleMetadata } from '@storybook/angular'
import { HostsPageComponent } from './hosts-page.component'
import { HttpClientModule } from '@angular/common/http'
import { MessageService } from 'primeng/api'
import { ToastModule } from 'primeng/toast'
import { BrowserAnimationsModule, NoopAnimationsModule } from '@angular/platform-browser/animations'
import { DHCPService, Hosts } from '../backend'
import { toastDecorator } from '../utils-stories'
import { RouterTestingModule } from '@angular/router/testing'
import { HelpTipComponent } from '../help-tip/help-tip.component'
import { ButtonModule } from 'primeng/button'
import { ChipsModule } from 'primeng/chips'
import { DividerModule } from 'primeng/divider'
import { FormsModule, ReactiveFormsModule } from '@angular/forms'
import { TableModule } from 'primeng/table'
import { TabMenuModule } from 'primeng/tabmenu'
import { BreadcrumbModule } from 'primeng/breadcrumb'
import { OverlayPanelModule } from 'primeng/overlaypanel'
import { TooltipModule } from 'primeng/tooltip'
import { FieldsetModule } from 'primeng/fieldset'
import { ProgressSpinnerModule } from 'primeng/progressspinner'
import { ToggleButtonModule } from 'primeng/togglebutton'
import { CheckboxModule } from 'primeng/checkbox'
import { DropdownModule } from 'primeng/dropdown'
import { MultiSelectModule } from 'primeng/multiselect'
import { ConfirmDialogModule } from 'primeng/confirmdialog'
import { TreeModule } from 'primeng/tree'
import { TagModule } from 'primeng/tag'
import { MessagesModule } from 'primeng/messages'
import { Observable, of } from 'rxjs'
import { HostsMigrationButtonComponent } from '../hosts-migration-button/hosts-migration-button.component'
import { ProgressButtonComponent } from '../progress-button/progress-button.component'
import { DialogModule } from 'primeng/dialog'
import { BreadcrumbsComponent } from '../breadcrumbs/breadcrumbs.component'
import { IdentifierComponent } from '../identifier/identifier.component'
import { HostsMigrationService } from '../hosts-migration-service/hosts-migration.service'
import { MockHostsMigrationService } from '../hosts-migration-service/hosts-migration-mock.service'
import { PlaceholderPipe } from '../pipes/placeholder.pipe'
import { SplitButtonModule } from 'primeng/splitbutton'
import { HostDataSourceLabelComponent } from '../host-data-source-label/host-data-source-label.component'
import { EntityLinkComponent } from '../entity-link/entity-link.component'

/**
 * Mocks the HostsMigrationService to use in the component's stories.
 */
class MockDHCPService implements Partial<DHCPService> {
    /**
     * Returns a fixed list of hosts.
     */
    getHosts() /** Not used arguments:
     * start?: number,
     * limit?: number,
     * appId?: number,
     * subnetId?: number,
     * localSubnetId?: number,
     * text?: string,
     * global?: boolean,
     * conflict?: boolean,
     * observe: any = 'body',
     * reportProgress: boolean = false,
     * options?: { httpHeaderAccept?: 'application/json' }
     */
    : Observable<any> {
        return of({
            total: 1,
            items: [
                {
                    id: 1,
                    hostIdentifiers: [
                        {
                            idType: 'duid',
                            idHexValue: '01:02:03:04',
                        },
                    ],
                    addressReservations: [
                        {
                            address: '192.0.2.1',
                        },
                    ],
                    localHosts: [
                        {
                            appId: 1,
                            appName: 'frog',
                            dataSource: 'config',
                        },
                    ],
                },
                {
                    id: 2,
                    hostIdentifiers: [
                        {
                            idType: 'duid',
                            idHexValue: '11:12:13:14',
                        },
                    ],
                    addressReservations: [
                        {
                            address: '192.0.2.2',
                        },
                    ],
                    localHosts: [
                        {
                            appId: 2,
                            appName: 'mouse',
                            dataSource: 'config',
                        },
                    ],
                },
            ],
        } as Hosts)
    }
}

/**
 * Describes the component's arguments.
 */
interface Args {}

export default {
    title: 'App/HostsPage',
    component: HostsPageComponent,
    argTypes: {},
    decorators: [
        applicationConfig({
            providers: [MessageService],
        }),
        moduleMetadata({
            imports: [
                BrowserAnimationsModule,
                ButtonModule,
                ChipsModule,
                DividerModule,
                FormsModule,
                TableModule,
                HttpClientModule,
                RouterTestingModule.withRoutes([
                    {
                        path: 'dhcp/hosts',
                        pathMatch: 'full',
                        redirectTo: 'dhcp/hosts/all',
                    },
                    {
                        path: 'dhcp/hosts/:id',
                        component: HostsPageComponent,
                    },
                ]),
                ,
                TabMenuModule,
                BreadcrumbModule,
                OverlayPanelModule,
                NoopAnimationsModule,
                TooltipModule,
                FormsModule,
                FieldsetModule,
                ProgressSpinnerModule,
                TableModule,
                ToggleButtonModule,
                ButtonModule,
                CheckboxModule,
                DropdownModule,
                FieldsetModule,
                MultiSelectModule,
                ReactiveFormsModule,
                ConfirmDialogModule,
                TreeModule,
                TagModule,
                MessagesModule,
                ToastModule,
                DialogModule,
                SplitButtonModule,
            ],
            declarations: [
                HelpTipComponent,
                HostsMigrationButtonComponent,
                ProgressButtonComponent,
                BreadcrumbsComponent,
                IdentifierComponent,
                PlaceholderPipe,
                HostDataSourceLabelComponent,
                EntityLinkComponent,
            ],
            providers: [
                {
                    provide: DHCPService,
                    useClass: MockDHCPService,
                },
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
 * The basic story.
 */
export const Primary: StoryObj<Args> = {
    args: {},
}
