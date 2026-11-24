import { type DocNode, type EditorDocument, isEditorDocument } from './document'

// extractPlainText walks the AST in document order, joining each top-level
// block with a space so adjacent blocks never fuse into one word.
// It mirrors the server word-count input (posts.content_plain).
export function extractPlainText(doc: EditorDocument): string {
  const blocks: string[] = []
  const collect = (node: DocNode, parts: string[]): void => {
    if (typeof node.text === 'string') parts.push(node.text)
    for (const child of node.content ?? []) collect(child, parts)
  }
  for (const child of doc.content ?? []) {
    const parts: string[] = []
    collect(child, parts)
    if (parts.length > 0) blocks.push(parts.join(''))
  }
  return blocks.join(' ').replace(/\s+/g, ' ').trim()
}

// countWords splits on whitespace like the server (strings.Fields), so the
// console estimate matches the stored word_count for the same text.
export function countWords(text: string): number {
  const trimmed = text.trim()
  if (!trimmed) return 0
  return trimmed.split(/\s+/).length
}

export function summarizeDocument(value: unknown): { text: string; words: number } {
  if (!isEditorDocument(value)) return { text: '', words: 0 }
  const text = extractPlainText(value)
  return { text, words: countWords(text) }
}
