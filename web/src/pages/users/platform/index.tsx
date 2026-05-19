import { useState, useEffect } from 'react'
import { Card, Table, Button, Space, Input, Select, Tag, Modal, Form, message, Popconfirm, Switch } from 'antd'
import { PlusOutlined, SearchOutlined, ReloadOutlined, KeyOutlined } from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import {
  getUsers, createUser, updateUser, deleteUser, resetPassword, toggleUserStatus,
  type UserItem, type CreateUserParams, type UpdateUserParams,
} from '../../../services/user'

const roleLabels: Record<string, { text: string; color: string }> = {
  admin: { text: '管理员', color: 'red' },
  operator: { text: '运维员', color: 'blue' },
  viewer: { text: '访客', color: 'default' },
}

const PlatformUsersPage = () => {
  const [data, setData] = useState<UserItem[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(false)
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(20)
  const [keyword, setKeyword] = useState('')
  const [roleFilter, setRoleFilter] = useState('')

  // Modal state
  const [modalOpen, setModalOpen] = useState(false)
  const [editingUser, setEditingUser] = useState<UserItem | null>(null)
  const [pwdModalOpen, setPwdModalOpen] = useState(false)
  const [pwdUserId, setPwdUserId] = useState<number>(0)
  const [form] = Form.useForm()
  const [pwdForm] = Form.useForm()

  const fetchData = async (p = page, ps = pageSize) => {
    setLoading(true)
    try {
      const params: Record<string, unknown> = { page: p, page_size: ps }
      if (keyword) params.keyword = keyword
      if (roleFilter) params.role = roleFilter
      const result = await getUsers(params as any)
      setData(result.list || [])
      setTotal(result.total)
    } catch {
      // handled
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { fetchData() }, [])

  const handleSearch = () => { setPage(1); fetchData(1, pageSize) }

  const handleCreate = () => {
    setEditingUser(null)
    form.resetFields()
    setModalOpen(true)
  }

  const handleEdit = (record: UserItem) => {
    setEditingUser(record)
    form.setFieldsValue({ nickname: record.nickname, email: record.email, phone: record.phone, role: record.role })
    setModalOpen(true)
  }

  const handleModalOk = async () => {
    try {
      const values = await form.validateFields()
      if (editingUser) {
        await updateUser(editingUser.id, values as UpdateUserParams)
        message.success('更新成功')
      } else {
        await createUser(values as CreateUserParams)
        message.success('创建成功')
      }
      setModalOpen(false)
      fetchData()
    } catch {
      // validation error
    }
  }

  const handleDelete = async (id: number) => {
    await deleteUser(id)
    message.success('删除成功')
    fetchData()
  }

  const handleToggleStatus = async (id: number, checked: boolean) => {
    await toggleUserStatus(id, checked ? 1 : 0)
    message.success(checked ? '已启用' : '已禁用')
    fetchData()
  }

  const handleResetPwd = (id: number) => {
    setPwdUserId(id)
    pwdForm.resetFields()
    setPwdModalOpen(true)
  }

  const handlePwdOk = async () => {
    try {
      const values = await pwdForm.validateFields()
      await resetPassword(pwdUserId, values.password)
      message.success('密码已重置')
      setPwdModalOpen(false)
    } catch {
      // validation error
    }
  }

  const columns: ColumnsType<UserItem> = [
    { title: 'ID', dataIndex: 'id', width: 60 },
    { title: '用户名', dataIndex: 'username', width: 120 },
    { title: '昵称', dataIndex: 'nickname', width: 120 },
    { title: '邮箱', dataIndex: 'email', width: 180, ellipsis: true },
    { title: '手机', dataIndex: 'phone', width: 130 },
    {
      title: '角色', dataIndex: 'role', width: 100,
      render: (role) => {
        const r = roleLabels[role] || { text: role, color: 'default' }
        return <Tag color={r.color}>{r.text}</Tag>
      },
    },
    {
      title: '状态', dataIndex: 'status', width: 80,
      render: (status, record) => (
        <Switch
          checked={status === 1}
          size="small"
          onChange={(checked) => handleToggleStatus(record.id, checked)}
          disabled={record.username === 'admin'}
        />
      ),
    },
    { title: 'MFA', dataIndex: 'mfa_enabled', width: 70, render: (v) => v ? <Tag color="green">已启用</Tag> : <Tag>未启用</Tag> },
    { title: '最后登录', dataIndex: 'last_login_at', width: 170, render: (v) => v || '-' },
    { title: '登录 IP', dataIndex: 'last_login_ip', width: 140, render: (v) => v || '-' },
    {
      title: '操作', width: 200, fixed: 'right',
      render: (_, record) => (
        <Space>
          <Button type="link" size="small" onClick={() => handleEdit(record)}>编辑</Button>
          <Button type="link" size="small" icon={<KeyOutlined />} onClick={() => handleResetPwd(record.id)}>重置密码</Button>
          {record.username !== 'admin' && (
            <Popconfirm title="确定删除？" onConfirm={() => handleDelete(record.id)}>
              <Button type="link" size="small" danger>删除</Button>
            </Popconfirm>
          )}
        </Space>
      ),
    },
  ]

  return (
    <>
      <Card title="平台用户管理" style={{ border: '1px solid #E7E9EF' }}>
        <Space style={{ marginBottom: 16 }} wrap>
          <Input
            placeholder="搜索用户名/昵称/邮箱"
            prefix={<SearchOutlined />}
            value={keyword}
            onChange={(e) => setKeyword(e.target.value)}
            onPressEnter={handleSearch}
            style={{ width: 220 }}
          />
          <Select
            value={roleFilter}
            onChange={setRoleFilter}
            style={{ width: 120 }}
            options={[
              { value: '', label: '全部角色' },
              { value: 'admin', label: '管理员' },
              { value: 'operator', label: '运维员' },
              { value: 'viewer', label: '访客' },
            ]}
          />
          <Button type="primary" icon={<SearchOutlined />} onClick={handleSearch}>搜索</Button>
          <Button icon={<ReloadOutlined />} onClick={() => { setKeyword(''); setRoleFilter(''); fetchData(1, pageSize) }}>重置</Button>
          <Button type="primary" icon={<PlusOutlined />} onClick={handleCreate}>新建用户</Button>
        </Space>

        <Table
          rowKey="id"
          columns={columns}
          dataSource={data}
          loading={loading}
          scroll={{ x: 1400 }}
          pagination={{
            current: page, pageSize, total,
            showSizeChanger: true,
            showTotal: (t) => `共 ${t} 条`,
            onChange: (p, ps) => { setPage(p); setPageSize(ps); fetchData(p, ps) },
          }}
        />
      </Card>

      {/* 创建/编辑用户 Modal */}
      <Modal
        title={editingUser ? '编辑用户' : '新建用户'}
        open={modalOpen}
        onOk={handleModalOk}
        onCancel={() => setModalOpen(false)}
        destroyOnHidden
      >
        <Form form={form} layout="vertical" style={{ marginTop: 16 }}>
          {!editingUser && (
            <>
              <Form.Item name="username" label="用户名" rules={[{ required: true, message: '请输入用户名' }, { min: 3, message: '至少3个字符' }]}>
                <Input />
              </Form.Item>
              <Form.Item name="password" label="密码" rules={[{ required: true, message: '请输入密码' }, { min: 6, message: '至少6个字符' }]}>
                <Input.Password />
              </Form.Item>
            </>
          )}
          <Form.Item name="nickname" label="昵称">
            <Input />
          </Form.Item>
          <Form.Item name="email" label="邮箱" rules={[{ type: 'email', message: '邮箱格式不正确' }]}>
            <Input />
          </Form.Item>
          <Form.Item name="phone" label="手机">
            <Input />
          </Form.Item>
          <Form.Item name="role" label="角色" rules={[{ required: true, message: '请选择角色' }]}>
            <Select options={[
              { value: 'admin', label: '管理员' },
              { value: 'operator', label: '运维员' },
              { value: 'viewer', label: '访客' },
            ]} />
          </Form.Item>
        </Form>
      </Modal>

      {/* 重置密码 Modal */}
      <Modal
        title="重置密码"
        open={pwdModalOpen}
        onOk={handlePwdOk}
        onCancel={() => setPwdModalOpen(false)}
        destroyOnHidden
      >
        <Form form={pwdForm} layout="vertical" style={{ marginTop: 16 }}>
          <Form.Item name="password" label="新密码" rules={[{ required: true, message: '请输入新密码' }, { min: 6, message: '至少6个字符' }]}>
            <Input.Password />
          </Form.Item>
        </Form>
      </Modal>
    </>
  )
}

export default PlatformUsersPage
