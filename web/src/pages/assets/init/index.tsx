import { useState, useEffect, useRef } from 'react'
import { Card, Steps, Table, Button, Space, Input, Select, Tag, Spin, Alert, message, Checkbox, Collapse, Badge } from 'antd'
import {
  SearchOutlined, CheckCircleOutlined, LoadingOutlined, RocketOutlined,
  BookOutlined,
} from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import { getAllAssets, type AssetSimple } from '../../../services/asset'
import { getAllProjects, type ProjectSimple } from '../../../services/project'
import {
  getPlaybooks, executePlaybook,
  type Playbook, type PlaybookAssetResult, type PlaybookStep, type StepExecResult,
} from '../../../services/playbook'
import { connectPlaybookWS, type ManagedWS, type WSState } from '../../../utils/websocket'

interface ExecuteResponse {
  total: number
  success: number
  failed: number
  results: PlaybookAssetResult[]
}

const statusColors: Record<string, string> = { online: 'green', offline: 'red', running: 'green', stopped: 'red', unknown: 'default', terminated: 'default' }
const statusLabels: Record<string, string> = { online: '在线', offline: '离线', running: '运行中', stopped: '已停止', unknown: '未知', terminated: '已终止' }

const AssetInitPage = () => {
  const [current, setCurrent] = useState(0)

  // Step 1: 资产选择
  const [allAssets, setAllAssets] = useState<AssetSimple[]>([])
  const [filteredAssets, setFilteredAssets] = useState<AssetSimple[]>([])
  const [assetsLoading, setAssetsLoading] = useState(false)
  const [selectedRowKeys, setSelectedRowKeys] = useState<React.Key[]>([])
  const [keyword, setKeyword] = useState('')
  const [projectFilter, setProjectFilter] = useState<number | ''>('')
  const [envFilter, setEnvFilter] = useState('')
  const [typeFilter, setTypeFilter] = useState('')
  const [projects, setProjects] = useState<ProjectSimple[]>([])

  // Step 2: 剧本选择
  const [playbooks, setPlaybooks] = useState<Playbook[]>([])
  const [playbooksLoading, setPlaybooksLoading] = useState(false)
  const [selectedPlaybookIds, setSelectedPlaybookIds] = useState<number[]>([])

  // Step 3: 执行结果
  const [executing, setExecuting] = useState(false)
  const [results, setResults] = useState<PlaybookAssetResult[]>([])
  const [summary, setSummary] = useState<ExecuteResponse | null>(null)
  const [wsState, setWsState] = useState<WSState | null>(null)
  const wsRef = useRef<ManagedWS | null>(null)

  useEffect(() => {
    setAssetsLoading(true)
    Promise.all([getAllAssets(), getAllProjects()]).then(([data, projs]) => {
      setAllAssets(data || [])
      setFilteredAssets(data || [])
      setProjects(projs || [])
    }).catch(() => {}).finally(() => setAssetsLoading(false))

    setPlaybooksLoading(true)
    getPlaybooks().then(data => {
      setPlaybooks((data || []).filter(p => p.enabled))
    }).catch(() => {}).finally(() => setPlaybooksLoading(false))

    return () => {
      if (wsRef.current) {
        wsRef.current.close()
        wsRef.current = null
      }
    }
  }, [])

  // 资产过滤
  useEffect(() => {
    let filtered = allAssets
    if (keyword) {
      const kw = keyword.toLowerCase()
      filtered = filtered.filter(a => a.hostname.toLowerCase().includes(kw) || a.ip.toLowerCase().includes(kw))
    }
    if (projectFilter !== '') filtered = filtered.filter(a => a.project_id === projectFilter)
    if (envFilter) filtered = filtered.filter(a => a.environment === envFilter)
    if (typeFilter) filtered = filtered.filter(a => a.type === typeFilter)
    setFilteredAssets(filtered)
  }, [keyword, projectFilter, envFilter, typeFilter, allAssets])

  const handleNext = () => {
    if (current === 0 && selectedRowKeys.length === 0) {
      message.warning('请先选择要初始化的资产')
      return
    }
    if (current === 1) {
      if (selectedPlaybookIds.length === 0) {
        message.warning('请至少选择一个剧本')
        return
      }
      handleExecute()
      return
    }
    setCurrent(current + 1)
  }

  const handlePrev = () => {
    if (current > 0 && !executing) setCurrent(current - 1)
  }

  const handleExecute = async () => {
    setCurrent(2)
    setExecuting(true)
    setResults([])
    setSummary(null)
    try {
      const assetIds = selectedRowKeys.map(Number)
      const job = await executePlaybook({ playbook_ids: selectedPlaybookIds, asset_ids: assetIds })

      const ws = connectPlaybookWS(job.job_id, (msg) => {
        if (msg.type === 'playbook_progress') {
          setResults(prev => [...prev, msg.data as PlaybookAssetResult])
        } else if (msg.type === 'playbook_done') {
          const data = msg.data as ExecuteResponse
          setSummary(data)
          setExecuting(false)
          if (wsRef.current) {
            wsRef.current.close()
            wsRef.current = null
          }
        }
      }, setWsState)
      if (!ws) {
        message.error('未登录或连接失败')
        setExecuting(false)
        return
      }
      wsRef.current = ws
    } catch {
      setExecuting(false)
      message.error('执行启动失败')
    }
  }

  const handleReset = () => {
    if (wsRef.current) {
      wsRef.current.close()
      wsRef.current = null
    }
    setCurrent(0)
    setSelectedRowKeys([])
    setSelectedPlaybookIds([])
    setResults([])
    setSummary(null)
    setExecuting(false)
    setWsState(null)
  }

  const togglePlaybook = (id: number) => {
    setSelectedPlaybookIds(prev =>
      prev.includes(id) ? prev.filter(x => x !== id) : [...prev, id]
    )
  }

  // 选择资产表格列
  const assetColumns: ColumnsType<AssetSimple> = [
    { title: '主机名', dataIndex: 'hostname', width: 160, ellipsis: true },
    { title: 'IP', dataIndex: 'ip', width: 140 },
    { title: '类型', dataIndex: 'type', width: 120, render: (v) => <Tag>{v}</Tag> },
    { title: '状态', dataIndex: 'status', width: 80, render: (v) => <Tag color={statusColors[v] || 'default'}>{statusLabels[v] || v}</Tag> },
    { title: '环境', dataIndex: 'environment', width: 80, render: (v) => v || '-' },
  ]

  // 执行结果表格列
  const resultColumns: ColumnsType<PlaybookAssetResult> = [
    {
      title: '主机', width: 180,
      render: (_, r) => (
        <div>
          <div style={{ fontWeight: 500 }}>{r.hostname || r.ip}</div>
          {r.hostname && <div style={{ fontSize: 12, color: '#979BA5' }}>{r.ip}</div>}
        </div>
      ),
    },
    {
      title: '状态', dataIndex: 'status', width: 80,
      render: (v) => {
        const map: Record<string, { color: string; text: string }> = {
          success: { color: 'success', text: '成功' },
          partial: { color: 'warning', text: '部分' },
          failed: { color: 'error', text: '失败' },
        }
        const m = map[v] || { color: 'default', text: v }
        return <Tag color={m.color}>{m.text}</Tag>
      },
    },
    {
      title: '步骤概况', width: 200,
      render: (_, r) => {
        if (!r.step_results) return '-'
        const s = r.step_results.filter(x => x.status === 'success').length
        const f = r.step_results.filter(x => x.status === 'failed').length
        const sk = r.step_results.filter(x => x.status === 'skipped').length
        return (
          <Space size={4}>
            {s > 0 && <Tag color="green">{s} 成功</Tag>}
            {f > 0 && <Tag color="red">{f} 失败</Tag>}
            {sk > 0 && <Tag color="default">{sk} 跳过</Tag>}
          </Space>
        )
      },
    },
    {
      title: '详情', ellipsis: true,
      render: (_, r) => {
        if (!r.step_results) return '-'
        const failedSteps = r.step_results.filter(x => x.status === 'failed')
        if (failedSteps.length === 0) return <span style={{ color: '#2DCB56', fontSize: 12 }}>全部成功</span>
        return (
          <span style={{ fontSize: 12, color: '#EA3636' }}>
            {failedSteps.map(s => s.name).join(', ')} 失败
          </span>
        )
      },
    },
  ]

  const selectedAssets = allAssets.filter(a => selectedRowKeys.includes(a.id))
  const selectedNames = playbooks.filter(p => selectedPlaybookIds.includes(p.id)).map(p => p.name)
  const totalSteps = playbooks
    .filter(p => selectedPlaybookIds.includes(p.id))
    .reduce((sum, p) => sum + (p.steps?.length || 0), 0)

  return (
    <Card title="资产初始化" style={{ border: '1px solid #E7E9EF' }}>
      <Steps
        current={current}
        style={{ marginBottom: 24 }}
        items={[
          { title: '选择资产', description: selectedRowKeys.length > 0 ? `已选 ${selectedRowKeys.length} 台` : '勾选要初始化的资产' },
          { title: '选择剧本', description: selectedPlaybookIds.length > 0 ? `已选 ${selectedPlaybookIds.length} 个剧本` : '选择要执行的剧本' },
          { title: '执行结果', description: executing ? '执行中...' : summary ? '已完成' : '等待执行' },
        ]}
      />

      {/* Step 0: 选择资产 */}
      {current === 0 && (
        <>
          <Space style={{ marginBottom: 16 }} wrap>
            <Input
              placeholder="搜索主机名/IP"
              prefix={<SearchOutlined />}
              value={keyword}
              onChange={e => setKeyword(e.target.value)}
              style={{ width: 200 }}
              allowClear
            />
            <Select
              value={projectFilter}
              onChange={setProjectFilter}
              style={{ width: 160 }}
              options={[
                { value: '' as const, label: '全部项目' },
                ...projects.map(p => ({ value: p.id, label: p.name })),
              ]}
            />
            <Select
              value={envFilter}
              onChange={setEnvFilter}
              style={{ width: 120 }}
              options={[
                { value: '', label: '全部环境' },
                { value: 'prod', label: '生产' },
                { value: 'test', label: '测试' },
                { value: 'dev', label: '开发' },
              ]}
            />
            <Select
              value={typeFilter}
              onChange={setTypeFilter}
              style={{ width: 140 }}
              options={[
                { value: '', label: '全部类型' },
                { value: 'ecs', label: 'ECS' },
                { value: 'physical_server', label: '物理服务器' },
                { value: 'vm', label: '虚拟机' },
              ]}
            />
            <span style={{ color: '#979BA5', fontSize: 13 }}>
              共 {filteredAssets.length} 台 | 已选 <span style={{ color: '#3A84FF', fontWeight: 500 }}>{selectedRowKeys.length}</span> 台
            </span>
          </Space>
          <Table
            rowKey="id"
            columns={assetColumns}
            dataSource={filteredAssets}
            loading={assetsLoading}
            size="middle"
            scroll={{ y: 420 }}
            rowSelection={{ selectedRowKeys, onChange: setSelectedRowKeys }}
            pagination={false}
          />
        </>
      )}

      {/* Step 1: 选择剧本 */}
      {current === 1 && (
        <div style={{ maxWidth: 720 }}>
          <Alert
            type="info"
            showIcon
            message={`将对 ${selectedRowKeys.length} 台资产执行选中的剧本，多个剧本的步骤将按顺序合并执行`}
            style={{ marginBottom: 20 }}
          />

          {/* 已选资产摘要 */}
          <div style={{ marginBottom: 20, padding: '12px 16px', background: '#F5F7FA', borderRadius: 2 }}>
            <div style={{ fontSize: 13, color: '#63656E', marginBottom: 8 }}>已选资产：</div>
            <div style={{ display: 'flex', flexWrap: 'wrap', gap: 6 }}>
              {selectedAssets.slice(0, 10).map(a => (
                <Tag key={a.id}>{a.hostname || a.ip}</Tag>
              ))}
              {selectedAssets.length > 10 && <Tag>...等共 {selectedAssets.length} 台</Tag>}
            </div>
          </div>

          {/* 剧本选择 */}
          <div style={{ marginBottom: 12, fontWeight: 500 }}>
            <BookOutlined style={{ marginRight: 8, color: '#3A84FF' }} />
            选择剧本（可多选，已选 <span style={{ color: '#3A84FF' }}>{selectedPlaybookIds.length}</span> 个，
            共 <span style={{ color: '#3A84FF' }}>{totalSteps}</span> 个步骤）
          </div>

          <Spin spinning={playbooksLoading}>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
              {playbooks.map(pb => {
                const checked = selectedPlaybookIds.includes(pb.id)
                return (
                  <div
                    key={pb.id}
                    style={{
                      border: `1px solid ${checked ? '#3A84FF' : '#E7E9EF'}`,
                      borderRadius: 2,
                      padding: '14px 16px',
                      cursor: 'pointer',
                      background: checked ? '#F0F5FF' : '#fff',
                      transition: 'all 0.2s ease',
                    }}
                    onClick={() => togglePlaybook(pb.id)}
                  >
                    <div style={{ display: 'flex', alignItems: 'center' }}>
                      <Checkbox checked={checked} style={{ marginRight: 12 }} />
                      <div style={{ flex: 1 }}>
                        <Space>
                          <span style={{ fontWeight: 500, color: '#313238' }}>{pb.name}</span>
                          {pb.is_builtin && <Tag color="blue">内置</Tag>}
                          <Tag>{pb.steps?.length || 0} 步骤</Tag>
                        </Space>
                        <div style={{ fontSize: 12, color: '#979BA5', marginTop: 4 }}>{pb.description}</div>
                      </div>
                    </div>

                    {/* 展开步骤预览 */}
                    {checked && pb.steps?.length > 0 && (
                      <Collapse
                        ghost
                        size="small"
                        style={{ marginTop: 8, marginLeft: 32 }}
                        items={[{
                          key: 'steps',
                          label: <span style={{ fontSize: 12, color: '#3A84FF' }}>查看步骤</span>,
                          children: (
                            <div>
                              {pb.steps.map((step: PlaybookStep, idx: number) => (
                                <div key={idx} style={{ display: 'flex', alignItems: 'center', gap: 8, padding: '4px 0', fontSize: 12, color: '#63656E' }}>
                                  <Badge count={idx + 1} style={{ backgroundColor: '#3A84FF' }} size="small" />
                                  <span>{step.name}</span>
                                  <Tag style={{ fontSize: 11 }}>{step.on_error === 'stop' ? '失败停止' : '失败继续'}</Tag>
                                </div>
                              ))}
                            </div>
                          ),
                        }]}
                      />
                    )}
                  </div>
                )
              })}
            </div>
          </Spin>
        </div>
      )}

      {/* Step 2: 执行结果 */}
      {current === 2 && (
        <>
          {wsState === 'reconnecting' && (
            <Alert type="warning" message="连接中断，正在尝试重连..." showIcon style={{ marginBottom: 16 }} />
          )}
          {wsState === 'disconnected' && executing && (
            <Alert type="error" message="连接已断开，无法接收实时结果" showIcon style={{ marginBottom: 16 }} />
          )}
          <div style={{ marginBottom: 16 }}>
            <Space size={16}>
              {executing ? (
                <>
                  <Spin indicator={<LoadingOutlined style={{ fontSize: 16, color: '#3A84FF' }} spin />} />
                  <span style={{ color: '#63656E' }}>正在执行 {selectedNames.join(' + ')}... {results.length}/{selectedRowKeys.length}</span>
                  <div style={{ width: 200, height: 6, borderRadius: 2, background: '#F0F1F5', overflow: 'hidden' }}>
                    <div style={{
                      height: '100%',
                      width: `${(results.length / selectedRowKeys.length) * 100}%`,
                      background: '#3A84FF',
                      borderRadius: 2,
                      transition: 'width 0.4s cubic-bezier(0.23,1,0.23,1)',
                    }} />
                  </div>
                </>
              ) : summary ? (
                <>
                  <CheckCircleOutlined style={{ fontSize: 16, color: '#2DCB56' }} />
                  <span style={{ color: '#313238', fontWeight: 500 }}>执行完成</span>
                  <span>共 {summary.total} 台</span>
                  <Tag color="success">{summary.success} 成功</Tag>
                  {summary.failed > 0 && <Tag color="error">{summary.failed} 失败</Tag>}
                </>
              ) : (
                <span style={{ color: '#979BA5' }}>等待结果...</span>
              )}
            </Space>
          </div>

          <Table
            rowKey="asset_id"
            columns={resultColumns}
            dataSource={results}
            size="middle"
            pagination={false}
            scroll={{ x: 800, y: 440 }}
            expandable={{
              expandedRowRender: (record: PlaybookAssetResult) => (
                <div>
                  {record.step_results?.map((sr: StepExecResult, i: number) => (
                    <div key={i} style={{ marginBottom: 8, padding: 8, background: '#F5F7FA', borderRadius: 2 }}>
                      <Space>
                        <Badge count={i + 1} style={{ backgroundColor: sr.status === 'success' ? '#2DCB56' : sr.status === 'failed' ? '#EA3636' : '#979BA5' }} />
                        <strong>{sr.name}</strong>
                        <Tag color={sr.status === 'success' ? 'green' : sr.status === 'failed' ? 'red' : 'default'}>{sr.status}</Tag>
                        <span style={{ color: '#979BA5', fontSize: 12 }}>{sr.latency}ms</span>
                      </Space>
                      {sr.output && (
                        <pre style={{ margin: '4px 0 0', fontSize: 12, maxHeight: 150, overflow: 'auto', background: '#F0F1F5', padding: 8, borderRadius: 2 }}>
                          {sr.output}
                        </pre>
                      )}
                      {sr.error && (
                        <pre style={{ margin: '4px 0 0', fontSize: 12, color: '#EA3636', maxHeight: 100, overflow: 'auto' }}>
                          {sr.error}
                        </pre>
                      )}
                    </div>
                  ))}
                </div>
              ),
            }}
          />
        </>
      )}

      {/* 底部按钮 */}
      <div style={{ marginTop: 24, display: 'flex', justifyContent: 'flex-end', borderTop: '1px solid #F0F1F5', paddingTop: 16 }}>
        <Space>
          {current === 2 && !executing && (
            <Button onClick={handleReset}>重新初始化</Button>
          )}
          {current > 0 && current < 2 && (
            <Button onClick={handlePrev}>上一步</Button>
          )}
          {current < 2 && (
            <Button type="primary" icon={current === 1 ? <RocketOutlined /> : undefined} onClick={handleNext}>
              {current === 0 ? '下一步：选择剧本' : `开始执行 (${selectedPlaybookIds.length} 个剧本 × ${selectedRowKeys.length} 台)`}
            </Button>
          )}
        </Space>
      </div>
    </Card>
  )
}

export default AssetInitPage
