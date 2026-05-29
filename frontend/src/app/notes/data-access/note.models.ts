export interface Note {
  id: string;
  title: string;
  content: string;
  pinned: boolean;
  createdAt?: string;
  updatedAt?: string;
}

export interface NotesPage {
  total: number;
  notes: Note[];
  page: number;
  limit: number;
  pages: number;
}

export interface CreateNoteRequest {
  title: string;
  content: string;
  pinned: boolean;
}

export type UpdateNoteRequest = Partial<CreateNoteRequest>;
