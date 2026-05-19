import { useEffect, useRef, forwardRef, useImperativeHandle } from 'react'
import { Terminal } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import { WebglAddon } from '@xterm/addon-webgl'
import '@xterm/xterm/css/xterm.css'

export interface TerminalHandle {
  fit: () => void
}

interface TerminalComponentProps {
  ws: WebSocket | null
  onStatusChange?: (status: 'connecting' | 'connected' | 'disconnected') => void
  onSizeChange?: (size: { cols: number; rows: number }) => void
}

const TerminalComponent = forwardRef<TerminalHandle, TerminalComponentProps>(
  ({ ws, onStatusChange, onSizeChange }, ref) => {
    const termRef = useRef<HTMLDivElement>(null)
    const termInstance = useRef<Terminal | null>(null)
    const fitAddonRef = useRef<FitAddon | null>(null)

    useImperativeHandle(ref, () => ({
      fit: () => {
        if (fitAddonRef.current) {
          fitAddonRef.current.fit()
        }
      },
    }))

    useEffect(() => {
      if (!termRef.current || !ws) return

      const term = new Terminal({
        cursorBlink: true,
        cursorStyle: 'bar',
        cursorWidth: 2,
        fontSize: 14,
        fontFamily: '"JetBrains Mono", "Fira Code", Menlo, Monaco, "Courier New", monospace',
        fontWeight: '400',
        lineHeight: 1.15,
        scrollback: 10000,
        allowProposedApi: true,
        theme: {
          background: '#282c34',
          foreground: '#abb2bf',
          cursor: '#528bff',
          cursorAccent: '#282c34',
          selectionBackground: 'rgba(62, 68, 81, 0.6)',
          black: '#3f4451',
          red: '#e06c75',
          green: '#98c379',
          yellow: '#e5c07b',
          blue: '#61afef',
          magenta: '#c678dd',
          cyan: '#56b6c2',
          white: '#abb2bf',
          brightBlack: '#4f5666',
          brightRed: '#be5046',
          brightGreen: '#98c379',
          brightYellow: '#d19a66',
          brightBlue: '#61afef',
          brightMagenta: '#c678dd',
          brightCyan: '#56b6c2',
          brightWhite: '#e6e6e6',
        },
      })

      const fit = new FitAddon()
      term.loadAddon(fit)
      term.open(termRef.current)

      try {
        const webgl = new WebglAddon()
        webgl.onContextLoss(() => webgl.dispose())
        term.loadAddon(webgl)
      } catch {
        // WebGL not available, canvas fallback
      }

      fit.fit()
      termInstance.current = term
      fitAddonRef.current = fit

      onStatusChange?.('connecting')

      if (ws.readyState === WebSocket.OPEN) {
        onStatusChange?.('connected')
      }

      ws.onmessage = (event) => {
        term.write(event.data)
        onStatusChange?.('connected')
      }

      ws.onclose = () => {
        term.write('\r\n\x1b[38;2;92;99;112m[ 会话已断开 ]\x1b[0m\r\n')
        onStatusChange?.('disconnected')
      }

      ws.onerror = () => {
        term.write('\r\n\x1b[38;2;224;108;117m[ 连接错误 ]\x1b[0m\r\n')
        onStatusChange?.('disconnected')
      }

      term.onData((data) => {
        if (ws.readyState === WebSocket.OPEN) {
          ws.send(data)
        }
      })

      const handleWindowResize = () => fit.fit()

      term.onResize(({ cols, rows }) => {
        onSizeChange?.({ cols, rows })
        if (ws.readyState === WebSocket.OPEN) {
          ws.send(JSON.stringify({ cols, rows }))
        }
      })

      window.addEventListener('resize', handleWindowResize)

      setTimeout(() => {
        fit.fit()
        onSizeChange?.({ cols: term.cols, rows: term.rows })
        if (ws.readyState === WebSocket.OPEN) {
          ws.send(JSON.stringify({ cols: term.cols, rows: term.rows }))
        }
      }, 100)

      return () => {
        window.removeEventListener('resize', handleWindowResize)
        term.dispose()
        termInstance.current = null
        fitAddonRef.current = null
      }
    }, [ws])

    return (
      <div
        ref={termRef}
        style={{
          width: '100%',
          height: '100%',
          backgroundColor: '#282c34',
        }}
      />
    )
  }
)

TerminalComponent.displayName = 'TerminalComponent'

export default TerminalComponent
