import { ComponentFixture, TestBed } from '@angular/core/testing';

import { NoteComposerComponent } from './note-composer.component';

describe('NoteComposerComponent', () => {
  let fixture: ComponentFixture<NoteComposerComponent>;
  let component: NoteComposerComponent;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [NoteComposerComponent],
    }).compileComponents();

    fixture = TestBed.createComponent(NoteComposerComponent);
    component = fixture.componentInstance;
  });

  it('emits a trimmed note and enters saving state', () => {
    const emitted: unknown[] = [];
    component.createNote.subscribe((value) => emitted.push(value));
    component.title = '  Title  ';
    component.content = '  Body  ';
    component.pinned = true;

    component.submit();

    expect(emitted).toEqual([{ title: 'Title', content: 'Body', pinned: true }]);
    expect(component.saving()).toBeTrue();
  });

  it('does not submit an empty note', () => {
    const emit = spyOn(component.createNote, 'emit');
    component.title = '  ';
    component.content = '';

    component.submit();

    expect(emit).not.toHaveBeenCalled();
    expect(component.saving()).toBeFalse();
  });

  it('resets the draft after a save succeeds', () => {
    component.title = 'Title';
    component.content = 'Body';
    component.pinned = true;
    component.saving.set(true);

    component.markSaved();

    expect(component.title).toBe('');
    expect(component.content).toBe('');
    expect(component.pinned).toBeFalse();
    expect(component.saving()).toBeFalse();
  });
});
