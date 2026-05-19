import { useState, useEffect } from 'react'
import { Card, Table, Checkbox, Button, Tag, message, Tabs, Space, Alert } from 'antd'
import { SaveOutlined, ReloadOutlined } from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import {
  getAllPermissions, getRolePermissions, setRolePermissions,
  type PermissionItem,
} from '../../../services/permission'

const moduleLabels: Record<string, string> = {
  asset: '资产管理',
  cloud: '云账号',
  sshkey: 'SSH 密钥',
  probe: '资产探针',
  task: '批量任务',
  file: '文件分发',
  terminal: 'Web 终端',
  linux_user: 'Linux 用户',
  project: '项目管理',
  ipam: 'IP 管理',
  audit: '审计中心',
  approval: '审批管理',
  playbook: '运维剧本',
  system: '系统管理',
}

const roles = [
  { key: 'admin', label: '管理员', color: 'red' },
  { key: 'operator', label: '运维员', color: 'blue' },
  { key: 'viewer', label: '访客', color: 'default' },
]

const PermissionsPage = () => {
  const [permissions, setPermissions] = useState<PermissionItem[]>([])
  const [loading, setLoading] = useState(false)
  const [saving, setSaving] = useState(false)
  const [activeRole, setActiveRole] = useState('admin')
  const [rolePerms, setRolePerms] = useState<Record<string, Set<string>>>({
    admin: new Set(),
    operator: new Set(),
    viewer: new Set(),
  })
  const [dirty, setDirty] = useState(false)

  const fetchAll = async () => {
    setLoading(true)
    try {
      const [perms, adminP, operatorP, viewerP] = await Promise.all([
        getAllPermissions(),
        getRolePermissions('admin'),
        getRolePermissions('operator'),
        getRolePermissions('viewer'),
      ])
      setPermissions(perms)
      setRolePerms({
        admin: new Set(adminP.permissions || []),
        operator: new Set(operatorP.permissions || []),
        viewer: new Set(viewerP.permissions || []),
      })
      setDirty(false)
    } catch { /* handled */ } finally { setLoading(false) }
  }

  useEffect(() => { fetchAll() }, [])

  const togglePermission = (code: string) => {
    setRolePerms(prev => {
      const newSet = new Set(prev[activeRole])
      if (newSet.has(code)) {
        newSet.delete(code)
      } else {
        newSet.add(code)
      }
      return { ...prev, [activeRole]: newSet }
    })
    setDirty(true)
  }

  const toggleModule = (module: string) => {
    const moduleCodes = permissions.filter(p => p.module === module).map(p => p.code)
    const currentPerms = rolePerms[activeRole]
    const allChecked = moduleCodes.every(c => currentPerms.has(c))

    setRolePerms(prev => {
      const newSet = new Set(prev[activeRole])
      for (const code of moduleCodes) {
        if (allChecked) {
          newSet.delete(code)
        } else {
          newSet.add(code)
        }
      }
      return { ...prev, [activeRole]: newSet }
    })
    setDirty(true)
  }

  const handleSave = async () => {
    setSaving(true)
    try {
      await setRolePermissions(activeRole, Array.from(rolePerms[activeRole]))
      message.success('权限保存成功')
      setDirty(false)
    } catch {
      message.error('保存失败')
    } finally { setSaving(false) }
  }

  // 按模块分组
  const modules = [...new Set(permissions.map(p => p.module))]

  const columns: ColumnsType<PermissionItem> = [
    {
      title: '权限', dataIndex: 'name', width: 160,
      render: (v, record) => (
        <span>
          <Checkbox
            checked={rolePerms[activeRole]?.has(record.code)}
            onChange={() => togglePermission(record.code)}
          >
            {v}
          </Checkbox>
        </span>
      ),
    },
    { title: '标识', dataIndex: 'code', width: 180, render: (v) => <Tag>{v}</Tag> },
    { title: '说明', dataIndex: 'description' },
  ]

  return (
    <Card title="权限管理" style={{ border: '1px solid #E7E9EF' }}>
      <Alert
        message="配置各角色可访问的功能权限。修改后点击保存生效。"
        type="info"
        showIcon
        style={{ marginBottom: 16 }}
      />

      <Tabs
        activeKey={activeRole}
        onChange={k => setActiveRole(k)}
        tabBarExtraContent={
          <Space>
            <Button icon={<ReloadOutlined />} onClick={fetchAll}>重置</Button>
            <Button type="primary" icon={<SaveOutlined />} onClick={handleSave} loading={saving} disabled={!dirty}>
              保存
            </Button>
          </Space>
        }
        items={roles.map(r => ({
          key: r.key,
          label: <span><Tag color={r.color}>{r.label}</Tag> {rolePerms[r.key]?.size || 0} 个权限</span>,
        }))}
      />

      {modules.map(mod => {
        const modPerms = permissions.filter(p => p.module === mod)
        const allChecked = modPerms.every(p => rolePerms[activeRole]?.has(p.code))
        const someChecked = modPerms.some(p => rolePerms[activeRole]?.has(p.code))

        return (
          <Card
            key={mod}
            size="small"
            title={
              <Checkbox
                checked={allChecked}
                indeterminate={someChecked && !allChecked}
                onChange={() => toggleModule(mod)}
              >
                <span style={{ fontWeight: 600 }}>{moduleLabels[mod] || mod}</span>
                <span style={{ color: '#979BA5', marginLeft: 8, fontWeight: 400 }}>
                  ({modPerms.filter(p => rolePerms[activeRole]?.has(p.code)).length}/{modPerms.length})
                </span>
              </Checkbox>
            }
            style={{ border: '1px solid #E7E9EF', marginBottom: 12 }}
          >
            <Table
              rowKey="code"
              columns={columns}
              dataSource={modPerms}
              loading={loading}
              pagination={false}
              size="small"
              showHeader={false}
            />
          </Card>
        )
      })}
    </Card>
  )
}

export default PermissionsPage
