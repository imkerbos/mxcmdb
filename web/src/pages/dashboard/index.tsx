import { useEffect, useState, useMemo } from 'react'
import { Card, Col, Row, Typography, List, Tag, Space, Tooltip, Modal, Checkbox, Button, message } from 'antd'
import {
  CloudServerOutlined,
  CloudOutlined,
  DatabaseOutlined,
  RadarChartOutlined,
  CodeOutlined,
  CodeSandboxOutlined,
  DesktopOutlined,
  ClockCircleOutlined,
  ArrowUpOutlined,
  CheckCircleOutlined,
  PauseCircleOutlined,
  ApartmentOutlined,
  PieChartOutlined,
  SettingOutlined,
  KeyOutlined,
  SafetyCertificateOutlined,
  UserOutlined,
  LinuxOutlined,
  FileOutlined,
  FileSearchOutlined,
  HistoryOutlined,
  GatewayOutlined,
  NodeIndexOutlined,
  VideoCameraOutlined,
  BookOutlined,
  AuditOutlined,
  SolutionOutlined,
  TagsOutlined,
  ToolOutlined,
} from '@ant-design/icons'
import { useNavigate } from 'react-router-dom'
import {
  getDashboardStats,
  getRecentActivities,
  getQuickActions,
  updateQuickActions,
  type DashboardStats,
  type RecentActivity,
  type DistributionItem,
} from '../../services/dashboard'

interface StatCardProps {
  title: string
  value: number
  icon: React.ReactNode
  color: string
  bgColor: string
  delay?: number
}

const StatCard = ({ title, value, icon, color, bgColor, delay = 0 }: StatCardProps) => {
  const [displayed, setDisplayed] = useState(0)

  useEffect(() => {
    if (value === 0) return
    const duration = 600
    const steps = 30
    const increment = value / steps
    let current = 0
    const timer = setTimeout(() => {
      const interval = setInterval(() => {
        current += increment
        if (current >= value) {
          setDisplayed(value)
          clearInterval(interval)
        } else {
          setDisplayed(Math.floor(current))
        }
      }, duration / steps)
    }, delay)
    return () => clearTimeout(timer)
  }, [value, delay])

  return (
    <Card
      hoverable
      style={{ border: '1px solid #E7E9EF' }}
      styles={{ body: { padding: '20px 24px' } }}
    >
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <div>
          <div style={{ color: '#979BA5', fontSize: 13, marginBottom: 8 }}>{title}</div>
          <div className="stat-value" style={{ fontSize: 28, fontWeight: 600, color: '#313238' }}>
            {displayed}
          </div>
        </div>
        <div
          style={{
            width: 56,
            height: 56,
            borderRadius: 8,
            background: bgColor,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            fontSize: 28,
            color,
            transition: 'transform 0.3s ease',
          }}
        >
          {icon}
        </div>
      </div>
    </Card>
  )
}

const barColors = ['#3A84FF', '#2DCB56', '#FF9C01', '#7B61FF', '#13C2C2', '#EB2F96', '#FA541C', '#FAAD14']

const DistributionBar = ({ data, title, icon }: { data: DistributionItem[]; title: string; icon: React.ReactNode }) => {
  const total = data.reduce((sum, item) => sum + item.value, 0)

  return (
    <Card
      title={<Space>{icon}<span>{title}</span></Space>}
      style={{ border: '1px solid #E7E9EF', height: '100%' }}
    >
      {total === 0 ? (
        <div style={{ textAlign: 'center', color: '#979BA5', padding: '20px 0' }}>暂无数据</div>
      ) : (
        <>
          {/* Stacked bar */}
          <div style={{ display: 'flex', height: 20, borderRadius: 2, overflow: 'hidden', marginBottom: 16 }}>
            {data.map((item, i) => {
              const pct = (item.value / total) * 100
              if (pct === 0) return null
              return (
                <Tooltip key={item.label} title={`${item.label}: ${item.value} (${pct.toFixed(1)}%)`}>
                  <div
                    style={{
                      width: `${pct}%`,
                      background: barColors[i % barColors.length],
                      transition: 'width 0.6s cubic-bezier(0.23,1,0.23,1)',
                      minWidth: pct > 0 ? 4 : 0,
                    }}
                  />
                </Tooltip>
              )
            })}
          </div>
          {/* Legend rows */}
          <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
            {data.map((item, i) => {
              const pct = total > 0 ? ((item.value / total) * 100).toFixed(1) : '0'
              return (
                <div key={item.label} style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                  <Space size={8}>
                    <div style={{ width: 10, height: 10, borderRadius: 2, background: barColors[i % barColors.length], flexShrink: 0 }} />
                    <span style={{ color: '#63656E', fontSize: 13 }}>{item.label}</span>
                  </Space>
                  <Space size={12}>
                    <span style={{ color: '#313238', fontWeight: 500, fontSize: 13 }}>{item.value}</span>
                    <span style={{ color: '#979BA5', fontSize: 12, width: 48, textAlign: 'right' }}>{pct}%</span>
                  </Space>
                </div>
              )
            })}
          </div>
        </>
      )}
    </Card>
  )
}

// 所有可选快捷操作注册表
interface QuickActionDef {
  key: string
  title: string
  desc: string
  icon: React.ReactNode
  color: string
  path: string
  category: string
}

const allQuickActions: QuickActionDef[] = [
  // 资产管理
  { key: 'assets_cloud', title: '云资源', desc: '查看云主机资源', icon: <CloudOutlined />, color: '#3A84FF', path: '/assets/cloud', category: '资产管理' },
  { key: 'assets_idc', title: 'IDC 资产', desc: '管理 IDC 物理资产', icon: <DatabaseOutlined />, color: '#2DCB56', path: '/assets/idc', category: '资产管理' },
  { key: 'assets_tags', title: '资产标签', desc: '管理资产标签体系', icon: <TagsOutlined />, color: '#13C2C2', path: '/assets/tags', category: '资产管理' },
  { key: 'assets_init', title: '资产初始化', desc: '批量初始化新资产', icon: <SettingOutlined />, color: '#7B61FF', path: '/assets/init', category: '资产管理' },
  { key: 'cloud_accounts', title: '云账号托管', desc: '管理云账号 AccessKey', icon: <SafetyCertificateOutlined />, color: '#FF9C01', path: '/cloud-accounts', category: '资产管理' },
  { key: 'probe', title: '资产探针', desc: '采集服务器信息', icon: <RadarChartOutlined />, color: '#3A84FF', path: '/probe', category: '资产管理' },
  { key: 'ipam_subnets', title: '网段管理', desc: '管理 CIDR 网段', icon: <GatewayOutlined />, color: '#2F54EB', path: '/ipam/subnets', category: '资产管理' },
  { key: 'ipam_addresses', title: 'IP 分配', desc: '分配与回收 IP', icon: <NodeIndexOutlined />, color: '#FA541C', path: '/ipam/addresses', category: '资产管理' },
  // 运维操作
  { key: 'task_execute', title: '命令执行', desc: '批量下发 Shell 命令', icon: <CodeSandboxOutlined />, color: '#2DCB56', path: '/tasks/execute', category: '运维操作' },
  { key: 'task_history', title: '执行历史', desc: '查看命令执行记录', icon: <HistoryOutlined />, color: '#979BA5', path: '/tasks/history', category: '运维操作' },
  { key: 'file_distribute', title: '文件下发', desc: '批量分发文件', icon: <FileOutlined />, color: '#7B61FF', path: '/files/distribute', category: '运维操作' },
  { key: 'file_history', title: '分发记录', desc: '查看文件分发历史', icon: <FileSearchOutlined />, color: '#979BA5', path: '/files/history', category: '运维操作' },
  { key: 'terminal', title: 'Web Terminal', desc: '浏览器内 SSH 终端', icon: <DesktopOutlined />, color: '#EB2F96', path: '/terminal', category: '运维操作' },
  { key: 'terminal_sessions', title: '终端会话', desc: '管理终端会话记录', icon: <VideoCameraOutlined />, color: '#EB2F96', path: '/terminal/sessions', category: '运维操作' },
  { key: 'sshkeys', title: 'SSH Key 管理', desc: '密钥下发与轮换', icon: <KeyOutlined />, color: '#87D068', path: '/sshkeys', category: '运维操作' },
  { key: 'playbooks', title: '运维剧本', desc: '编排自动化剧本', icon: <BookOutlined />, color: '#FAAD14', path: '/playbooks', category: '运维操作' },
  // 管理
  { key: 'projects', title: '项目管理', desc: '管理业务项目', icon: <ApartmentOutlined />, color: '#7B61FF', path: '/projects', category: '管理' },
  { key: 'approvals', title: '审批管理', desc: '处理审批工单', icon: <SolutionOutlined />, color: '#3A84FF', path: '/approvals', category: '管理' },
  { key: 'audit_operation', title: '操作审计', desc: '查看操作审计日志', icon: <AuditOutlined />, color: '#FAAD14', path: '/audit/operation', category: '管理' },
  { key: 'audit_command', title: '命令审计', desc: '查看命令审计日志', icon: <CodeOutlined />, color: '#FAAD14', path: '/audit/command', category: '管理' },
  { key: 'users_platform', title: '平台用户', desc: '管理平台账号', icon: <UserOutlined />, color: '#2F54EB', path: '/users/platform', category: '管理' },
  { key: 'users_linux', title: 'Linux 用户', desc: '管理 Linux 账号', icon: <LinuxOutlined />, color: '#2F54EB', path: '/users/linux', category: '管理' },
  { key: 'settings_general', title: '基础配置', desc: '系统基础参数设置', icon: <ToolOutlined />, color: '#979BA5', path: '/settings/general', category: '管理' },
  { key: 'settings_security', title: '安全设置', desc: '安全策略与限流', icon: <SafetyCertificateOutlined />, color: '#EA3636', path: '/settings/security', category: '管理' },
]

const quickActionMap = new Map(allQuickActions.map(a => [a.key, a]))

const moduleColors: Record<string, string> = {
  auth: '#3A84FF', asset: '#2DCB56', cloud: '#13C2C2', ssh: '#87D068',
  task: '#FF9C01', file: '#7B61FF', terminal: '#EB2F96', user: '#2F54EB',
  settings: '#FAAD14', ipam: '#FA541C',
}

const DashboardPage = () => {
  const navigate = useNavigate()
  const [stats, setStats] = useState<DashboardStats | null>(null)
  const [activities, setActivities] = useState<RecentActivity[]>([])
  const [actionKeys, setActionKeys] = useState<string[]>([])
  const [editOpen, setEditOpen] = useState(false)
  const [editKeys, setEditKeys] = useState<string[]>([])
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    getDashboardStats().then(setStats).catch(() => {})
    getRecentActivities().then(setActivities).catch(() => {})
    getQuickActions().then(r => setActionKeys(r.actions || [])).catch(() => {})
  }, [])

  const visibleActions = useMemo(
    () => actionKeys.map(k => quickActionMap.get(k)).filter((a): a is QuickActionDef => !!a),
    [actionKeys],
  )

  const openEdit = () => {
    setEditKeys([...actionKeys])
    setEditOpen(true)
  }

  const handleToggle = (key: string, checked: boolean) => {
    setEditKeys(prev => {
      if (checked) {
        if (prev.length >= 12) {
          message.warning('最多选择 12 个快捷操作')
          return prev
        }
        return [...prev, key]
      }
      return prev.filter(k => k !== key)
    })
  }

  const handleSave = async () => {
    setSaving(true)
    try {
      await updateQuickActions(editKeys)
      setActionKeys(editKeys)
      setEditOpen(false)
      message.success('快捷操作已更新')
    } catch {
      // error handled by request interceptor
    } finally {
      setSaving(false)
    }
  }

  // 按分类分组
  const categories = useMemo(() => {
    const map = new Map<string, QuickActionDef[]>()
    for (const action of allQuickActions) {
      const list = map.get(action.category) || []
      list.push(action)
      map.set(action.category, list)
    }
    return map
  }, [])

  return (
    <div>
      {/* Stat cards */}
      <Row gutter={[16, 16]}>
        <Col xs={24} sm={12} xl={6}>
          <StatCard title="总资产" value={stats?.total_assets || 0} icon={<CloudServerOutlined />} color="#3A84FF" bgColor="#E1ECFF" delay={0} />
        </Col>
        <Col xs={24} sm={12} xl={6}>
          <StatCard title="云主机" value={stats?.cloud_assets || 0} icon={<CloudServerOutlined />} color="#13C2C2" bgColor="#E6FFFB" delay={80} />
        </Col>
        <Col xs={24} sm={12} xl={6}>
          <StatCard title="IDC 资产" value={stats?.idc_assets || 0} icon={<DatabaseOutlined />} color="#2DCB56" bgColor="#DCFFE2" delay={160} />
        </Col>
        <Col xs={24} sm={12} xl={6}>
          <StatCard title="项目数" value={stats?.total_projects || 0} icon={<ApartmentOutlined />} color="#7B61FF" bgColor="#F0EBFF" delay={240} />
        </Col>
      </Row>

      <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
        <Col xs={24} sm={12} xl={6}>
          <StatCard title="在线" value={stats?.online_assets || 0} icon={<CheckCircleOutlined />} color="#2DCB56" bgColor="#DCFFE2" delay={320} />
        </Col>
        <Col xs={24} sm={12} xl={6}>
          <StatCard title="离线" value={stats?.offline_assets || 0} icon={<PauseCircleOutlined />} color="#EA3636" bgColor="#FFDDDD" delay={400} />
        </Col>
        <Col xs={24} sm={12} xl={6}>
          <StatCard title="已探测" value={stats?.probed_assets || 0} icon={<RadarChartOutlined />} color="#2F54EB" bgColor="#E1ECFF" delay={480} />
        </Col>
        <Col span={6} />
      </Row>

      {/* Distribution charts */}
      <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
        <Col xs={24} lg={12}>
          <DistributionBar
            data={stats?.env_distribution || []}
            title="环境分布"
            icon={<PieChartOutlined style={{ color: '#3A84FF' }} />}
          />
        </Col>
        <Col xs={24} lg={12}>
          <DistributionBar
            data={stats?.type_distribution || []}
            title="资产类型分布"
            icon={<PieChartOutlined style={{ color: '#3A84FF' }} />}
          />
        </Col>
      </Row>

      {/* Quick actions + recent activity */}
      <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
        <Col xs={24} lg={10}>
          <Card
            title={<Space><ArrowUpOutlined style={{ color: '#3A84FF' }} /><span>快捷操作</span></Space>}
            extra={
              <Button type="link" size="small" icon={<SettingOutlined />} onClick={openEdit} style={{ color: '#979BA5' }}>
                编辑
              </Button>
            }
            style={{ border: '1px solid #E7E9EF', height: '100%' }}
          >
            {visibleActions.length === 0 ? (
              <div style={{ textAlign: 'center', color: '#979BA5', padding: '20px 0' }}>
                暂未配置快捷操作，点击右上角编辑添加
              </div>
            ) : (
              <Row gutter={[12, 12]}>
                {visibleActions.map((item, i) => (
                  <Col span={12} key={item.key}>
                    <Card
                      hoverable
                      size="small"
                      style={{
                        border: '1px solid #E7E9EF',
                        textAlign: 'center',
                        cursor: 'pointer',
                        animationDelay: `${i * 0.08}s`,
                      }}
                      styles={{ body: { padding: '20px 12px' } }}
                      onClick={() => navigate(item.path)}
                    >
                      <div style={{ fontSize: 26, color: item.color, marginBottom: 8, transition: 'transform 0.3s ease' }}>
                        {item.icon}
                      </div>
                      <div style={{ fontWeight: 500, color: '#313238', marginBottom: 2 }}>{item.title}</div>
                      <Typography.Text type="secondary" style={{ fontSize: 12 }}>
                        {item.desc}
                      </Typography.Text>
                    </Card>
                  </Col>
                ))}
              </Row>
            )}
          </Card>
        </Col>

        <Col xs={24} lg={14}>
          <Card
            title={<Space><ClockCircleOutlined style={{ color: '#3A84FF' }} /><span>最近活动</span></Space>}
            style={{ border: '1px solid #E7E9EF', height: '100%' }}
          >
            <List
              dataSource={activities}
              locale={{ emptyText: '暂无活动' }}
              renderItem={(item) => (
                <List.Item style={{ padding: '10px 0', borderBottom: '1px solid #F0F1F5' }}>
                  <List.Item.Meta
                    title={
                      <Space size={8}>
                        <Tag
                          color={moduleColors[item.module] || '#979BA5'}
                          style={{ fontSize: 12, lineHeight: '20px', padding: '0 8px', margin: 0 }}
                        >
                          {item.module}
                        </Tag>
                        <span style={{ color: '#313238', fontSize: 13 }}>{item.action} {item.path}</span>
                      </Space>
                    }
                    description={
                      <span style={{ fontSize: 12, color: '#979BA5' }}>
                        {item.username || '-'} &middot; {item.client_ip} &middot; {item.created_at}
                      </span>
                    }
                  />
                </List.Item>
              )}
            />
          </Card>
        </Col>
      </Row>

      {/* 编辑快捷操作弹窗 */}
      <Modal
        title="编辑快捷操作"
        open={editOpen}
        onCancel={() => setEditOpen(false)}
        onOk={handleSave}
        confirmLoading={saving}
        okText="保存"
        cancelText="取消"
        width={640}
      >
        <div style={{ color: '#979BA5', fontSize: 12, marginBottom: 16 }}>
          勾选需要在仪表盘显示的快捷操作（最多 12 个，已选 {editKeys.length} 个）
        </div>
        {Array.from(categories.entries()).map(([category, actions]) => (
          <div key={category} style={{ marginBottom: 20 }}>
            <div style={{ fontWeight: 500, color: '#313238', marginBottom: 10, fontSize: 13 }}>{category}</div>
            <Row gutter={[0, 8]}>
              {actions.map(action => (
                <Col span={12} key={action.key}>
                  <Checkbox
                    checked={editKeys.includes(action.key)}
                    onChange={e => handleToggle(action.key, e.target.checked)}
                  >
                    <Space size={6}>
                      <span style={{ color: action.color, fontSize: 14 }}>{action.icon}</span>
                      <span style={{ color: '#313238', fontSize: 13 }}>{action.title}</span>
                      <span style={{ color: '#979BA5', fontSize: 12 }}>{action.desc}</span>
                    </Space>
                  </Checkbox>
                </Col>
              ))}
            </Row>
          </div>
        ))}
      </Modal>
    </div>
  )
}

export default DashboardPage
