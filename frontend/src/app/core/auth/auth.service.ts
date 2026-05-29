import { HttpClient } from '@angular/common/http';
import { computed, inject, Injectable, signal } from '@angular/core';
import { catchError, finalize, Observable, of, tap } from 'rxjs';

import { environment } from '../../../environments/environment';
import { apiErrorMessage } from '../api/api-error';
import { User } from './auth.models';

@Injectable({ providedIn: 'root' })
export class AuthService {
  private readonly http = inject(HttpClient);
  private readonly baseUrl = `${environment.apiBaseUrl}/api/auth`;
  private readonly userState = signal<User | null>(null);
  private readonly loadingState = signal(true);
  private readonly errorState = signal<string | null>(null);

  readonly user = this.userState.asReadonly();
  readonly loading = this.loadingState.asReadonly();
  readonly error = this.errorState.asReadonly();
  readonly isAuthenticated = computed(() => this.userState() !== null);

  loadSession(): Observable<User | null> {
    this.loadingState.set(true);
    this.errorState.set(null);

    return this.http.get<User>(`${this.baseUrl}/me`, { withCredentials: true }).pipe(
      tap((user) => this.userState.set(user)),
      catchError((error) => {
        const status = typeof error === 'object' && error !== null && 'status' in error ? error.status : undefined;
        this.userState.set(null);
        if (status !== 401) {
          this.errorState.set(apiErrorMessage(error, 'Could not load your session'));
        }
        return of(null);
      }),
      finalize(() => this.loadingState.set(false)),
    );
  }

  login(): void {
    window.location.href = `${this.baseUrl}/login`;
  }

  logout(): Observable<void> {
    this.errorState.set(null);
    return this.http.post<void>(`${this.baseUrl}/logout`, null, { withCredentials: true }).pipe(
      tap(() => this.userState.set(null)),
      catchError((error) => {
        this.errorState.set(apiErrorMessage(error, 'Could not log out'));
        throw error;
      }),
    );
  }
}
