import { useState } from 'react'
import { Card, Table, Input, DatePicker, Select, Space, Tag, Button } from 'antd'
import { SearchOutlined, ReloadOutlined } from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import dayjs from 'dayjs'
import { getAuditLogs, type AuditLog } from '../../../services/audit'

const moduleOptions = [
  { value: '', label: '全部模块' },
  { value: 'user', label: '用户管理' },
  { value: 'asset', label: '资产管理' },
  { value: 'cloud', label: '云账号' },
  { value: 'ssh', label: 'SSH Key' },
  { value: 'probe', label: '资产探针' },
  { value: 'file', label: '文件分发' },
  { value: 'ipam', label: 'IPAM' },
  { value: 'settings', label: '系统设置' },
]

const actionColors: Record<string, string> = {
  create: 'blue',
  update: 'orange',
  delete: 'red',
  execute: 'purple',
  deploy: 'cyan',
  sync: 'green',
  login: 'geekblue',
  query: 'default',
}

const OperationAuditPage = () => {
  const [data, setData] = useState<AuditLog[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(false)
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(20)
  const [username, setUsername] = useState('')
  const [module, setModule] = useState('')
  const [dateRange, setDateRange] = useState<[dayjs.Dayjs | null, dayjs.Dayjs | null]>([null, null])

  const fetchData = async (p = page, ps = pageSize) => {
    setLoading(true)
    try {
      const params: Record<string, unknown> = { page: p, page_size: ps }
      if (username) params.username = username
      if (module) params.module = module
      if (dateRange[0]) params.start_date = dateRange[0].format('YYYY-MM-DD')
      if (dateRange[1]) params.end_date = dateRange[1].format('YYYY-MM-DD')

      const result = await getAuditLogs(params as any)
      setData(result.list || [])
      setTotal(result.total)
    } catch {
      // handled by interceptor
    } finally {
      setLoading(false)
    }
  }

  const handleSearch = () => {
    setPage(1)
    fetchData(1, pageSize)
  }

  useState(() => {
    fetchData()
  })

  const columns: ColumnsType<AuditLog> = [
    { title: '时间', dataIndex: 'created_at', width: 180, render: (v) => dayjs(v).format('YYYY-MM-DD HH:mm:ss') },
    { title: '用户', dataIndex: 'username', width: 100 },
    { title: '模块', dataIndex: 'module', width: 100 },
    {
      title: '操作',
      dataIndex: 'action',
      width: 100,
      render: (action) => <Tag color={actionColors[action] || 'default'}>{action}</Tag>,
    },
    { title: '方法', dataIndex: 'method', width: 80 },
    { title: '路径', dataIndex: 'path', width: 250, ellipsis: true },
    { title: 'IP', dataIndex: 'client_ip', width: 140 },
    {
      title: '状态',
      dataIndex: 'response_code',
      width: 80,
      render: (code) => <Tag color={code === 200 ? 'success' : 'error'}>{code}</Tag>,
    },
    { title: '耗时(ms)', dataIndex: 'duration', width: 100 },
  ]

  return (
    <Card title="操作审计" style={{ border: '1px solid #E7E9EF' }}>
      <Space style={{ marginBottom: 16 }} wrap>
        <Input
          placeholder="用户名"
          prefix={<SearchOutlined />}
          value={username}
          onChange={(e) => setUsername(e.target.value)}
          onPressEnter={handleSearch}
          style={{ width: 160 }}
        />
        <Select
          value={module}
          onChange={setModule}
          options={moduleOptions}
          style={{ width: 140 }}
        />
        <DatePicker.RangePicker
          value={dateRange}
          onChange={(dates) => setDateRange(dates as [dayjs.Dayjs, dayjs.Dayjs])}
        />
        <Button type="primary" icon={<SearchOutlined />} onClick={handleSearch}>
          搜索
        </Button>
        <Button icon={<ReloadOutlined />} onClick={() => { setUsername(''); setModule(''); setDateRange([null, null]); fetchData(1, pageSize) }}>
          重置
        </Button>
      </Space>

      <Table
        rowKey="id"
        columns={columns}
        dataSource={data}
        loading={loading}
        pagination={{
          current: page,
          pageSize,
          total,
          showSizeChanger: true,
          showTotal: (t) => `共 ${t} 条`,
          onChange: (p, ps) => { setPage(p); setPageSize(ps); fetchData(p, ps) },
        }}
      />
    </Card>
  )
}

export default OperationAuditPage
