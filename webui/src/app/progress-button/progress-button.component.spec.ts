import { ComponentFixture, TestBed } from '@angular/core/testing'

import { ProgressButtonComponent } from './progress-button.component'

describe('ProgressButtonComponent', () => {
    let component: ProgressButtonComponent
    let fixture: ComponentFixture<ProgressButtonComponent>

    beforeEach(() => {
        TestBed.configureTestingModule({
            declarations: [ProgressButtonComponent],
        })
        fixture = TestBed.createComponent(ProgressButtonComponent)
        component = fixture.componentInstance
        fixture.detectChanges()
    })

    it('should create', () => {
        expect(component).toBeTruthy()
    })
})
