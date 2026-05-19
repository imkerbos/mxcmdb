import { useEffect, useState } from 'react'
import { Card, Form, Input, Button, message, Spin, Descriptions, Typography, Space, Select, Table, Tag } from 'antd'
import { SaveOutlined, ReloadOutlined, SafetyCertificateOutlined } from '@ant-design/icons'
import { getConfigs, batchUpdateConfigs, type SystemConfig } from '../../../services/settings'
import { getUsers, type UserItem } from '../../../services/user'

const mfaPolicyOptions = [
  { value: 'optional', label: '用户自愿 — 用户可自行选择是否启用 MFA' },
  { value: 'required_admin', label: '管理员必须 — 管理员角色登录时必须绑定 MFA' },
  { value: 'required_all', label: '全员必须 — 所有用户登录时必须绑定 MFA' },
]

const SecuritySettingsPage = () => {
  const [configs, setConfigs] = useState<SystemConfig[]>([])
  const [loading, setLoading] = useState(false)
  const [saving, setSaving] = useState(false)
  const [users, setUsers] = useState<UserItem[]>([])
  const [usersLoading, setUsersLoading] = useState(false)
  const [form] = Form.useForm()

  const fetchConfigs = async () => {
    setLoading(true)
    try {
      const data = await getConfigs('security')
      setConfigs(data)
      const values: Record<string, string> = {}
      data.forEach((c) => {
        values[c.key] = c.value
      })
      form.setFieldsValue(values)
    } catch {
      // error handled by interceptor
    } finally {
      setLoading(false)
    }
  }

  const fetchUsers = async () => {
    setUsersLoading(true)
    try {
      const data = await getUsers({ page: 1, page_size: 200 })
      setUsers(data.list || [])
    } catch {
      // handled
    } finally {
      setUsersLoading(false)
    }
  }

  useEffect(() => {
    fetchConfigs()
    fetchUsers()
  }, [])

  const handleSave = async () => {
    setSaving(true)
    try {
      const values = form.getFieldsValue()
      const items = configs
        .filter((c) => values[c.key] !== c.value)
        .map((c) => ({ key: c.key, value: values[c.key] }))
      if (items.length === 0) {
        message.info('没有修改')
        return
      }
      await batchUpdateConfigs(items)
      message.success('保存成功')
      fetchConfigs()
    } catch {
      // error handled by interceptor
    } finally {
      setSaving(false)
    }
  }

  const renderConfigInput = (item: SystemConfig) => {
    if (item.key === 'security.mfa_policy') {
      return (
        <Select options={mfaPolicyOptions} style={{ width: 400 }} />
      )
    }
    if (item.key === 'security.dangerous_commands') {
      return <Input.TextArea rows={3} style={{ width: 400 }} />
    }
    return <Input style={{ width: 300 }} />
  }

  const mfaColumns = [
    { title: '用户名', dataIndex: 'username', key: 'username', width: 120 },
    { title: '昵称', dataIndex: 'nickname', key: 'nickname', width: 120 },
    {
      title: '角色', dataIndex: 'role', key: 'role', width: 100,
      render: (role: string) => {
        const map: Record<string, string> = { admin: '管理员', operator: '运维员', viewer: '访客' }
        return map[role] || role
      },
    },
    {
      title: 'MFA 状态', dataIndex: 'mfa_enabled', key: 'mfa_enabled', width: 100,
      render: (enabled: boolean) =>
        enabled ? <Tag color="green">已启用</Tag> : <Tag color="default">未启用</Tag>,
    },
  ]

  return (
    <Spin spinning={loading}>
      <Form form={form} layout="vertical">
        <Card title="安全设置" style={{ marginBottom: 16, border: '1px solid #E7E9EF' }}>
          <Descriptions column={1} bordered size="small">
            {configs.map((item) => (
              <Descriptions.Item
                key={item.key}
                label={
                  <div>
                    <div>{item.description}</div>
                    <Typography.Text type="secondary" style={{ fontSize: 12 }}>
                      {item.key}
                    </Typography.Text>
                  </div>
                }
              >
                <Form.Item name={item.key} noStyle>
                  {renderConfigInput(item)}
                </Form.Item>
              </Descriptions.Item>
            ))}
          </Descriptions>
        </Card>

        <Card style={{ marginBottom: 16, border: '1px solid #E7E9EF' }}>
          <Space>
            <Button type="primary" icon={<SaveOutlined />} onClick={handleSave} loading={saving}>
              保存修改
            </Button>
            <Button icon={<ReloadOutlined />} onClick={fetchConfigs}>
              重新加载
            </Button>
          </Space>
        </Card>
      </Form>

      <Card
        title={<><SafetyCertificateOutlined style={{ marginRight: 8, color: '#3A84FF' }} />用户 MFA 状态</>}
        style={{ border: '1px solid #E7E9EF' }}
      >
        <Table
          dataSource={users}
          columns={mfaColumns}
          rowKey="id"
          size="middle"
          loading={usersLoading}
          pagination={{ showSizeChanger: true, showTotal: (t) => `共 ${t} 条` }}
        />
      </Card>
    </Spin>
  )
}

export default SecuritySettingsPage
