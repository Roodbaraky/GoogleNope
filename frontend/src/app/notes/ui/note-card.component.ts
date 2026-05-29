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
          <span class="pin" title="Pinned" aria-label="Pinned">Pinned</span>
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
      background: var(--notion-canvas);
      border: 1px solid var(--notion-hairline);
      border-radius: 12px;
      cursor: pointer;
      display: grid;
      gap: 12px;
      min-height: 148px;
      padding: 16px;
      transition: border-color 120ms ease, box-shadow 120ms ease, transform 120ms ease, background 120ms ease;
      width: 100%;
    }

    .note-card:hover,
    .note-card:focus {
      border-color: var(--notion-hairline-strong);
      box-shadow: var(--notion-shadow-card);
      outline: 0;
      transform: translateY(-1px);
    }

    .note-card.pinned {
      background: var(--notion-tint-yellow);
      border-color: #f5d75e;
    }

    .note-header {
      align-items: start;
      display: flex;
      gap: 8px;
      justify-content: space-between;
    }

    h2 {
      color: var(--notion-ink);
      font-size: 1rem;
      font-weight: 600;
      line-height: 1.35;
      margin: 0;
      overflow-wrap: anywhere;
    }

    .pin {
      background: var(--notion-tint-lavender);
      border-radius: 6px;
      color: #391c57;
      flex: 0 0 auto;
      font-size: 0.75rem;
      font-weight: 600;
      line-height: 1.4;
      padding: 2px 8px;
    }

    p {
      color: var(--notion-charcoal);
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
      color: var(--notion-steel);
      font-size: 0.78rem;
    }

    @media (max-width: 640px) {
      .note-card {
        border-radius: 10px;
        min-height: 132px;
        padding: 14px;
      }

      h2 {
        font-size: 0.96rem;
      }

      p {
        -webkit-line-clamp: 5;
      }
    }
  `,
})
export class NoteCardComponent {
  @Input({ required: true }) note!: Note;
  @Output() open = new EventEmitter<Note>();
}
