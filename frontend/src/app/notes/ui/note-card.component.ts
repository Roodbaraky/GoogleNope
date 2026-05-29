import { DatePipe } from '@angular/common';
import { Component, EventEmitter, Input, Output } from '@angular/core';

import { Note } from '../data-access/note.models';

@Component({
  selector: 'app-note-card',
  imports: [DatePipe],
  template: `
    <article class="note-card" [class.pinned]="note.pinned" tabindex="0" (click)="open.emit(note)" (keydown.enter)="open.emit(note)">
      <div class="note-header">
        <h2>{{ note.title || 'Untitled' }}</h2>
        @if (note.pinned) {
          <span class="pin" title="Pinned" aria-label="Pinned">★</span>
        }
      </div>
      <p>{{ note.content || 'No content' }}</p>
      <time [dateTime]="note.updatedAt || note.createdAt">
        {{ note.updatedAt || note.createdAt | date: 'mediumDate' }}
      </time>
    </article>
  `,
  styles: `
    .note-card {
      background: #ffffff;
      border: 1px solid #dadce0;
      border-radius: 8px;
      cursor: pointer;
      display: grid;
      gap: 12px;
      min-height: 148px;
      padding: 14px;
      transition: border-color 120ms ease, box-shadow 120ms ease, transform 120ms ease;
    }

    .note-card:hover,
    .note-card:focus {
      border-color: #9aa0a6;
      box-shadow: 0 2px 8px rgb(60 64 67 / 16%);
      outline: 0;
      transform: translateY(-1px);
    }

    .note-card.pinned {
      border-color: #fbbc04;
    }

    .note-header {
      align-items: start;
      display: flex;
      gap: 8px;
      justify-content: space-between;
    }

    h2 {
      color: #202124;
      font-size: 1rem;
      line-height: 1.35;
      margin: 0;
      overflow-wrap: anywhere;
    }

    .pin {
      color: #f9ab00;
      flex: 0 0 auto;
      font-size: 0.9rem;
    }

    p {
      color: #3c4043;
      display: -webkit-box;
      line-height: 1.45;
      margin: 0;
      overflow: hidden;
      overflow-wrap: anywhere;
      -webkit-box-orient: vertical;
      -webkit-line-clamp: 7;
      white-space: pre-wrap;
    }

    time {
      align-self: end;
      color: #80868b;
      font-size: 0.78rem;
    }
  `,
})
export class NoteCardComponent {
  @Input({ required: true }) note!: Note;
  @Output() open = new EventEmitter<Note>();
}
