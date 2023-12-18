import { TestBed } from '@angular/core/testing';

import { HostsMigrationService } from './hosts-migration.service';

describe('HostsMigrationServiceService', () => {
  let service: HostsMigrationService;

  beforeEach(() => {
    TestBed.configureTestingModule({});
    service = TestBed.inject(HostsMigrationService);
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });
});
