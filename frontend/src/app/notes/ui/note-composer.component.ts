import { Component, EventEmitter, Output, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';

import { CreateNoteRequest } from '../data-access/note.models';

@Component({
  selector: 'app-note-composer',
  imports: [FormsModule],
  template: `
    <form class="composer" (ngSubmit)="submit()">
      <div class="composer-chrome">
        <span class="dot peach"></span>
        <span class="dot mint"></span>
        <span class="dot lavender"></span>
      </div>
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
          <span class="toggle-track" aria-hidden="true"></span>
          <span>Pin note</span>
        </label>
        <button type="submit" [disabled]="saving() || !canSubmit()">
          {{ saving() ? 'Saving' : 'Add note' }}
        </button>
      </div>
    </form>
  `,
  styles: `
    .composer {
      background: var(--notion-canvas);
      border: 1px solid var(--notion-hairline);
      border-radius: 12px;
      box-shadow: var(--notion-shadow-card);
      display: grid;
      gap: 10px;
      margin: 0 auto 32px;
      max-width: 680px;
      padding: 16px;
    }

    .composer-chrome {
      display: flex;
      gap: 6px;
      padding-bottom: 2px;
    }

    .dot {
      border-radius: 9999px;
      height: 8px;
      width: 8px;
    }

    .peach {
      background: var(--notion-tint-peach);
    }

    .mint {
      background: var(--notion-tint-mint);
    }

    .lavender {
      background: var(--notion-tint-lavender);
    }

    input,
    textarea {
      border: 0;
      color: var(--notion-ink);
      outline: 0;
      resize: vertical;
      width: 100%;
    }

    input::placeholder,
    textarea::placeholder {
      color: var(--notion-stone);
    }

    input {
      font-size: 1.125rem;
      font-weight: 600;
      line-height: 1.35;
    }

    textarea {
      color: var(--notion-charcoal);
      line-height: 1.55;
      min-height: 84px;
    }

    .composer-actions {
      align-items: center;
      display: flex;
      justify-content: space-between;
      gap: 12px;
    }

    .pin-toggle {
      align-items: center;
      color: var(--notion-slate);
      cursor: pointer;
      display: inline-flex;
      font-size: 0.875rem;
      font-weight: 500;
      gap: 10px;
    }

    .pin-toggle input {
      inline-size: 1px;
      opacity: 0;
      position: absolute;
    }

    .toggle-track {
      background: var(--notion-surface);
      border: 1px solid var(--notion-hairline-strong);
      border-radius: 9999px;
      height: 22px;
      position: relative;
      transition: background 120ms ease, border-color 120ms ease;
      width: 38px;
    }

    .toggle-track::after {
      background: var(--notion-canvas);
      border: 1px solid var(--notion-hairline);
      border-radius: 50%;
      box-shadow: var(--notion-shadow-subtle);
      content: '';
      height: 16px;
      left: 2px;
      position: absolute;
      top: 2px;
      transition: transform 120ms ease;
      width: 16px;
    }

    .pin-toggle input:checked + .toggle-track {
      background: var(--notion-primary);
      border-color: var(--notion-primary);
    }

    .pin-toggle input:checked + .toggle-track::after {
      transform: translateX(16px);
    }

    .pin-toggle input:focus-visible + .toggle-track {
      outline: 2px solid var(--notion-primary);
      outline-offset: 2px;
    }

    button {
      background: var(--notion-primary);
      border: 0;
      border-radius: 8px;
      color: var(--notion-canvas);
      cursor: pointer;
      font-size: 0.875rem;
      font-weight: 500;
      min-height: 40px;
      padding: 0 18px;
    }

    button:hover:not(:disabled) {
      background: var(--notion-primary-pressed);
    }

    button:disabled {
      background: var(--notion-hairline);
      color: var(--notion-muted);
      cursor: not-allowed;
    }

    @media (max-width: 520px) {
      .composer-actions {
        align-items: stretch;
        flex-direction: column;
      }

      button {
        width: 100%;
      }
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
