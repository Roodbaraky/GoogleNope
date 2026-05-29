import { ComponentFixture, TestBed } from '@angular/core/testing';

import { Note } from '../data-access/note.models';
import { NoteEditorComponent } from './note-editor.component';

describe('NoteEditorComponent', () => {
  let fixture: ComponentFixture<NoteEditorComponent>;
  let component: NoteEditorComponent;

  const note: Note = {
    id: '1',
    title: 'Original',
    content: 'Body',
    pinned: false,
    createdAt: '2026-05-29T09:00:00Z',
    updatedAt: '2026-05-29T09:00:00Z',
  };

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [NoteEditorComponent],
    }).compileComponents();

    fixture = TestBed.createComponent(NoteEditorComponent);
    component = fixture.componentInstance;
  });

  it('copies the selected note into editable fields', () => {
    fixture.componentRef.setInput('note', note);
    fixture.detectChanges();

    expect(component.title).toBe('Original');
    expect(component.content).toBe('Body');
    expect(component.pinned).toBeFalse();
  });

  it('emits a trimmed update when closed', () => {
    const emitted: unknown[] = [];
    component.saveAndClose.subscribe((value) => emitted.push(value));
    fixture.componentRef.setInput('note', note);
    fixture.detectChanges();

    component.title = '  Updated  ';
    component.content = '  New body  ';
    component.pinned = true;
    component.close();

    expect(emitted).toEqual([{ title: 'Updated', content: 'New body', pinned: true }]);
  });

  it('emits the selected note for deletion', () => {
    const emit = spyOn(component.deleteNote, 'emit');
    fixture.componentRef.setInput('note', note);
    fixture.detectChanges();

    component.delete();

    expect(emit).toHaveBeenCalledOnceWith(note);
  });
});
