import { useState, useEffect } from 'react'
import {
  Card, Table, Button, Space, Tag, Modal, Form, Input, message, Popconfirm,
  Switch, Tabs, Select, Collapse, InputNumber, Drawer, Badge, Typography,
} from 'antd'
import {
  BookOutlined, PlusOutlined, PlayCircleOutlined, EditOutlined,
  DeleteOutlined, EyeOutlined, HistoryOutlined, ThunderboltOutlined,
} from '@ant-design/icons'
import {
  getPlaybooks, createPlaybook, updatePlaybook, deletePlaybook,
  executePlaybook, getExecutions, getExecutionResults,
  type Playbook, type PlaybookStep, type ExecutionListItem, type PlaybookAssetResult,
} from '../../services/playbook'
import BatchAssetSelector from '../../components/BatchAssetSelector'

const { TextArea } = Input

const PlaybooksPage = () => {
  const [playbooks, setPlaybooks] = useState<Playbook[]>([])
  const [loading, setLoading] = useState(false)
  const [editOpen, setEditOpen] = useState(false)
  const [editingPlaybook, setEditingPlaybook] = useState<Playbook | null>(null)
  const [form] = Form.useForm()
  const [steps, setSteps] = useState<PlaybookStep[]>([{ name: '', script: '', timeout: 60, on_error: 'stop' }])

  // 执行相关
  const [execOpen, setExecOpen] = useState(false)
  const [execPlaybook, setExecPlaybook] = useState<Playbook | null>(null)
  const [selectedAssetIds, setSelectedAssetIds] = useState<number[]>([])
  const [executing, setExecuting] = useState(false)

  // 查看详情
  const [viewOpen, setViewOpen] = useState(false)
  const [viewPlaybook, setViewPlaybook] = useState<Playbook | null>(null)

  // 执行历史
  const [executions, setExecutions] = useState<ExecutionListItem[]>([])
  const [execTotal, setExecTotal] = useState(0)
  const [execPage, setExecPage] = useState(1)
  const [execLoading, setExecLoading] = useState(false)

  // 执行结果详情
  const [resultOpen, setResultOpen] = useState(false)
  const [resultData, setResultData] = useState<PlaybookAssetResult[]>([])
  const [resultLoading, setResultLoading] = useState(false)

  const fetchPlaybooks = async () => {
    setLoading(true)
    try {
      const data = await getPlaybooks()
      setPlaybooks(data || [])
    } catch { /* handled */ } finally { setLoading(false) }
  }

  const fetchExecutions = async (page = 1) => {
    setExecLoading(true)
    try {
      const data = await getExecutions({ page, page_size: 20 })
      setExecutions(data.list || [])
      setExecTotal(data.total)
      setExecPage(page)
    } catch { /* handled */ } finally { setExecLoading(false) }
  }

  useEffect(() => {
    fetchPlaybooks()
    fetchExecutions()
  }, [])

  // 创建/编辑
  const handleOpenEdit = (pb?: Playbook) => {
    if (pb) {
      setEditingPlaybook(pb)
      form.setFieldsValue({ name: pb.name, description: pb.description })
      setSteps(pb.steps.length > 0 ? pb.steps : [{ name: '', script: '', timeout: 60, on_error: 'stop' }])
    } else {
      setEditingPlaybook(null)
      form.resetFields()
      setSteps([{ name: '', script: '', timeout: 60, on_error: 'stop' }])
    }
    setEditOpen(true)
  }

  const handleSave = async () => {
    try {
      const values = await form.validateFields()
      const validSteps = steps.filter((s) => s.name && s.script)
      if (validSteps.length === 0) {
        message.warning('请至少添加一个有效步骤')
        return
      }

      if (editingPlaybook) {
        await updatePlaybook(editingPlaybook.id, { ...values, steps: validSteps })
        message.success('更新成功')
      } else {
        await createPlaybook({ ...values, steps: validSteps })
        message.success('创建成功')
      }
      setEditOpen(false)
      fetchPlaybooks()
    } catch { /* validation failed */ }
  }

  const handleDelete = async (id: number) => {
    try {
      await deletePlaybook(id)
      message.success('删除成功')
      fetchPlaybooks()
    } catch { /* handled */ }
  }

  const handleToggle = async (pb: Playbook, checked: boolean) => {
    try {
      await updatePlaybook(pb.id, { name: pb.name, description: pb.description, enabled: checked, steps: pb.steps })
      message.success(checked ? '已启用' : '已禁用')
      fetchPlaybooks()
    } catch { /* handled */ }
  }

  // 执行
  const handleOpenExec = (pb: Playbook) => {
    setExecPlaybook(pb)
    setSelectedAssetIds([])
    setExecOpen(true)
  }

  const handleExecute = async () => {
    if (!execPlaybook || selectedAssetIds.length === 0) {
      message.warning('请选择目标资产')
      return
    }
    setExecuting(true)
    try {
      const result = await executePlaybook({ playbook_ids: [execPlaybook.id], asset_ids: selectedAssetIds })
      message.success(`剧本执行已启动，任务 ID: ${result.job_id}`)
      setExecOpen(false)
      fetchExecutions()
    } catch { /* handled */ } finally { setExecuting(false) }
  }

  // 查看结果
  const handleViewResults = async (executionId: number) => {
    setResultOpen(true)
    setResultLoading(true)
    try {
      const data = await getExecutionResults(executionId)
      setResultData(data || [])
    } catch { /* handled */ } finally { setResultLoading(false) }
  }

  // 步骤管理
  const addStep = () => setSteps([...steps, { name: '', script: '', timeout: 60, on_error: 'stop' }])
  const removeStep = (idx: number) => setSteps(steps.filter((_, i) => i !== idx))
  const updateStep = (idx: number, field: keyof PlaybookStep, value: string | number) => {
    const newSteps = [...steps]
    const step = { ...newSteps[idx] }
    if (field === 'timeout') {
      step.timeout = value as number
    } else {
      step[field] = value as string
    }
    newSteps[idx] = step
    setSteps(newSteps)
  }

  const playbookColumns = [
    {
      title: '剧本名称', dataIndex: 'name', key: 'name', width: 180,
      render: (name: string, record: Playbook) => (
        <Space>
          {record.is_builtin && <Tag color="blue">内置</Tag>}
          <span style={{ fontWeight: 500 }}>{name}</span>
        </Space>
      ),
    },
    { title: '描述', dataIndex: 'description', key: 'description', ellipsis: true },
    { title: '步骤数', key: 'steps', width: 80, render: (_: unknown, r: Playbook) => r.steps?.length || 0 },
    {
      title: '状态', dataIndex: 'enabled', key: 'enabled', width: 80,
      render: (enabled: boolean, record: Playbook) => (
        <Switch size="small" checked={enabled} onChange={(v) => handleToggle(record, v)} />
      ),
    },
    { title: '更新时间', dataIndex: 'updated_at', key: 'updated_at', width: 170 },
    {
      title: '操作', key: 'action', width: 220,
      render: (_: unknown, record: Playbook) => (
        <Space size={4}>
          <Button type="link" size="small" icon={<EyeOutlined />} onClick={() => { setViewPlaybook(record); setViewOpen(true) }}>
            查看
          </Button>
          <Button type="link" size="small" icon={<PlayCircleOutlined />} onClick={() => handleOpenExec(record)} disabled={!record.enabled}>
            执行
          </Button>
          {!record.is_builtin && (
            <>
              <Button type="link" size="small" icon={<EditOutlined />} onClick={() => handleOpenEdit(record)}>编辑</Button>
              <Popconfirm title="确认删除该剧本？" onConfirm={() => handleDelete(record.id)}>
                <Button type="link" size="small" danger icon={<DeleteOutlined />}>删除</Button>
              </Popconfirm>
            </>
          )}
        </Space>
      ),
    },
  ]

  const executionColumns = [
    { title: '剧本', dataIndex: 'playbook_name', key: 'playbook_name', width: 160 },
    {
      title: '状态', dataIndex: 'status', key: 'status', width: 100,
      render: (s: string) => {
        const map: Record<string, { color: string; text: string }> = {
          running: { color: 'blue', text: '执行中' },
          success: { color: 'green', text: '成功' },
          partial: { color: 'orange', text: '部分成功' },
          failed: { color: 'red', text: '失败' },
        }
        const item = map[s] || { color: 'default', text: s }
        return <Tag color={item.color}>{item.text}</Tag>
      },
    },
    { title: '资产数', dataIndex: 'total_assets', key: 'total_assets', width: 80 },
    { title: '成功', dataIndex: 'success', key: 'success', width: 60, render: (v: number) => <span style={{ color: '#2DCB56' }}>{v}</span> },
    { title: '失败', dataIndex: 'failed', key: 'failed', width: 60, render: (v: number) => <span style={{ color: v > 0 ? '#EA3636' : undefined }}>{v}</span> },
    { title: '开始时间', dataIndex: 'started_at', key: 'started_at', width: 170 },
    { title: '完成时间', dataIndex: 'finished_at', key: 'finished_at', width: 170 },
    {
      title: '操作', key: 'action', width: 80,
      render: (_: unknown, record: ExecutionListItem) => (
        <Button type="link" size="small" icon={<EyeOutlined />} onClick={() => handleViewResults(record.id)}>详情</Button>
      ),
    },
  ]

  // assetColumns removed - BatchAssetSelector handles display

  return (
    <Tabs
      defaultActiveKey="playbooks"
      items={[
        {
          key: 'playbooks',
          label: <><BookOutlined /> 剧本管理</>,
          children: (
            <>
              <Card
                title={<><ThunderboltOutlined style={{ marginRight: 8, color: '#3A84FF' }} />运维剧本</>}
                extra={
                  <Button type="primary" icon={<PlusOutlined />} onClick={() => handleOpenEdit()}>
                    新建剧本
                  </Button>
                }
                style={{ border: '1px solid #E7E9EF' }}
              >
                <Table
                  dataSource={playbooks}
                  columns={playbookColumns}
                  rowKey="id"
                  size="middle"
                  loading={loading}
                  pagination={false}
                />
              </Card>

              {/* 创建/编辑弹窗 */}
              <Modal
                title={editingPlaybook ? '编辑剧本' : '新建剧本'}
                open={editOpen}
                onCancel={() => setEditOpen(false)}
                onOk={handleSave}
                width={720}
                destroyOnHidden
              >
                <Form form={form} layout="vertical" style={{ marginTop: 16 }}>
                  <Form.Item name="name" label="剧本名称" rules={[{ required: true, message: '请输入剧本名称' }]}>
                    <Input placeholder="如：内核优化" />
                  </Form.Item>
                  <Form.Item name="description" label="描述">
                    <TextArea rows={2} placeholder="描述该剧本的用途" />
                  </Form.Item>
                </Form>

                <div style={{ marginBottom: 8, fontWeight: 500 }}>执行步骤</div>
                <Collapse
                  accordion
                  items={steps.map((step, idx) => ({
                    key: String(idx),
                    label: (
                      <Space>
                        <Badge count={idx + 1} style={{ backgroundColor: '#3A84FF' }} />
                        <span>{step.name || `步骤 ${idx + 1}`}</span>
                      </Space>
                    ),
                    extra: steps.length > 1 ? (
                      <DeleteOutlined style={{ color: '#EA3636' }} onClick={(e) => { e.stopPropagation(); removeStep(idx) }} />
                    ) : null,
                    children: (
                      <Space direction="vertical" style={{ width: '100%' }}>
                        <Input placeholder="步骤名称" value={step.name} onChange={(e) => updateStep(idx, 'name', e.target.value)} />
                        <TextArea
                          rows={4}
                          placeholder="Shell 脚本"
                          value={step.script}
                          onChange={(e) => updateStep(idx, 'script', e.target.value)}
                          style={{ fontFamily: 'monospace', fontSize: 13 }}
                        />
                        <Space>
                          <span>超时(秒):</span>
                          <InputNumber min={5} max={3600} value={step.timeout} onChange={(v) => updateStep(idx, 'timeout', v || 60)} />
                          <span>失败策略:</span>
                          <Select
                            value={step.on_error}
                            onChange={(v) => updateStep(idx, 'on_error', v)}
                            options={[
                              { value: 'stop', label: '停止后续步骤' },
                              { value: 'continue', label: '继续执行' },
                            ]}
                            style={{ width: 140 }}
                          />
                        </Space>
                      </Space>
                    ),
                  }))}
                />
                <Button type="dashed" block style={{ marginTop: 12 }} icon={<PlusOutlined />} onClick={addStep}>
                  添加步骤
                </Button>
              </Modal>

              {/* 查看详情 */}
              <Drawer
                title={viewPlaybook?.name}
                open={viewOpen}
                onClose={() => setViewOpen(false)}
                width={600}
              >
                {viewPlaybook && (
                  <>
                    <Typography.Paragraph type="secondary">{viewPlaybook.description}</Typography.Paragraph>
                    <div style={{ marginBottom: 8 }}>
                      {viewPlaybook.is_builtin && <Tag color="blue">内置剧本</Tag>}
                      <Tag color={viewPlaybook.enabled ? 'green' : 'default'}>{viewPlaybook.enabled ? '已启用' : '已禁用'}</Tag>
                    </div>
                    <Typography.Title level={5} style={{ marginTop: 16 }}>执行步骤 ({viewPlaybook.steps?.length || 0})</Typography.Title>
                    <Collapse
                      items={viewPlaybook.steps?.map((step, idx) => ({
                        key: String(idx),
                        label: (
                          <Space>
                            <Badge count={idx + 1} style={{ backgroundColor: '#3A84FF' }} />
                            <span>{step.name}</span>
                            <Tag>{step.on_error === 'stop' ? '失败停止' : '失败继续'}</Tag>
                            <Tag>超时 {step.timeout}s</Tag>
                          </Space>
                        ),
                        children: (
                          <pre style={{ background: '#F5F7FA', padding: 12, borderRadius: 2, fontSize: 13, overflow: 'auto', maxHeight: 300 }}>
                            {step.script}
                          </pre>
                        ),
                      }))}
                    />
                  </>
                )}
              </Drawer>

              {/* 执行弹窗 */}
              <Modal
                title={`执行剧本：${execPlaybook?.name}`}
                open={execOpen}
                onCancel={() => setExecOpen(false)}
                onOk={handleExecute}
                okText="开始执行"
                okButtonProps={{ loading: executing, disabled: selectedAssetIds.length === 0 }}
                width={900}
                destroyOnHidden
              >
                <BatchAssetSelector value={selectedAssetIds} onChange={setSelectedAssetIds} maxHeight={360} />
              </Modal>
            </>
          ),
        },
        {
          key: 'history',
          label: <><HistoryOutlined /> 执行历史</>,
          children: (
            <>
              <Card title="执行历史" style={{ border: '1px solid #E7E9EF' }}>
                <Table
                  dataSource={executions}
                  columns={executionColumns}
                  rowKey="id"
                  size="middle"
                  loading={execLoading}
                  pagination={{
                    current: execPage,
                    total: execTotal,
                    pageSize: 20,
                    showSizeChanger: true,
                    showTotal: (t) => `共 ${t} 条`,
                    onChange: fetchExecutions,
                  }}
                />
              </Card>

              {/* 执行结果详情 */}
              <Drawer
                title="执行结果详情"
                open={resultOpen}
                onClose={() => setResultOpen(false)}
                width={700}
              >
                <Table
                  dataSource={resultData}
                  rowKey="asset_id"
                  size="small"
                  loading={resultLoading}
                  pagination={false}
                  columns={[
                    { title: '主机名', dataIndex: 'hostname', key: 'hostname', width: 120, ellipsis: true },
                    { title: 'IP', dataIndex: 'ip', key: 'ip', width: 130 },
                    {
                      title: '状态', dataIndex: 'status', key: 'status', width: 80,
                      render: (s: string) => {
                        const map: Record<string, { color: string; text: string }> = {
                          success: { color: 'green', text: '成功' },
                          partial: { color: 'orange', text: '部分成功' },
                          failed: { color: 'red', text: '失败' },
                        }
                        const item = map[s] || { color: 'default', text: s }
                        return <Tag color={item.color}>{item.text}</Tag>
                      },
                    },
                  ]}
                  expandable={{
                    expandedRowRender: (record: PlaybookAssetResult) => (
                      <div>
                        {record.step_results?.map((sr, i) => (
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
              </Drawer>
            </>
          ),
        },
      ]}
    />
  )
}

export default PlaybooksPage
