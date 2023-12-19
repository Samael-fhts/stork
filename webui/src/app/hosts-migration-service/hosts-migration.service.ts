import { Injectable } from '@angular/core';
import { Observable, throwError } from 'rxjs';


export interface Migration {
  id: number
  progress: number
  errors: number
  inProgress: boolean
  filter: string
}


@Injectable({
  providedIn: 'root'
})
export class HostsMigrationService {

  constructor() { }

  getCurrentMigration(): Observable<Migration> {
    return throwError(() => new Error('Not implemented'));
  }

  startMigration(filter: string): Observable<Migration> {
    return throwError(() => new Error('Not implemented'));
  }

  removeMigration(migrationId: number): Observable<void> {
    return throwError(() => new Error('Not implemented'));
  }

  getMigrationUpdates(migrationId: number): Observable<Migration> {
    return throwError(() => new Error('Not implemented'));
  }
}
