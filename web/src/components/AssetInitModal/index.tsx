import { useState, useEffect, useRef } from 'react'
import { Modal, Steps, Table, Tag, Space, Spin, Checkbox, Collapse, Badge, Button, Alert, message } from 'antd'
import {
  CheckCircleOutlined, LoadingOutlined, BookOutlined,
} from '@ant-design/icons'
import {
  getPlaybooks, executePlaybook,
  type Playbook, type PlaybookStep, type PlaybookAssetResult, type StepExecResult,
} from '../../services/playbook'
import { connectPlaybookWS, type ManagedWS, type WSState } from '../../utils/websocket'

interface AssetInitModalProps {
  open: boolean
  assetIds: number[]
  onClose: () => void
  onSuccess?: () => void
}

interface ExecuteResponse {
  total: number
  success: number
  failed: number
  results: PlaybookAssetResult[]
}

const AssetInitModal = ({ open, assetIds, onClose, onSuccess }: AssetInitModalProps) => {
  const [currentStep, setCurrentStep] = useState(0)
  const [loading, setLoading] = useState(false)
  const [playbooks, setPlaybooks] = useState<Playbook[]>([])
  const [playbooksLoading, setPlaybooksLoading] = useState(false)
  const [selectedIds, setSelectedIds] = useState<number[]>([])
  const [results, setResults] = useState<PlaybookAssetResult[]>([])
  const [summary, setSummary] = useState<ExecuteResponse | null>(null)
  const [wsState, setWsState] = useState<WSState | null>(null)
  const wsRef = useRef<ManagedWS | null>(null)

  useEffect(() => {
    if (open) {
      setCurrentStep(0)
      setResults([])
      setSummary(null)
      setSelectedIds([])
      setWsState(null)
      setPlaybooksLoading(true)
      getPlaybooks().then(data => {
        setPlaybooks((data || []).filter(p => p.enabled))
      }).catch(() => {}).finally(() => setPlaybooksLoading(false))
    }
    return () => {
      if (wsRef.current) {
        wsRef.current.close()
        wsRef.current = null
      }
    }
  }, [open])

  const togglePlaybook = (id: number) => {
    setSelectedIds(prev => prev.includes(id) ? prev.filter(x => x !== id) : [...prev, id])
  }

  const handleExecute = async () => {
    if (selectedIds.length === 0) {
      message.warning('请至少选择一个剧本')
      return
    }
    setLoading(true)
    setCurrentStep(1)
    setResults([])
    setSummary(null)
    try {
      const job = await executePlaybook({ playbook_ids: selectedIds, asset_ids: assetIds })

      const ws = connectPlaybookWS(job.job_id, (msg) => {
        if (msg.type === 'playbook_progress') {
          setResults(prev => [...prev, msg.data as PlaybookAssetResult])
        } else if (msg.type === 'playbook_done') {
          const data = msg.data as ExecuteResponse
          setSummary(data)
          setLoading(false)
          if (data.success > 0 && onSuccess) onSuccess()
          if (wsRef.current) {
            wsRef.current.close()
            wsRef.current = null
          }
        }
      }, setWsState)
      if (!ws) {
        message.error('未登录或连接失败')
        setLoading(false)
        return
      }
      wsRef.current = ws
    } catch {
      setLoading(false)
    }
  }

  const handleClose = () => {
    if (wsRef.current) {
      wsRef.current.close()
      wsRef.current = null
    }
    onClose()
  }

  return (
    <Modal
      title="资产初始化"
      open={open}
      onCancel={handleClose}
      width={currentStep === 0 ? 560 : 800}
      footer={currentStep === 0 ? [
        <Button key="cancel" onClick={handleClose}>取消</Button>,
        <Button key="submit" type="primary" onClick={handleExecute} disabled={selectedIds.length === 0}>
          开始执行 ({selectedIds.length} 个剧本 × {assetIds.length} 台)
        </Button>,
      ] : [
        <Button key="close" onClick={handleClose}>关闭</Button>,
      ]}
      destroyOnHidden
    >
      <Steps
        current={currentStep}
        size="small"
        style={{ marginBottom: 24, marginTop: 8 }}
        items={[{ title: '选择剧本' }, { title: '执行结果' }]}
      />

      {currentStep === 0 && (
        <Spin spinning={playbooksLoading}>
          <div style={{ color: '#979BA5', fontSize: 13, marginBottom: 16 }}>
            <BookOutlined style={{ marginRight: 4 }} />
            选择要对 {assetIds.length} 台资产执行的剧本（可多选）：
          </div>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
            {playbooks.map(pb => {
              const checked = selectedIds.includes(pb.id)
              return (
                <div
                  key={pb.id}
                  style={{
                    border: `1px solid ${checked ? '#3A84FF' : '#E7E9EF'}`,
                    borderRadius: 2,
                    padding: '12px 14px',
                    cursor: 'pointer',
                    background: checked ? '#F0F5FF' : '#fff',
                    transition: 'all 0.2s ease',
                  }}
                  onClick={() => togglePlaybook(pb.id)}
                >
                  <div style={{ display: 'flex', alignItems: 'center' }}>
                    <Checkbox checked={checked} style={{ marginRight: 10 }} />
                    <div style={{ flex: 1 }}>
                      <Space size={4}>
                        <span style={{ fontWeight: 500 }}>{pb.name}</span>
                        {pb.is_builtin && <Tag color="blue">内置</Tag>}
                        <Tag>{pb.steps?.length || 0} 步骤</Tag>
                      </Space>
                      <div style={{ fontSize: 12, color: '#979BA5', marginTop: 2 }}>{pb.description}</div>
                    </div>
                  </div>
                  {checked && pb.steps?.length > 0 && (
                    <Collapse
                      ghost size="small" style={{ marginTop: 6, marginLeft: 28 }}
                      items={[{
                        key: 'steps',
                        label: <span style={{ fontSize: 12, color: '#3A84FF' }}>查看步骤</span>,
                        children: pb.steps.map((step: PlaybookStep, idx: number) => (
                          <div key={idx} style={{ display: 'flex', alignItems: 'center', gap: 6, padding: '3px 0', fontSize: 12, color: '#63656E' }}>
                            <Badge count={idx + 1} style={{ backgroundColor: '#3A84FF' }} size="small" />
                            <span>{step.name}</span>
                          </div>
                        )),
                      }]}
                    />
                  )}
                </div>
              )
            })}
          </div>
        </Spin>
      )}

      {currentStep === 1 && (
        <>
          {wsState === 'reconnecting' && (
            <Alert type="warning" message="连接中断，正在尝试重连..." showIcon style={{ marginBottom: 16 }} />
          )}
          {wsState === 'disconnected' && loading && (
            <Alert type="error" message="连接已断开，无法接收实时结果" showIcon style={{ marginBottom: 16 }} />
          )}
          <div style={{ marginBottom: 16 }}>
            <Space size={16}>
              {loading ? (
                <>
                  <Spin indicator={<LoadingOutlined style={{ fontSize: 16, color: '#3A84FF' }} spin />} />
                  <span style={{ color: '#63656E' }}>正在执行... {results.length}/{assetIds.length}</span>
                </>
              ) : summary ? (
                <>
                  <CheckCircleOutlined style={{ fontSize: 16, color: '#2DCB56' }} />
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
            columns={[
              { title: '主机', width: 160, render: (_: unknown, r: PlaybookAssetResult) => <span>{r.hostname || r.ip}<br /><span style={{ fontSize: 12, color: '#979BA5' }}>{r.ip}</span></span> },
              {
                title: '状态', dataIndex: 'status', width: 80,
                render: (v: string) => {
                  const map: Record<string, { color: string; text: string }> = {
                    success: { color: 'success', text: '成功' }, partial: { color: 'warning', text: '部分' }, failed: { color: 'error', text: '失败' },
                  }
                  const m = map[v] || { color: 'default', text: v }
                  return <Tag color={m.color}>{m.text}</Tag>
                },
              },
              {
                title: '步骤', width: 180,
                render: (_: unknown, r: PlaybookAssetResult) => {
                  if (!r.step_results) return '-'
                  const s = r.step_results.filter(x => x.status === 'success').length
                  const f = r.step_results.filter(x => x.status === 'failed').length
                  return <Space size={4}>{s > 0 && <Tag color="green">{s} 成功</Tag>}{f > 0 && <Tag color="red">{f} 失败</Tag>}</Space>
                },
              },
            ]}
            dataSource={results}
            size="middle"
            pagination={false}
            scroll={{ y: 360 }}
            expandable={{
              expandedRowRender: (record: PlaybookAssetResult) => (
                <div>
                  {record.step_results?.map((sr: StepExecResult, i: number) => (
                    <div key={i} style={{ marginBottom: 6, padding: 6, background: '#F5F7FA', borderRadius: 2 }}>
                      <Space>
                        <Badge count={i + 1} style={{ backgroundColor: sr.status === 'success' ? '#2DCB56' : sr.status === 'failed' ? '#EA3636' : '#979BA5' }} />
                        <strong style={{ fontSize: 12 }}>{sr.name}</strong>
                        <Tag color={sr.status === 'success' ? 'green' : sr.status === 'failed' ? 'red' : 'default'}>{sr.status}</Tag>
                        <span style={{ color: '#979BA5', fontSize: 11 }}>{sr.latency}ms</span>
                      </Space>
                      {sr.output && <pre style={{ margin: '4px 0 0', fontSize: 11, maxHeight: 100, overflow: 'auto', background: '#F0F1F5', padding: 6, borderRadius: 2 }}>{sr.output}</pre>}
                      {sr.error && <pre style={{ margin: '4px 0 0', fontSize: 11, color: '#EA3636', maxHeight: 60, overflow: 'auto' }}>{sr.error}</pre>}
                    </div>
                  ))}
                </div>
              ),
            }}
          />
        </>
      )}
    </Modal>
  )
}

export default AssetInitModal
