import { describe, expect, it } from 'vitest'
import {
  countWords,
  DEFAULT_LIMITS,
  extractPlainText,
  isEditorDocument,
  summarizeDocument,
  validateDocument,
  type EditorDocument
} from './index'

const doc = (content: EditorDocument['content']): EditorDocument => ({ type: 'doc', content })

describe('isEditorDocument', () => {
  it('accepts doc shapes and rejects the rest', () => {
    expect(isEditorDocument(doc([]))).toBe(true)
    expect(isEditorDocument(null)).toBe(false)
    expect(isEditorDocument({})).toBe(false)
    expect(isEditorDocument({ type: 'doc' })).toBe(false)
    expect(isEditorDocument({ type: 'text', content: [] })).toBe(false)
  })
})

describe('validateDocument', () => {
  it('accepts a typical post', () => {
    const result = validateDocument(
      doc([
        { type: 'heading', attrs: { level: 2 }, content: [{ type: 'text', text: '제목' }] },
        {
          type: 'paragraph',
          content: [
            { type: 'text', text: '본문 ' },
            { type: 'text', text: '강조', marks: [{ type: 'bold' }] },
            { type: 'text', text: '링크', marks: [{ type: 'link', attrs: { href: 'https://example.com' } }] }
          ]
        },
        { type: 'codeBlock', attrs: { language: 'go' }, content: [{ type: 'text', text: 'fmt.Println()' }] },
        {
          type: 'table',
          content: [
            {
              type: 'tableRow',
              content: [
                { type: 'tableHeader', content: [{ type: 'paragraph', content: [{ type: 'text', text: 'H' }] }] },
                { type: 'tableCell', content: [{ type: 'paragraph', content: [{ type: 'text', text: 'C' }] }] }
              ]
            }
          ]
        },
        { type: 'image', attrs: { src: 'https://cdn.example.com/a.png', alt: 'a' } }
      ])
    )
    expect(result.ok).toBe(true)
    expect(result.issues).toEqual([])
  })

  it('rejects unknown blocks and marks with paths', () => {
    const result = validateDocument(
      doc([
        { type: 'paragraph', content: [{ type: 'text', text: 'x', marks: [{ type: 'fontSize' }] }] },
        { type: 'script', content: [{ type: 'text', text: 'alert(1)' }] }
      ])
    )
    expect(result.ok).toBe(false)
    expect(result.issues.map((issue) => issue.message)).toContain('허용되지 않은 서식: fontSize')
    expect(result.issues.map((issue) => issue.message)).toContain('허용되지 않은 블록: script')
    expect(result.issues[0]?.path).toMatch(/^\$\.content\[0\]/)
  })

  it('rejects dangerous links and images', () => {
    const bad = validateDocument(
      doc([
        {
          type: 'paragraph',
          content: [{ type: 'text', text: 'x', marks: [{ type: 'link', attrs: { href: 'javascript:alert(1)' } }] }]
        }
      ])
    )
    expect(bad.ok).toBe(false)
    const img = validateDocument(doc([{ type: 'image', attrs: { src: 'ftp://x/y.png' } }]))
    expect(img.ok).toBe(false)
    const dataImg = validateDocument(doc([{ type: 'image', attrs: { src: 'data:image/png;base64,xx' } }]))
    expect(dataImg.ok).toBe(true)
  })

  it('enforces heading levels and limits', () => {
    const heading = validateDocument(doc([{ type: 'heading', attrs: { level: 9 }, content: [] }]))
    expect(heading.ok).toBe(false)
    const many = validateDocument(doc(Array.from({ length: 10 }, () => ({ type: 'paragraph' }))), {
      ...DEFAULT_LIMITS,
      maxNodes: 5
    })
    expect(many.ok).toBe(false)
    const deep = (depth: number): { type: string; content?: unknown[] } =>
      depth <= 0 ? { type: 'paragraph' } : { type: 'blockquote', content: [deep(depth - 1) as never] }
    const nested = validateDocument(doc([deep(20) as never]))
    expect(nested.ok).toBe(false)
  })
})

describe('text helpers', () => {
  it('extracts and counts like the server', () => {
    const text = extractPlainText(
      doc([
        { type: 'heading', content: [{ type: 'text', text: 'Hello' }] },
        { type: 'paragraph', content: [{ type: 'text', text: 'world  wide' }] }
      ])
    )
    expect(text).toBe('Hello world wide')
    expect(countWords(text)).toBe(3)
    expect(countWords('  ')).toBe(0)
    expect(summarizeDocument(null)).toEqual({ text: '', words: 0 })
  })
})
