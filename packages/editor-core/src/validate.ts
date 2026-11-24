import { type DocNode, type EditorDocument, isEditorDocument } from './document'

// Node and mark types the consoles can produce. The server HTML sanitizer
// (bluemonday UGC + code classes + YouTube/Vimeo embeds) accepts everything
// rendered from this set; anything outside it is rejected before save.
export const ALLOWED_NODE_TYPES = new Set([
  'doc',
  'paragraph',
  'heading',
  'blockquote',
  'codeBlock',
  'bulletList',
  'orderedList',
  'listItem',
  'taskList',
  'taskItem',
  'table',
  'tableRow',
  'tableHeader',
  'tableCell',
  'image',
  'horizontalRule',
  'hardBreak',
  'text',
  'videoEmbed'
])

export const ALLOWED_MARK_TYPES = new Set(['bold', 'italic', 'strike', 'code', 'link', 'underline'])

export interface DocumentLimits {
  maxNodes: number
  maxDepth: number
  maxTextChars: number
}

export const DEFAULT_LIMITS: DocumentLimits = {
  maxNodes: 5000,
  maxDepth: 12,
  maxTextChars: 200000
}

export interface DocumentIssue {
  path: string
  message: string
}

export interface ValidationResult {
  ok: boolean
  issues: DocumentIssue[]
  nodeCount: number
  textChars: number
}

export function validateDocument(value: unknown, limits: DocumentLimits = DEFAULT_LIMITS): ValidationResult {
  const issues: DocumentIssue[] = []
  if (!isEditorDocument(value)) {
    return { ok: false, issues: [{ path: '$', message: '문서는 doc/content 형태여야 합니다' }], nodeCount: 0, textChars: 0 }
  }
  let nodeCount = 0
  let textChars = 0
  const visit = (node: DocNode, path: string, depth: number): void => {
    nodeCount += 1
    if (depth > limits.maxDepth) {
      issues.push({ path, message: `중첩 깊이 초과(최대 ${limits.maxDepth})` })
      return
    }
    if (!ALLOWED_NODE_TYPES.has(node.type)) {
      issues.push({ path, message: `허용되지 않은 블록: ${node.type}` })
      return
    }
    for (const mark of node.marks ?? []) {
      if (!ALLOWED_MARK_TYPES.has(mark.type)) {
        issues.push({ path, message: `허용되지 않은 서식: ${mark.type}` })
      }
      if (mark.type === 'link') {
        const href = typeof mark.attrs?.href === 'string' ? (mark.attrs.href as string) : ''
        if (!/^https?:\/\//i.test(href) && !href.startsWith('/') && !href.startsWith('#') && !href.startsWith('mailto:')) {
          issues.push({ path, message: '링크 주소는 http(s)/상대경로/앵커만 허용됩니다' })
        }
      }
    }
    if (typeof node.text === 'string') textChars += [...node.text].length
    if (node.type === 'image') {
      const src = typeof node.attrs?.src === 'string' ? (node.attrs.src as string) : ''
      if (!src || (!/^https?:\/\//i.test(src) && !src.startsWith('/') && !src.startsWith('data:image/'))) {
        issues.push({ path, message: '이미지 주소가 올바르지 않습니다' })
      }
    }
    if (node.type === 'heading') {
      const level = node.attrs?.level as number | undefined
      if (level !== undefined && (level < 1 || level > 6)) {
        issues.push({ path, message: '제목 수준은 1~6이어야 합니다' })
      }
    }
    ;(node.content ?? []).forEach((child, index) => {
      if (child && typeof child === 'object') visit(child as DocNode, `${path}.content[${index}]`, depth + 1)
    })
  }
  ;(value.content ?? []).forEach((child, index) => {
    if (child && typeof child === 'object') visit(child as DocNode, `$.content[${index}]`, 1)
  })
  if (nodeCount > limits.maxNodes) {
    issues.push({ path: '$', message: `블록 수 초과(최대 ${limits.maxNodes})` })
  }
  if (textChars > limits.maxTextChars) {
    issues.push({ path: '$', message: `텍스트 길이 초과(최대 ${limits.maxTextChars}자)` })
  }
  return { ok: issues.length === 0, issues, nodeCount, textChars }
}
