import { TestBed } from '@angular/core/testing';

import { HostsMigrationServiceService } from './hosts-migration-service.service';

describe('HostsMigrationServiceService', () => {
  let service: HostsMigrationServiceService;

  beforeEach(() => {
    TestBed.configureTestingModule({});
    service = TestBed.inject(HostsMigrationServiceService);
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });
});
