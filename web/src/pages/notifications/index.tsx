import { useState, useEffect } from 'react'
import { Card, Table, Tag, Button, Space, Select, message, Popconfirm } from 'antd'
import {
  CheckOutlined,
  InfoCircleOutlined, CheckCircleOutlined, WarningOutlined, CloseCircleOutlined,
} from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import {
  getNotifications, markRead, markAllRead, deleteNotification,
  type NotificationItem,
} from '../../services/notification'

const typeIcons: Record<string, React.ReactNode> = {
  info: <InfoCircleOutlined style={{ color: '#3A84FF' }} />,
  success: <CheckCircleOutlined style={{ color: '#2DCB56' }} />,
  warning: <WarningOutlined style={{ color: '#FF9C01' }} />,
  error: <CloseCircleOutlined style={{ color: '#EA3636' }} />,
}

const sourceLabels: Record<string, string> = {
  probe: '探针', task: '任务', ssh_key: 'SSH 密钥', terminal: '终端', system: '系统',
}

const NotificationsPage = () => {
  const [data, setData] = useState<NotificationItem[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(false)
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(20)
  const [readFilter, setReadFilter] = useState<string>('')
  const [selectedRowKeys, setSelectedRowKeys] = useState<React.Key[]>([])

  const fetchData = async (p = page, ps = pageSize) => {
    setLoading(true)
    try {
      const params: Record<string, unknown> = { page: p, page_size: ps }
      if (readFilter) params.read = readFilter
      const result = await getNotifications(params as { read?: string; page?: number; page_size?: number })
      setData(result.list || [])
      setTotal(result.total)
    } catch { /* handled */ } finally { setLoading(false) }
  }

  useEffect(() => { fetchData() }, [])

  const handleMarkAllRead = async () => {
    await markAllRead()
    message.success('已全部标记为已读')
    fetchData()
  }

  const handleDelete = async (id: number) => {
    await deleteNotification(id)
    message.success('已删除')
    fetchData()
  }

  const handleBatchMarkRead = async () => {
    for (const id of selectedRowKeys) {
      await markRead(Number(id))
    }
    message.success(`已标记 ${selectedRowKeys.length} 条为已读`)
    setSelectedRowKeys([])
    fetchData()
  }

  const columns: ColumnsType<NotificationItem> = [
    {
      title: '类型', dataIndex: 'type', width: 46, align: 'center',
      render: (v) => typeIcons[v] || typeIcons.info,
    },
    {
      title: '标题', dataIndex: 'title', width: 180, ellipsis: true,
      render: (v, record) => (
        <span style={{ fontWeight: record.read ? 400 : 600 }}>{v}</span>
      ),
    },
    { title: '内容', dataIndex: 'content', ellipsis: true, render: (v) => v || '-' },
    {
      title: '来源', dataIndex: 'source', width: 70,
      render: (v) => sourceLabels[v] || v || '-',
    },
    {
      title: '状态', dataIndex: 'read', width: 60,
      render: (v) => v ? <Tag>已读</Tag> : <Tag color="blue">未读</Tag>,
    },
    { title: '时间', dataIndex: 'created_at', width: 150 },
    {
      title: '操作', width: 120,
      render: (_, record) => (
        <Space size={0}>
          {!record.read && (
            <Button type="link" size="small" onClick={() => { markRead(record.id).then(() => fetchData()) }}>
              已读
            </Button>
          )}
          <Popconfirm title="确定删除？" onConfirm={() => handleDelete(record.id)}>
            <Button type="link" size="small" danger>删除</Button>
          </Popconfirm>
        </Space>
      ),
    },
  ]

  return (
    <Card title="通知中心" style={{ border: '1px solid #E7E9EF' }}>
      <Space style={{ marginBottom: 16 }} wrap>
        <Select
          value={readFilter}
          onChange={v => { setReadFilter(v); setPage(1); fetchData(1, pageSize) }}
          style={{ width: 120 }}
          options={[
            { value: '', label: '全部' },
            { value: 'false', label: '未读' },
            { value: 'true', label: '已读' },
          ]}
        />
        <Button icon={<CheckOutlined />} onClick={handleMarkAllRead}>全部已读</Button>
        {selectedRowKeys.length > 0 && (
          <Button type="primary" ghost icon={<CheckOutlined />} onClick={handleBatchMarkRead}>
            标记已读 ({selectedRowKeys.length})
          </Button>
        )}
      </Space>
      <Table
        rowKey="id"
        columns={columns}
        dataSource={data}
        loading={loading}
        rowSelection={{
          selectedRowKeys,
          onChange: setSelectedRowKeys,
          getCheckboxProps: (record) => ({ disabled: record.read }),
        }}
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

export default NotificationsPage
