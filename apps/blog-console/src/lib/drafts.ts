// Local draft cache. The server autosave stays the source of truth while
// reachable; this store keeps a per-post fallback so offline edits survive
// reloads and network drops.

export interface LocalDraft {
  postId: string
  blogId: string
  title: string
  subtitle: string
  content: unknown
  clientTime: string
  serverRevisionNo: number
}

export interface DraftStore {
  load(postId: string): Promise<LocalDraft | null>
  save(draft: LocalDraft): Promise<void>
  clear(postId: string): Promise<void>
}

const DB_NAME = 'treeblog-drafts'
const STORE_NAME = 'autosaves'

function openDatabase(): Promise<IDBDatabase> {
  return new Promise((resolve, reject) => {
    const request = indexedDB.open(DB_NAME, 1)
    request.onupgradeneeded = () => {
      request.result.createObjectStore(STORE_NAME, { keyPath: 'postId' })
    }
    request.onsuccess = () => resolve(request.result)
    request.onerror = () => reject(request.error ?? new Error('indexeddb open failed'))
  })
}

function requestToPromise<T>(request: IDBRequest<T>): Promise<T> {
  return new Promise((resolve, reject) => {
    request.onsuccess = () => resolve(request.result)
    request.onerror = () => reject(request.error ?? new Error('indexeddb request failed'))
  })
}

// IndexedDBDraftStore persists drafts in the browser. Every method rejects
// when IndexedDB is unavailable; callers must treat the cache as optional.
export class IndexedDBDraftStore implements DraftStore {
  static supported(): boolean {
    return typeof indexedDB !== 'undefined'
  }

  private async store(mode: IDBTransactionMode): Promise<{ tx: IDBTransaction; store: IDBObjectStore }> {
    const db = await openDatabase()
    const tx = db.transaction(STORE_NAME, mode)
    const store = tx.objectStore(STORE_NAME)
    tx.oncomplete = () => db.close()
    return { tx, store }
  }

  async load(postId: string): Promise<LocalDraft | null> {
    const { store } = await this.store('readonly')
    const found = await requestToPromise(store.get(postId) as IDBRequest<LocalDraft | undefined>)
    return found ?? null
  }

  async save(draft: LocalDraft): Promise<void> {
    const { store } = await this.store('readwrite')
    await requestToPromise(store.put(draft))
  }

  async clear(postId: string): Promise<void> {
    const { store } = await this.store('readwrite')
    await requestToPromise(store.delete(postId))
  }
}

// MemoryDraftStore is the test double and the fallback when IndexedDB is
// unavailable (private mode, old browsers, SSR).
export class MemoryDraftStore implements DraftStore {
  private readonly drafts = new Map<string, LocalDraft>()

  async load(postId: string): Promise<LocalDraft | null> {
    return this.drafts.get(postId) ?? null
  }

  async save(draft: LocalDraft): Promise<void> {
    // Structured clone through JSON so later mutations of the editor
    // document cannot corrupt the cached snapshot.
    this.drafts.set(draft.postId, JSON.parse(JSON.stringify(draft)) as LocalDraft)
  }

  async clear(postId: string): Promise<void> {
    this.drafts.delete(postId)
  }
}

export function createDraftStore(): DraftStore {
  return IndexedDBDraftStore.supported() ? new IndexedDBDraftStore() : new MemoryDraftStore()
}

// shouldRestoreLocal decides whether the cached draft wins over the loaded
// server state: only when the cache targets the same revision base and is
// strictly newer than the last confirmed server save.
export function shouldRestoreLocal(local: LocalDraft | null, serverRevisionNo: number, serverSavedAt: string | null): boolean {
  if (!local) return false
  if (local.serverRevisionNo !== serverRevisionNo) return false
  if (!serverSavedAt) return true
  return Date.parse(local.clientTime) > Date.parse(serverSavedAt)
}
