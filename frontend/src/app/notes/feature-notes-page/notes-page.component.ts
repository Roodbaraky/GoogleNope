import { Component, OnInit, ViewChild, computed, inject, signal } from '@angular/core';

import { apiErrorMessage } from '../../core/api/api-error';
import { AuthService } from '../../core/auth/auth.service';
import { Note, UpdateNoteRequest } from '../data-access/note.models';
import { NotesApiService } from '../data-access/notes-api.service';
import { NoteCardComponent } from '../ui/note-card.component';
import { NoteComposerComponent } from '../ui/note-composer.component';
import { NoteEditorComponent } from '../ui/note-editor.component';

@Component({
  selector: 'app-notes-page',
  imports: [NoteCardComponent, NoteComposerComponent, NoteEditorComponent],
  template: `
    <main class="shell">
      <header class="topbar">
        <div>
          <h1>GoogleNope</h1>
          @if (auth.user(); as user) {
            <p>{{ user.email || user.name }}</p>
          } @else {
            <p>Private notes, backed by your account.</p>
          }
        </div>
        <div class="session-actions">
          @if (auth.user(); as user) {
            @if (user.picture) {
              <img [src]="user.picture" alt="" />
            }
            <button type="button" class="secondary" (click)="logout()">Log out</button>
          } @else if (!auth.loading()) {
            <button type="button" (click)="auth.login()">Log in</button>
          }
        </div>
      </header>

      @if (auth.loading()) {
        <section class="state">Checking your session...</section>
      } @else if (!auth.isAuthenticated()) {
        <section class="login-panel">
          <h2>Sign in to open your notes</h2>
          <p>Notes are stored per user, so the backend session has to be active before the workspace loads.</p>
          @if (auth.error()) {
            <p class="error">{{ auth.error() }}</p>
          }
          <button type="button" (click)="auth.login()">Continue with OAuth</button>
        </section>
      } @else {
        <app-note-composer #composer (createNote)="create($event)" />

        @if (error()) {
          <section class="error state">{{ error() }}</section>
        }

        @if (loading()) {
          <section class="state">Loading notes...</section>
        } @else if (notes().length === 0) {
          <section class="empty state">
            <h2>No notes yet</h2>
            <p>Create a note above and it will appear here.</p>
          </section>
        } @else {
          @if (pinnedNotes().length > 0) {
            <section class="note-section" aria-label="Pinned notes">
              <h2>Pinned</h2>
              <div class="note-grid">
                @for (note of pinnedNotes(); track note.id) {
                  <app-note-card [note]="note" (open)="open($event)" />
                }
              </div>
            </section>
          }

          @if (otherNotes().length > 0) {
            <section class="note-section" aria-label="Other notes">
              <h2>Others</h2>
              <div class="note-grid">
                @for (note of otherNotes(); track note.id) {
                  <app-note-card [note]="note" (open)="open($event)" />
                }
              </div>
            </section>
          }
        }
      }

      <app-note-editor
        [note]="selectedNote()"
        (saveAndClose)="saveAndClose($event)"
        (deleteNote)="delete($event)"
      />
    </main>
  `,
  styleUrl: './notes-page.component.css',
})
export class NotesPageComponent implements OnInit {
  @ViewChild('composer') private composer?: NoteComposerComponent;

  readonly auth = inject(AuthService);
  private readonly notesApi = inject(NotesApiService);

  readonly notes = signal<Note[]>([]);
  readonly loading = signal(false);
  readonly error = signal<string | null>(null);
  readonly selectedNote = signal<Note | null>(null);
  readonly pinnedNotes = computed(() => this.notes().filter((note) => note.pinned));
  readonly otherNotes = computed(() => this.notes().filter((note) => !note.pinned));

  ngOnInit(): void {
    this.auth.loadSession().subscribe((user) => {
      if (user) {
        this.loadNotes();
      }
    });
  }

  loadNotes(): void {
    this.loading.set(true);
    this.error.set(null);
    this.notesApi.list().subscribe({
      next: (page) => this.notes.set(page.notes),
      error: (error) => this.error.set(apiErrorMessage(error, 'Could not load notes')),
      complete: () => this.loading.set(false),
    });
  }

  create(input: { title: string; content: string; pinned: boolean }): void {
    this.error.set(null);
    this.notesApi.create(input).subscribe({
      next: (note) => {
        this.notes.update((notes) => [note, ...notes]);
        this.composer?.markSaved();
      },
      error: (error) => {
        this.error.set(apiErrorMessage(error, 'Could not create note'));
        this.composer?.markIdle();
      },
    });
  }

  open(note: Note): void {
    this.selectedNote.set(note);
  }

  saveAndClose(update: UpdateNoteRequest): void {
    const note = this.selectedNote();
    this.selectedNote.set(null);
    if (!note || !this.hasChanged(note, update)) {
      return;
    }

    this.error.set(null);
    this.notesApi.update(note.id, update).subscribe({
      next: (saved) => this.replaceNote(saved),
      error: (error) => this.error.set(apiErrorMessage(error, 'Could not save note')),
    });
  }

  delete(note: Note): void {
    if (!window.confirm('Delete this note?')) {
      return;
    }

    this.selectedNote.set(null);
    this.error.set(null);
    this.notesApi.delete(note.id).subscribe({
      next: () => this.notes.update((notes) => notes.filter((item) => item.id !== note.id)),
      error: (error) => this.error.set(apiErrorMessage(error, 'Could not delete note')),
    });
  }

  logout(): void {
    this.auth.logout().subscribe({
      next: () => this.notes.set([]),
      error: (error) => this.error.set(apiErrorMessage(error, 'Could not log out')),
    });
  }

  private replaceNote(saved: Note): void {
    this.notes.update((notes) => notes.map((note) => (note.id === saved.id ? saved : note)));
  }

  private hasChanged(note: Note, update: UpdateNoteRequest): boolean {
    return note.title !== update.title || note.content !== update.content || note.pinned !== update.pinned;
  }
}
