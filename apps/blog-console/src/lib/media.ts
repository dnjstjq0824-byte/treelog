import { api, apiBase } from './api'
import type { MediaAsset, UploadSession } from './types'

async function sha256Hex(file: File): Promise<string> {
  const digest = await crypto.subtle.digest('SHA-256', await file.arrayBuffer())
  return [...new Uint8Array(digest)].map((byte) => byte.toString(16).padStart(2, '0')).join('')
}

function firstBytesBase64(file: File, length = 16): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => {
      const buffer = new Uint8Array(reader.result as ArrayBuffer)
      let binary = ''
      buffer.forEach((byte) => {
        binary += String.fromCharCode(byte)
      })
      resolve(btoa(binary))
    }
    reader.onerror = () => reject(reader.error)
    reader.readAsArrayBuffer(file.slice(0, length))
  })
}

/** Full upload flow: session → PUT bytes → complete. Returns the asset. */
export async function uploadImage(blogId: string, file: File): Promise<MediaAsset> {
  const session = await api<UploadSession>(`/api/v1/blogs/${blogId}/media/upload-sessions`, {
    method: 'POST',
    body: {
      original_filename: file.name,
      mime_type: file.type || 'application/octet-stream',
      size_bytes: file.size,
      sha256_hash: await sha256Hex(file)
    }
  })
  if (session.deduplicated && session.existing_asset) {
    return session.existing_asset
  }
  const put = await fetch(session.presigned_url, {
    method: 'PUT',
    headers: { 'Content-Type': file.type || 'application/octet-stream' },
    body: file
  })
  if (!put.ok) {
    throw new Error(`업로드 실패 (${put.status})`)
  }
  return await api<MediaAsset>(
    `/api/v1/blogs/${blogId}/media/upload-sessions/${session.session_id}/complete`,
    { method: 'POST', body: { magic_bytes_base64: await firstBytesBase64(file) } }
  )
}

export function publicMediaURL(assetId: string, variant?: string, format?: string): string {
  const params = new URLSearchParams()
  if (variant) params.set('variant', variant)
  if (format) params.set('format', format)
  const query = params.toString()
  return `${apiBase()}/api/v1/public/media/${assetId}${query ? `?${query}` : ''}`
}

export interface MediaVariantInfo {
  variant_name: string
  format: string
  width: number
  height: number
  size_bytes: number
  public_cdn_url: string
}

// buildSrcset renders width-described srcset entries from the variants list
// endpoint, preferring WebP twins where the pipeline emitted them.
export function buildSrcset(assetId: string, variants: MediaVariantInfo[]): string {
  const byWidth = new Map<number, MediaVariantInfo>()
  for (const variant of variants) {
    if (variant.variant_name === 'THUMB_400') continue
    const current = byWidth.get(variant.width)
    if (!current || (variant.format === 'WEBP' && current.format !== 'WEBP')) {
      byWidth.set(variant.width, variant)
    }
  }
  return [...byWidth.entries()]
    .sort(([a], [b]) => a - b)
    .map(([width, variant]) => `${publicMediaURL(assetId, variant.variant_name, variant.format.toLowerCase())} ${width}w`)
    .join(', ')
}
