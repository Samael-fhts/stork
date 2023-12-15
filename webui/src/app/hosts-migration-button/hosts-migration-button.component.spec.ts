import { ComponentFixture, TestBed } from '@angular/core/testing'

import { HostsMigrationButtonComponent } from './hosts-migration-button.component'

describe('HostsMigrationButtonComponent', () => {
    let component: HostsMigrationButtonComponent
    let fixture: ComponentFixture<HostsMigrationButtonComponent>

    beforeEach(() => {
        TestBed.configureTestingModule({
            declarations: [HostsMigrationButtonComponent],
        })
        fixture = TestBed.createComponent(HostsMigrationButtonComponent)
        component = fixture.componentInstance
        fixture.detectChanges()
    })

    it('should create', () => {
        expect(component).toBeTruthy()
    })
})
