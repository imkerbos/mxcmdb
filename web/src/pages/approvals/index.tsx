import { useState, useEffect } from 'react'
import { Card, Table, Tag, Button, Space, Select, Modal, Input, message, Tabs, Descriptions, Popconfirm } from 'antd'
import {
  CheckOutlined, CloseOutlined,
} from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import {
  getApprovals, getMyApprovals, reviewApproval, cancelApproval,
  type ApprovalItem,
} from '../../services/approval'

const typeLabels: Record<string, string> = {
  batch_delete: '批量删除资产',
  ssh_key_revoke: 'SSH 密钥撤销',
  departure_cleanup: '离职清理',
  ssh_key_rotate: 'SSH 密钥轮换',
}

const statusColors: Record<string, string> = {
  pending: 'gold', approved: 'green', rejected: 'red', cancelled: 'default',
}

const statusLabels: Record<string, string> = {
  pending: '待审批', approved: '已通过', rejected: '已驳回', cancelled: '已取消',
}

const ApprovalsPage = () => {
  const [activeTab, setActiveTab] = useState('all')
  const [data, setData] = useState<ApprovalItem[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(false)
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(20)
  const [statusFilter, setStatusFilter] = useState('')

  // 审批弹窗
  const [reviewOpen, setReviewOpen] = useState(false)
  const [reviewItem, setReviewItem] = useState<ApprovalItem | null>(null)
  const [reviewAction, setReviewAction] = useState<'approve' | 'reject'>('approve')
  const [reviewNote, setReviewNote] = useState('')
  const [reviewLoading, setReviewLoading] = useState(false)

  // 详情弹窗
  const [detailOpen, setDetailOpen] = useState(false)
  const [detailItem, setDetailItem] = useState<ApprovalItem | null>(null)

  const fetchData = async (p = page, ps = pageSize) => {
    setLoading(true)
    try {
      const params: Record<string, unknown> = { page: p, page_size: ps }
      if (statusFilter) params.status = statusFilter

      const result = activeTab === 'mine'
        ? await getMyApprovals({ page: p, page_size: ps })
        : await getApprovals(params as { status?: string; page?: number; page_size?: number })

      setData(result.list || [])
      setTotal(result.total)
    } catch { /* handled */ } finally { setLoading(false) }
  }

  useEffect(() => { fetchData(1, pageSize) }, [activeTab])

  const handleReview = async () => {
    if (!reviewItem) return
    setReviewLoading(true)
    try {
      await reviewApproval(reviewItem.id, { action: reviewAction, note: reviewNote })
      message.success(reviewAction === 'approve' ? '已通过' : '已驳回')
      setReviewOpen(false)
      setReviewNote('')
      fetchData()
    } catch { /* handled */ } finally { setReviewLoading(false) }
  }

  const handleCancel = async (id: number) => {
    await cancelApproval(id)
    message.success('已取消')
    fetchData()
  }

  const openReview = (item: ApprovalItem, action: 'approve' | 'reject') => {
    setReviewItem(item)
    setReviewAction(action)
    setReviewNote('')
    setReviewOpen(true)
  }

  const parsePayload = (payload: string) => {
    try {
      return JSON.stringify(JSON.parse(payload), null, 2)
    } catch {
      return payload
    }
  }

  const columns: ColumnsType<ApprovalItem> = [
    { title: 'ID', dataIndex: 'id', width: 60 },
    { title: '标题', dataIndex: 'title', ellipsis: true },
    {
      title: '类型', dataIndex: 'type', width: 140,
      render: (v) => <Tag>{typeLabels[v] || v}</Tag>,
    },
    {
      title: '状态', dataIndex: 'status', width: 100,
      render: (v) => <Tag color={statusColors[v] || 'default'}>{statusLabels[v] || v}</Tag>,
    },
    { title: '发起人', dataIndex: 'requester_name', width: 100 },
    { title: '审批人', dataIndex: 'reviewer_name', width: 100, render: (v) => v || '-' },
    { title: '审批时间', dataIndex: 'reviewed_at', width: 170, render: (v) => v || '-' },
    { title: '创建时间', dataIndex: 'created_at', width: 170 },
    {
      title: '操作', width: 220, fixed: 'right',
      render: (_, record) => (
        <Space>
          <Button type="link" size="small" onClick={() => { setDetailItem(record); setDetailOpen(true) }}>详情</Button>
          {record.status === 'pending' && (
            <>
              <Button type="link" size="small" icon={<CheckOutlined />} style={{ color: '#2DCB56' }} onClick={() => openReview(record, 'approve')}>通过</Button>
              <Button type="link" size="small" icon={<CloseOutlined />} danger onClick={() => openReview(record, 'reject')}>驳回</Button>
              {activeTab === 'mine' && (
                <Popconfirm title="确定取消？" onConfirm={() => handleCancel(record.id)}>
                  <Button type="link" size="small">取消</Button>
                </Popconfirm>
              )}
            </>
          )}
        </Space>
      ),
    },
  ]

  return (
    <>
      <Card title="审批管理" style={{ border: '1px solid #E7E9EF' }}>
        <Tabs
          activeKey={activeTab}
          onChange={k => { setActiveTab(k); setPage(1) }}
          items={[
            { key: 'all', label: '全部审批' },
            { key: 'mine', label: '我发起的' },
          ]}
        />
        <Space style={{ marginBottom: 16 }}>
          {activeTab === 'all' && (
            <Select
              value={statusFilter}
              onChange={v => { setStatusFilter(v); setPage(1); fetchData(1, pageSize) }}
              style={{ width: 120 }}
              options={[
                { value: '', label: '全部状态' },
                { value: 'pending', label: '待审批' },
                { value: 'approved', label: '已通过' },
                { value: 'rejected', label: '已驳回' },
              ]}
            />
          )}
        </Space>
        <Table
          rowKey="id"
          columns={columns}
          dataSource={data}
          loading={loading}
          scroll={{ x: 1200 }}
          pagination={{
            current: page, pageSize, total,
            showSizeChanger: true,
            showTotal: (t) => `共 ${t} 条`,
            onChange: (p, ps) => { setPage(p); setPageSize(ps); fetchData(p, ps) },
          }}
        />
      </Card>

      {/* 审批弹窗 */}
      <Modal
        title={reviewAction === 'approve' ? '审批通过' : '驳回审批'}
        open={reviewOpen}
        onOk={handleReview}
        onCancel={() => setReviewOpen(false)}
        confirmLoading={reviewLoading}
        okType={reviewAction === 'approve' ? 'primary' : 'default'}
        okButtonProps={reviewAction === 'reject' ? { danger: true } : undefined}
      >
        {reviewItem && (
          <div style={{ marginBottom: 16 }}>
            <Descriptions column={1} size="small">
              <Descriptions.Item label="标题">{reviewItem.title}</Descriptions.Item>
              <Descriptions.Item label="类型">{typeLabels[reviewItem.type] || reviewItem.type}</Descriptions.Item>
              <Descriptions.Item label="发起人">{reviewItem.requester_name}</Descriptions.Item>
            </Descriptions>
          </div>
        )}
        <Input.TextArea
          rows={3}
          value={reviewNote}
          onChange={e => setReviewNote(e.target.value)}
          placeholder="审批备注（可选）"
        />
      </Modal>

      {/* 详情弹窗 */}
      <Modal
        title="审批详情"
        open={detailOpen}
        onCancel={() => setDetailOpen(false)}
        footer={null}
        width={640}
      >
        {detailItem && (
          <>
            <Descriptions column={2} size="small" bordered style={{ marginBottom: 16 }}>
              <Descriptions.Item label="标题" span={2}>{detailItem.title}</Descriptions.Item>
              <Descriptions.Item label="类型">{typeLabels[detailItem.type] || detailItem.type}</Descriptions.Item>
              <Descriptions.Item label="状态">
                <Tag color={statusColors[detailItem.status]}>{statusLabels[detailItem.status] || detailItem.status}</Tag>
              </Descriptions.Item>
              <Descriptions.Item label="发起人">{detailItem.requester_name}</Descriptions.Item>
              <Descriptions.Item label="创建时间">{detailItem.created_at}</Descriptions.Item>
              {detailItem.reviewer_name && (
                <>
                  <Descriptions.Item label="审批人">{detailItem.reviewer_name}</Descriptions.Item>
                  <Descriptions.Item label="审批时间">{detailItem.reviewed_at}</Descriptions.Item>
                </>
              )}
              {detailItem.review_note && (
                <Descriptions.Item label="审批备注" span={2}>{detailItem.review_note}</Descriptions.Item>
              )}
              {detailItem.description && (
                <Descriptions.Item label="说明" span={2}>{detailItem.description}</Descriptions.Item>
              )}
            </Descriptions>
            <Card size="small" title="操作参数" style={{ border: '1px solid #E7E9EF' }}>
              <pre style={{ fontSize: 12, color: '#63656E', margin: 0, maxHeight: 300, overflow: 'auto', whiteSpace: 'pre-wrap' }}>
                {parsePayload(detailItem.payload)}
              </pre>
            </Card>
          </>
        )}
      </Modal>
    </>
  )
}

export default ApprovalsPage
