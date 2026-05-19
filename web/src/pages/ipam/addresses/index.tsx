import { useState, useEffect } from 'react'
import { Card, Table, Button, Space, Tag, Select, Modal, Form, Input, message } from 'antd'
import { ReloadOutlined } from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import { useSearchParams } from 'react-router-dom'
import { getSubnetIPs, getSubnets, allocateIP, releaseIP, type IPAddressItem, type SubnetItem } from '../../../services/ipam'

const IPAddressesPage = () => {
  const [searchParams] = useSearchParams()
  const initialSubnet = searchParams.get('subnet')
  const [subnetId, setSubnetId] = useState<number>(initialSubnet ? Number(initialSubnet) : 0)
  const [subnets, setSubnets] = useState<SubnetItem[]>([])
  const [data, setData] = useState<IPAddressItem[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(false)
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(50)
  const [statusFilter, setStatusFilter] = useState<string>('')
  const [allocateOpen, setAllocateOpen] = useState(false)
  const [allocateIpId, setAllocateIpId] = useState(0)
  const [form] = Form.useForm()

  useEffect(() => {
    getSubnets({ page: 1, page_size: 100 }).then((r) => setSubnets(r.list || [])).catch(() => {})
  }, [])

  useEffect(() => {
    if (subnetId > 0) fetchData()
  }, [subnetId])

  const fetchData = async (p = page, ps = pageSize) => {
    if (subnetId <= 0) return
    setLoading(true)
    try {
      const params: Record<string, unknown> = { page: p, page_size: ps }
      if (statusFilter) params.status = statusFilter
      const result = await getSubnetIPs(subnetId, params)
      setData(result.list || [])
      setTotal(result.total)
    } catch { /* handled */ } finally { setLoading(false) }
  }

  const handleAllocate = async () => {
    try {
      const values = await form.validateFields()
      await allocateIP(allocateIpId, values)
      message.success('IP 分配成功')
      setAllocateOpen(false)
      form.resetFields()
      fetchData()
    } catch { /* validation */ }
  }

  const handleRelease = async (ipId: number) => {
    await releaseIP(ipId)
    message.success('IP 已释放')
    fetchData()
  }

  const statusColors: Record<string, string> = { available: 'green', allocated: 'blue', reserved: 'orange' }
  const statusTexts: Record<string, string> = { available: '可用', allocated: '已分配', reserved: '保留' }

  const columns: ColumnsType<IPAddressItem> = [
    { title: 'IP 地址', dataIndex: 'address', width: 140 },
    { title: '状态', dataIndex: 'status', width: 90, render: (v) => <Tag color={statusColors[v] || 'default'}>{statusTexts[v] || v}</Tag> },
    { title: '主机名', dataIndex: 'hostname', width: 140, render: (v) => v || '-' },
    { title: '备注', dataIndex: 'comment', ellipsis: true, render: (v) => v || '-' },
    {
      title: '操作', width: 120, fixed: 'right',
      render: (_, record) => (
        <Space>
          {record.status === 'available' && (
            <Button type="link" size="small" onClick={() => { setAllocateIpId(record.id); form.resetFields(); setAllocateOpen(true) }}>分配</Button>
          )}
          {record.status === 'allocated' && (
            <Button type="link" size="small" danger onClick={() => handleRelease(record.id)}>释放</Button>
          )}
        </Space>
      ),
    },
  ]

  return (
    <>
      <Card title="IP 分配" style={{ border: '1px solid #E7E9EF' }}>
        <Space style={{ marginBottom: 16 }} wrap>
          <Select
            placeholder="选择网段"
            value={subnetId || undefined}
            onChange={(v) => { setSubnetId(v); setPage(1) }}
            style={{ width: 240 }}
            options={subnets.map((s) => ({ value: s.id, label: `${s.name} (${s.cidr})` }))}
          />
          <Select
            placeholder="状态"
            allowClear
            value={statusFilter || undefined}
            onChange={(v) => setStatusFilter(v || '')}
            style={{ width: 120 }}
            options={[
              { value: 'available', label: '可用' },
              { value: 'allocated', label: '已分配' },
              { value: 'reserved', label: '保留' },
            ]}
          />
          <Button type="primary" onClick={() => { setPage(1); fetchData(1, pageSize) }}>筛选</Button>
          <Button icon={<ReloadOutlined />} onClick={() => fetchData()}>刷新</Button>
        </Space>
        <Table rowKey="id" columns={columns} dataSource={data} loading={loading} pagination={{ current: page, pageSize, total, showSizeChanger: true, showTotal: (t) => `共 ${t} 条`, onChange: (p, ps) => { setPage(p); setPageSize(ps); fetchData(p, ps) } }} size="small" />
      </Card>

      <Modal title="分配 IP" open={allocateOpen} onOk={handleAllocate} onCancel={() => setAllocateOpen(false)} destroyOnHidden>
        <Form form={form} layout="vertical" style={{ marginTop: 16 }}>
          <Form.Item name="hostname" label="主机名"><Input placeholder="关联主机名" /></Form.Item>
          <Form.Item name="comment" label="备注"><Input placeholder="用途说明" /></Form.Item>
        </Form>
      </Modal>
    </>
  )
}

export default IPAddressesPage
