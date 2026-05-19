import { useState, useEffect } from 'react'
import { Card, Table, Button, Space, Tag, Select, Modal, Collapse, message } from 'antd'
import { SearchOutlined, ReloadOutlined } from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import { getTasks, getTaskResults, type TaskItem, type TaskResultItem } from '../../../services/task'

const TaskHistoryPage = () => {
  const [data, setData] = useState<TaskItem[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(false)
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(20)
  const [status, setStatus] = useState<string>('')
  const [detailOpen, setDetailOpen] = useState(false)
  const [detailResults, setDetailResults] = useState<TaskResultItem[]>([])
  const [detailLoading, setDetailLoading] = useState(false)
  const [detailTask, setDetailTask] = useState<TaskItem | null>(null)

  const fetchData = async (p = page, ps = pageSize) => {
    setLoading(true)
    try {
      const params: Record<string, unknown> = { page: p, page_size: ps }
      if (status) params.status = status
      const result = await getTasks(params)
      setData(result.list || [])
      setTotal(result.total)
    } catch { /* handled */ } finally { setLoading(false) }
  }

  useEffect(() => { fetchData() }, [])

  const showDetail = async (record: TaskItem) => {
    setDetailTask(record)
    setDetailOpen(true)
    setDetailLoading(true)
    try {
      const results = await getTaskResults(record.id)
      setDetailResults(results || [])
    } catch {
      setDetailResults([])
      message.error('查询任务结果失败')
    } finally { setDetailLoading(false) }
  }

  const statusColors: Record<string, string> = { completed: 'green', running: 'processing', failed: 'red', pending: 'gold' }

  const columns: ColumnsType<TaskItem> = [
    { title: 'ID', dataIndex: 'id', width: 60 },
    { title: '任务名称', dataIndex: 'name', width: 180, ellipsis: true },
    { title: '命令', dataIndex: 'command', width: 200, ellipsis: true },
    {
      title: '状态', dataIndex: 'status', width: 100,
      render: (v) => <Tag color={statusColors[v] || 'default'}>{v}</Tag>,
    },
    { title: '总数', dataIndex: 'total_count', width: 60 },
    { title: '成功', dataIndex: 'success_count', width: 60, render: (v) => <span style={{ color: '#52c41a' }}>{v}</span> },
    { title: '失败', dataIndex: 'fail_count', width: 60, render: (v) => <span style={{ color: v > 0 ? '#ff4d4f' : undefined }}>{v}</span> },
    { title: '开始时间', dataIndex: 'started_at', width: 170, render: (v) => v || '-' },
    { title: '完成时间', dataIndex: 'finished_at', width: 170, render: (v) => v || '-' },
    {
      title: '操作', width: 80, fixed: 'right',
      render: (_, record) => (
        <Button type="link" size="small" onClick={() => showDetail(record)}>详情</Button>
      ),
    },
  ]

  const resultColumns: ColumnsType<TaskResultItem> = [
    { title: '主机名', dataIndex: 'hostname', width: 140, ellipsis: true },
    { title: 'IP', dataIndex: 'ip', width: 130 },
    {
      title: '状态', dataIndex: 'status', width: 80,
      render: (v) => <Tag color={statusColors[v] || 'default'}>{v}</Tag>,
    },
    { title: '退出码', dataIndex: 'exit_code', width: 80 },
    { title: '耗时(ms)', dataIndex: 'duration', width: 90 },
  ]

  return (
    <>
      <Card title="执行历史" style={{ border: '1px solid #E7E9EF' }}>
        <Space style={{ marginBottom: 16 }} wrap>
          <Select
            placeholder="状态筛选"
            allowClear
            value={status || undefined}
            onChange={(v) => setStatus(v || '')}
            style={{ width: 140 }}
            options={[
              { value: 'running', label: '执行中' },
              { value: 'completed', label: '已完成' },
              { value: 'failed', label: '失败' },
            ]}
          />
          <Button type="primary" icon={<SearchOutlined />} onClick={() => { setPage(1); fetchData(1, pageSize) }}>搜索</Button>
          <Button icon={<ReloadOutlined />} onClick={() => { setStatus(''); fetchData(1, pageSize) }}>重置</Button>
        </Space>
        <Table rowKey="id" columns={columns} dataSource={data} loading={loading} scroll={{ x: 1300 }} pagination={{ current: page, pageSize, total, showSizeChanger: true, showTotal: (t) => `共 ${t} 条`, onChange: (p, ps) => { setPage(p); setPageSize(ps); fetchData(p, ps) } }} />
      </Card>

      <Modal title={`任务详情 - ${detailTask?.name || ''}`} open={detailOpen} onCancel={() => setDetailOpen(false)} footer={null} width={800}>
        {detailTask && (
          <Space direction="vertical" style={{ width: '100%', marginBottom: 16 }}>
            <div><strong>命令：</strong><code>{detailTask.command}</code></div>
            <div><strong>状态：</strong><Tag color={statusColors[detailTask.status]}>{detailTask.status}</Tag></div>
          </Space>
        )}
        <Table
          rowKey="id"
          columns={resultColumns}
          dataSource={detailResults}
          loading={detailLoading}
          pagination={false}
          size="small"
          expandable={{
            expandedRowRender: (record) => (
              <Collapse
                size="small"
                items={[
                  ...(record.stdout ? [{ key: 'stdout', label: 'Stdout', children: <pre style={{ maxHeight: 200, overflow: 'auto', fontSize: 12, margin: 0 }}>{record.stdout}</pre> }] : []),
                  ...(record.stderr ? [{ key: 'stderr', label: 'Stderr', children: <pre style={{ maxHeight: 200, overflow: 'auto', fontSize: 12, margin: 0, color: '#ff4d4f' }}>{record.stderr}</pre> }] : []),
                ]}
                defaultActiveKey={['stdout']}
              />
            ),
          }}
        />
      </Modal>
    </>
  )
}

export default TaskHistoryPage
