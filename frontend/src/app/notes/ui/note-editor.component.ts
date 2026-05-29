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
      background: rgb(32 33 36 / 36%);
      display: flex;
      inset: 0;
      justify-content: center;
      padding: 20px;
      position: fixed;
      z-index: 20;
    }

    .editor {
      background: #ffffff;
      border-radius: 8px;
      box-shadow: 0 10px 32px rgb(32 33 36 / 32%);
      display: grid;
      gap: 12px;
      max-height: min(760px, calc(100vh - 40px));
      max-width: 720px;
      padding: 18px;
      width: min(100%, 720px);
    }

    input,
    textarea {
      border: 0;
      color: #202124;
      font: inherit;
      outline: 0;
      width: 100%;
    }

    input {
      font-size: 1.2rem;
      font-weight: 600;
    }

    textarea {
      min-height: 320px;
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
      color: #5f6368;
      font-size: 0.9rem;
    }

    button {
      background: transparent;
      border: 1px solid #dadce0;
      border-radius: 6px;
      color: #202124;
      cursor: pointer;
      font: inherit;
      min-height: 36px;
      padding: 0 14px;
    }

    button:hover {
      background: #f1f3f4;
    }

    .danger {
      border-color: #f4c7c3;
      color: #b3261e;
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
