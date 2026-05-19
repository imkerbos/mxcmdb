import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { Card, Table, Button, Space, Input, Select, Tag, Modal, Divider } from 'antd'
import { SearchOutlined, ReloadOutlined, ApiOutlined, CheckCircleOutlined, CloseCircleOutlined, LoadingOutlined, SettingOutlined } from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import { getAssets, testConnection, type AssetItem, type TestConnectionResult } from '../../../services/asset'
import AssetInitModal from '../../../components/AssetInitModal'
import { message } from 'antd'

const statusColors: Record<string, string> = { running: 'green', stopped: 'red', unknown: 'default', terminated: 'default' }
const envColors: Record<string, string> = { prod: 'red', test: 'orange', dev: 'blue' }

const CloudAssetsPage = () => {
  const navigate = useNavigate()
  const [data, setData] = useState<AssetItem[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(false)
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(20)
  const [keyword, setKeyword] = useState('')
  const [statusFilter, setStatusFilter] = useState('')
  const [selectedRowKeys, setSelectedRowKeys] = useState<React.Key[]>([])

  // 测试连接
  const [testOpen, setTestOpen] = useState(false)
  const [testLoading, setTestLoading] = useState(false)
  const [testResults, setTestResults] = useState<TestConnectionResult[]>([])

  // 初始化
  const [initOpen, setInitOpen] = useState(false)

  const fetchData = async (p = page, ps = pageSize) => {
    setLoading(true)
    try {
      const params: Record<string, unknown> = { source: 'cloud_sync', page: p, page_size: ps }
      if (keyword) params.keyword = keyword
      if (statusFilter) params.status = statusFilter
      const result = await getAssets(params as any)
      setData(result.list || [])
      setTotal(result.total)
    } catch { /* handled */ } finally { setLoading(false) }
  }

  useEffect(() => { fetchData() }, [])

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

  const columns: ColumnsType<AssetItem> = [
    { title: '主机名', dataIndex: 'hostname', width: 150, ellipsis: true },
    { title: 'IP', dataIndex: 'ip', width: 140 },
    { title: '实例 ID', dataIndex: 'instance_id', width: 200, ellipsis: true },
    { title: '类型', dataIndex: 'type', width: 80, render: (v) => <Tag>{v}</Tag> },
    { title: '规格', dataIndex: 'spec', width: 140, ellipsis: true },
    { title: '区域', dataIndex: 'region', width: 120 },
    { title: '状态', dataIndex: 'status', width: 80, render: (v) => <Tag color={statusColors[v] || 'default'}>{v}</Tag> },
    { title: '环境', dataIndex: 'environment', width: 80, render: (v) => v ? <Tag color={envColors[v] || 'default'}>{v}</Tag> : '-' },
    { title: '项目', dataIndex: 'project_name', width: 100, render: (v) => v || '-' },
    { title: '操作系统', dataIndex: 'os', width: 120, ellipsis: true },
    { title: '最后探测', dataIndex: 'probe_last_at', width: 170, render: (v) => v || '-' },
    {
      title: '操作', width: 80, fixed: 'right',
      render: (_, record) => (
        <Button type="link" size="small" onClick={() => navigate(`/assets/${record.id}/detail`)}>详情</Button>
      ),
    },
  ]

  const authMethodLabels: Record<string, string> = {
    password: '密码', key: '密钥', 'password+key': '密码+密钥', none: '无',
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
      <Card title="云资源" style={{ border: '1px solid #E7E9EF' }}>
        <Space style={{ marginBottom: 16 }} wrap>
          <Input placeholder="搜索主机名/IP" prefix={<SearchOutlined />} value={keyword} onChange={(e) => setKeyword(e.target.value)} onPressEnter={() => { setPage(1); fetchData(1, pageSize) }} style={{ width: 200 }} />
          <Select value={statusFilter} onChange={setStatusFilter} style={{ width: 120 }} options={[{ value: '', label: '全部状态' }, { value: 'running', label: '运行中' }, { value: 'stopped', label: '已停止' }]} />
          <Button type="primary" icon={<SearchOutlined />} onClick={() => { setPage(1); fetchData(1, pageSize) }}>搜索</Button>
          <Button icon={<ReloadOutlined />} onClick={() => { setKeyword(''); setStatusFilter(''); fetchData(1, pageSize) }}>重置</Button>
          <Divider type="vertical" />
          <Button type="primary" ghost icon={<SettingOutlined />} onClick={() => setInitOpen(true)} disabled={selectedRowKeys.length === 0}>
            初始化{selectedRowKeys.length > 0 ? ` (${selectedRowKeys.length})` : ''}
          </Button>
          <Button icon={<ApiOutlined />} onClick={handleTestConnection} disabled={selectedRowKeys.length === 0}>
            测试连接{selectedRowKeys.length > 0 ? ` (${selectedRowKeys.length})` : ''}
          </Button>
        </Space>
        <Table
          rowKey="id"
          columns={columns}
          dataSource={data}
          loading={loading}
          scroll={{ x: 1400 }}
          rowSelection={{ selectedRowKeys, onChange: setSelectedRowKeys }}
          pagination={{ current: page, pageSize, total, showSizeChanger: true, showTotal: (t) => `共 ${t} 条`, onChange: (p, ps) => { setPage(p); setPageSize(ps); fetchData(p, ps) } }}
        />
      </Card>

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

export default CloudAssetsPage
