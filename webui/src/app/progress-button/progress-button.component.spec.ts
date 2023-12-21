import { ComponentFixture, TestBed } from '@angular/core/testing'

import { ProgressButtonComponent } from './progress-button.component'
import { SplitButtonModule } from 'primeng/splitbutton'
import { ButtonModule } from 'primeng/button'

describe('ProgressButtonComponent', () => {
    let component: ProgressButtonComponent
    let fixture: ComponentFixture<ProgressButtonComponent>

    beforeEach(() => {
        TestBed.configureTestingModule({
            declarations: [ProgressButtonComponent],
            imports: [ButtonModule, SplitButtonModule],
        })
        fixture = TestBed.createComponent(ProgressButtonComponent)
        component = fixture.componentInstance
        fixture.detectChanges()
    })

    it('should create', () => {
        expect(component).toBeTruthy()
    })
})
