// Shared Tiptap JSON AST types. The server persists this shape in
// posts.content_json and revisions; consoles and validators share it here.

export interface DocMark {
  type: string
  // deno-lint-ignore no-explicit-any
  attrs?: Record<string, any>
}

export interface DocNode {
  type: string
  // deno-lint-ignore no-explicit-any
  attrs?: Record<string, any>
  content?: DocNode[]
  marks?: DocMark[]
  text?: string
}

export interface EditorDocument {
  type: 'doc'
  content: DocNode[]
}

export function isEditorDocument(value: unknown): value is EditorDocument {
  if (typeof value !== 'object' || value === null) return false
  const doc = value as { type?: unknown; content?: unknown }
  return doc.type === 'doc' && Array.isArray(doc.content)
}
