import { useState, useEffect, useRef, useCallback } from 'react'
import { useSearchParams } from 'react-router-dom'
import { Card, Table, Button, Tag, Modal, Tabs, Space, Popconfirm, message } from 'antd'
import {
  ReloadOutlined, PlayCircleOutlined, StopOutlined,
  UserOutlined, DesktopOutlined, ClockCircleOutlined,
} from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import {
  getTerminalSessions, getActiveSessions, killSession, getSessionRecording,
  type TerminalSession, type ActiveSession,
} from '../../../services/terminal'
import SessionPlayer from '../../../components/SessionPlayer'

const TerminalSessionsPage = () => {
  // 在线会话
  const [activeSessions, setActiveSessions] = useState<ActiveSession[]>([])
  const [activeLoading, setActiveLoading] = useState(false)
  const activeTimer = useRef<ReturnType<typeof setInterval>>()

  // 历史会话
  const [sessions, setSessions] = useState<TerminalSession[]>([])
  const [sessionsTotal, setSessionsTotal] = useState(0)
  const [sessionsLoading, setSessionsLoading] = useState(false)
  const [sessionsPage, setSessionsPage] = useState(1)

  // 回放
  const [replayOpen, setReplayOpen] = useState(false)
  const [replayLoading, setReplayLoading] = useState(false)
  const [replayData, setReplayData] = useState<string | null>(null)
  const [replaySession, setReplaySession] = useState<TerminalSession | null>(null)

  const [searchParams, setSearchParams] = useSearchParams()
  const [activeTab, setActiveTab] = useState(searchParams.get('tab') || 'active')

  const handleTabChange = useCallback((key: string) => {
    setActiveTab(key)
    setSearchParams({ tab: key }, { replace: true })
  }, [setSearchParams])

  const fetchActive = async () => {
    setActiveLoading(true)
    try {
      const list = await getActiveSessions()
      setActiveSessions(list || [])
    } catch { /* handled */ } finally { setActiveLoading(false) }
  }

  const fetchHistory = async (page = 1) => {
    setSessionsLoading(true)
    try {
      const r = await getTerminalSessions({ page, page_size: 20, status: 'disconnected' })
      setSessions(r.list || [])
      setSessionsTotal(r.total)
      setSessionsPage(page)
    } catch { /* handled */ } finally { setSessionsLoading(false) }
  }

  useEffect(() => {
    fetchActive()
    // 每 5 秒刷新在线会话
    activeTimer.current = setInterval(fetchActive, 5000)
    return () => clearInterval(activeTimer.current)
  }, [])

  useEffect(() => {
    if (activeTab === 'history') fetchHistory()
  }, [activeTab])

  const handleKill = async (sessionId: number) => {
    try {
      await killSession(sessionId)
      message.success('会话已终止')
      fetchActive()
    } catch {
      message.error('终止会话失败')
    }
  }

  const openReplay = async (session: TerminalSession) => {
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

  const activeColumns: ColumnsType<ActiveSession> = [
    {
      title: '用户', dataIndex: 'username', width: 120,
      render: (v) => (
        <span style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
          <UserOutlined style={{ color: '#3A84FF', fontSize: 13 }} />
          {v}
        </span>
      ),
    },
    {
      title: '目标主机', width: 200,
      render: (_, r) => (
        <span style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
          <DesktopOutlined style={{ color: '#979BA5', fontSize: 13 }} />
          <span>{r.hostname || r.ip}</span>
          {r.hostname && <span style={{ color: '#979BA5', fontSize: 12 }}>({r.ip})</span>}
        </span>
      ),
    },
    { title: '客户端 IP', dataIndex: 'client_ip', width: 140 },
    { title: '开始时间', dataIndex: 'started_at', width: 170 },
    {
      title: '持续时间', dataIndex: 'duration', width: 120,
      render: (v) => (
        <span style={{ display: 'flex', alignItems: 'center', gap: 4 }}>
          <ClockCircleOutlined style={{ color: '#FF9C01', fontSize: 12 }} />
          <span style={{ fontFamily: 'monospace' }}>{v}</span>
        </span>
      ),
    },
    {
      title: '状态', width: 80,
      render: () => <Tag color="green">在线</Tag>,
    },
    {
      title: '操作', width: 100, fixed: 'right',
      render: (_, record) => (
        <Popconfirm
          title="确认终止该会话？"
          description="终止后用户将被强制断开连接"
          onConfirm={() => handleKill(record.session_id)}
          okText="终止"
          cancelText="取消"
          okButtonProps={{ danger: true }}
        >
          <Button type="link" size="small" danger icon={<StopOutlined />}>
            终止
          </Button>
        </Popconfirm>
      ),
    },
  ]

  const historyColumns: ColumnsType<TerminalSession> = [
    { title: 'ID', dataIndex: 'id', width: 60 },
    {
      title: '用户', dataIndex: 'username', width: 120,
      render: (v) => (
        <span style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
          <UserOutlined style={{ color: '#3A84FF', fontSize: 13 }} />
          {v}
        </span>
      ),
    },
    {
      title: '目标主机', width: 200,
      render: (_, r) => (
        <span style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
          <DesktopOutlined style={{ color: '#979BA5', fontSize: 13 }} />
          <span>{r.hostname || r.ip}</span>
          {r.hostname && <span style={{ color: '#979BA5', fontSize: 12 }}>({r.ip})</span>}
        </span>
      ),
    },
    { title: '客户端 IP', dataIndex: 'client_ip', width: 140 },
    { title: '开始时间', dataIndex: 'started_at', width: 170 },
    { title: '结束时间', dataIndex: 'finished_at', width: 170, render: (v) => v || '-' },
    {
      title: '操作', width: 80, fixed: 'right',
      render: (_, record) => (
        <Button
          type="link"
          size="small"
          icon={<PlayCircleOutlined />}
          onClick={() => openReplay(record)}
        >
          回放
        </Button>
      ),
    },
  ]

  return (
    <>
      <Card style={{ border: '1px solid #E7E9EF' }}>
        <Tabs
          activeKey={activeTab}
          onChange={handleTabChange}
          tabBarExtraContent={
            <Space>
              <Button
                icon={<ReloadOutlined />}
                onClick={() => activeTab === 'active' ? fetchActive() : fetchHistory(sessionsPage)}
              >
                刷新
              </Button>
            </Space>
          }
          items={[
            {
              key: 'active',
              label: (
                <span>
                  在线会话
                  {activeSessions.length > 0 && (
                    <Tag color="green" style={{ marginLeft: 8, borderRadius: 10 }}>
                      {activeSessions.length}
                    </Tag>
                  )}
                </span>
              ),
              children: (
                <Table
                  rowKey="session_id"
                  columns={activeColumns}
                  dataSource={activeSessions}
                  loading={activeLoading}
                  size="middle"
                  scroll={{ x: 900 }}
                  pagination={false}
                  locale={{ emptyText: '当前没有在线会话' }}
                />
              ),
            },
            {
              key: 'history',
              label: '历史会话',
              children: (
                <Table
                  rowKey="id"
                  columns={historyColumns}
                  dataSource={sessions}
                  loading={sessionsLoading}
                  size="middle"
                  scroll={{ x: 1000 }}
                  pagination={{
                    current: sessionsPage,
                    pageSize: 20,
                    total: sessionsTotal,
                    showSizeChanger: true,
                    showTotal: (t) => `共 ${t} 条`,
                    onChange: (p) => fetchHistory(p),
                  }}
                />
              ),
            },
          ]}
        />
      </Card>

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
    </>
  )
}

export default TerminalSessionsPage
