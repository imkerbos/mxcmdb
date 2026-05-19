import { useState, useEffect, useRef, useCallback } from 'react'
import { useSearchParams } from 'react-router-dom'
import { Input, Tag, Drawer, Table, Tooltip, Modal, Button, message } from 'antd'
import {
  SearchOutlined, DesktopOutlined, HistoryOutlined,
  CloseOutlined, ExpandOutlined, CompressOutlined,
  PlayCircleOutlined,
} from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import { connectTerminalWS, getTerminalSessions, getSessionRecording, type TerminalSession } from '../../services/terminal'
import { getAllAssets, type AssetSimple } from '../../services/asset'
import TerminalComponent, { type TerminalHandle } from '../../components/Terminal'
import SessionPlayer from '../../components/SessionPlayer'

interface TerminalTab {
  key: string
  assetId: number
  title: string
  ip: string
  ws: WebSocket
  status: 'connecting' | 'connected' | 'disconnected'
}

const statusDotColor = (status: string) => {
  switch (status) {
    case 'running': case 'connected': return '#2DCB56'
    case 'connecting': return '#FF9C01'
    default: return '#C4C6CC'
  }
}

const statusLabel = (status: string) => {
  switch (status) {
    case 'connected': return '已连接'
    case 'connecting': return '连接中...'
    default: return '已断开'
  }
}

const TerminalPage = () => {
  const [assets, setAssets] = useState<AssetSimple[]>([])
  const [keyword, setKeyword] = useState('')
  const [tabs, setTabs] = useState<TerminalTab[]>([])
  const [activeKey, setActiveKey] = useState<string | null>(null)
  const [isFullscreen, setIsFullscreen] = useState(false)
  const [termSize, setTermSize] = useState<{ cols: number; rows: number } | null>(null)

  // 会话记录 Drawer
  const [sessionsOpen, setSessionsOpen] = useState(false)
  const [sessions, setSessions] = useState<TerminalSession[]>([])
  const [sessionsTotal, setSessionsTotal] = useState(0)
  const [sessionsLoading, setSessionsLoading] = useState(false)

  // 会话回放
  const [replayOpen, setReplayOpen] = useState(false)
  const [replayLoading, setReplayLoading] = useState(false)
  const [replayData, setReplayData] = useState<string | null>(null)
  const [replaySession, setReplaySession] = useState<TerminalSession | null>(null)

  const [searchParams] = useSearchParams()
  const termRefs = useRef<Map<string, TerminalHandle>>(new Map())
  const tabsRef = useRef(tabs)
  tabsRef.current = tabs
  const autoConnected = useRef(false)

  useEffect(() => {
    getAllAssets().then(list => {
      setAssets(list)
      // URL 参数自动连接（Shift+Click 新窗口场景）
      const connectId = searchParams.get('connect')
      if (connectId && !autoConnected.current) {
        autoConnected.current = true
        const target = list.find(a => a.id === Number(connectId))
        if (target) createTab(target)
      }
    }).catch(() => {})
    return () => {
      tabsRef.current.forEach(t => t.ws.close())
    }
  }, [])

  // 切换 tab / 全屏时重新 fit
  useEffect(() => {
    if (activeKey) {
      setTimeout(() => termRefs.current.get(activeKey)?.fit(), 60)
    }
  }, [activeKey, isFullscreen])

  const filteredAssets = assets.filter(a => {
    if (!keyword) return true
    const kw = keyword.toLowerCase()
    return (a.hostname || '').toLowerCase().includes(kw) || a.ip.includes(kw)
  })

  // 创建新 tab（跳过复用检查）
  const createTab = useCallback((asset: AssetSimple) => {
    const ws = connectTerminalWS(asset.id)
    if (!ws) return

    const key = `term_${asset.id}_${Date.now()}`
    const tab: TerminalTab = {
      key,
      assetId: asset.id,
      title: asset.hostname || asset.ip,
      ip: asset.ip,
      ws,
      status: 'connecting',
    }

    setTabs(prev => [...prev, tab])
    setActiveKey(key)
  }, [])

  const openTerminal = (asset: AssetSimple, e?: React.MouseEvent) => {
    // Shift+Click: 新浏览器窗口打开
    if (e?.shiftKey) {
      window.open(`/terminal?connect=${asset.id}`, '_blank')
      return
    }

    // Ctrl/Cmd+Click: 强制新建 tab
    if (e?.ctrlKey || e?.metaKey) {
      createTab(asset)
      return
    }

    // 普通点击: 复用已有 tab
    const existing = tabs.find(t => t.assetId === asset.id && t.status !== 'disconnected')
    if (existing) {
      setActiveKey(existing.key)
      return
    }
    createTab(asset)
  }

  const closeTab = (key: string) => {
    const tab = tabs.find(t => t.key === key)
    if (tab) tab.ws.close()
    termRefs.current.delete(key)

    const remaining = tabs.filter(t => t.key !== key)
    setTabs(remaining)

    if (activeKey === key) {
      setActiveKey(remaining.length > 0 ? remaining[remaining.length - 1].key : null)
    }
  }

  const handleStatusChange = (key: string, status: TerminalTab['status']) => {
    setTabs(prev => prev.map(t => t.key === key ? { ...t, status } : t))
  }

  const fetchSessions = async (page = 1) => {
    setSessionsLoading(true)
    try {
      const r = await getTerminalSessions({ page, page_size: 10 })
      setSessions(r.list || [])
      setSessionsTotal(r.total)
    } catch { /* handled */ } finally { setSessionsLoading(false) }
  }

  const openSessions = () => {
    setSessionsOpen(true)
    fetchSessions()
  }

  const openReplay = async (session: TerminalSession) => {
    if (session.status === 'connected') {
      message.warning('该会话正在进行中，无法回放')
      return
    }
    setReplaySession(session)
    setReplayData(null)
    setReplayOpen(true)
    setReplayLoading(true)
    try {
      const data = await getSessionRecording(session.id)
      if (!data || data.trim().length === 0) {
        message.warning('该会话没有录制数据')
        setReplayOpen(false)
        return
      }
      setReplayData(data)
    } catch {
      message.error('获取录制数据失败')
      setReplayOpen(false)
    } finally {
      setReplayLoading(false)
    }
  }

  const activeTab = tabs.find(t => t.key === activeKey)

  const sessionColumns: ColumnsType<TerminalSession> = [
    { title: 'ID', dataIndex: 'id', width: 60 },
    { title: '主机', dataIndex: 'hostname', width: 140, render: (v, r) => v || r.ip },
    { title: 'IP', dataIndex: 'ip', width: 130 },
    { title: '用户', dataIndex: 'username', width: 100 },
    {
      title: '状态', dataIndex: 'status', width: 100,
      render: (v) => <Tag color={v === 'connected' ? 'green' : 'default'}>{v === 'connected' ? '在线' : '已断开'}</Tag>,
    },
    { title: '开始时间', dataIndex: 'started_at', width: 170 },
    { title: '结束时间', dataIndex: 'finished_at', width: 170, render: (v) => v || '-' },
    {
      title: '操作', width: 80, fixed: 'right',
      render: (_, record) => (
        <Button
          type="link"
          size="small"
          icon={<PlayCircleOutlined />}
          disabled={record.status === 'connected'}
          onClick={() => openReplay(record)}
        >
          回放
        </Button>
      ),
    },
  ]

  return (
    <div style={{
      margin: '-20px -24px',
      height: 'calc(100vh - 52px)',
      display: 'flex',
      overflow: 'hidden',
    }}>
      {/* ===== Left Sidebar ===== */}
      {!isFullscreen && (
        <div style={{
          width: 280,
          flexShrink: 0,
          background: '#fff',
          borderRight: '1px solid #E7E9EF',
          display: 'flex',
          flexDirection: 'column',
        }}>
          {/* Header */}
          <div style={{ padding: '16px 16px 12px', borderBottom: '1px solid #E7E9EF' }}>
            <div style={{
              fontSize: 15, fontWeight: 600, color: '#313238',
              marginBottom: 12, display: 'flex', alignItems: 'center', gap: 8,
            }}>
              <DesktopOutlined style={{ color: '#3A84FF' }} />
              Web Terminal
            </div>
            <Input
              placeholder="搜索主机名 / IP"
              prefix={<SearchOutlined style={{ color: '#C4C6CC' }} />}
              value={keyword}
              onChange={e => setKeyword(e.target.value)}
              allowClear
              size="small"
            />
          </div>

          {/* Asset List */}
          <div style={{ flex: 1, overflow: 'auto', padding: '4px 0' }}>
            {filteredAssets.map(asset => {
              const hasTab = tabs.some(t => t.assetId === asset.id && t.status !== 'disconnected')
              return (
                <div
                  key={asset.id}
                  onClick={(e) => openTerminal(asset, e)}
                  style={{
                    padding: '10px 16px',
                    cursor: 'pointer',
                    display: 'flex',
                    alignItems: 'center',
                    gap: 10,
                    transition: 'background 0.15s ease',
                    borderLeft: hasTab ? '3px solid #3A84FF' : '3px solid transparent',
                  }}
                  onMouseEnter={e => { e.currentTarget.style.background = '#F0F5FF' }}
                  onMouseLeave={e => { e.currentTarget.style.background = 'transparent' }}
                >
                  <span style={{
                    width: 8, height: 8, borderRadius: '50%',
                    background: statusDotColor(asset.status), flexShrink: 0,
                  }} />
                  <div style={{ flex: 1, minWidth: 0 }}>
                    <div style={{
                      fontSize: 13, fontWeight: 500, color: '#313238',
                      overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap',
                    }}>
                      {asset.hostname || asset.ip}
                    </div>
                    <div style={{ fontSize: 12, color: '#979BA5', marginTop: 2, display: 'flex', alignItems: 'center', gap: 6 }}>
                      {asset.ip}
                      {asset.type && (
                        <Tag style={{ fontSize: 11, lineHeight: '16px', padding: '0 4px', margin: 0 }} color="default">{asset.type}</Tag>
                      )}
                    </div>
                  </div>
                </div>
              )
            })}
            {filteredAssets.length === 0 && (
              <div style={{ padding: '40px 16px', textAlign: 'center', color: '#C4C6CC', fontSize: 13 }}>
                {keyword ? '未找到匹配资产' : '暂无可用资产'}
              </div>
            )}
          </div>

          {/* Footer */}
          <div
            onClick={openSessions}
            style={{
              padding: '12px 16px', borderTop: '1px solid #E7E9EF',
              cursor: 'pointer', display: 'flex', alignItems: 'center', gap: 8,
              color: '#63656E', fontSize: 13, transition: 'background 0.15s ease',
            }}
            onMouseEnter={e => { e.currentTarget.style.background = '#F0F5FF' }}
            onMouseLeave={e => { e.currentTarget.style.background = 'transparent' }}
          >
            <HistoryOutlined />
            会话记录
          </div>
        </div>
      )}

      {/* ===== Right Terminal Area ===== */}
      <div style={{ flex: 1, display: 'flex', flexDirection: 'column', minWidth: 0 }}>
        {tabs.length > 0 ? (
          <>
            {/* Tab Bar */}
            <div style={{
              height: 38, background: '#21252b', flexShrink: 0,
              display: 'flex', alignItems: 'end', paddingLeft: 8, gap: 1,
            }}>
              {tabs.map(tab => {
                const isActive = tab.key === activeKey
                return (
                  <div
                    key={tab.key}
                    onClick={() => setActiveKey(tab.key)}
                    style={{
                      height: 32, padding: '0 14px',
                      display: 'flex', alignItems: 'center', gap: 8,
                      cursor: 'pointer', fontSize: 12, userSelect: 'none',
                      color: isActive ? '#d7dae0' : '#6b717d',
                      background: isActive ? '#282c34' : 'transparent',
                      borderRadius: '6px 6px 0 0',
                      transition: 'all 0.15s ease',
                      maxWidth: 200,
                    }}
                    onMouseEnter={e => { if (!isActive) e.currentTarget.style.background = '#2c313a' }}
                    onMouseLeave={e => { if (!isActive) e.currentTarget.style.background = 'transparent' }}
                  >
                    <span style={{
                      width: 7, height: 7, borderRadius: '50%',
                      background: statusDotColor(tab.status), flexShrink: 0,
                    }} />
                    <span style={{ overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                      {tab.title}
                    </span>
                    <CloseOutlined
                      style={{ fontSize: 9, color: '#6b717d', flexShrink: 0, padding: 2 }}
                      onClick={e => { e.stopPropagation(); closeTab(tab.key) }}
                    />
                  </div>
                )
              })}

              <div style={{ flex: 1 }} />
              <Tooltip title={isFullscreen ? '退出全屏' : '全屏'}>
                <div
                  onClick={() => setIsFullscreen(f => !f)}
                  style={{
                    width: 32, height: 32, display: 'flex', alignItems: 'center', justifyContent: 'center',
                    cursor: 'pointer', color: '#6b717d', transition: 'color 0.15s',
                  }}
                  onMouseEnter={e => { e.currentTarget.style.color = '#abb2bf' }}
                  onMouseLeave={e => { e.currentTarget.style.color = '#6b717d' }}
                >
                  {isFullscreen ? <CompressOutlined style={{ fontSize: 13 }} /> : <ExpandOutlined style={{ fontSize: 13 }} />}
                </div>
              </Tooltip>
            </div>

            {/* Terminal Body */}
            <div style={{ flex: 1, position: 'relative', background: '#282c34' }}>
              {tabs.map(tab => (
                <div
                  key={tab.key}
                  style={{
                    position: 'absolute', inset: 0,
                    padding: '6px 6px 0',
                    display: tab.key === activeKey ? 'block' : 'none',
                  }}
                >
                  <TerminalComponent
                    ref={handle => {
                      if (handle) termRefs.current.set(tab.key, handle)
                      else termRefs.current.delete(tab.key)
                    }}
                    ws={tab.ws}
                    onStatusChange={status => handleStatusChange(tab.key, status)}
                    onSizeChange={setTermSize}
                  />
                </div>
              ))}
            </div>

            {/* Status Bar */}
            <div style={{
              height: 26, background: '#21252b', flexShrink: 0,
              display: 'flex', alignItems: 'center', padding: '0 14px',
              fontSize: 11, color: '#5c6370', gap: 16,
              borderTop: '1px solid #181a1f',
            }}>
              <span style={{ fontFamily: 'monospace' }}>SSH</span>
              {termSize && <span style={{ fontFamily: 'monospace' }}>{termSize.cols} × {termSize.rows}</span>}
              <span style={{ fontFamily: 'monospace' }}>UTF-8</span>
              <div style={{ flex: 1 }} />
              {activeTab && (
                <span style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
                  <span style={{
                    width: 6, height: 6, borderRadius: '50%',
                    background: statusDotColor(activeTab.status),
                    boxShadow: activeTab.status === 'connected' ? '0 0 6px rgba(45,203,86,0.4)' : 'none',
                  }} />
                  <span style={{ color: activeTab.status === 'connected' ? '#98c379' : '#5c6370' }}>
                    {statusLabel(activeTab.status)}
                  </span>
                  <span style={{ color: '#3e4451' }}>·</span>
                  <span>{activeTab.ip}</span>
                </span>
              )}
            </div>
          </>
        ) : (
          /* Empty State */
          <div style={{
            flex: 1, display: 'flex', flexDirection: 'column',
            alignItems: 'center', justifyContent: 'center',
            background: '#282c34',
          }}>
            <div style={{
              width: 80, height: 80, borderRadius: '50%',
              background: '#2c313a', display: 'flex',
              alignItems: 'center', justifyContent: 'center',
              marginBottom: 24,
            }}>
              <DesktopOutlined style={{ fontSize: 36, color: '#4b5263' }} />
            </div>
            <div style={{ fontSize: 16, color: '#abb2bf', marginBottom: 8 }}>选择左侧资产开始连接</div>
            <div style={{ fontSize: 13, color: '#5c6370' }}>支持多终端标签页，实时 SSH 会话</div>
          </div>
        )}
      </div>

      {/* Sessions Drawer */}
      <Drawer
        title="会话记录"
        open={sessionsOpen}
        onClose={() => setSessionsOpen(false)}
        width={720}
      >
        <Table
          rowKey="id"
          columns={sessionColumns}
          dataSource={sessions}
          loading={sessionsLoading}
          size="small"
          scroll={{ x: 900 }}
          pagination={{
            pageSize: 10,
            total: sessionsTotal,
            showTotal: (t) => `共 ${t} 条`,
            onChange: (p) => fetchSessions(p),
          }}
        />
      </Drawer>

      {/* 会话回放 Modal */}
      <Modal
        title={replaySession ? `会话回放 — ${replaySession.hostname || replaySession.ip} (${replaySession.username})` : '会话回放'}
        open={replayOpen}
        onCancel={() => { setReplayOpen(false); setReplayData(null) }}
        footer={null}
        width={960}
        destroyOnHidden
        styles={{ body: { padding: 0 } }}
      >
        <SessionPlayer recording={replayData} loading={replayLoading} />
      </Modal>
    </div>
  )
}

export default TerminalPage
