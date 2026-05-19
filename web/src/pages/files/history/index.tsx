import { useState, useEffect } from 'react'
import { Card, Table, Button, Tag, Modal, Tooltip, message } from 'antd'
import { ReloadOutlined } from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import { getFileTasks, getFileTaskResults, type FileTaskItem, type FileTaskResultItem } from '../../../services/file-distribution'

const FileHistoryPage = () => {
  const [data, setData] = useState<FileTaskItem[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(false)
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(20)
  const [detailOpen, setDetailOpen] = useState(false)
  const [detailResults, setDetailResults] = useState<FileTaskResultItem[]>([])
  const [detailLoading, setDetailLoading] = useState(false)

  const fetchData = async (p = page, ps = pageSize) => {
    setLoading(true)
    try {
      const result = await getFileTasks({ page: p, page_size: ps })
      setData(result.list || [])
      setTotal(result.total)
    } catch { /* handled */ } finally { setLoading(false) }
  }

  useEffect(() => { fetchData() }, [])

  const showDetail = async (record: FileTaskItem) => {
    setDetailOpen(true)
    setDetailLoading(true)
    try {
      const results = await getFileTaskResults(record.id)
      setDetailResults(results || [])
    } catch {
      setDetailResults([])
      message.error('查询结果失败')
    } finally { setDetailLoading(false) }
  }

  const formatSize = (bytes: number) => {
    if (bytes < 1024) return `${bytes} B`
    if (bytes < 1048576) return `${(bytes / 1024).toFixed(1)} KB`
    return `${(bytes / 1048576).toFixed(1)} MB`
  }

  const statusColors: Record<string, string> = { completed: 'green', running: 'processing', failed: 'red', pending: 'gold' }

  const columns: ColumnsType<FileTaskItem> = [
    { title: 'ID', dataIndex: 'id', width: 60 },
    { title: '任务名称', dataIndex: 'name', width: 160, ellipsis: true },
    { title: '文件名', dataIndex: 'file_name', width: 160, ellipsis: true },
    { title: '大小', dataIndex: 'file_size', width: 80, render: (v) => formatSize(v) },
    { title: '远程路径', dataIndex: 'remote_path', width: 200, ellipsis: true },
    { title: '状态', dataIndex: 'status', width: 90, render: (v) => <Tag color={statusColors[v] || 'default'}>{v}</Tag> },
    { title: '成功/失败', width: 90, render: (_, r) => <span><span style={{ color: '#52c41a' }}>{r.success_count}</span> / <span style={{ color: r.fail_count > 0 ? '#ff4d4f' : undefined }}>{r.fail_count}</span></span> },
    { title: '创建时间', dataIndex: 'created_at', width: 170 },
    { title: '操作', width: 80, render: (_, record) => <Button type="link" size="small" onClick={() => showDetail(record)}>详情</Button> },
  ]

  const resultColumns: ColumnsType<FileTaskResultItem> = [
    { title: '主机名', dataIndex: 'hostname', width: 140, ellipsis: true },
    { title: 'IP', dataIndex: 'ip', width: 130 },
    { title: '状态', dataIndex: 'status', width: 80, render: (v) => <Tag color={statusColors[v] || 'default'}>{v}</Tag> },
    { title: '错误信息', dataIndex: 'error_msg', ellipsis: { showTitle: false }, render: (v) => v ? <Tooltip title={v} placement="topLeft"><span>{v}</span></Tooltip> : '-' },
    { title: '耗时(ms)', dataIndex: 'duration', width: 90 },
  ]

  return (
    <>
      <Card title="分发记录" style={{ border: '1px solid #E7E9EF' }} extra={<Button icon={<ReloadOutlined />} onClick={() => fetchData()}>刷新</Button>}>
        <Table rowKey="id" columns={columns} dataSource={data} loading={loading} pagination={{ current: page, pageSize, total, showSizeChanger: true, showTotal: (t) => `共 ${t} 条`, onChange: (p, ps) => { setPage(p); setPageSize(ps); fetchData(p, ps) } }} />
      </Card>

      <Modal title="分发详情" open={detailOpen} onCancel={() => setDetailOpen(false)} footer={null} width={700}>
        <Table rowKey="id" columns={resultColumns} dataSource={detailResults} loading={detailLoading} pagination={false} size="small" />
      </Modal>
    </>
  )
}

export default FileHistoryPage
