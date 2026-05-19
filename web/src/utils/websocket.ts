type MessageHandler = (data: { type: string; data: unknown }) => void

type WSState = 'connecting' | 'connected' | 'disconnected' | 'reconnecting'

interface WSOptions {
  /** 消息回调 */
  onMessage: MessageHandler
  /** 连接状态变化回调 */
  onStateChange?: (state: WSState) => void
  /** 最大重连次数，默认 5 */
  maxRetries?: number
  /** 重连间隔基数（毫秒），默认 2000，每次翻倍 */
  retryInterval?: number
  /** 心跳间隔（毫秒），默认 30000 */
  heartbeatInterval?: number
}

interface ManagedWS {
  /** 主动关闭连接，不再重连 */
  close: () => void
  /** 当前连接状态 */
  getState: () => WSState
}

function createManagedWS(url: string, options: WSOptions): ManagedWS {
  const {
    onMessage,
    onStateChange,
    maxRetries = 5,
    retryInterval = 2000,
    heartbeatInterval = 30000,
  } = options

  let ws: WebSocket | null = null
  let state: WSState = 'connecting'
  let retryCount = 0
  let closed = false
  let heartbeatTimer: ReturnType<typeof setInterval> | null = null
  let retryTimer: ReturnType<typeof setTimeout> | null = null

  const setState = (newState: WSState) => {
    state = newState
    onStateChange?.(newState)
  }

  const clearTimers = () => {
    if (heartbeatTimer) { clearInterval(heartbeatTimer); heartbeatTimer = null }
    if (retryTimer) { clearTimeout(retryTimer); retryTimer = null }
  }

  const startHeartbeat = () => {
    if (heartbeatTimer) clearInterval(heartbeatTimer)
    heartbeatTimer = setInterval(() => {
      if (ws && ws.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify({ type: 'ping' }))
      }
    }, heartbeatInterval)
  }

  const connect = () => {
    if (closed) return

    ws = new WebSocket(url)

    ws.onopen = () => {
      retryCount = 0
      setState('connected')
      startHeartbeat()
    }

    ws.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data)
        if (msg.type === 'pong') return
        onMessage(msg)
      } catch {
        // ignore parse errors
      }
    }

    ws.onerror = () => {
      // onerror 之后会触发 onclose，在 onclose 中处理重连
    }

    ws.onclose = () => {
      clearTimers()
      if (closed) {
        setState('disconnected')
        return
      }
      if (retryCount < maxRetries) {
        setState('reconnecting')
        const delay = retryInterval * Math.pow(2, retryCount)
        retryCount++
        retryTimer = setTimeout(connect, delay)
      } else {
        setState('disconnected')
      }
    }
  }

  connect()

  return {
    close: () => {
      closed = true
      clearTimers()
      if (ws) {
        ws.close()
        ws = null
      }
      setState('disconnected')
    },
    getState: () => state,
  }
}

function buildWSUrl(path: string): string | null {
  const token = localStorage.getItem('access_token')
  if (!token) return null

  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const host = window.location.host
  return `${protocol}//${host}${path}?token=${token}`
}

export function connectTaskWS(taskId: number, onMessage: MessageHandler, onStateChange?: (state: WSState) => void): ManagedWS | null {
  const url = buildWSUrl(`/ws/tasks/${taskId}`)
  if (!url) return null

  return createManagedWS(url, { onMessage, onStateChange })
}

export function connectPlaybookWS(jobId: string, onMessage: MessageHandler, onStateChange?: (state: WSState) => void): ManagedWS | null {
  const url = buildWSUrl(`/ws/playbooks/${jobId}`)
  if (!url) return null

  return createManagedWS(url, { onMessage, onStateChange })
}

export type { ManagedWS, WSState }
