import { useEffect, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import {
  Card, Row, Col, Tabs, Table, Tag, Space, Button, Descriptions, Typography, Tooltip, Progress, Badge, Modal, message,
} from 'antd'
import {
  ArrowLeftOutlined, CloudServerOutlined, DesktopOutlined,
  GlobalOutlined, SafetyOutlined, KeyOutlined,
  CodeOutlined, PlayCircleOutlined, ReloadOutlined, ApiOutlined,
} from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import {
  getAssetDetail, testConnection,
  type AssetDetail, type AssetDetailLinuxUser, type AssetDetailSSHKeyBinding,
  type AssetDetailTerminalSession, type AssetDetailProbeHistory,
} from '../../../services/asset'
import { getSessionRecording } from '../../../services/terminal'
import { executeProbe } from '../../../services/probe'
import SessionPlayer from '../../../components/SessionPlayer'

const { Text } = Typography

const statusColors: Record<string, string> = { online: 'green', offline: 'red', running: 'green', stopped: 'red', unknown: 'default', terminated: 'default' }
const statusLabels: Record<string, string> = { online: '在线', offline: '离线', running: '运行中', stopped: '已停止', unknown: '未知', terminated: '已终止' }
const envColors: Record<string, string> = { prod: 'red', test: 'orange', dev: 'blue' }

// ====== CPU 解析 ======
const parseCPU = (raw: string) => {
  const lines = raw.split('\n')
  const get = (key: string) => {
    const line = lines.find(l => l.includes(key))
    return line ? line.split(':').slice(1).join(':').trim() : '-'
  }
  return {
    model: get('Model name'),
    arch: get('Architecture'),
    cores: get('CPU(s)'),
    sockets: get('Socket(s)'),
    threadsPerCore: get('Thread(s) per core'),
    coresPerSocket: get('Core(s) per socket'),
    hypervisor: get('Hypervisor vendor'),
    virtualization: get('Virtualization type'),
    l1d: get('L1d cache'),
    l1i: get('L1i cache'),
    l2: get('L2 cache'),
    l3: get('L3 cache'),
  }
}

// ====== Memory 解析 ======
const parseMemory = (raw: string) => {
  const lines = raw.split('\n').filter(l => l.startsWith('Mem:'))
  if (lines.length === 0) return null
  const parts = lines[0].split(/\s+/)
  const total = parseInt(parts[1]) || 0
  const used = parseInt(parts[2]) || 0
  const free = parseInt(parts[3]) || 0
  const available = parseInt(parts[6]) || 0
  return { total, used, free, available, usedPct: total > 0 ? Math.round((used / total) * 100) : 0 }
}

// ====== Disk 解析 ======
interface DiskItem { filesystem: string; size: string; used: string; avail: string; usePct: number; mount: string }
const parseDisk = (raw: string): DiskItem[] => {
  return raw.split('\n').slice(1).filter(l => l.trim()).map(line => {
    const p = line.split(/\s+/)
    return {
      filesystem: p[0] || '',
      size: p[1] || '',
      used: p[2] || '',
      avail: p[3] || '',
      usePct: parseInt(p[4]) || 0,
      mount: p[5] || '',
    }
  }).filter(d => !d.filesystem.startsWith('tmpfs') && !d.filesystem.startsWith('devtmpfs'))
}

// ====== Network 解析 ======
interface NetIface { name: string; state: string; mac: string; ips: string[] }
const parseNetwork = (raw: string): NetIface[] => {
  const ifaces: NetIface[] = []
  let current: NetIface | null = null
  for (const line of raw.split('\n')) {
    const ifMatch = line.match(/^\d+:\s+(\S+):\s+<([^>]*)>/)
    if (ifMatch) {
      if (current) ifaces.push(current)
      current = { name: ifMatch[1].replace(':', ''), state: ifMatch[2].includes('UP') ? 'UP' : 'DOWN', mac: '', ips: [] }
    }
    if (current) {
      const macMatch = line.match(/link\/\w+\s+([\da-f:]+)/)
      if (macMatch) current.mac = macMatch[1]
      const ipMatch = line.match(/inet6?\s+(\S+)/)
      if (ipMatch) current.ips.push(ipMatch[1])
    }
  }
  if (current) ifaces.push(current)
  return ifaces.filter(i => i.name !== 'lo')
}

// ====== SSH Users 解析 ======
interface SSHUser { username: string; uid: string; gid: string; home: string; shell: string }
const parseSSHUsers = (raw: string): SSHUser[] => {
  return raw.split('\n').filter(l => l.trim()).map(line => {
    const p = line.split(':')
    return { username: p[0] || '', uid: p[2] || '', gid: p[3] || '', home: p[5] || '', shell: p[6] || '' }
  })
}

// ====== Running Services 解析 ======
const parseServices = (raw: string): string[] => {
  return raw.split('\n').filter(l => l.includes('.service')).map(line => {
    const p = line.trim().split(/\s+/)
    return p[0]?.replace('.service', '') || ''
  }).filter(Boolean)
}

// ====== Processes 解析 ======
interface ProcessItem { pid: string; ppid: string; user: string; cpu: string; mem: string; rss: string; stat: string; startTime: string; command: string }
const parseProcesses = (raw: string): ProcessItem[] => {
  const lines = raw.split('\n').filter(l => l.trim())
  if (lines.length <= 1) return []
  return lines.slice(1).map(line => {
    const p = line.trim().split(/\s+/)
    return {
      pid: p[0] || '', ppid: p[1] || '', user: p[2] || '', cpu: p[3] || '0',
      mem: p[4] || '0', rss: p[5] || '0', stat: p[6] || '', startTime: p[7] || '',
      command: p.slice(8).join(' ') || '',
    }
  }).filter(p => p.pid && p.command)
}

// ====== Listeners 解析 ======
interface ListenerItem { protocol: string; listenAddr: string; port: string; process: string }
const parseListeners = (raw: string): ListenerItem[] => {
  return raw.split('\n').filter(l => l.trim()).map(line => {
    const p = line.trim().split(/\s+/)
    const local = p[3] || ''
    const lastColon = local.lastIndexOf(':')
    const addr = lastColon > 0 ? local.substring(0, lastColon) : '*'
    const port = lastColon > 0 ? local.substring(lastColon + 1) : local
    const procField = p.slice(5).join(' ')
    const procMatch = procField.match(/users:\(\("([^"]+)"/)
    return {
      protocol: (p[0] || 'tcp').toUpperCase(),
      listenAddr: addr === '*' || addr === '0.0.0.0' || addr === '[::]' ? '全部' : addr,
      port,
      process: procMatch ? procMatch[1] : procField.replace(/users:\(\(/, '').replace(/\)\)/, '') || '-',
    }
  }).filter(l => l.port)
}

const formatRSS = (kb: string) => {
  const n = parseInt(kb) || 0
  if (n >= 1048576) return `${(n / 1048576).toFixed(1)} GB`
  if (n >= 1024) return `${(n / 1024).toFixed(0)} MB`
  return `${n} KB`
}

const formatMB = (mb: number) => mb >= 1024 ? `${(mb / 1024).toFixed(1)} GB` : `${mb} MB`

const AssetDetailPage = () => {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [detail, setDetail] = useState<AssetDetail | null>(null)
  const [probing, setProbing] = useState(false)
  const [testing, setTesting] = useState(false)
  const [replayOpen, setReplayOpen] = useState(false)
  const [replayLoading, setReplayLoading] = useState(false)
  const [replayData, setReplayData] = useState<string | null>(null)
  const [replayTitle, setReplayTitle] = useState('')

  const assetId = Number(id)

  const fetchDetail = async () => {
    try {
      const data = await getAssetDetail(assetId)
      setDetail(data)
    } catch { /* handled */ }
  }

  useEffect(() => {
    if (assetId) fetchDetail()
  }, [assetId])

  const handleProbe = async () => {
    setProbing(true)
    try {
      await executeProbe([assetId])
      setTimeout(fetchDetail, 6000)
    } catch { /* handled */ } finally { setTimeout(() => setProbing(false), 6000) }
  }

  const handleTestConn = async () => {
    setTesting(true)
    try {
      const resp = await testConnection([assetId])
      const r = resp.results?.[0]
      if (r) {
        const { Modal } = await import('antd')
        Modal.info({
          title: '连接测试结果',
          content: r.status === 'success'
            ? `连接成功，延迟 ${r.latency}ms，认证方式: ${r.auth_method}`
            : `连接失败: ${r.error}`,
        })
      }
    } catch { /* handled */ } finally { setTesting(false) }
  }

  const handleReplay = async (session: AssetDetailTerminalSession) => {
    if (session.status === 'connected') {
      message.warning('该会话正在进行中，无法回放')
      return
    }
    setReplayTitle(`${detail?.hostname || detail?.ip || ''} (${session.username})`)
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

  if (!detail) return null

  const cpu = detail.probe ? parseCPU(detail.probe.cpu) : null
  const mem = detail.probe ? parseMemory(detail.probe.memory) : null
  const disks = detail.probe ? parseDisk(detail.probe.disk) : []
  const nets = detail.probe ? parseNetwork(detail.probe.network) : []
  const sshUsers = detail.probe ? parseSSHUsers(detail.probe.ssh_users) : []
  const services = detail.probe ? parseServices(detail.probe.running_services) : []
  const processes = detail.probe ? parseProcesses(detail.probe.processes) : []
  const listeners = detail.probe ? parseListeners(detail.probe.listeners) : []

  // ====== Tab: 概览 ======
  const descLabelStyle = { color: '#979BA5', fontSize: 13 }
  const descContentStyle = { color: '#313238', fontSize: 13 }

  const OverviewTab = (
    <Row gutter={[16, 16]}>
      {/* 资源概览卡片 */}
      {mem && (
        <>
          <Col xs={24} sm={6}>
            <Card size="small" style={{ border: '1px solid #E7E9EF' }} styles={{ body: { padding: '16px 20px' } }}>
              <div style={{ color: '#979BA5', fontSize: 12, marginBottom: 6 }}>CPU</div>
              <div style={{ fontSize: 22, fontWeight: 600, color: '#313238' }}>{cpu?.cores || '-'} <span style={{ fontSize: 13, fontWeight: 400, color: '#979BA5' }}>核</span></div>
              <div style={{ fontSize: 12, color: '#979BA5', marginTop: 4, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{cpu?.model || '-'}</div>
            </Card>
          </Col>
          <Col xs={24} sm={6}>
            <Card size="small" style={{ border: '1px solid #E7E9EF' }} styles={{ body: { padding: '16px 20px' } }}>
              <div style={{ color: '#979BA5', fontSize: 12, marginBottom: 6 }}>内存</div>
              <div style={{ fontSize: 22, fontWeight: 600, color: '#313238' }}>{formatMB(mem.total)}</div>
              <Progress percent={mem.usedPct} strokeColor={mem.usedPct > 80 ? '#EA3636' : '#3A84FF'} size="small" style={{ marginTop: 4 }} />
              <div style={{ fontSize: 12, color: '#979BA5' }}>已用 {formatMB(mem.used)} / 可用 {formatMB(mem.available)}</div>
            </Card>
          </Col>
          <Col xs={24} sm={6}>
            <Card size="small" style={{ border: '1px solid #E7E9EF' }} styles={{ body: { padding: '16px 20px' } }}>
              <div style={{ color: '#979BA5', fontSize: 12, marginBottom: 6 }}>磁盘</div>
              <div style={{ fontSize: 22, fontWeight: 600, color: '#313238' }}>{disks.length} <span style={{ fontSize: 13, fontWeight: 400, color: '#979BA5' }}>分区</span></div>
              <div style={{ fontSize: 12, color: '#979BA5', marginTop: 4 }}>
                {disks.length > 0 ? `${disks[0].mount} ${disks[0].size}，已用 ${disks[0].usePct}%` : '-'}
              </div>
            </Card>
          </Col>
          <Col xs={24} sm={6}>
            <Card size="small" style={{ border: '1px solid #E7E9EF' }} styles={{ body: { padding: '16px 20px' } }}>
              <div style={{ color: '#979BA5', fontSize: 12, marginBottom: 6 }}>运行时间</div>
              <div style={{ fontSize: 22, fontWeight: 600, color: '#313238' }}>{detail.probe?.uptime ? detail.probe.uptime.replace('T', ' ').substring(0, 16) : '-'}</div>
              <div style={{ fontSize: 12, color: '#979BA5', marginTop: 4 }}>自上次启动</div>
            </Card>
          </Col>
        </>
      )}

      {/* 基本信息 */}
      <Col xs={24} lg={12}>
        <Card size="small" title="基本信息" style={{ border: '1px solid #E7E9EF' }}>
          <Descriptions column={2} size="small" labelStyle={descLabelStyle} contentStyle={descContentStyle}>
            <Descriptions.Item label="主机名">{detail.hostname || '-'}</Descriptions.Item>
            <Descriptions.Item label="IP 地址"><Text copyable style={{ fontSize: 13 }}>{detail.ip}</Text></Descriptions.Item>
            <Descriptions.Item label="公网 IP">{detail.probe?.public_ip || '-'}</Descriptions.Item>
            <Descriptions.Item label="SSH 端口">{detail.port}</Descriptions.Item>
            <Descriptions.Item label="操作系统">{detail.os || '-'}</Descriptions.Item>
            <Descriptions.Item label="内核">{detail.probe?.kernel || '-'}</Descriptions.Item>
            <Descriptions.Item label="Docker">{detail.probe?.docker_version && detail.probe.docker_version !== 'not installed' ? detail.probe.docker_version : '-'}</Descriptions.Item>
            <Descriptions.Item label="SSH 认证">
              {detail.has_password && <Tag color="blue" style={{ margin: 0 }}>密码</Tag>}
              {detail.ssh_key_id && <Tag color="green" style={{ margin: 0, marginLeft: detail.has_password ? 4 : 0 }}>密钥</Tag>}
              {!detail.has_password && !detail.ssh_key_id && <Tag style={{ margin: 0 }}>未配置</Tag>}
            </Descriptions.Item>
            <Descriptions.Item label="最后探测">{detail.probe_last_at || '未探测'}</Descriptions.Item>
            <Descriptions.Item label="类型"><Tag style={{ margin: 0 }}>{detail.type}</Tag></Descriptions.Item>
          </Descriptions>
        </Card>
      </Col>

      {/* 归属信息 + 设备信息 */}
      <Col xs={24} lg={12}>
        <Card size="small" title="归属与设备" style={{ border: '1px solid #E7E9EF' }}>
          <Descriptions column={2} size="small" labelStyle={descLabelStyle} contentStyle={descContentStyle}>
            <Descriptions.Item label="环境">{detail.environment ? <Tag color={envColors[detail.environment] || 'default'} style={{ margin: 0 }}>{detail.environment}</Tag> : '-'}</Descriptions.Item>
            <Descriptions.Item label="部门">{detail.department || '-'}</Descriptions.Item>
            <Descriptions.Item label="项目">{detail.project_name || '-'}</Descriptions.Item>
            <Descriptions.Item label="负责人">{detail.owner || '-'}</Descriptions.Item>
            <Descriptions.Item label="生产商">{detail.probe?.manufacturer || '-'}</Descriptions.Item>
            <Descriptions.Item label="设备型号">{detail.probe?.product_model || '-'}</Descriptions.Item>
            <Descriptions.Item label="序列号"><Text copyable={!!detail.probe?.serial_number} style={{ fontSize: 13 }}>{detail.probe?.serial_number || '-'}</Text></Descriptions.Item>
            <Descriptions.Item label="来源"><Tag style={{ margin: 0 }}>{detail.source}</Tag></Descriptions.Item>
          </Descriptions>
        </Card>
      </Col>

      {/* 网络摘要 */}
      <Col span={24}>
        <Card size="small" title="网络摘要" style={{ border: '1px solid #E7E9EF' }}>
          <Descriptions column={{ xs: 1, sm: 2, md: 4 }} size="small" labelStyle={descLabelStyle} contentStyle={descContentStyle}>
            <Descriptions.Item label="默认网关">{detail.probe?.gateway ? detail.probe.gateway.split(/\s+/).find(s => /\d+\.\d+\.\d+\.\d+/.test(s)) || detail.probe.gateway : '-'}</Descriptions.Item>
            <Descriptions.Item label="DNS 服务器">{detail.probe?.dns ? detail.probe.dns.split('\n').filter(Boolean).join(' / ') : '-'}</Descriptions.Item>
            <Descriptions.Item label="公网 IP"><Text copyable={!!detail.probe?.public_ip} style={{ fontSize: 13 }}>{detail.probe?.public_ip || '-'}</Text></Descriptions.Item>
            <Descriptions.Item label="网络接口">{nets.length} 个</Descriptions.Item>
          </Descriptions>
        </Card>
      </Col>

      {/* 标签 */}
      {detail.tags && detail.tags.length > 0 && (
        <Col span={24}>
          <Card size="small" title="标签" style={{ border: '1px solid #E7E9EF' }}>
            <Space wrap size={[8, 8]}>
              {detail.tags.map(t => <Tag key={`${t.key}:${t.value}`} color="blue">{t.key}: {t.value}</Tag>)}
            </Space>
          </Card>
        </Col>
      )}
    </Row>
  )

  // ====== Tab: 硬件信息 ======
  const diskColumns: ColumnsType<DiskItem> = [
    { title: '文件系统', dataIndex: 'filesystem', ellipsis: true },
    { title: '大小', dataIndex: 'size', width: 100 },
    { title: '已用', dataIndex: 'used', width: 100 },
    { title: '可用', dataIndex: 'avail', width: 100 },
    { title: '使用率', dataIndex: 'usePct', width: 120, render: v => <Progress percent={v} size="small" strokeColor={v > 80 ? '#EA3636' : '#3A84FF'} /> },
    { title: '挂载点', dataIndex: 'mount' },
  ]

  const HardwareTab = (
    <Row gutter={[16, 16]}>
      <Col span={24}>
        <Card size="small" title="CPU 信息" style={{ border: '1px solid #E7E9EF' }}>
          {cpu ? (
            <Descriptions column={{ xs: 1, sm: 2, md: 3 }} size="small">
              <Descriptions.Item label="型号">{cpu.model}</Descriptions.Item>
              <Descriptions.Item label="架构">{cpu.arch}</Descriptions.Item>
              <Descriptions.Item label="核心数">{cpu.cores}</Descriptions.Item>
              <Descriptions.Item label="Socket 数">{cpu.sockets}</Descriptions.Item>
              <Descriptions.Item label="每 Socket 核数">{cpu.coresPerSocket}</Descriptions.Item>
              <Descriptions.Item label="每核线程数">{cpu.threadsPerCore}</Descriptions.Item>
              <Descriptions.Item label="虚拟化平台">{cpu.hypervisor}</Descriptions.Item>
              <Descriptions.Item label="虚拟化类型">{cpu.virtualization}</Descriptions.Item>
              <Descriptions.Item label="L1d 缓存">{cpu.l1d}</Descriptions.Item>
              <Descriptions.Item label="L1i 缓存">{cpu.l1i}</Descriptions.Item>
              <Descriptions.Item label="L2 缓存">{cpu.l2}</Descriptions.Item>
              <Descriptions.Item label="L3 缓存">{cpu.l3}</Descriptions.Item>
            </Descriptions>
          ) : <Text type="secondary">暂无数据，请先执行探针采集</Text>}
        </Card>
      </Col>
      <Col span={24}>
        <Card size="small" title="内存信息" style={{ border: '1px solid #E7E9EF' }}>
          {mem ? (
            <div>
              <Row gutter={16} style={{ marginBottom: 16 }}>
                <Col span={6}><Text type="secondary">总量</Text><div style={{ fontSize: 18, fontWeight: 600 }}>{formatMB(mem.total)}</div></Col>
                <Col span={6}><Text type="secondary">已用</Text><div style={{ fontSize: 18, fontWeight: 600, color: '#EA3636' }}>{formatMB(mem.used)}</div></Col>
                <Col span={6}><Text type="secondary">空闲</Text><div style={{ fontSize: 18, fontWeight: 600, color: '#2DCB56' }}>{formatMB(mem.free)}</div></Col>
                <Col span={6}><Text type="secondary">可用</Text><div style={{ fontSize: 18, fontWeight: 600, color: '#3A84FF' }}>{formatMB(mem.available)}</div></Col>
              </Row>
              <Progress percent={mem.usedPct} strokeColor={mem.usedPct > 80 ? '#EA3636' : '#3A84FF'} format={pct => `${pct}% 已用`} />
            </div>
          ) : <Text type="secondary">暂无数据</Text>}
        </Card>
      </Col>
      <Col span={24}>
        <Card size="small" title="磁盘分区" style={{ border: '1px solid #E7E9EF' }}>
          <Table rowKey="mount" columns={diskColumns} dataSource={disks} pagination={false} size="middle" />
        </Card>
      </Col>
    </Row>
  )

  // ====== Tab: 网络 ======
  const netColumns: ColumnsType<NetIface> = [
    { title: '网卡', dataIndex: 'name', width: 120 },
    { title: '状态', dataIndex: 'state', width: 80, render: v => <Badge status={v === 'UP' ? 'success' : 'error'} text={v} /> },
    { title: 'MAC', dataIndex: 'mac', width: 180 },
    { title: 'IP 地址', dataIndex: 'ips', render: (ips: string[]) => ips.map(ip => <Tag key={ip}>{ip}</Tag>) },
  ]

  const NetworkTab = (
    <Row gutter={[16, 16]}>
      <Col span={24}>
        <Card size="small" title="网络概况" style={{ border: '1px solid #E7E9EF' }}>
          <Descriptions column={{ xs: 1, sm: 2, md: 4 }} size="small" labelStyle={descLabelStyle} contentStyle={descContentStyle}>
            <Descriptions.Item label="默认网关">{detail.probe?.gateway ? detail.probe.gateway.split(/\s+/).find(s => /\d+\.\d+\.\d+\.\d+/.test(s)) || '-' : '-'}</Descriptions.Item>
            <Descriptions.Item label="DNS 服务器">
              {detail.probe?.dns ? (
                <Space size={4} wrap>{detail.probe.dns.split('\n').filter(Boolean).map(d => <Tag key={d} style={{ margin: 0 }}>{d}</Tag>)}</Space>
              ) : '-'}
            </Descriptions.Item>
            <Descriptions.Item label="公网 IP"><Text copyable={!!detail.probe?.public_ip} style={{ fontSize: 13 }}>{detail.probe?.public_ip || '-'}</Text></Descriptions.Item>
            <Descriptions.Item label="接口数量">{nets.length} 个</Descriptions.Item>
          </Descriptions>
        </Card>
      </Col>
      <Col span={24}>
        <Card size="small" title="网络接口" style={{ border: '1px solid #E7E9EF' }}>
          <Table rowKey="name" columns={netColumns} dataSource={nets} pagination={false} size="middle" />
        </Card>
      </Col>
    </Row>
  )

  // ====== Tab: 进程与服务 ======
  const processColumns: ColumnsType<ProcessItem> = [
    { title: 'PID', dataIndex: 'pid', width: 70 },
    { title: 'PPID', dataIndex: 'ppid', width: 70 },
    { title: '用户', dataIndex: 'user', width: 90, ellipsis: true },
    { title: 'CPU%', dataIndex: 'cpu', width: 70, sorter: (a, b) => parseFloat(a.cpu) - parseFloat(b.cpu), render: v => { const n = parseFloat(v); return <span style={{ color: n > 50 ? '#EA3636' : n > 10 ? '#FF9C01' : '#313238' }}>{v}%</span> } },
    { title: 'MEM%', dataIndex: 'mem', width: 70, sorter: (a, b) => parseFloat(a.mem) - parseFloat(b.mem), render: v => { const n = parseFloat(v); return <span style={{ color: n > 50 ? '#EA3636' : n > 10 ? '#FF9C01' : '#313238' }}>{v}%</span> } },
    { title: 'RSS', dataIndex: 'rss', width: 80, render: v => formatRSS(v) },
    { title: '状态', dataIndex: 'stat', width: 60 },
    { title: '启动时间', dataIndex: 'startTime', width: 80 },
    { title: '进程名', dataIndex: 'command', ellipsis: true },
  ]

  const listenerColumns: ColumnsType<ListenerItem> = [
    { title: '协议', dataIndex: 'protocol', width: 70 },
    { title: '监听地址', dataIndex: 'listenAddr', width: 120 },
    { title: '端口', dataIndex: 'port', width: 80, sorter: (a, b) => parseInt(a.port) - parseInt(b.port) },
    { title: '进程', dataIndex: 'process', ellipsis: true },
  ]

  const ServicesTab = (
    <Row gutter={[16, 16]}>
      <Col span={24}>
        <Card size="small" title={<span>Top 进程 <span style={{ color: '#979BA5', fontWeight: 400 }}>（按内存排序，共 {processes.length} 条）</span></span>} style={{ border: '1px solid #E7E9EF' }}>
          {processes.length > 0 ? (
            <Table rowKey={(_, i) => `p${i}`} columns={processColumns} dataSource={processes} pagination={false} size="middle" scroll={{ y: 360 }} />
          ) : <Text type="secondary">暂无数据，请先执行探针采集</Text>}
        </Card>
      </Col>
      <Col xs={24} lg={10}>
        <Card size="small" title={<span>监听端口 <span style={{ color: '#979BA5', fontWeight: 400 }}>（{listeners.length}）</span></span>} style={{ border: '1px solid #E7E9EF' }}>
          {listeners.length > 0 ? (
            <Table rowKey={(_, i) => `l${i}`} columns={listenerColumns} dataSource={listeners} pagination={false} size="middle" scroll={{ y: 320 }} />
          ) : <Text type="secondary">暂无数据</Text>}
        </Card>
      </Col>
      <Col xs={24} lg={14}>
        <Card size="small" title={<span>Systemd 服务 <span style={{ color: '#979BA5', fontWeight: 400 }}>（{services.length}）</span></span>} style={{ border: '1px solid #E7E9EF' }}>
          {services.length > 0 ? (
            <div style={{ maxHeight: 360, overflow: 'auto', display: 'flex', flexWrap: 'wrap', gap: 6, alignContent: 'flex-start' }}>
              {services.map(s => <Tag key={s} color="blue" style={{ margin: 0 }}>{s}</Tag>)}
            </div>
          ) : <Text type="secondary">暂无数据</Text>}
        </Card>
      </Col>
    </Row>
  )

  // ====== Tab: 系统用户 ======
  const sshUserColumns: ColumnsType<SSHUser> = [
    { title: '用户名', dataIndex: 'username', width: 120 },
    { title: 'UID', dataIndex: 'uid', width: 80 },
    { title: 'GID', dataIndex: 'gid', width: 80 },
    { title: '主目录', dataIndex: 'home' },
    { title: 'Shell', dataIndex: 'shell' },
  ]

  const SSHUsersTab = (
    <Card size="small" title={`系统用户 (${sshUsers.length})`} style={{ border: '1px solid #E7E9EF' }}>
      <Table rowKey="username" columns={sshUserColumns} dataSource={sshUsers} pagination={false} size="middle" />
    </Card>
  )

  // ====== Tab: 关联信息 ======
  const bindingColumns: ColumnsType<AssetDetailSSHKeyBinding> = [
    { title: '密钥名称', dataIndex: 'key_name' },
    { title: '部署用户', dataIndex: 'username' },
    { title: '状态', dataIndex: 'status', render: v => <Tag color={v === 'deployed' ? 'green' : v === 'failed' ? 'red' : 'default'}>{v}</Tag> },
    { title: '部署时间', dataIndex: 'deployed_at', render: v => v || '-' },
  ]

  const linuxUserColumns: ColumnsType<AssetDetailLinuxUser> = [
    { title: '用户名', dataIndex: 'username' },
    { title: 'UID', dataIndex: 'uid', width: 80 },
    { title: 'GID', dataIndex: 'gid', width: 80 },
    { title: '主目录', dataIndex: 'home' },
    { title: 'Shell', dataIndex: 'shell' },
    { title: 'Sudo', dataIndex: 'sudo', width: 80, render: v => v ? <Tag color="blue">是</Tag> : <Tag>否</Tag> },
    { title: '状态', dataIndex: 'status', render: v => <Tag color={v === 'active' ? 'green' : 'red'}>{v}</Tag> },
  ]

  const sessionColumns: ColumnsType<AssetDetailTerminalSession> = [
    { title: '用户', dataIndex: 'username' },
    { title: '状态', dataIndex: 'status', render: v => <Badge status={v === 'connected' ? 'processing' : 'default'} text={v} /> },
    { title: '客户端 IP', dataIndex: 'client_ip' },
    { title: '开始时间', dataIndex: 'started_at' },
    { title: '结束时间', dataIndex: 'finished_at', render: v => v || '-' },
    {
      title: '操作', width: 80,
      render: (_, record) => (
        <Button
          type="link"
          size="small"
          icon={<PlayCircleOutlined />}
          disabled={record.status === 'connected'}
          onClick={() => handleReplay(record)}
        >
          回放
        </Button>
      ),
    },
  ]

  const historyColumns: ColumnsType<AssetDetailProbeHistory> = [
    { title: '状态', dataIndex: 'status', render: v => <Tag color={v === 'success' ? 'green' : 'red'}>{v}</Tag> },
    { title: '操作系统', dataIndex: 'os' },
    { title: '内核', dataIndex: 'kernel' },
    { title: '采集时间', dataIndex: 'collected_at' },
  ]

  const RelationsTab = (
    <Row gutter={[16, 16]}>
      <Col span={24}>
        <Card size="small" title={`SSH 密钥绑定 (${detail.ssh_key_bindings.length})`} style={{ border: '1px solid #E7E9EF' }}>
          <Table rowKey="id" columns={bindingColumns} dataSource={detail.ssh_key_bindings} pagination={false} size="middle" />
        </Card>
      </Col>
      <Col span={24}>
        <Card size="small" title={`Linux 用户 (${detail.linux_users.length})`} style={{ border: '1px solid #E7E9EF' }}>
          <Table rowKey="id" columns={linuxUserColumns} dataSource={detail.linux_users} pagination={false} size="middle" />
        </Card>
      </Col>
      <Col span={24}>
        <Card size="small" title={`终端会话 (${detail.terminal_sessions.length})`} style={{ border: '1px solid #E7E9EF' }}>
          <Table rowKey="id" columns={sessionColumns} dataSource={detail.terminal_sessions} pagination={false} size="middle" />
        </Card>
      </Col>
      <Col span={24}>
        <Card size="small" title={`探针历史 (${detail.probe_history.length})`} style={{ border: '1px solid #E7E9EF' }}>
          <Table rowKey="id" columns={historyColumns} dataSource={detail.probe_history} pagination={false} size="middle" />
        </Card>
      </Col>
    </Row>
  )

  const tabItems = [
    { key: 'overview', label: <span><CloudServerOutlined /> 概览</span>, children: OverviewTab },
    { key: 'hardware', label: <span><DesktopOutlined /> 硬件信息</span>, children: HardwareTab },
    { key: 'network', label: <span><GlobalOutlined /> 网络</span>, children: NetworkTab },
    { key: 'services', label: <span><ApiOutlined /> 进程与服务</span>, children: ServicesTab },
    { key: 'sshusers', label: <span><SafetyOutlined /> 系统用户</span>, children: SSHUsersTab },
    { key: 'relations', label: <span><KeyOutlined /> 关联信息</span>, children: RelationsTab },
  ]

  return (
    <div>
      {/* Header */}
      <Card style={{ border: '1px solid #E7E9EF', marginBottom: 16 }} styles={{ body: { padding: '12px 24px' } }}>
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
          <Space size={16}>
            <Button type="text" icon={<ArrowLeftOutlined />} onClick={() => navigate(-1)} />
            <div>
              <Space size={8} align="center">
                <Typography.Title level={4} style={{ margin: 0 }}>
                  {detail.hostname || detail.ip}
                </Typography.Title>
                <Tag color={statusColors[detail.status] || 'default'}>{statusLabels[detail.status] || detail.status}</Tag>
                {detail.environment && <Tag color={envColors[detail.environment] || 'default'}>{detail.environment}</Tag>}
                <Tag>{detail.type}</Tag>
              </Space>
              <div style={{ color: '#979BA5', fontSize: 13, marginTop: 2 }}>
                <Text copyable={{ text: detail.ip }} style={{ color: '#979BA5', fontSize: 13 }}>{detail.ip}:{detail.port}</Text>
                <span style={{ margin: '0 12px' }}>|</span>
                {detail.os || '未知 OS'}
                {detail.owner && <><span style={{ margin: '0 12px' }}>|</span>负责人: {detail.owner}</>}
              </div>
            </div>
          </Space>
          <Space>
            <Tooltip title="测试连接">
              <Button icon={<PlayCircleOutlined />} loading={testing} onClick={handleTestConn}>连接测试</Button>
            </Tooltip>
            <Tooltip title="执行探针采集">
              <Button icon={<ReloadOutlined />} loading={probing} onClick={handleProbe}>探针采集</Button>
            </Tooltip>
            <Tooltip title="打开终端">
              <Button type="primary" icon={<CodeOutlined />} onClick={() => navigate(`/terminal?asset=${assetId}`)}>终端</Button>
            </Tooltip>
          </Space>
        </div>
      </Card>

      {/* Tabs */}
      <Card style={{ border: '1px solid #E7E9EF' }}>
        <Tabs items={tabItems} />
      </Card>

      {/* 会话回放 Modal */}
      <Modal
        title={`会话回放 — ${replayTitle}`}
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

export default AssetDetailPage
