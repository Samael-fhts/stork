import { ComponentFixture, TestBed } from '@angular/core/testing'

import { StorkTabViewComponent } from './stork-tab-view.component'

describe('StorkTabViewComponent', () => {
    let component: StorkTabViewComponent
    let fixture: ComponentFixture<StorkTabViewComponent>

    beforeEach(async () => {
        await TestBed.configureTestingModule({
            imports: [StorkTabViewComponent],
        }).compileComponents()

        fixture = TestBed.createComponent(StorkTabViewComponent)
        component = fixture.componentInstance
        fixture.detectChanges()
    })

    it('should create', () => {
        expect(component).toBeTruthy()
    })
})
