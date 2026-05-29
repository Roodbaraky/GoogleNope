import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';

import { environment } from '../../../environments/environment';
import { AuthService } from './auth.service';
import { User } from './auth.models';

describe('AuthService', () => {
  let service: AuthService;
  let http: HttpTestingController;

  const user: User = {
    id: 'oauth:user-1',
    email: 'user@example.com',
    name: 'User One',
    picture: '',
  };

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [provideHttpClient(), provideHttpClientTesting()],
    });

    service = TestBed.inject(AuthService);
    http = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    http.verify();
  });

  it('loads the current session', () => {
    let loaded: unknown;

    service.loadSession().subscribe((result) => {
      loaded = result;
    });

    const request = http.expectOne(`${environment.apiBaseUrl}/api/auth/me`);
    expect(request.request.method).toBe('GET');
    expect(request.request.withCredentials).toBeTrue();

    request.flush(user);

    expect(loaded).toEqual(user);
    expect(service.user()).toEqual(user);
    expect(service.loading()).toBeFalse();
    expect(service.error()).toBeNull();
    expect(service.isAuthenticated()).toBeTrue();
  });

  it('treats an unauthorized session check as signed out without surfacing an error', () => {
    service.loadSession().subscribe();

    const request = http.expectOne(`${environment.apiBaseUrl}/api/auth/me`);
    request.flush({ error: 'Unauthorized' }, { status: 401, statusText: 'Unauthorized' });

    expect(service.user()).toBeNull();
    expect(service.loading()).toBeFalse();
    expect(service.error()).toBeNull();
    expect(service.isAuthenticated()).toBeFalse();
  });

  it('clears the session on logout', () => {
    service.loadSession().subscribe();
    http.expectOne(`${environment.apiBaseUrl}/api/auth/me`).flush(user);

    service.logout().subscribe();
    const request = http.expectOne(`${environment.apiBaseUrl}/api/auth/logout`);
    expect(request.request.method).toBe('POST');
    expect(request.request.withCredentials).toBeTrue();
    request.flush(null);

    expect(service.user()).toBeNull();
    expect(service.isAuthenticated()).toBeFalse();
  });
});
