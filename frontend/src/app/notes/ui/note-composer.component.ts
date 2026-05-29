import { Component, EventEmitter, Output, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';

import { CreateNoteRequest } from '../data-access/note.models';

@Component({
  selector: 'app-note-composer',
  imports: [FormsModule],
  template: `
    <form class="composer" (ngSubmit)="submit()">
      <input
        name="title"
        type="text"
        maxlength="120"
        placeholder="Title"
        [(ngModel)]="title"
        [disabled]="saving()"
      />
      <textarea
        name="content"
        rows="3"
        maxlength="20000"
        placeholder="Take a note..."
        [(ngModel)]="content"
        [disabled]="saving()"
      ></textarea>
      <div class="composer-actions">
        <label class="pin-toggle">
          <input name="pinned" type="checkbox" [(ngModel)]="pinned" [disabled]="saving()" />
          <span>Pinned</span>
        </label>
        <button type="submit" [disabled]="saving() || !canSubmit()">Add note</button>
      </div>
    </form>
  `,
  styles: `
    .composer {
      background: #ffffff;
      border: 1px solid #dadce0;
      border-radius: 8px;
      box-shadow: 0 1px 3px rgb(60 64 67 / 18%);
      display: grid;
      gap: 8px;
      margin: 0 auto 24px;
      max-width: 680px;
      padding: 14px;
    }

    input,
    textarea {
      border: 0;
      color: #202124;
      font: inherit;
      outline: 0;
      resize: vertical;
      width: 100%;
    }

    input {
      font-weight: 600;
    }

    textarea {
      min-height: 68px;
    }

    .composer-actions {
      align-items: center;
      display: flex;
      justify-content: space-between;
      gap: 12px;
    }

    .pin-toggle {
      align-items: center;
      color: #5f6368;
      display: inline-flex;
      font-size: 0.9rem;
      gap: 8px;
    }

    button {
      background: #1a73e8;
      border: 0;
      border-radius: 6px;
      color: #ffffff;
      cursor: pointer;
      font: inherit;
      font-weight: 600;
      min-height: 36px;
      padding: 0 16px;
    }

    button:disabled {
      background: #d2e3fc;
      cursor: not-allowed;
    }
  `,
})
export class NoteComposerComponent {
  @Output() createNote = new EventEmitter<CreateNoteRequest>();

  readonly saving = signal(false);
  title = '';
  content = '';
  pinned = false;

  canSubmit(): boolean {
    return this.title.trim().length > 0 || this.content.trim().length > 0;
  }

  submit(): void {
    if (!this.canSubmit()) {
      return;
    }

    this.saving.set(true);
    this.createNote.emit({
      title: this.title.trim(),
      content: this.content.trim(),
      pinned: this.pinned,
    });
  }

  markSaved(): void {
    this.title = '';
    this.content = '';
    this.pinned = false;
    this.saving.set(false);
  }

  markIdle(): void {
    this.saving.set(false);
  }
}
