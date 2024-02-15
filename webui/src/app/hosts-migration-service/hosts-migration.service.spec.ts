import { TestBed } from '@angular/core/testing'

import { HostsMigrationService } from './hosts-migration.service'
import { HttpClientTestingModule } from '@angular/common/http/testing'

describe('HostsMigrationServiceService', () => {
    let service: HostsMigrationService

    beforeEach(() => {
        TestBed.configureTestingModule({
            imports: [HttpClientTestingModule],
        })
        service = TestBed.inject(HostsMigrationService)
    })

    it('should be created', () => {
        expect(service).toBeTruthy()
    })
})
