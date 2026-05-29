import { HttpClient, HttpParams } from '@angular/common/http';
import { inject, Injectable } from '@angular/core';
import { Observable } from 'rxjs';

import { environment } from '../../../environments/environment';
import { CreateNoteRequest, Note, NotesPage, UpdateNoteRequest } from './note.models';

@Injectable({ providedIn: 'root' })
export class NotesApiService {
  private readonly http = inject(HttpClient);
  private readonly baseUrl = `${environment.apiBaseUrl}/api/notes`;

  list(page = 1, limit = 30): Observable<NotesPage> {
    const params = new HttpParams().set('page', page).set('limit', limit);
    return this.http.get<NotesPage>(this.baseUrl, { params, withCredentials: true });
  }

  create(input: CreateNoteRequest): Observable<Note> {
    return this.http.post<Note>(this.baseUrl, input, { withCredentials: true });
  }

  update(id: string, input: UpdateNoteRequest): Observable<Note> {
    return this.http.patch<Note>(`${this.baseUrl}/${id}`, input, { withCredentials: true });
  }

  delete(id: string): Observable<void> {
    return this.http.delete<void>(`${this.baseUrl}/${id}`, { withCredentials: true });
  }
}
