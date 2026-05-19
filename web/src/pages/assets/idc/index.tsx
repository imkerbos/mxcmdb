import { useState, useEffect, useRef } from 'react'
import { useNavigate } from 'react-router-dom'
import { Card, Table, Button, Space, Input, Select, Tag, Modal, Form, message, Popconfirm, Upload, Alert, Divider, Result, Tooltip } from 'antd'
import { PlusOutlined, SearchOutlined, ReloadOutlined, ImportOutlined, ApiOutlined, DeleteOutlined, UploadOutlined, DownloadOutlined, CheckCircleOutlined, CloseCircleOutlined, LoadingOutlined, SettingOutlined, KeyOutlined, LockOutlined } from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import {
  getAssets, createAsset, updateAsset, deleteAsset, importAssets, testConnection, batchDeleteAssets,
  type AssetItem, type CreateAssetParams, type TestConnectionResult
} from '../../../services/asset'
import { getAllProjects, type ProjectSimple } from '../../../services/project'
import { getAllSSHKeys, type SSHKeyItem } from '../../../services/ssh-key'
import AssetInitModal from '../../../components/AssetInitModal'

const statusColors: Record<string, string> = { online: 'green', offline: 'red', unknown: 'default' }
const statusLabels: Record<string, string> = { online: '在线', offline: '离线', unknown: '未知' }
const envColors: Record<string, string> = { prod: 'red', test: 'orange', dev: 'blue' }

const IDCAssetsPage = () => {
  const navigate = useNavigate()
  const [data, setData] = useState<AssetItem[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(false)
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(20)
  const [keyword, setKeyword] = useState('')
  const [envFilter, setEnvFilter] = useState('')
  const [modalOpen, setModalOpen] = useState(false)
  const [editing, setEditing] = useState<AssetItem | null>(null)
  const [form] = Form.useForm()
  const [selectedRowKeys, setSelectedRowKeys] = useState<React.Key[]>([])

  // 批量导入
  const [importOpen, setImportOpen] = useState(false)
  const [importData, setImportData] = useState<CreateAssetParams[]>([])
  const [importLoading, setImportLoading] = useState(false)
  const [importResult, setImportResult] = useState<{ total: number; success: number; failed: number; errors?: { index: number; ip: string; message: string }[] } | null>(null)
  const textAreaRef = useRef<HTMLTextAreaElement>(null)

  // 测试连接
  const [testOpen, setTestOpen] = useState(false)
  const [testLoading, setTestLoading] = useState(false)
  const [testResults, setTestResults] = useState<TestConnectionResult[]>([])

  // 项目列表
  const [projects, setProjects] = useState<ProjectSimple[]>([])

  // SSH 密钥列表
  const [sshKeys, setSSHKeys] = useState<SSHKeyItem[]>([])

  // CSV 模板下载
  const handleDownloadTemplate = () => {
    const headers = ['ip', 'hostname', 'port', 'type', 'os', 'ssh_user', 'ssh_password', 'environment', 'department', 'owner', 'spec', 'business_group']
    const example = ['10.0.0.1', 'web-01', '22', 'physical_server', 'CentOS 7', 'root', '', 'prod', '运维部', 'zhangsan', '4C8G', '']
    const csv = [headers.join(','), example.join(',')].join('\n')
    const blob = new Blob(['\uFEFF' + csv], { type: 'text/csv;charset=utf-8;' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = 'asset_import_template.csv'
    a.click()
    URL.revokeObjectURL(url)
  }

  // 初始化
  const [initOpen, setInitOpen] = useState(false)

  const fetchData = async (p = page, ps = pageSize) => {
    setLoading(true)
    try {
      const params: Record<string, unknown> = { source: 'manual', page: p, page_size: ps }
      if (keyword) params.keyword = keyword
      if (envFilter) params.environment = envFilter
      const result = await getAssets(params as any)
      setData(result.list || [])
      setTotal(result.total)
    } catch { /* handled */ } finally { setLoading(false) }
  }

  useEffect(() => {
    fetchData()
    getAllProjects().then(setProjects).catch(() => {})
    getAllSSHKeys().then(setSSHKeys).catch(() => {})
  }, [])

  const handleCreate = () => {
    setEditing(null)
    form.resetFields()
    form.setFieldsValue({ type: 'physical_server', port: 22, ssh_user: 'root' })
    setModalOpen(true)
  }

  const handleEdit = (record: AssetItem) => {
    setEditing(record)
    form.setFieldsValue({
      hostname: record.hostname,
      ip: record.ip,
      port: record.port,
      type: record.type,
      os: record.os,
      status: record.status,
      spec: record.spec,
      department: record.department,
      project_id: record.project_id,
      owner: record.owner,
      environment: record.environment,
      ssh_user: record.ssh_user,
      ssh_key_id: record.ssh_key_id,
    })
    setModalOpen(true)
  }

  const handleModalOk = async () => {
    try {
      const values = await form.validateFields()
      if (editing) {
        const updateData: Record<string, unknown> = { ...values }
        // 如果清除了密钥选择
        if (editing.ssh_key_id && !values.ssh_key_id) {
          updateData.clear_ssh_key = true
        }
        await updateAsset(editing.id, updateData)
        message.success('更新成功')
      } else {
        await createAsset({ ...values, source: 'manual' } as CreateAssetParams)
        message.success('创建成功')
      }
      setModalOpen(false)
      fetchData()
    } catch { /* validation */ }
  }

  // 批量导入
  const handleImportOpen = () => {
    setImportData([])
    setImportResult(null)
    setImportOpen(true)
  }

  const parseCSV = (text: string): CreateAssetParams[] => {
    const lines = text.trim().split('\n').filter(l => l.trim())
    if (lines.length < 2) return []
    const headers = lines[0].split(',').map(h => h.trim().toLowerCase())
    return lines.slice(1).map(line => {
      const values = line.split(',').map(v => v.trim())
      const item: Record<string, unknown> = {}
      headers.forEach((h, i) => {
        if (h === 'port') item[h] = parseInt(values[i]) || 22
        else item[h] = values[i] || ''
      })
      if (!item.type) item.type = 'physical_server'
      if (!item.source) item.source = 'manual'
      return item as unknown as CreateAssetParams
    })
  }

  const handleFileUpload = (file: File) => {
    const reader = new FileReader()
    reader.onload = (e) => {
      const text = e.target?.result as string
      try {
        // 尝试 JSON 格式
        const parsed = JSON.parse(text)
        const items = Array.isArray(parsed) ? parsed : parsed.assets || []
        setImportData(items)
        message.success(`解析成功，共 ${items.length} 条记录`)
      } catch {
        // 尝试 CSV 格式
        const items = parseCSV(text)
        if (items.length > 0) {
          setImportData(items)
          message.success(`解析成功，共 ${items.length} 条记录`)
        } else {
          message.error('文件格式错误，支持 JSON 或 CSV 格式')
        }
      }
    }
    reader.readAsText(file)
    return false
  }

  const handleImportSubmit = async () => {
    if (importData.length === 0) {
      message.warning('请先上传或粘贴资产数据')
      return
    }
    setImportLoading(true)
    try {
      const result = await importAssets(importData)
      setImportResult(result)
      if (result.success > 0) {
        message.success(`成功导入 ${result.success} 条`)
        fetchData()
      }
    } catch { /* handled */ } finally { setImportLoading(false) }
  }

  const handleParseText = () => {
    const text = textAreaRef.current?.value || ''
    if (!text.trim()) { message.warning('请输入数据'); return }
    try {
      const parsed = JSON.parse(text)
      const items = Array.isArray(parsed) ? parsed : parsed.assets || []
      setImportData(items)
      message.success(`解析成功，共 ${items.length} 条记录`)
    } catch {
      const items = parseCSV(text)
      if (items.length > 0) {
        setImportData(items)
        message.success(`解析成功，共 ${items.length} 条记录`)
      } else {
        message.error('格式错误，支持 JSON 数组或 CSV 格式')
      }
    }
  }

  // 测试连接
  const handleTestConnection = async () => {
    const ids = selectedRowKeys.map(Number)
    if (ids.length === 0) { message.warning('请先选择要测试的资产'); return }
    setTestResults([])
    setTestOpen(true)
    setTestLoading(true)
    try {
      const result = await testConnection(ids)
      setTestResults(result.results || [])
    } catch { /* handled */ } finally { setTestLoading(false) }
  }

  // 批量删除
  const handleBatchDelete = async () => {
    const ids = selectedRowKeys.map(Number)
    if (ids.length === 0) return
    try {
      const result = await batchDeleteAssets(ids)
      message.success(`成功删除 ${result.success} 条${result.failed > 0 ? `，失败 ${result.failed} 条` : ''}`)
      setSelectedRowKeys([])
      fetchData()
    } catch { /* handled */ }
  }

  // 认证方式渲染（紧凑图标模式）
  const renderAuthMethod = (record: AssetItem) => {
    const methods: React.ReactNode[] = []
    if (record.has_password) {
      methods.push(<Tooltip key="pwd" title="密码认证"><LockOutlined style={{ color: '#979BA5', fontSize: 15 }} /></Tooltip>)
    }
    if (record.ssh_key_id) {
      methods.push(<Tooltip key="key" title={record.ssh_key_name || '密钥认证'}><KeyOutlined style={{ color: '#3A84FF', fontSize: 15 }} /></Tooltip>)
    }
    if (methods.length === 0) {
      return <Tag color="orange">未配置</Tag>
    }
    return <Space size={8}>{methods}</Space>
  }

  const columns: ColumnsType<AssetItem> = [
    { title: '主机名', dataIndex: 'hostname', width: 160, ellipsis: true },
    { title: 'IP', dataIndex: 'ip', width: 140 },
    { title: '状态', dataIndex: 'status', width: 80, render: (v) => <Tag color={statusColors[v] || 'default'}>{statusLabels[v] || v}</Tag> },
    { title: '环境', dataIndex: 'environment', width: 80, render: (v) => v ? <Tag color={envColors[v] || 'default'}>{v}</Tag> : '-' },
    {
      title: '认证', width: 100, key: 'auth',
      render: (_, record) => renderAuthMethod(record),
    },
    { title: '部门', dataIndex: 'department', width: 100, ellipsis: true },
    { title: '项目', dataIndex: 'project_name', width: 100, ellipsis: true, render: (v) => v || '-' },
    {
      title: '操作', width: 150,
      render: (_, record) => (
        <Space size={0}>
          <Button type="link" size="small" onClick={() => navigate(`/assets/${record.id}/detail`)}>详情</Button>
          <Button type="link" size="small" onClick={() => handleEdit(record)}>编辑</Button>
          <Popconfirm title="确定删除？" onConfirm={async () => { await deleteAsset(record.id); message.success('已删除'); fetchData() }}>
            <Button type="link" size="small" danger>删除</Button>
          </Popconfirm>
        </Space>
      ),
    },
  ]

  const importPreviewColumns: ColumnsType<CreateAssetParams> = [
    { title: 'IP', dataIndex: 'ip', width: 140 },
    { title: '主机名', dataIndex: 'hostname', width: 140 },
    { title: '端口', dataIndex: 'port', width: 60 },
    { title: '类型', dataIndex: 'type', width: 100 },
    { title: 'SSH 用户', dataIndex: 'ssh_user', width: 100 },
    { title: '环境', dataIndex: 'environment', width: 80 },
    { title: '部门', dataIndex: 'department', width: 100 },
  ]

  const authMethodLabels: Record<string, string> = {
    password: '密码',
    key: '密钥',
    'password+key': '密码+密钥',
    none: '无',
  }

  const testResultColumns: ColumnsType<TestConnectionResult> = [
    { title: '主机名', dataIndex: 'hostname', width: 140, ellipsis: true },
    { title: 'IP', dataIndex: 'ip', width: 140 },
    {
      title: '状态', dataIndex: 'status', width: 100,
      render: (v) => v === 'success'
        ? <Tag icon={<CheckCircleOutlined />} color="success">成功</Tag>
        : <Tag icon={<CloseCircleOutlined />} color="error">失败</Tag>,
    },
    {
      title: '认证方式', dataIndex: 'auth_method', width: 100,
      render: (v: string) => authMethodLabels[v] || v,
    },
    { title: '延迟', dataIndex: 'latency', width: 80, render: (v) => `${v} ms` },
    { title: '错误信息', dataIndex: 'error', ellipsis: true, render: (v) => v || '-' },
  ]

  return (
    <>
      <Card title="IDC 资产" style={{ border: '1px solid #E7E9EF' }}>
        <Space style={{ marginBottom: 16 }} wrap>
          <Input placeholder="搜索主机名/IP" prefix={<SearchOutlined />} value={keyword} onChange={(e) => setKeyword(e.target.value)} onPressEnter={() => { setPage(1); fetchData(1, pageSize) }} style={{ width: 200 }} />
          <Select value={envFilter} onChange={setEnvFilter} style={{ width: 120 }} options={[{ value: '', label: '全部环境' }, { value: 'prod', label: '生产' }, { value: 'test', label: '测试' }, { value: 'dev', label: '开发' }]} />
          <Button type="primary" icon={<SearchOutlined />} onClick={() => { setPage(1); fetchData(1, pageSize) }}>搜索</Button>
          <Button icon={<ReloadOutlined />} onClick={() => { setKeyword(''); setEnvFilter(''); fetchData(1, pageSize) }}>重置</Button>
          <Divider type="vertical" />
          <Button type="primary" icon={<PlusOutlined />} onClick={handleCreate}>添加资产</Button>
          <Button icon={<ImportOutlined />} onClick={handleImportOpen}>批量导入</Button>
          <Button type="primary" ghost icon={<SettingOutlined />} onClick={() => setInitOpen(true)} disabled={selectedRowKeys.length === 0}>
            初始化{selectedRowKeys.length > 0 ? ` (${selectedRowKeys.length})` : ''}
          </Button>
          <Button icon={<ApiOutlined />} onClick={handleTestConnection} disabled={selectedRowKeys.length === 0}>
            测试连接{selectedRowKeys.length > 0 ? ` (${selectedRowKeys.length})` : ''}
          </Button>
          <Popconfirm title={`确定删除选中的 ${selectedRowKeys.length} 条资产？`} onConfirm={handleBatchDelete} disabled={selectedRowKeys.length === 0}>
            <Button icon={<DeleteOutlined />} danger disabled={selectedRowKeys.length === 0}>
              批量删除{selectedRowKeys.length > 0 ? ` (${selectedRowKeys.length})` : ''}
            </Button>
          </Popconfirm>
        </Space>
        <Table
          rowKey="id"
          columns={columns}
          dataSource={data}
          loading={loading}
          rowSelection={{ selectedRowKeys, onChange: setSelectedRowKeys }}
          pagination={{ current: page, pageSize, total, showSizeChanger: true, showTotal: (t) => `共 ${t} 条`, onChange: (p, ps) => { setPage(p); setPageSize(ps); fetchData(p, ps) } }}
        />
      </Card>

      {/* 添加/编辑 Modal */}
      <Modal title={editing ? '编辑资产' : '添加资产'} open={modalOpen} onOk={handleModalOk} onCancel={() => setModalOpen(false)} destroyOnHidden width={600}>
        <Form form={form} layout="vertical" style={{ marginTop: 16 }}>
          <Form.Item name="hostname" label="主机名"><Input /></Form.Item>
          <Form.Item name="ip" label="IP 地址" rules={[{ required: !editing, message: '请输入 IP 地址' }]}><Input disabled={!!editing} /></Form.Item>
          <Form.Item name="port" label="SSH 端口"><Input type="number" /></Form.Item>
          <Form.Item name="type" label="类型" rules={[{ required: true }]}>
            <Select options={[{ value: 'physical_server', label: '物理服务器' }, { value: 'vm', label: '虚拟机' }, { value: 'network_device', label: '网络设备' }, { value: 'cabinet', label: '机柜' }]} />
          </Form.Item>
          <Form.Item name="os" label="操作系统"><Input /></Form.Item>
          {/* 状态由系统探活自动判定，不再手动选择 */}
          <Form.Item name="spec" label="规格"><Input placeholder="如 4C8G" /></Form.Item>
          <Form.Item name="environment" label="环境"><Select allowClear options={[{ value: 'prod', label: '生产' }, { value: 'test', label: '测试' }, { value: 'dev', label: '开发' }]} /></Form.Item>
          <Form.Item name="department" label="部门"><Input /></Form.Item>
          <Form.Item name="project_id" label="项目">
            <Select allowClear placeholder="选择项目" options={projects.map(p => ({ value: p.id, label: `${p.name} (${p.code})` }))} />
          </Form.Item>
          <Form.Item name="owner" label="负责人"><Input /></Form.Item>

          <Divider plain style={{ margin: '8px 0 16px', color: '#979BA5', fontSize: 13 }}>SSH 认证</Divider>

          <Form.Item name="ssh_user" label="SSH 用户"><Input /></Form.Item>
          <Form.Item name="ssh_password" label="SSH 密码">
            <Input.Password placeholder={editing ? '留空不修改' : ''} />
          </Form.Item>
          <Form.Item name="ssh_key_id" label="SSH 密钥">
            <Select
              allowClear
              placeholder="选择托管密钥（可选）"
              options={sshKeys.map(k => ({
                value: k.id,
                label: (
                  <Space size={4}>
                    <KeyOutlined style={{ color: '#3A84FF' }} />
                    <span>{k.name}</span>
                    <span style={{ color: '#979BA5', fontSize: 12 }}>({k.key_type})</span>
                  </Space>
                ),
              }))}
            />
          </Form.Item>
          <div style={{ fontSize: 12, color: '#979BA5', marginTop: -16, marginBottom: 16 }}>
            密码和密钥可同时配置，连接时优先使用密钥认证
          </div>
        </Form>
      </Modal>

      {/* 批量导入 Modal */}
      <Modal
        title="批量导入资产"
        open={importOpen}
        onCancel={() => setImportOpen(false)}
        width={800}
        footer={importResult ? [
          <Button key="close" onClick={() => setImportOpen(false)}>关闭</Button>,
        ] : [
          <Button key="cancel" onClick={() => setImportOpen(false)}>取消</Button>,
          <Button key="submit" type="primary" loading={importLoading} onClick={handleImportSubmit} disabled={importData.length === 0}>
            导入 ({importData.length} 条)
          </Button>,
        ]}
        destroyOnHidden
      >
        {importResult ? (
          <Result
            status={importResult.failed === 0 ? 'success' : 'warning'}
            title={`导入完成：成功 ${importResult.success} 条，失败 ${importResult.failed} 条`}
            subTitle={importResult.errors?.map((e, i) => (
              <div key={i} style={{ color: '#EA3636', fontSize: 13 }}>第 {e.index + 1} 行 ({e.ip}): {e.message}</div>
            ))}
          />
        ) : (
          <>
            <Alert
              type="info"
              showIcon
              message="支持 JSON 和 CSV 两种格式"
              description={
                <div>
                  <div>JSON 格式：<code>{`[{"ip":"1.2.3.4","type":"physical_server","hostname":"web-01","port":22}]`}</code></div>
                  <div style={{ marginTop: 4 }}>CSV 格式：首行为表头，必填字段 <code>ip</code>、<code>type</code></div>
                  <div style={{ marginTop: 8 }}>
                    <Button type="link" icon={<DownloadOutlined />} onClick={handleDownloadTemplate} style={{ padding: 0 }}>下载 CSV 模板</Button>
                  </div>
                </div>
              }
              style={{ marginBottom: 16 }}
            />
            <Space direction="vertical" style={{ width: '100%' }}>
              <Upload.Dragger
                accept=".json,.csv,.txt"
                showUploadList={false}
                beforeUpload={handleFileUpload}
                style={{ padding: '16px 0' }}
              >
                <p><UploadOutlined style={{ fontSize: 28, color: '#3A84FF' }} /></p>
                <p>点击或拖拽文件到此区域上传</p>
                <p style={{ color: '#979BA5', fontSize: 12 }}>支持 .json / .csv / .txt 文件</p>
              </Upload.Dragger>

              <Divider plain style={{ margin: '12px 0', color: '#979BA5' }}>或 手动粘贴数据</Divider>

              <Input.TextArea
                ref={textAreaRef as any}
                rows={5}
                placeholder={`粘贴 JSON 数组或 CSV 格式数据...\n\n示例 CSV:\nip,hostname,type,port,ssh_user,environment\n10.0.0.1,web-01,physical_server,22,root,prod`}
              />
              <Button onClick={handleParseText}>解析数据</Button>

              {importData.length > 0 && (
                <>
                  <Divider plain style={{ margin: '12px 0' }}>预览 ({importData.length} 条)</Divider>
                  <Table
                    rowKey={(_, i) => `${i}`}
                    columns={importPreviewColumns}
                    dataSource={importData}
                    size="small"
                    scroll={{ y: 200 }}
                    pagination={false}
                  />
                </>
              )}
            </Space>
          </>
        )}
      </Modal>

      {/* 测试连接结果 Modal */}
      <Modal
        title="SSH 连接测试"
        open={testOpen}
        onCancel={() => setTestOpen(false)}
        footer={<Button onClick={() => setTestOpen(false)}>关闭</Button>}
        width={750}
        destroyOnHidden
      >
        {testLoading ? (
          <div style={{ textAlign: 'center', padding: '40px 0' }}>
            <LoadingOutlined style={{ fontSize: 32, color: '#3A84FF' }} />
            <div style={{ marginTop: 12, color: '#979BA5' }}>正在测试 {selectedRowKeys.length} 台主机的 SSH 连接...</div>
          </div>
        ) : (
          <>
            {testResults.length > 0 && (
              <div style={{ marginBottom: 12 }}>
                <Space>
                  <Tag color="success">{testResults.filter(r => r.status === 'success').length} 成功</Tag>
                  <Tag color="error">{testResults.filter(r => r.status !== 'success').length} 失败</Tag>
                </Space>
              </div>
            )}
            <Table rowKey="asset_id" columns={testResultColumns} dataSource={testResults} size="middle" pagination={false} scroll={{ y: 400 }} />
          </>
        )}
      </Modal>

      {/* 资产初始化 Modal */}
      <AssetInitModal
        open={initOpen}
        assetIds={selectedRowKeys.map(Number)}
        onClose={() => setInitOpen(false)}
        onSuccess={() => fetchData()}
      />
    </>
  )
}

export default IDCAssetsPage
