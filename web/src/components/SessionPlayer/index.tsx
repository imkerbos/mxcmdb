import { useEffect, useRef, useState, useCallback } from 'react'
import { Terminal } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import '@xterm/xterm/css/xterm.css'
import {
  CaretRightOutlined,
  PauseOutlined,
  StepForwardOutlined,
  StepBackwardOutlined,
  DesktopOutlined,
} from '@ant-design/icons'
import { Slider, Select, Spin } from 'antd'

interface AsciicastHeader {
  version: number
  width: number
  height: number
  timestamp?: number
}

interface AsciicastEvent {
  time: number
  type: string
  data: string
}

interface SessionPlayerProps {
  recording: string | null
  loading?: boolean
}

function parseAsciicast(raw: string): { header: AsciicastHeader; events: AsciicastEvent[] } {
  const lines = raw.trim().split('\n')
  if (lines.length === 0) return { header: { version: 2, width: 120, height: 40 }, events: [] }

  const header: AsciicastHeader = JSON.parse(lines[0])
  const events: AsciicastEvent[] = []

  for (let i = 1; i < lines.length; i++) {
    const line = lines[i].trim()
    if (!line) continue
    try {
      const arr = JSON.parse(line)
      if (Array.isArray(arr) && arr.length >= 3) {
        events.push({ time: arr[0], type: arr[1], data: arr[2] })
      }
    } catch { /* skip malformed lines */ }
  }

  return { header, events }
}

const SessionPlayer = ({ recording, loading }: SessionPlayerProps) => {
  const containerRef = useRef<HTMLDivElement>(null)
  const termRef = useRef<Terminal | null>(null)
  const fitRef = useRef<FitAddon | null>(null)
  const timerRef = useRef<number | null>(null)
  const eventsRef = useRef<AsciicastEvent[]>([])

  const [playing, setPlaying] = useState(false)
  const [speed, setSpeed] = useState(1)
  const [currentTime, setCurrentTime] = useState(0)
  const [duration, setDuration] = useState(0)
  const [currentIndex, setCurrentIndex] = useState(0)
  const [ready, setReady] = useState(false)

  const speedRef = useRef(speed)
  speedRef.current = speed

  // 初始化终端
  useEffect(() => {
    if (!containerRef.current) return

    const term = new Terminal({
      cursorBlink: false,
      cursorStyle: 'bar',
      fontSize: 14,
      fontFamily: '"JetBrains Mono", "Fira Code", Menlo, Monaco, "Courier New", monospace',
      fontWeight: '400',
      lineHeight: 1.15,
      scrollback: 10000,
      disableStdin: true,
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
    term.open(containerRef.current)
    fit.fit()

    termRef.current = term
    fitRef.current = fit

    const handleResize = () => fit.fit()
    window.addEventListener('resize', handleResize)

    return () => {
      window.removeEventListener('resize', handleResize)
      if (timerRef.current) cancelAnimationFrame(timerRef.current)
      term.dispose()
      termRef.current = null
      fitRef.current = null
    }
  }, [])

  // 加载录制数据
  useEffect(() => {
    if (!recording || !termRef.current) return

    const { header, events } = parseAsciicast(recording)
    eventsRef.current = events

    if (header.width && header.height) {
      termRef.current.resize(header.width, header.height)
    }

    const dur = events.length > 0 ? events[events.length - 1].time : 0
    setDuration(dur)
    setCurrentTime(0)
    setCurrentIndex(0)
    setPlaying(false)
    setReady(true)

    termRef.current.reset()

    setTimeout(() => fitRef.current?.fit(), 50)
  }, [recording])

  // 从头播放到指定时间点
  const replayTo = useCallback((targetTime: number) => {
    const term = termRef.current
    if (!term) return

    term.reset()
    const events = eventsRef.current
    let idx = 0
    for (; idx < events.length; idx++) {
      if (events[idx].time > targetTime) break
      if (events[idx].type === 'o') {
        term.write(events[idx].data)
      }
    }
    setCurrentIndex(idx)
    setCurrentTime(targetTime)
  }, [])

  // 播放循环
  useEffect(() => {
    if (!playing || !ready) return

    let lastFrameTime = performance.now()
    let accTime = currentTime

    const tick = (now: number) => {
      const delta = (now - lastFrameTime) / 1000 * speedRef.current
      lastFrameTime = now
      accTime += delta

      const events = eventsRef.current
      const term = termRef.current
      if (!term) return

      let idx = currentIndex
      while (idx < events.length && events[idx].time <= accTime) {
        if (events[idx].type === 'o') {
          term.write(events[idx].data)
        }
        idx++
      }

      setCurrentIndex(idx)
      setCurrentTime(accTime)

      if (idx >= events.length) {
        setPlaying(false)
        setCurrentTime(duration)
        return
      }

      timerRef.current = requestAnimationFrame(tick)
    }

    timerRef.current = requestAnimationFrame(tick)

    return () => {
      if (timerRef.current) {
        cancelAnimationFrame(timerRef.current)
        timerRef.current = null
      }
    }
  }, [playing, ready])

  const handlePlayPause = () => {
    if (!ready) return
    if (currentTime >= duration && !playing) {
      // 播放结束，从头开始
      replayTo(0)
      setPlaying(true)
      return
    }
    setPlaying(p => !p)
  }

  const handleSeek = (value: number) => {
    setPlaying(false)
    replayTo(value)
  }

  const handleSkip = (delta: number) => {
    const target = Math.max(0, Math.min(duration, currentTime + delta))
    setPlaying(false)
    replayTo(target)
  }

  const formatTime = (t: number) => {
    const mins = Math.floor(t / 60)
    const secs = Math.floor(t % 60)
    return `${mins.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`
  }

  if (loading) {
    return (
      <Spin tip="加载录制数据...">
        <div style={{ height: 400, background: '#282c34' }} />
      </Spin>
    )
  }

  if (recording !== null && ready && eventsRef.current.length === 0) {
    return (
      <div style={{
        display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center',
        height: 400, background: '#282c34', color: '#5c6370', gap: 12,
      }}>
        <DesktopOutlined style={{ fontSize: 40, color: '#3e4451' }} />
        <span style={{ fontSize: 14 }}>该会话没有录制内容</span>
        <span style={{ fontSize: 12, color: '#4b5263' }}>可能是会话异常断开导致录制数据未保存</span>
      </div>
    )
  }

  return (
    <div style={{ background: '#282c34', borderRadius: 2, overflow: 'hidden' }}>
      {/* 终端区域 */}
      <div
        ref={containerRef}
        style={{ width: '100%', height: 400, background: '#282c34', padding: '4px 6px' }}
      />

      {/* 控制栏 */}
      <div style={{
        height: 48,
        background: '#21252b',
        display: 'flex',
        alignItems: 'center',
        padding: '0 16px',
        gap: 12,
        borderTop: '1px solid #181a1f',
      }}>
        {/* 后退 5s */}
        <div
          onClick={() => handleSkip(-5)}
          style={{ cursor: 'pointer', color: '#6b717d', display: 'flex', padding: 4, transition: 'color 0.15s' }}
          onMouseEnter={e => { e.currentTarget.style.color = '#abb2bf' }}
          onMouseLeave={e => { e.currentTarget.style.color = '#6b717d' }}
        >
          <StepBackwardOutlined style={{ fontSize: 14 }} />
        </div>

        {/* 播放/暂停 */}
        <div
          onClick={handlePlayPause}
          style={{
            cursor: 'pointer',
            width: 32, height: 32,
            borderRadius: '50%',
            background: '#3A84FF',
            display: 'flex', alignItems: 'center', justifyContent: 'center',
            transition: 'background 0.2s',
          }}
          onMouseEnter={e => { e.currentTarget.style.background = '#5594FA' }}
          onMouseLeave={e => { e.currentTarget.style.background = '#3A84FF' }}
        >
          {playing
            ? <PauseOutlined style={{ fontSize: 14, color: '#fff' }} />
            : <CaretRightOutlined style={{ fontSize: 14, color: '#fff', marginLeft: 1 }} />}
        </div>

        {/* 前进 5s */}
        <div
          onClick={() => handleSkip(5)}
          style={{ cursor: 'pointer', color: '#6b717d', display: 'flex', padding: 4, transition: 'color 0.15s' }}
          onMouseEnter={e => { e.currentTarget.style.color = '#abb2bf' }}
          onMouseLeave={e => { e.currentTarget.style.color = '#6b717d' }}
        >
          <StepForwardOutlined style={{ fontSize: 14 }} />
        </div>

        {/* 时间 */}
        <span style={{ fontSize: 12, color: '#6b717d', fontFamily: 'monospace', minWidth: 90, textAlign: 'center' }}>
          {formatTime(currentTime)} / {formatTime(duration)}
        </span>

        {/* 进度条 */}
        <div style={{ flex: 1 }}>
          <Slider
            min={0}
            max={duration || 1}
            step={0.1}
            value={currentTime}
            onChange={handleSeek}
            tooltip={{ formatter: (v) => v !== undefined ? formatTime(v) : '' }}
            styles={{
              track: { background: '#3A84FF' },
              rail: { background: '#3e4451' },
            }}
          />
        </div>

        {/* 倍速 */}
        <Select
          value={speed}
          onChange={v => { setSpeed(v); speedRef.current = v }}
          size="small"
          variant="borderless"
          popupMatchSelectWidth={70}
          style={{ width: 70 }}
          options={[
            { value: 0.5, label: '0.5x' },
            { value: 1, label: '1x' },
            { value: 2, label: '2x' },
            { value: 4, label: '4x' },
            { value: 8, label: '8x' },
          ]}
        />
      </div>
    </div>
  )
}

export default SessionPlayer
