import { useEffect, useState, useMemo } from 'react'
import { Table, Input, Select, Space, Tag, Checkbox } from 'antd'
import { SearchOutlined } from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import { getAllAssets, type AssetSimple } from '../../services/asset'
import { getAllProjects, type ProjectSimple } from '../../services/project'

interface BatchAssetSelectorProps {
  value?: number[]
  onChange?: (ids: number[]) => void
  maxHeight?: number
}

const statusColors: Record<string, string> = {
  online: 'green', offline: 'red', running: 'green', stopped: 'red', unknown: 'default', terminated: 'default',
}
const statusLabels: Record<string, string> = {
  online: '在线', offline: '离线', running: '运行中', stopped: '已停止', unknown: '未知', terminated: '已终止',
}
const envLabels: Record<string, string> = { prod: '生产', test: '测试', dev: '开发' }
const typeLabels: Record<string, string> = { ecs: 'ECS', physical_server: '物理服务器', vm: '虚拟机', network_device: '网络设备' }

const BatchAssetSelector = ({ value = [], onChange, maxHeight = 400 }: BatchAssetSelectorProps) => {
  const [allAssets, setAllAssets] = useState<AssetSimple[]>([])
  const [projects, setProjects] = useState<ProjectSimple[]>([])
  const [loading, setLoading] = useState(false)

  // Filters
  const [keyword, setKeyword] = useState('')
  const [projectFilter, setProjectFilter] = useState<number | ''>('')
  const [envFilter, setEnvFilter] = useState('')
  const [typeFilter, setTypeFilter] = useState('')

  useEffect(() => {
    setLoading(true)
    Promise.all([getAllAssets(), getAllProjects()])
      .then(([assets, projs]) => {
        setAllAssets(assets || [])
        setProjects(projs || [])
      })
      .catch(() => {})
      .finally(() => setLoading(false))
  }, [])

  // Build project map for display
  const projectMap = useMemo(() => {
    const m: Record<number, string> = {}
    projects.forEach(p => { m[p.id] = p.name })
    return m
  }, [projects])

  // Filtered assets
  const filteredAssets = useMemo(() => {
    let list = allAssets
    if (keyword) {
      const kw = keyword.toLowerCase()
      list = list.filter(a => a.hostname?.toLowerCase().includes(kw) || a.ip?.toLowerCase().includes(kw))
    }
    if (projectFilter !== '') list = list.filter(a => a.project_id === projectFilter)
    if (envFilter) list = list.filter(a => a.environment === envFilter)
    if (typeFilter) list = list.filter(a => a.type === typeFilter)
    return list
  }, [allAssets, keyword, projectFilter, envFilter, typeFilter])

  const filteredIds = useMemo(() => filteredAssets.map(a => a.id), [filteredAssets])

  // "Select all filtered" state
  const allFilteredSelected = filteredIds.length > 0 && filteredIds.every(id => value.includes(id))
  const someFilteredSelected = filteredIds.some(id => value.includes(id)) && !allFilteredSelected

  const handleSelectAllFiltered = (checked: boolean) => {
    if (!onChange) return
    if (checked) {
      // Add all filtered IDs to selection (keep existing non-filtered selections)
      const newSet = new Set([...value, ...filteredIds])
      onChange(Array.from(newSet))
    } else {
      // Remove all filtered IDs from selection
      const removeSet = new Set(filteredIds)
      onChange(value.filter(id => !removeSet.has(id)))
    }
  }

  const columns: ColumnsType<AssetSimple> = [
    {
      title: '主机名', dataIndex: 'hostname', width: 180, ellipsis: true,
      render: (v: string, r) => v || r.ip,
    },
    { title: 'IP', dataIndex: 'ip', width: 140 },
    {
      title: '项目', dataIndex: 'project_id', width: 120, ellipsis: true,
      render: (v: number | null) => v ? projectMap[v] || '-' : '-',
    },
    {
      title: '环境', dataIndex: 'environment', width: 70,
      render: (v: string) => envLabels[v] || v || '-',
    },
    {
      title: '类型', dataIndex: 'type', width: 100,
      render: (v: string) => typeLabels[v] || v || '-',
    },
    {
      title: '状态', dataIndex: 'status', width: 80,
      render: (v: string) => <Tag color={statusColors[v] || 'default'}>{statusLabels[v] || v}</Tag>,
    },
  ]

  return (
    <div>
      <Space wrap style={{ marginBottom: 12 }}>
        <Input
          placeholder="搜索主机名 / IP"
          prefix={<SearchOutlined style={{ color: '#979BA5' }} />}
          value={keyword}
          onChange={e => setKeyword(e.target.value)}
          style={{ width: 200 }}
          allowClear
        />
        <Select
          value={projectFilter}
          onChange={setProjectFilter}
          style={{ width: 160 }}
          options={[
            { value: '' as const, label: '全部项目' },
            ...projects.map(p => ({ value: p.id, label: p.name })),
          ]}
        />
        <Select
          value={envFilter}
          onChange={setEnvFilter}
          style={{ width: 110 }}
          options={[
            { value: '', label: '全部环境' },
            { value: 'prod', label: '生产' },
            { value: 'test', label: '测试' },
            { value: 'dev', label: '开发' },
          ]}
        />
        <Select
          value={typeFilter}
          onChange={setTypeFilter}
          style={{ width: 130 }}
          options={[
            { value: '', label: '全部类型' },
            { value: 'ecs', label: 'ECS' },
            { value: 'physical_server', label: '物理服务器' },
            { value: 'vm', label: '虚拟机' },
          ]}
        />
      </Space>
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 8 }}>
        <Checkbox
          checked={allFilteredSelected}
          indeterminate={someFilteredSelected}
          onChange={e => handleSelectAllFiltered(e.target.checked)}
        >
          全选当前筛选结果（{filteredAssets.length} 台）
        </Checkbox>
        <span style={{ color: '#979BA5', fontSize: 13 }}>
          已选 <span style={{ color: '#3A84FF', fontWeight: 500 }}>{value.length}</span> 台
        </span>
      </div>
      <Table
        rowKey="id"
        columns={columns}
        dataSource={filteredAssets}
        loading={loading}
        size="small"
        scroll={{ y: maxHeight }}
        rowSelection={{
          selectedRowKeys: value,
          onChange: (keys) => onChange?.(keys as number[]),
        }}
        pagination={false}
      />
    </div>
  )
}

export default BatchAssetSelector
