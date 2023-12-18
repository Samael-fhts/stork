import { Injectable } from '@angular/core';
import { Observable, throwError } from 'rxjs';


interface Migration {
  id: number
  progress: number
  errors: number
  inProgress: boolean
}


@Injectable({
  providedIn: 'root'
})
export class HostsMigrationServiceService {

  constructor() { }

  getCurrentMigration(): Observable<Migration> {
    return throwError(() => new Error('Not implemented'));
  }

  startMigration(): Observable<Migration> {
    return throwError(() => new Error('Not implemented'));
  }

  cancelMigration(migrationId: number): Observable<Migration> {
    return throwError(() => new Error('Not implemented'));
  }

  removeMigration(migrationId: number): Observable<void> {
    return throwError(() => new Error('Not implemented'));
  }

  getMigrationUpdates(migrationId: number): Observable<Migration> {
    return throwError(() => new Error('Not implemented'));
  }
}
