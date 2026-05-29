import { Component, EventEmitter, Input, OnChanges, Output, SimpleChanges } from '@angular/core';
import { FormsModule } from '@angular/forms';

import { Note, UpdateNoteRequest } from '../data-access/note.models';

@Component({
  selector: 'app-note-editor',
  imports: [FormsModule],
  template: `
    @if (note) {
      <div class="backdrop" (click)="close()">
        <section class="editor" role="dialog" aria-modal="true" aria-label="Edit note" (click)="$event.stopPropagation()">
          <input name="edit-title" maxlength="120" placeholder="Title" [(ngModel)]="title" />
          <textarea name="edit-content" maxlength="20000" placeholder="Note" [(ngModel)]="content"></textarea>
          <div class="editor-actions">
            <label>
              <input name="edit-pinned" type="checkbox" [(ngModel)]="pinned" />
              <span class="toggle-track" aria-hidden="true"></span>
              <span>Pinned</span>
            </label>
            <div class="buttons">
              <button type="button" class="danger" (click)="delete()">Delete</button>
              <button type="button" (click)="close()">Close</button>
            </div>
          </div>
        </section>
      </div>
    }
  `,
  styles: `
    .backdrop {
      align-items: center;
      background: rgb(15 15 15 / 36%);
      display: flex;
      inset: 0;
      justify-content: center;
      padding: 20px;
      position: fixed;
      z-index: 20;
    }

    .editor {
      background: var(--notion-canvas);
      border: 1px solid var(--notion-hairline);
      border-radius: 12px;
      box-shadow: var(--notion-shadow-modal);
      display: grid;
      gap: 14px;
      max-height: min(760px, calc(100vh - 40px));
      max-width: 720px;
      padding: 20px;
      width: min(100%, 720px);
    }

    input,
    textarea {
      border: 0;
      color: var(--notion-ink);
      outline: 0;
      width: 100%;
    }

    input::placeholder,
    textarea::placeholder {
      color: var(--notion-stone);
    }

    input {
      font-size: 1.35rem;
      font-weight: 600;
      line-height: 1.3;
    }

    textarea {
      color: var(--notion-charcoal);
      line-height: 1.55;
      min-height: 320px;
      max-height: 56vh;
      resize: vertical;
      white-space: pre-wrap;
    }

    .editor-actions,
    label,
    .buttons {
      align-items: center;
      display: flex;
      gap: 10px;
    }

    .editor-actions {
      justify-content: space-between;
    }

    label {
      color: var(--notion-slate);
      cursor: pointer;
      font-size: 0.875rem;
      font-weight: 500;
    }

    label input {
      inline-size: 1px;
      opacity: 0;
      position: absolute;
    }

    .toggle-track {
      background: var(--notion-surface);
      border: 1px solid var(--notion-hairline-strong);
      border-radius: 9999px;
      flex: 0 0 auto;
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

    label input:checked + .toggle-track {
      background: var(--notion-primary);
      border-color: var(--notion-primary);
    }

    label input:checked + .toggle-track::after {
      transform: translateX(16px);
    }

    label input:focus-visible + .toggle-track {
      outline: 2px solid var(--notion-primary);
      outline-offset: 2px;
    }

    button {
      background: transparent;
      border: 1px solid var(--notion-hairline-strong);
      border-radius: 8px;
      color: var(--notion-ink);
      cursor: pointer;
      font-size: 0.875rem;
      font-weight: 500;
      min-height: 40px;
      padding: 0 16px;
      white-space: nowrap;
    }

    button:hover {
      background: var(--notion-surface);
    }

    .danger {
      border-color: #f4b8b8;
      color: var(--notion-error);
    }

    @media (max-width: 820px) {
      .editor {
        max-width: 680px;
      }

      textarea {
        min-height: 280px;
      }
    }

    @media (max-width: 560px) {
      .backdrop {
        align-items: flex-end;
        padding: 0;
      }

      .editor {
        border-bottom: 0;
        border-radius: 12px 12px 0 0;
        max-height: min(92vh, 760px);
        padding: 16px 14px max(16px, env(safe-area-inset-bottom));
        width: 100%;
      }

      input {
        font-size: 1.18rem;
      }

      textarea {
        min-height: 44vh;
        max-height: 58vh;
      }

      .editor-actions {
        align-items: stretch;
        flex-direction: column;
        gap: 14px;
      }

      label {
        min-height: 28px;
      }

      .buttons {
        display: grid;
        gap: 10px;
        grid-template-columns: 1fr 1fr;
      }

      button {
        min-height: 44px;
      }
    }
  `,
})
export class NoteEditorComponent implements OnChanges {
  @Input() note: Note | null = null;
  @Output() saveAndClose = new EventEmitter<UpdateNoteRequest>();
  @Output() deleteNote = new EventEmitter<Note>();

  title = '';
  content = '';
  pinned = false;

  ngOnChanges(changes: SimpleChanges): void {
    if (changes['note'] && this.note) {
      this.title = this.note.title;
      this.content = this.note.content;
      this.pinned = this.note.pinned;
    }
  }

  close(): void {
    if (!this.note) {
      return;
    }

    this.saveAndClose.emit({
      title: this.title.trim(),
      content: this.content.trim(),
      pinned: this.pinned,
    });
  }

  delete(): void {
    if (this.note) {
      this.deleteNote.emit(this.note);
    }
  }
}
