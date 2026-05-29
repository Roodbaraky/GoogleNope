import { computed, signal } from '@angular/core';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { of, throwError } from 'rxjs';

import { AuthService } from '../../core/auth/auth.service';
import { User } from '../../core/auth/auth.models';
import { Note, NotesPage } from '../data-access/note.models';
import { NotesApiService } from '../data-access/notes-api.service';
import { NotesPageComponent } from './notes-page.component';

class FakeAuthService {
  readonly user = signal<User | null>({
    id: 'oauth:user-1',
    email: 'user@example.com',
    name: 'User One',
    picture: '',
  });
  readonly loading = signal(false);
  readonly error = signal<string | null>(null);
  readonly isAuthenticated = computed(() => this.user() !== null);

  loadSession() {
    return of(this.user());
  }

  login(): void {}

  logout() {
    this.user.set(null);
    return of(undefined);
  }
}

class FakeNotesApiService {
  notes: Note[] = [];
  createError: unknown = null;

  list() {
    return of<NotesPage>({
      total: this.notes.length,
      notes: this.notes,
      page: 1,
      limit: 30,
      pages: this.notes.length > 0 ? 1 : 0,
    });
  }

  create(input: { title: string; content: string; pinned: boolean }) {
    if (this.createError) {
      return throwError(() => this.createError);
    }

    return of<Note>({
      id: 'new-note',
      title: input.title,
      content: input.content,
      pinned: input.pinned,
      createdAt: '2026-05-29T09:00:00Z',
      updatedAt: '2026-05-29T09:00:00Z',
    });
  }

  update(id: string, input: Partial<Note>) {
    return of<Note>({
      id,
      title: input.title ?? '',
      content: input.content ?? '',
      pinned: input.pinned ?? false,
      createdAt: '2026-05-29T09:00:00Z',
      updatedAt: '2026-05-29T09:10:00Z',
    });
  }

  delete() {
    return of(undefined);
  }
}

describe('NotesPageComponent', () => {
  let fixture: ComponentFixture<NotesPageComponent>;
  let component: NotesPageComponent;
  let notesApi: FakeNotesApiService;

  const pinned: Note = {
    id: 'pinned',
    title: 'Pinned',
    content: 'Pinned body',
    pinned: true,
    createdAt: '2026-05-29T09:00:00Z',
    updatedAt: '2026-05-29T09:00:00Z',
  };

  const other: Note = {
    id: 'other',
    title: 'Other',
    content: 'Other body',
    pinned: false,
    createdAt: '2026-05-29T09:00:00Z',
    updatedAt: '2026-05-29T09:00:00Z',
  };

  beforeEach(async () => {
    notesApi = new FakeNotesApiService();
    notesApi.notes = [pinned, other];

    await TestBed.configureTestingModule({
      imports: [NotesPageComponent],
      providers: [
        { provide: AuthService, useClass: FakeAuthService },
        { provide: NotesApiService, useValue: notesApi },
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(NotesPageComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('loads notes for an authenticated session and splits pinned notes', () => {
    expect(component.notes()).toEqual([pinned, other]);
    expect(component.pinnedNotes()).toEqual([pinned]);
    expect(component.otherNotes()).toEqual([other]);
    expect(component.loading()).toBeFalse();
  });

  it('prepends a created note', () => {
    component.create({ title: 'New', content: 'Body', pinned: false });

    expect(component.notes()[0]).toEqual(jasmine.objectContaining({ id: 'new-note', title: 'New' }));
    expect(component.error()).toBeNull();
  });

  it('surfaces create failures without changing the note list', () => {
    notesApi.createError = { error: { error: 'Create failed' } };

    component.create({ title: 'New', content: 'Body', pinned: false });

    expect(component.notes()).toEqual([pinned, other]);
    expect(component.error()).toBe('Create failed');
  });
});
