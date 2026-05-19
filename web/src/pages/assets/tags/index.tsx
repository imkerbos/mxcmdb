import { useState, useEffect } from 'react'
import { Card, Table, Input, Space, Tag, Button } from 'antd'
import { SearchOutlined, ReloadOutlined } from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import { getAssets, type AssetItem } from '../../../services/asset'

const AssetTagsPage = () => {
  const [data, setData] = useState<AssetItem[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(false)
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(20)
  const [keyword, setKeyword] = useState('')

  const fetchData = async (p = page, ps = pageSize) => {
    setLoading(true)
    try {
      const params: Record<string, unknown> = { page: p, page_size: ps }
      if (keyword) params.keyword = keyword
      const result = await getAssets(params as any)
      setData(result.list || [])
      setTotal(result.total)
    } catch { /* handled */ } finally { setLoading(false) }
  }

  useEffect(() => { fetchData() }, [])

  const columns: ColumnsType<AssetItem> = [
    { title: '主机名', dataIndex: 'hostname', width: 150 },
    { title: 'IP', dataIndex: 'ip', width: 140 },
    { title: '类型', dataIndex: 'type', width: 100 },
    {
      title: '标签', dataIndex: 'tags',
      render: (tags: AssetItem['tags']) => (
        <Space wrap>
          {(tags || []).map((t) => (
            <Tag key={`${t.key}:${t.value}`} color="blue">{t.key}: {t.value}</Tag>
          ))}
          {(!tags || tags.length === 0) && <span style={{ color: '#999' }}>无标签</span>}
        </Space>
      ),
    },
  ]

  return (
    <Card title="资产标签" style={{ border: '1px solid #E7E9EF' }}>
      <Space style={{ marginBottom: 16 }} wrap>
        <Input placeholder="搜索主机名/IP" prefix={<SearchOutlined />} value={keyword} onChange={(e) => setKeyword(e.target.value)} onPressEnter={() => { setPage(1); fetchData(1, pageSize) }} style={{ width: 200 }} />
        <Button type="primary" icon={<SearchOutlined />} onClick={() => { setPage(1); fetchData(1, pageSize) }}>搜索</Button>
        <Button icon={<ReloadOutlined />} onClick={() => { setKeyword(''); fetchData(1, pageSize) }}>重置</Button>
      </Space>
      <Table rowKey="id" columns={columns} dataSource={data} loading={loading} pagination={{ current: page, pageSize, total, showSizeChanger: true, showTotal: (t) => `共 ${t} 条`, onChange: (p, ps) => { setPage(p); setPageSize(ps); fetchData(p, ps) } }} />
    </Card>
  )
}

export default AssetTagsPage
