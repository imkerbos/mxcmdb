import { useEffect, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { Card, Row, Col, Table, Tag, Space, Button, Tooltip, Typography, Progress } from 'antd'
import {
  ArrowLeftOutlined,
  CloudServerOutlined,
  ClusterOutlined,
  DatabaseOutlined,
  HddOutlined,
} from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import {
  getProjectSummary,
  getProjectAssets,
  type ProjectSummary,
  type DistributionItem,
} from '../../services/project'
import type { AssetItem } from '../../services/asset'

const statusColors: Record<string, string> = { online: 'green', offline: 'red', running: 'green', stopped: 'red', unknown: 'default', terminated: 'default' }
const statusLabels: Record<string, string> = { online: '在线', offline: '离线', running: '运行中', stopped: '已停止', unknown: '未知', terminated: '已终止' }
const envColors: Record<string, string> = { prod: 'red', test: 'orange', dev: 'blue' }

const barColors = ['#3A84FF', '#2DCB56', '#FF9C01', '#7B61FF', '#13C2C2', '#EB2F96', '#FA541C', '#FAAD14']

// 通用标签翻译
const labelMap: Record<string, string> = {
  online: '在线', offline: '离线', running: '运行中', stopped: '已停止', unknown: '未知',
  ecs: 'ECS', idc: 'IDC', physical_server: '物理服务器', vm: '虚拟机', network_device: '网络设备',
  prod: '生产', test: '测试', dev: '开发',
}
const t = (key: string) => labelMap[key] || key

const MiniBar = ({ data, title, icon }: { data: DistributionItem[]; title: string; icon: React.ReactNode }) => {
  const total = data.reduce((s, i) => s + i.value, 0)
  return (
    <div>
      <div style={{ display: 'flex', alignItems: 'center', gap: 6, marginBottom: 10 }}>
        {icon}
        <span style={{ fontWeight: 500, color: '#313238', fontSize: 13 }}>{title}</span>
      </div>
      {total === 0 ? (
        <div style={{ color: '#C4C6CC', fontSize: 12 }}>暂无数据</div>
      ) : (
        <>
          <div style={{ display: 'flex', height: 6, borderRadius: 2, overflow: 'hidden', marginBottom: 8 }}>
            {data.map((item, i) => {
              const pct = (item.value / total) * 100
              if (pct === 0) return null
              return (
                <Tooltip key={item.label} title={`${t(item.label)}: ${item.value}`}>
                  <div style={{ width: `${pct}%`, background: barColors[i % barColors.length], minWidth: pct > 0 ? 3 : 0 }} />
                </Tooltip>
              )
            })}
          </div>
          <div style={{ display: 'flex', flexWrap: 'wrap', gap: '4px 14px' }}>
            {data.map((item, i) => (
              <Space key={item.label} size={4}>
                <div style={{ width: 6, height: 6, borderRadius: 1, background: barColors[i % barColors.length] }} />
                <span style={{ fontSize: 12, color: '#63656E' }}>{t(item.label)}</span>
                <span style={{ fontSize: 12, color: '#313238', fontWeight: 500 }}>{item.value}</span>
              </Space>
            ))}
          </div>
        </>
      )}
    </div>
  )
}

const formatMemory = (mb: number) => {
  if (mb >= 1024) return `${(mb / 1024).toFixed(1)} GB`
  return `${mb} MB`
}

const formatDisk = (gb: number) => {
  if (gb >= 1024) return `${(gb / 1024).toFixed(1)} TB`
  return `${gb} GB`
}

const formatTime = (v: string) => {
  if (!v) return '-'
  return v.replace('T', ' ').replace(/\.\d+Z$/, '').substring(0, 19)
}

const ProjectDetailPage = () => {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [summary, setSummary] = useState<ProjectSummary | null>(null)
  const [assets, setAssets] = useState<AssetItem[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(20)
  const [loading, setLoading] = useState(false)

  const projectId = Number(id)

  const fetchSummary = async () => {
    try {
      const data = await getProjectSummary(projectId)
      setSummary(data)
    } catch { /* handled */ }
  }

  const fetchAssets = async (p = page, ps = pageSize) => {
    setLoading(true)
    try {
      const result = await getProjectAssets(projectId, { page: p, page_size: ps })
      setAssets(result.list || [])
      setTotal(result.total)
    } catch { /* handled */ } finally { setLoading(false) }
  }

  useEffect(() => {
    if (!projectId) return
    fetchSummary()
    fetchAssets()
  }, [projectId])

  const s = summary?.asset_stats
  const probePct = (s?.total_assets || 0) > 0 ? Math.round(((s?.probed || 0) / (s?.total_assets || 1)) * 100) : 0

  const columns: ColumnsType<AssetItem> = [
    { title: '主机名', dataIndex: 'hostname', width: 160, ellipsis: true },
    { title: 'IP', dataIndex: 'ip', width: 140 },
    { title: '状态', dataIndex: 'status', width: 80, render: (v) => <Tag color={statusColors[v] || 'default'}>{statusLabels[v] || v}</Tag> },
    { title: '环境', dataIndex: 'environment', width: 80, render: (v) => v ? <Tag color={envColors[v] || 'default'}>{v}</Tag> : '-' },
    { title: '操作系统', dataIndex: 'os', ellipsis: true },
    { title: '部门', dataIndex: 'department', width: 100, ellipsis: true },
    { title: '负责人', dataIndex: 'owner', width: 90 },
    { title: '最后探测', dataIndex: 'probe_last_at', width: 160, render: (v) => formatTime(v) },
  ]

  return (
    <div>
      {/* Header */}
      <Card style={{ border: '1px solid #E7E9EF', marginBottom: 16 }} styles={{ body: { padding: '14px 24px' } }}>
        <div style={{ display: 'flex', alignItems: 'center' }}>
          <Button type="text" icon={<ArrowLeftOutlined />} onClick={() => navigate('/projects')} style={{ marginRight: 12 }} />
          <div style={{ flex: 1 }}>
            <Space size={8} align="center">
              <Typography.Title level={4} style={{ margin: 0 }}>{summary?.project.name || '-'}</Typography.Title>
              <Tag style={{ margin: 0 }}>{summary?.project.code || '-'}</Tag>
              <Tag color={summary?.project.status === 'active' ? 'green' : 'default'} style={{ margin: 0 }}>
                {summary?.project.status === 'active' ? '活跃' : '已归档'}
              </Tag>
            </Space>
            <div style={{ color: '#979BA5', fontSize: 13, marginTop: 2 }}>
              {summary?.project.description || '暂无描述'}
              {summary?.project.owner && <span style={{ marginLeft: 16 }}>负责人：{summary.project.owner}</span>}
            </div>
          </div>
        </div>
      </Card>

      {/* 资源指纹 + 分布 */}
      <Row gutter={[16, 16]}>
        {/* 左侧：4 个指标 2x2 */}
        <Col xs={24} lg={12}>
          <Card style={{ border: '1px solid #E7E9EF', height: '100%' }} styles={{ body: { padding: 20 } }}>
            <Row gutter={[16, 20]}>
              <Col span={12}>
                <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
                  <div style={{ width: 40, height: 40, borderRadius: 8, background: '#E1ECFF', display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: 18, color: '#3A84FF', flexShrink: 0 }}>
                    <CloudServerOutlined />
                  </div>
                  <div>
                    <div style={{ fontSize: 12, color: '#979BA5' }}>服务器</div>
                    <div style={{ fontSize: 22, fontWeight: 600, color: '#313238', lineHeight: 1.2 }}>{s?.total_assets || 0}</div>
                    <div style={{ fontSize: 12, color: '#979BA5' }}>
                      <span style={{ color: '#2DCB56' }}>{s?.running || 0} 在线</span>
                      <span style={{ margin: '0 4px', color: '#DCDEE5' }}>|</span>
                      <span style={{ color: '#EA3636' }}>{s?.stopped || 0} 离线</span>
                    </div>
                  </div>
                </div>
              </Col>
              <Col span={12}>
                <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
                  <div style={{ width: 40, height: 40, borderRadius: 8, background: '#DCFFE2', display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: 18, color: '#2DCB56', flexShrink: 0 }}>
                    <ClusterOutlined />
                  </div>
                  <div>
                    <div style={{ fontSize: 12, color: '#979BA5' }}>CPU 核心</div>
                    <div style={{ fontSize: 22, fontWeight: 600, color: '#313238', lineHeight: 1.2 }}>{s?.total_cpu || 0}</div>
                    <div style={{ fontSize: 12, color: '#979BA5' }}>
                      {(s?.total_assets || 0) > 0 ? `均 ${((s?.total_cpu || 0) / (s?.total_assets || 1)).toFixed(1)} 核/台` : '-'}
                    </div>
                  </div>
                </div>
              </Col>
              <Col span={12}>
                <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
                  <div style={{ width: 40, height: 40, borderRadius: 8, background: '#F0EBFF', display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: 18, color: '#7B61FF', flexShrink: 0 }}>
                    <DatabaseOutlined />
                  </div>
                  <div>
                    <div style={{ fontSize: 12, color: '#979BA5' }}>内存</div>
                    <div style={{ fontSize: 22, fontWeight: 600, color: '#313238', lineHeight: 1.2 }}>{formatMemory(s?.total_memory || 0)}</div>
                    <div style={{ fontSize: 12, color: '#979BA5' }}>
                      {(s?.total_assets || 0) > 0 ? `均 ${formatMemory(Math.round((s?.total_memory || 0) / (s?.total_assets || 1)))}/台` : '-'}
                    </div>
                  </div>
                </div>
              </Col>
              <Col span={12}>
                <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
                  <div style={{ width: 40, height: 40, borderRadius: 8, background: '#E6FFFB', display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: 18, color: '#13C2C2', flexShrink: 0 }}>
                    <HddOutlined />
                  </div>
                  <div>
                    <div style={{ fontSize: 12, color: '#979BA5' }}>磁盘</div>
                    <div style={{ fontSize: 22, fontWeight: 600, color: '#313238', lineHeight: 1.2 }}>{formatDisk(s?.total_disk || 0)}</div>
                    <div style={{ fontSize: 12, color: '#979BA5' }}>
                      {(s?.total_assets || 0) > 0 ? `均 ${formatDisk(Math.round((s?.total_disk || 0) / (s?.total_assets || 1)))}/台` : '-'}
                    </div>
                  </div>
                </div>
              </Col>
            </Row>
          </Card>
        </Col>

        {/* 右侧：分布 + 探测覆盖 */}
        <Col xs={24} lg={12}>
          <Card style={{ border: '1px solid #E7E9EF', height: '100%' }} styles={{ body: { padding: 20 } }}>
            <Row gutter={[0, 20]}>
              <Col span={12}>
                <MiniBar data={summary?.status_distribution || []} title="状态分布" icon={<div style={{ width: 6, height: 6, borderRadius: 1, background: '#3A84FF' }} />} />
              </Col>
              <Col span={12}>
                <MiniBar data={summary?.type_distribution || []} title="类型分布" icon={<div style={{ width: 6, height: 6, borderRadius: 1, background: '#2DCB56' }} />} />
              </Col>
              <Col span={24}>
                <div style={{ borderTop: '1px solid #F0F1F5', paddingTop: 16 }}>
                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 8 }}>
                    <span style={{ fontSize: 13, fontWeight: 500, color: '#313238' }}>探测覆盖</span>
                    <span style={{ fontSize: 12, color: '#979BA5' }}>{s?.probed || 0} / {s?.total_assets || 0}</span>
                  </div>
                  <Progress percent={probePct} strokeColor="#3A84FF" size="small" format={pct => `${pct}%`} />
                </div>
              </Col>
            </Row>
          </Card>
        </Col>
      </Row>

      {/* Asset list */}
      <Card title="项目资产" style={{ border: '1px solid #E7E9EF', marginTop: 16 }}>
        <Table
          rowKey="id"
          columns={columns}
          dataSource={assets}
          loading={loading}
          pagination={{
            current: page,
            pageSize,
            total,
            showSizeChanger: true,
            showTotal: (t) => `共 ${t} 条`,
            onChange: (p, ps) => { setPage(p); setPageSize(ps); fetchAssets(p, ps) },
          }}
        />
      </Card>
    </div>
  )
}

export default ProjectDetailPage
