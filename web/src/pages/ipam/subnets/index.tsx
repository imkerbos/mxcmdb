import { useState, useEffect } from 'react'
import { Card, Table, Button, Space, Modal, Form, Input, InputNumber, Progress, message, Popconfirm } from 'antd'
import { PlusOutlined, ReloadOutlined } from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import { getSubnets, createSubnet, deleteSubnet, type SubnetItem } from '../../../services/ipam'
import { useNavigate } from 'react-router-dom'

const SubnetsPage = () => {
  const [data, setData] = useState<SubnetItem[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(false)
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(20)
  const [createOpen, setCreateOpen] = useState(false)
  const [form] = Form.useForm()
  const navigate = useNavigate()

  const fetchData = async (p = page, ps = pageSize) => {
    setLoading(true)
    try {
      const result = await getSubnets({ page: p, page_size: ps })
      setData(result.list || [])
      setTotal(result.total)
    } catch { /* handled */ } finally { setLoading(false) }
  }

  useEffect(() => { fetchData() }, [])

  const handleCreate = async () => {
    try {
      const values = await form.validateFields()
      await createSubnet(values)
      message.success('网段创建成功')
      setCreateOpen(false)
      form.resetFields()
      fetchData()
    } catch { /* validation */ }
  }

  const columns: ColumnsType<SubnetItem> = [
    { title: 'ID', dataIndex: 'id', width: 60 },
    { title: '名称', dataIndex: 'name', width: 150 },
    { title: 'CIDR', dataIndex: 'cidr', width: 150 },
    { title: '网关', dataIndex: 'gateway', width: 130, render: (v) => v || '-' },
    { title: 'VLAN', dataIndex: 'vlan', width: 70, render: (v) => v || '-' },
    { title: '总 IP', dataIndex: 'total_ips', width: 80 },
    { title: '已用', dataIndex: 'used_ips', width: 60 },
    {
      title: '使用率', width: 140,
      render: (_, r) => (
        <Progress
          percent={Math.round(r.usage_percent)}
          size="small"
          status={r.usage_percent > 80 ? 'exception' : 'normal'}
        />
      ),
    },
    {
      title: '操作', width: 150, fixed: 'right',
      render: (_, record) => (
        <Space>
          <Button type="link" size="small" onClick={() => navigate(`/ipam/addresses?subnet=${record.id}`)}>IP 列表</Button>
          <Popconfirm title="删除网段将同时删除所有 IP 记录，确定？" onConfirm={async () => { await deleteSubnet(record.id); message.success('已删除'); fetchData() }}>
            <Button type="link" size="small" danger>删除</Button>
          </Popconfirm>
        </Space>
      ),
    },
  ]

  return (
    <>
      <Card title="网段管理" style={{ border: '1px solid #E7E9EF' }}>
        <Space style={{ marginBottom: 16 }}>
          <Button type="primary" icon={<PlusOutlined />} onClick={() => { form.resetFields(); setCreateOpen(true) }}>添加网段</Button>
          <Button icon={<ReloadOutlined />} onClick={() => fetchData()}>刷新</Button>
        </Space>
        <Table rowKey="id" columns={columns} dataSource={data} loading={loading} scroll={{ x: 1000 }} pagination={{ current: page, pageSize, total, showSizeChanger: true, showTotal: (t) => `共 ${t} 条`, onChange: (p, ps) => { setPage(p); setPageSize(ps); fetchData(p, ps) } }} />
      </Card>

      <Modal title="添加网段" open={createOpen} onOk={handleCreate} onCancel={() => setCreateOpen(false)} destroyOnHidden>
        <Form form={form} layout="vertical" style={{ marginTop: 16 }}>
          <Form.Item name="name" label="名称" rules={[{ required: true, message: '请输入名称' }]}><Input placeholder="如 办公网" /></Form.Item>
          <Form.Item name="cidr" label="CIDR" rules={[{ required: true, message: '请输入 CIDR' }]}><Input placeholder="如 192.168.1.0/24" /></Form.Item>
          <Form.Item name="gateway" label="网关"><Input placeholder="如 192.168.1.1" /></Form.Item>
          <Form.Item name="vlan" label="VLAN ID"><InputNumber min={0} max={4095} style={{ width: '100%' }} /></Form.Item>
          <Form.Item name="comment" label="备注"><Input.TextArea rows={2} /></Form.Item>
        </Form>
      </Modal>
    </>
  )
}

export default SubnetsPage
