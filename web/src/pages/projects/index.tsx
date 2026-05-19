import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { Card, Table, Button, Space, Input, Select, Tag, Modal, Form, message, Popconfirm } from 'antd'
import { PlusOutlined, SearchOutlined, ReloadOutlined } from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import {
  getProjects, createProject, updateProject, deleteProject,
  type ProjectItem, type CreateProjectParams,
} from '../../services/project'

const statusMap: Record<string, { color: string; text: string }> = {
  active: { color: 'green', text: '活跃' },
  archived: { color: 'default', text: '已归档' },
}

const ProjectsPage = () => {
  const navigate = useNavigate()
  const [data, setData] = useState<ProjectItem[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(false)
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(20)
  const [keyword, setKeyword] = useState('')
  const [statusFilter, setStatusFilter] = useState('')
  const [modalOpen, setModalOpen] = useState(false)
  const [editing, setEditing] = useState<ProjectItem | null>(null)
  const [form] = Form.useForm()

  const fetchData = async (p = page, ps = pageSize) => {
    setLoading(true)
    try {
      const params: Record<string, unknown> = { page: p, page_size: ps }
      if (keyword) params.keyword = keyword
      if (statusFilter) params.status = statusFilter
      const result = await getProjects(params as any)
      setData(result.list || [])
      setTotal(result.total)
    } catch { /* handled */ } finally { setLoading(false) }
  }

  useEffect(() => { fetchData() }, [])

  const handleCreate = () => {
    setEditing(null)
    form.resetFields()
    setModalOpen(true)
  }

  const handleEdit = (record: ProjectItem) => {
    setEditing(record)
    form.setFieldsValue({
      name: record.name,
      code: record.code,
      description: record.description,
      owner: record.owner,
      status: record.status,
    })
    setModalOpen(true)
  }

  const handleModalOk = async () => {
    try {
      const values = await form.validateFields()
      if (editing) {
        await updateProject(editing.id, values)
        message.success('更新成功')
      } else {
        await createProject(values as CreateProjectParams)
        message.success('创建成功')
      }
      setModalOpen(false)
      fetchData()
    } catch { /* validation */ }
  }

  const columns: ColumnsType<ProjectItem> = [
    { title: '项目名称', dataIndex: 'name', width: 160, render: (v, record) => <Button type="link" size="small" style={{ padding: 0 }} onClick={() => navigate(`/projects/${record.id}`)}>{v}</Button> },
    { title: '项目代码', dataIndex: 'code', width: 120, render: (v) => <Tag>{v}</Tag> },
    { title: '描述', dataIndex: 'description', ellipsis: true },
    { title: '负责人', dataIndex: 'owner', width: 100 },
    { title: '资产数', dataIndex: 'asset_count', width: 80, render: (v) => <Tag color="blue">{v}</Tag> },
    {
      title: '状态', dataIndex: 'status', width: 80,
      render: (v) => {
        const s = statusMap[v] || { color: 'default', text: v }
        return <Tag color={s.color}>{s.text}</Tag>
      },
    },
    { title: '创建时间', dataIndex: 'created_at', width: 170 },
    {
      title: '操作', width: 150, fixed: 'right',
      render: (_, record) => (
        <Space>
          <Button type="link" size="small" onClick={() => handleEdit(record)}>编辑</Button>
          <Popconfirm
            title="确定删除该项目？"
            description={record.asset_count > 0 ? '该项目下仍有资产，需先解除关联' : undefined}
            onConfirm={async () => {
              try {
                await deleteProject(record.id)
                message.success('已删除')
                fetchData()
              } catch { /* handled */ }
            }}
          >
            <Button type="link" size="small" danger>删除</Button>
          </Popconfirm>
        </Space>
      ),
    },
  ]

  return (
    <>
      <Card title="项目管理" style={{ border: '1px solid #E7E9EF' }}>
        <Space style={{ marginBottom: 16 }} wrap>
          <Input
            placeholder="搜索项目名称/代码"
            prefix={<SearchOutlined />}
            value={keyword}
            onChange={e => setKeyword(e.target.value)}
            onPressEnter={() => { setPage(1); fetchData(1, pageSize) }}
            style={{ width: 200 }}
          />
          <Select
            value={statusFilter}
            onChange={setStatusFilter}
            style={{ width: 120 }}
            options={[
              { value: '', label: '全部状态' },
              { value: 'active', label: '活跃' },
              { value: 'archived', label: '已归档' },
            ]}
          />
          <Button type="primary" icon={<SearchOutlined />} onClick={() => { setPage(1); fetchData(1, pageSize) }}>搜索</Button>
          <Button icon={<ReloadOutlined />} onClick={() => { setKeyword(''); setStatusFilter(''); fetchData(1, pageSize) }}>重置</Button>
          <Button type="primary" icon={<PlusOutlined />} onClick={handleCreate}>创建项目</Button>
        </Space>
        <Table
          rowKey="id"
          columns={columns}
          dataSource={data}
          loading={loading}
          scroll={{ x: 1000 }}
          pagination={{
            current: page,
            pageSize,
            total,
            showSizeChanger: true,
            showTotal: (t) => `共 ${t} 个项目`,
            onChange: (p, ps) => { setPage(p); setPageSize(ps); fetchData(p, ps) },
          }}
        />
      </Card>

      <Modal
        title={editing ? '编辑项目' : '创建项目'}
        open={modalOpen}
        onOk={handleModalOk}
        onCancel={() => setModalOpen(false)}
        destroyOnHidden
        width={520}
      >
        <Form form={form} layout="vertical" style={{ marginTop: 16 }}>
          <Form.Item name="name" label="项目名称" rules={[{ required: true, message: '请输入项目名称' }]}>
            <Input placeholder="如：电商平台" />
          </Form.Item>
          <Form.Item name="code" label="项目代码" rules={[{ required: true, message: '请输入项目代码' }]}>
            <Input placeholder="如：EC-MALL" disabled={!!editing} />
          </Form.Item>
          <Form.Item name="description" label="描述">
            <Input.TextArea rows={3} placeholder="项目描述（可选）" />
          </Form.Item>
          <Form.Item name="owner" label="负责人">
            <Input placeholder="项目负责人" />
          </Form.Item>
          {editing && (
            <Form.Item name="status" label="状态">
              <Select options={[
                { value: 'active', label: '活跃' },
                { value: 'archived', label: '已归档' },
              ]} />
            </Form.Item>
          )}
        </Form>
      </Modal>
    </>
  )
}

export default ProjectsPage
