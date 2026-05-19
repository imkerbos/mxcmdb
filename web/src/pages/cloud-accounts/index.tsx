import { useState, useEffect } from 'react'
import { Card, Table, Button, Space, Input, Tag, Modal, Form, Select, message, Popconfirm } from 'antd'
import { PlusOutlined, SearchOutlined, ReloadOutlined } from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import {
  getCloudAccounts, createCloudAccount, updateCloudAccount, deleteCloudAccount,
  type CloudAccount, type CreateCloudAccountParams,
} from '../../services/cloud-account'

const CloudAccountsPage = () => {
  const [data, setData] = useState<CloudAccount[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(false)
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(20)
  const [keyword, setKeyword] = useState('')
  const [modalOpen, setModalOpen] = useState(false)
  const [editing, setEditing] = useState<CloudAccount | null>(null)
  const [form] = Form.useForm()

  const fetchData = async (p = page, ps = pageSize) => {
    setLoading(true)
    try {
      const params: Record<string, unknown> = { page: p, page_size: ps }
      if (keyword) params.keyword = keyword
      const result = await getCloudAccounts(params)
      setData(result.list || [])
      setTotal(result.total)
    } catch { /* handled */ } finally { setLoading(false) }
  }

  useEffect(() => { fetchData() }, [])

  const handleCreate = () => { setEditing(null); form.resetFields(); form.setFieldsValue({ provider: 'aliyun' }); setModalOpen(true) }

  const handleEdit = (record: CloudAccount) => {
    setEditing(record)
    form.setFieldsValue({ name: record.name, provider: record.provider, access_key_id: record.access_key_id, region: record.region })
    setModalOpen(true)
  }

  const handleModalOk = async () => {
    try {
      const values = await form.validateFields()
      if (editing) {
        await updateCloudAccount(editing.id, values)
        message.success('更新成功')
      } else {
        await createCloudAccount(values as CreateCloudAccountParams)
        message.success('创建成功')
      }
      setModalOpen(false)
      fetchData()
    } catch { /* validation */ }
  }

  const syncStatusColors: Record<string, string> = { success: 'green', failed: 'red', syncing: 'processing' }

  const columns: ColumnsType<CloudAccount> = [
    { title: 'ID', dataIndex: 'id', width: 60 },
    { title: '名称', dataIndex: 'name', width: 150 },
    { title: '云商', dataIndex: 'provider', width: 100, render: (v) => <Tag>{v === 'aliyun' ? '阿里云' : v}</Tag> },
    { title: 'AccessKey ID', dataIndex: 'access_key_id', width: 200, ellipsis: true },
    { title: '区域', dataIndex: 'region', width: 120 },
    { title: '状态', dataIndex: 'status', width: 80, render: (v) => <Tag color={v === 1 ? 'green' : 'default'}>{v === 1 ? '启用' : '禁用'}</Tag> },
    { title: '最后同步', dataIndex: 'last_sync_at', width: 170, render: (v) => v || '-' },
    { title: '同步状态', dataIndex: 'last_sync_status', width: 100, render: (v) => v ? <Tag color={syncStatusColors[v] || 'default'}>{v}</Tag> : '-' },
    {
      title: '操作', width: 150, fixed: 'right',
      render: (_, record) => (
        <Space>
          <Button type="link" size="small" onClick={() => handleEdit(record)}>编辑</Button>
          <Popconfirm title="确定删除？" onConfirm={async () => { await deleteCloudAccount(record.id); message.success('已删除'); fetchData() }}>
            <Button type="link" size="small" danger>删除</Button>
          </Popconfirm>
        </Space>
      ),
    },
  ]

  return (
    <>
      <Card title="云账号托管" style={{ border: '1px solid #E7E9EF' }}>
        <Space style={{ marginBottom: 16 }} wrap>
          <Input placeholder="搜索名称/AccessKey" prefix={<SearchOutlined />} value={keyword} onChange={(e) => setKeyword(e.target.value)} onPressEnter={() => { setPage(1); fetchData(1, pageSize) }} style={{ width: 220 }} />
          <Button type="primary" icon={<SearchOutlined />} onClick={() => { setPage(1); fetchData(1, pageSize) }}>搜索</Button>
          <Button icon={<ReloadOutlined />} onClick={() => { setKeyword(''); fetchData(1, pageSize) }}>重置</Button>
          <Button type="primary" icon={<PlusOutlined />} onClick={handleCreate}>添加云账号</Button>
        </Space>
        <Table rowKey="id" columns={columns} dataSource={data} loading={loading} scroll={{ x: 1200 }} pagination={{ current: page, pageSize, total, showSizeChanger: true, showTotal: (t) => `共 ${t} 条`, onChange: (p, ps) => { setPage(p); setPageSize(ps); fetchData(p, ps) } }} />
      </Card>

      <Modal title={editing ? '编辑云账号' : '添加云账号'} open={modalOpen} onOk={handleModalOk} onCancel={() => setModalOpen(false)} destroyOnHidden>
        <Form form={form} layout="vertical" style={{ marginTop: 16 }}>
          <Form.Item name="name" label="名称" rules={[{ required: true, message: '请输入名称' }]}><Input /></Form.Item>
          <Form.Item name="provider" label="云服务商" rules={[{ required: true }]}><Select options={[{ value: 'aliyun', label: '阿里云' }]} /></Form.Item>
          <Form.Item name="access_key_id" label="AccessKey ID" rules={[{ required: !editing, message: '请输入 AccessKey ID' }]}><Input /></Form.Item>
          <Form.Item name="access_key_secret" label="AccessKey Secret" rules={[{ required: !editing, message: '请输入 AccessKey Secret' }]}><Input.Password placeholder={editing ? '留空不修改' : ''} /></Form.Item>
          <Form.Item name="region" label="默认区域"><Input placeholder="如 cn-hangzhou" /></Form.Item>
        </Form>
      </Modal>
    </>
  )
}

export default CloudAccountsPage
