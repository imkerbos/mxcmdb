import { useState, useEffect } from 'react'
import { Card, Table, Button, Space, Input, Tag, Modal, Form, Switch, message, Popconfirm } from 'antd'
import { PlusOutlined, SearchOutlined, ReloadOutlined } from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import { getLinuxUsers, createLinuxUser, deleteLinuxUser, type LinuxUserItem } from '../../../services/linux-user'
import AssetSelect from '../../../components/AssetSelect'

const LinuxUsersPage = () => {
  const [data, setData] = useState<LinuxUserItem[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(false)
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(20)
  const [keyword, setKeyword] = useState('')
  const [createOpen, setCreateOpen] = useState(false)
  const [form] = Form.useForm()

  const fetchData = async (p = page, ps = pageSize) => {
    setLoading(true)
    try {
      const params: Record<string, unknown> = { page: p, page_size: ps }
      if (keyword) params.username = keyword
      const result = await getLinuxUsers(params)
      setData(result.list || [])
      setTotal(result.total)
    } catch { /* handled */ } finally { setLoading(false) }
  }

  useEffect(() => { fetchData() }, [])

  const handleCreate = async () => {
    try {
      const values = await form.validateFields()
      await createLinuxUser({
        username: values.username,
        asset_ids: values.asset_ids,
        shell: values.shell || '/bin/bash',
        sudo: values.sudo || false,
      })
      message.success('用户创建任务已提交')
      setCreateOpen(false)
      form.resetFields()
      fetchData()
    } catch { /* validation */ }
  }

  const handleDelete = async (record: LinuxUserItem) => {
    await deleteLinuxUser(record.username, [record.asset_id])
    message.success('用户删除任务已提交')
    fetchData()
  }

  const statusColors: Record<string, string> = { active: 'green', locked: 'orange', deleted: 'default', failed: 'red' }

  const columns: ColumnsType<LinuxUserItem> = [
    { title: 'ID', dataIndex: 'id', width: 60 },
    { title: '用户名', dataIndex: 'username', width: 120 },
    { title: '主机名', dataIndex: 'hostname', width: 140, ellipsis: true },
    { title: 'IP', dataIndex: 'ip', width: 130 },
    { title: 'UID', dataIndex: 'uid', width: 70 },
    { title: 'Shell', dataIndex: 'shell', width: 120 },
    { title: 'Sudo', dataIndex: 'sudo', width: 60, render: (v) => v ? <Tag color="blue">yes</Tag> : '-' },
    { title: '状态', dataIndex: 'status', width: 80, render: (v) => <Tag color={statusColors[v] || 'default'}>{v}</Tag> },
    {
      title: '操作', width: 80, fixed: 'right',
      render: (_, record) => (
        <Popconfirm title={`确定在 ${record.ip} 上删除用户 ${record.username}？`} onConfirm={() => handleDelete(record)}>
          <Button type="link" size="small" danger>删除</Button>
        </Popconfirm>
      ),
    },
  ]

  return (
    <>
      <Card title="Linux 用户管理" style={{ border: '1px solid #E7E9EF' }}>
        <Space style={{ marginBottom: 16 }} wrap>
          <Input placeholder="搜索用户名" prefix={<SearchOutlined />} value={keyword} onChange={(e) => setKeyword(e.target.value)} onPressEnter={() => { setPage(1); fetchData(1, pageSize) }} style={{ width: 200 }} />
          <Button type="primary" icon={<SearchOutlined />} onClick={() => { setPage(1); fetchData(1, pageSize) }}>搜索</Button>
          <Button icon={<ReloadOutlined />} onClick={() => { setKeyword(''); fetchData(1, pageSize) }}>重置</Button>
          <Button type="primary" icon={<PlusOutlined />} onClick={() => { form.resetFields(); setCreateOpen(true) }}>创建用户</Button>
        </Space>
        <Table rowKey="id" columns={columns} dataSource={data} loading={loading} scroll={{ x: 900 }} pagination={{ current: page, pageSize, total, showSizeChanger: true, showTotal: (t) => `共 ${t} 条`, onChange: (p, ps) => { setPage(p); setPageSize(ps); fetchData(p, ps) } }} />
      </Card>

      <Modal title="创建 Linux 用户" open={createOpen} onOk={handleCreate} onCancel={() => setCreateOpen(false)} destroyOnHidden>
        <Form form={form} layout="vertical" style={{ marginTop: 16 }}>
          <Form.Item name="username" label="用户名" rules={[{ required: true, message: '请输入用户名' }]}>
            <Input placeholder="如 deploy" />
          </Form.Item>
          <Form.Item name="asset_ids" label="目标资产" rules={[{ required: true, message: '请选择资产' }]}>
            <AssetSelect style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="shell" label="Shell">
            <Input placeholder="/bin/bash" />
          </Form.Item>
          <Form.Item name="sudo" label="Sudo 权限" valuePropName="checked">
            <Switch />
          </Form.Item>
        </Form>
      </Modal>
    </>
  )
}

export default LinuxUsersPage
