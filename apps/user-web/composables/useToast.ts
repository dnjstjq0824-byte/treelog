export interface Toast {
  id: number
  message: string
  kind: 'info' | 'success' | 'error'
}

let nextId = 1

export function useToast() {
  const toasts = useState<Toast[]>('tb_toasts', () => [])

  function dismiss(id: number): void {
    toasts.value = toasts.value.filter((toast) => toast.id !== id)
  }

  function push(message: string, kind: Toast['kind'] = 'info', timeoutMs = 3600): void {
    const id = nextId++
    toasts.value = [...toasts.value, { id, message, kind }]
    if (timeoutMs > 0) {
      setTimeout(() => dismiss(id), timeoutMs)
    }
  }

  return {
    toasts: readonly(toasts),
    push,
    dismiss,
    success: (message: string) => push(message, 'success'),
    error: (message: string) => push(message, 'error')
  }
}
