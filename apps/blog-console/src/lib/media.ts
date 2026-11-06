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

export function publicMediaURL(assetId: string): string {
  return `${apiBase()}/api/v1/public/media/${assetId}`
}
