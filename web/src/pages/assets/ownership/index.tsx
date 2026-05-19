import { useState, useEffect } from 'react'
import { Card, Table, Input, Select, Space, Button } from 'antd'
import { SearchOutlined, ReloadOutlined } from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import { getAssets, type AssetItem } from '../../../services/asset'

const AssetOwnershipPage = () => {
  const [data, setData] = useState<AssetItem[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(false)
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(20)
  const [keyword, setKeyword] = useState('')
  const [deptFilter, setDeptFilter] = useState('')

  const fetchData = async (p = page, ps = pageSize) => {
    setLoading(true)
    try {
      const params: Record<string, unknown> = { page: p, page_size: ps }
      if (keyword) params.keyword = keyword
      if (deptFilter) params.department = deptFilter
      const result = await getAssets(params as any)
      setData(result.list || [])
      setTotal(result.total)
    } catch { /* handled */ } finally { setLoading(false) }
  }

  useEffect(() => { fetchData() }, [])

  const columns: ColumnsType<AssetItem> = [
    { title: '主机名', dataIndex: 'hostname', width: 150, ellipsis: true },
    { title: 'IP', dataIndex: 'ip', width: 140 },
    { title: '类型', dataIndex: 'type', width: 100 },
    { title: '部门', dataIndex: 'department', width: 120 },
    { title: '项目', dataIndex: 'project', width: 120 },
    { title: '负责人', dataIndex: 'owner', width: 100 },
    { title: '环境', dataIndex: 'environment', width: 80 },
    { title: '业务组', dataIndex: 'business_group', width: 120 },
  ]

  return (
    <Card title="资产归属" style={{ border: '1px solid #E7E9EF' }}>
      <Space style={{ marginBottom: 16 }} wrap>
        <Input placeholder="搜索主机名/IP" prefix={<SearchOutlined />} value={keyword} onChange={(e) => setKeyword(e.target.value)} onPressEnter={() => { setPage(1); fetchData(1, pageSize) }} style={{ width: 200 }} />
        <Select value={deptFilter} onChange={setDeptFilter} style={{ width: 140 }} allowClear placeholder="按部门筛选" options={[{ value: '', label: '全部部门' }]} />
        <Button type="primary" icon={<SearchOutlined />} onClick={() => { setPage(1); fetchData(1, pageSize) }}>搜索</Button>
        <Button icon={<ReloadOutlined />} onClick={() => { setKeyword(''); setDeptFilter(''); fetchData(1, pageSize) }}>重置</Button>
      </Space>
      <Table rowKey="id" columns={columns} dataSource={data} loading={loading} pagination={{ current: page, pageSize, total, showSizeChanger: true, showTotal: (t) => `共 ${t} 条`, onChange: (p, ps) => { setPage(p); setPageSize(ps); fetchData(p, ps) } }} />
    </Card>
  )
}

export default AssetOwnershipPage
