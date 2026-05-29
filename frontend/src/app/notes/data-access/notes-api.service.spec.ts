import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';

import { environment } from '../../../environments/environment';
import { NotesApiService } from './notes-api.service';

describe('NotesApiService', () => {
  let service: NotesApiService;
  let http: HttpTestingController;

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [provideHttpClient(), provideHttpClientTesting()],
    });

    service = TestBed.inject(NotesApiService);
    http = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    http.verify();
  });

  it('lists notes with pagination and credentials', () => {
    let total = 0;

    service.list(2, 10).subscribe((page) => {
      total = page.total;
    });

    const request = http.expectOne(`${environment.apiBaseUrl}/api/notes?page=2&limit=10`);
    expect(request.request.method).toBe('GET');
    expect(request.request.withCredentials).toBeTrue();

    request.flush({ total: 3, notes: [], page: 2, limit: 10, pages: 1 });
    expect(total).toBe(3);
  });

  it('creates, updates, and deletes notes with credentials', () => {
    service.create({ title: 'Title', content: 'Body', pinned: true }).subscribe();
    const create = http.expectOne(`${environment.apiBaseUrl}/api/notes`);
    expect(create.request.method).toBe('POST');
    expect(create.request.withCredentials).toBeTrue();
    expect(create.request.body).toEqual({ title: 'Title', content: 'Body', pinned: true });
    create.flush({ id: '1', title: 'Title', content: 'Body', pinned: true });

    service.update('1', { content: 'Updated' }).subscribe();
    const update = http.expectOne(`${environment.apiBaseUrl}/api/notes/1`);
    expect(update.request.method).toBe('PATCH');
    expect(update.request.withCredentials).toBeTrue();
    expect(update.request.body).toEqual({ content: 'Updated' });
    update.flush({ id: '1', title: 'Title', content: 'Updated', pinned: true });

    service.delete('1').subscribe();
    const deleteRequest = http.expectOne(`${environment.apiBaseUrl}/api/notes/1`);
    expect(deleteRequest.request.method).toBe('DELETE');
    expect(deleteRequest.request.withCredentials).toBeTrue();
    deleteRequest.flush(null);
  });
});
