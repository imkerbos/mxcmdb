import { useState, useEffect } from 'react'
import { Card, Table, Button, Space, Input, Tag, Modal, Form, Radio, Select, Dropdown, Tooltip, message, Popconfirm, Descriptions, Alert } from 'antd'
import {
  PlusOutlined, SearchOutlined, ReloadOutlined, SendOutlined, LinkOutlined,
  SyncOutlined, UserDeleteOutlined, RollbackOutlined,
  HistoryOutlined, MoreOutlined, DeleteOutlined, KeyOutlined, FileTextOutlined,
} from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import type { MenuProps } from 'antd'
import {
  getSSHKeys, getAllSSHKeys, createSSHKey, deleteSSHKey, deploySSHKey, getSSHKeyBindings,
  revokeSSHKey, rotateSSHKey, departureCleanup, downloadSSHKey, getSSHKeyDeployLogs,
  type SSHKeyItem, type SSHKeyBinding, type CreateSSHKeyParams,
  type RevokeResponse, type DepartureCleanupResponse, type SSHKeyDeployLog,
} from '../../services/ssh-key'
import BatchAssetSelector from '../../components/BatchAssetSelector'

// 触发浏览器下载
const triggerDownload = (content: string, filename: string) => {
  const blob = new Blob([content], { type: 'application/octet-stream' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = filename
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
  URL.revokeObjectURL(url)
}

const SSHKeysPage = () => {
  const [data, setData] = useState<SSHKeyItem[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(false)
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(20)
  const [keyword, setKeyword] = useState('')
  const [createOpen, setCreateOpen] = useState(false)
  const [createMode, setCreateMode] = useState<'generate' | 'import'>('generate')
  const [deployOpen, setDeployOpen] = useState(false)
  const [deployKeyId, setDeployKeyId] = useState<number>(0)
  const [bindingsOpen, setBindingsOpen] = useState(false)
  const [bindingsKeyId, setBindingsKeyId] = useState<number>(0)
  const [bindings, setBindings] = useState<SSHKeyBinding[]>([])
  const [bindingsLoading, setBindingsLoading] = useState(false)
  const [createForm] = Form.useForm()
  const [deployForm] = Form.useForm()

  // 轮换
  const [rotateOpen, setRotateOpen] = useState(false)
  const [rotateKeyId, setRotateKeyId] = useState<number>(0)
  const [rotateKeyName, setRotateKeyName] = useState('')
  const [rotateLoading, setRotateLoading] = useState(false)
  const [rotateMode, setRotateMode] = useState<'generate' | 'select'>('select')
  const [allKeys, setAllKeys] = useState<SSHKeyItem[]>([])
  const [rotateForm] = Form.useForm()

  // 离职清理
  const [departureOpen, setDepartureOpen] = useState(false)
  const [departureLoading, setDepartureLoading] = useState(false)
  const [departureResult, setDepartureResult] = useState<DepartureCleanupResponse | null>(null)
  const [departureForm] = Form.useForm()

  // 撤销
  const [revokeLoading, setRevokeLoading] = useState(false)

  // 部署历史
  const [logsOpen, setLogsOpen] = useState(false)
  const [logsKeyId, setLogsKeyId] = useState<number>(0)
  const [logsKeyName, setLogsKeyName] = useState('')
  const [logs, setLogs] = useState<SSHKeyDeployLog[]>([])
  const [logsTotal, setLogsTotal] = useState(0)
  const [logsPage, setLogsPage] = useState(1)
  const [logsPageSize, setLogsPageSize] = useState(15)
  const [logsLoading, setLogsLoading] = useState(false)

  const fetchData = async (p = page, ps = pageSize) => {
    setLoading(true)
    try {
      const params: Record<string, unknown> = { page: p, page_size: ps }
      if (keyword) params.keyword = keyword
      const result = await getSSHKeys(params)
      setData(result.list || [])
      setTotal(result.total)
    } catch { /* handled */ } finally { setLoading(false) }
  }

  useEffect(() => { fetchData() }, [])

  const handleCreate = async () => {
    try {
      const values = await createForm.validateFields()
      const params: CreateSSHKeyParams = { name: values.name, comment: values.comment }
      if (createMode === 'import') {
        params.public_key = values.public_key
        params.private_key = values.private_key
      } else {
        params.key_type = values.key_type || 'ed25519'
      }
      await createSSHKey(params)
      message.success('创建成功')
      setCreateOpen(false)
      createForm.resetFields()
      fetchData()
    } catch { /* validation */ }
  }

  const handleDeploy = async () => {
    try {
      const values = await deployForm.validateFields()
      await deploySSHKey(deployKeyId, { asset_ids: values.asset_ids, username: values.username })
      message.success('部署任务已提交')
      setDeployOpen(false)
      deployForm.resetFields()
    } catch { /* validation */ }
  }

  const showBindings = async (keyId: number) => {
    setBindingsKeyId(keyId)
    setBindingsOpen(true)
    setBindingsLoading(true)
    try {
      const result = await getSSHKeyBindings(keyId)
      setBindings(result || [])
    } catch { setBindings([]) } finally { setBindingsLoading(false) }
  }

  const handleRevokeBinding = async (binding: SSHKeyBinding) => {
    setRevokeLoading(true)
    try {
      const result = await revokeSSHKey(binding.ssh_key_id, {
        asset_ids: [binding.asset_id],
        username: binding.username,
      })
      if (result.success > 0) {
        message.success(`已从 ${binding.hostname || binding.ip} 撤销密钥`)
      } else {
        message.error(`撤销失败: ${result.results[0]?.error || '未知错误'}`)
      }
      showBindings(bindingsKeyId)
    } catch {
      message.error('撤销失败')
    } finally {
      setRevokeLoading(false)
    }
  }

  const openRotateModal = async (record: SSHKeyItem) => {
    setRotateKeyId(record.id)
    setRotateKeyName(record.name)
    setRotateMode('select')
    rotateForm.resetFields()
    setRotateOpen(true)
    try {
      const keys = await getAllSSHKeys()
      setAllKeys(keys.filter(k => k.id !== record.id))
    } catch { setAllKeys([]) }
  }

  const handleRotate = async () => {
    try {
      const values = await rotateForm.validateFields()
      setRotateLoading(true)
      const params: Record<string, unknown> = {}
      if (rotateMode === 'select') {
        params.new_key_id = values.new_key_id
      } else {
        params.new_name = values.new_name
        params.new_comment = values.new_comment
      }
      const result = await rotateSSHKey(rotateKeyId, params)
      message.success(`轮换已启动，新密钥: ${result.new_key.name}。新密钥正在部署，旧密钥将在部署完成后自动撤销。`)
      setRotateOpen(false)
      rotateForm.resetFields()
      fetchData()
    } catch { /* validation */ } finally { setRotateLoading(false) }
  }

  const handleDepartureCleanup = async () => {
    try {
      const values = await departureForm.validateFields()
      setDepartureLoading(true)
      setDepartureResult(null)
      const result = await departureCleanup({
        key_ids: values.key_ids,
        asset_ids: values.asset_ids,
      })
      setDepartureResult(result)
      message.success(`离职清理完成: ${result.total_keys} 个密钥, ${result.total_assets} 台资产`)
    } catch (e: unknown) {
      const err = e as { message?: string }
      if (err.message) message.error(err.message)
    } finally { setDepartureLoading(false) }
  }

  // 下载私钥
  const handleDownloadPrivateKey = async (record: SSHKeyItem) => {
    try {
      const result = await downloadSSHKey(record.id)
      triggerDownload(result.private_key, `${result.name}.pem`)
      message.success('私钥下载成功，请妥善保管')
    } catch {
      message.error('下载失败')
    }
  }

  // 下载公钥（直接使用已有数据）
  const handleDownloadPublicKey = (record: SSHKeyItem) => {
    triggerDownload(record.public_key, `${record.name}.pub`)
    message.success('公钥下载成功')
  }

  // 部署历史
  const showDeployLogs = async (keyId: number, keyName: string, p = 1, ps = 15) => {
    setLogsKeyId(keyId)
    setLogsKeyName(keyName)
    setLogsPage(p)
    setLogsPageSize(ps)
    setLogsOpen(true)
    setLogsLoading(true)
    try {
      const result = await getSSHKeyDeployLogs(keyId, { page: p, page_size: ps })
      setLogs(result.list || [])
      setLogsTotal(result.total)
    } catch { setLogs([]); setLogsTotal(0) } finally { setLogsLoading(false) }
  }

  // 删除确认
  const handleDelete = (record: SSHKeyItem) => {
    Modal.confirm({
      title: '确认删除',
      content: `确定删除密钥「${record.name}」？关联的绑定记录也将被清除。`,
      okText: '删除',
      okButtonProps: { danger: true },
      cancelText: '取消',
      onOk: async () => {
        await deleteSSHKey(record.id)
        message.success('已删除')
        fetchData()
      },
    })
  }

  const statusColors: Record<string, string> = { deployed: 'green', pending: 'gold', failed: 'red', revoked: 'default' }
  const statusTexts: Record<string, string> = { deployed: '已部署', pending: '部署中', failed: '失败', revoked: '已撤销' }

  const actionTexts: Record<string, string> = { deploy: '部署', revoke: '撤销', rotate: '轮换' }
  const actionColors: Record<string, string> = { deploy: 'blue', revoke: 'orange', rotate: 'gold' }

  // 操作栏「更多」下拉菜单
  const getMoreMenu = (record: SSHKeyItem): MenuProps => ({
    items: [
      { key: 'bindings', icon: <LinkOutlined />, label: '绑定详情' },
      { key: 'logs', icon: <HistoryOutlined />, label: '操作历史' },
      { type: 'divider' },
      { key: 'download-private', icon: <KeyOutlined />, label: '下载私钥 (.pem)' },
      { key: 'download-public', icon: <FileTextOutlined />, label: '下载公钥 (.pub)' },
      { type: 'divider' },
      { key: 'delete', icon: <DeleteOutlined />, label: '删除', danger: true },
    ],
    onClick: ({ key }) => {
      switch (key) {
        case 'bindings': showBindings(record.id); break
        case 'logs': showDeployLogs(record.id, record.name); break
        case 'download-private': handleDownloadPrivateKey(record); break
        case 'download-public': handleDownloadPublicKey(record); break
        case 'delete': handleDelete(record); break
      }
    },
  })

  const columns: ColumnsType<SSHKeyItem> = [
    { title: '名称', dataIndex: 'name', width: 180, ellipsis: true },
    {
      title: '类型', dataIndex: 'key_type', width: 120,
      render: (v) => <Tag style={{ margin: 0 }}>{v}</Tag>,
    },
    {
      title: '指纹', dataIndex: 'fingerprint', width: 300, ellipsis: true,
      render: (v) => <span style={{ color: '#63656E', fontFamily: 'monospace', fontSize: 12 }}>{v}</span>,
    },
    { title: '备注', dataIndex: 'comment', width: 160, ellipsis: true, render: (v) => v || <span style={{ color: '#C4C6CC' }}>-</span> },
    { title: '创建时间', dataIndex: 'created_at', width: 170 },
    {
      title: '操作', width: 200, fixed: 'right',
      render: (_, record) => (
        <Space size={0}>
          <Tooltip title="部署到资产">
            <Button type="link" size="small" icon={<SendOutlined />} onClick={() => { setDeployKeyId(record.id); deployForm.resetFields(); setDeployOpen(true) }}>部署</Button>
          </Tooltip>
          <Tooltip title="轮换密钥">
            <Button type="link" size="small" icon={<SyncOutlined />} onClick={() => openRotateModal(record)}>轮换</Button>
          </Tooltip>
          <Dropdown menu={getMoreMenu(record)} trigger={['click']}>
            <Button type="link" size="small" icon={<MoreOutlined style={{ fontSize: 16 }} />} style={{ color: '#63656E' }} />
          </Dropdown>
        </Space>
      ),
    },
  ]

  // 绑定详情表格列
  const bindingColumns: ColumnsType<SSHKeyBinding> = [
    { title: '主机名', dataIndex: 'hostname', width: 150, render: (v, r) => v || r.ip },
    { title: 'IP', dataIndex: 'ip', width: 140 },
    { title: '用户', dataIndex: 'username', width: 100 },
    {
      title: '状态', dataIndex: 'status', width: 90,
      render: (v) => <Tag color={statusColors[v] || 'default'}>{statusTexts[v] || v}</Tag>,
    },
    { title: '部署时间', dataIndex: 'deployed_at', width: 170, render: (v) => v || '-' },
    {
      title: '操作', width: 80, fixed: 'right',
      render: (_, record) => record.status === 'deployed' ? (
        <Popconfirm
          title="确定撤销？"
          description={`将从 ${record.hostname || record.ip} 的 ${record.username} 用户移除此密钥`}
          onConfirm={() => handleRevokeBinding(record)}
        >
          <Button type="link" size="small" danger icon={<RollbackOutlined />} loading={revokeLoading}>撤销</Button>
        </Popconfirm>
      ) : null,
    },
  ]

  // 部署历史表格列
  const logColumns: ColumnsType<SSHKeyDeployLog> = [
    {
      title: '操作', dataIndex: 'action', width: 80,
      render: (v) => <Tag color={actionColors[v] || 'default'}>{actionTexts[v] || v}</Tag>,
    },
    { title: '主机名', dataIndex: 'hostname', width: 140, render: (v, r) => v || r.ip },
    { title: 'IP', dataIndex: 'ip', width: 130 },
    { title: '用户', dataIndex: 'username', width: 100 },
    {
      title: '结果', dataIndex: 'status', width: 80,
      render: (v) => <Tag color={v === 'success' ? 'green' : 'red'}>{v === 'success' ? '成功' : '失败'}</Tag>,
    },
    { title: '错误信息', dataIndex: 'error', width: 200, ellipsis: true, render: (v) => v || '-' },
    { title: '时间', dataIndex: 'created_at', width: 170 },
  ]

  const revokeResultSummary = (results: RevokeResponse[]) => {
    let totalSuccess = 0, totalFailed = 0
    for (const r of results) {
      totalSuccess += r.success
      totalFailed += r.failed
    }
    return { totalSuccess, totalFailed }
  }

  return (
    <>
      <Card title="SSH Key 管理" style={{ border: '1px solid #E7E9EF' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16, flexWrap: 'wrap', gap: 8 }}>
          <Space wrap>
            <Input
              placeholder="搜索名称 / 备注"
              prefix={<SearchOutlined style={{ color: '#979BA5' }} />}
              value={keyword}
              onChange={(e) => setKeyword(e.target.value)}
              onPressEnter={() => { setPage(1); fetchData(1, pageSize) }}
              style={{ width: 240 }}
              allowClear
              onClear={() => { setKeyword(''); fetchData(1, pageSize) }}
            />
            <Button type="primary" icon={<SearchOutlined />} onClick={() => { setPage(1); fetchData(1, pageSize) }}>搜索</Button>
            <Button icon={<ReloadOutlined />} onClick={() => { setKeyword(''); setPage(1); fetchData(1, pageSize) }}>重置</Button>
          </Space>
          <Space>
            <Button type="primary" icon={<PlusOutlined />} onClick={() => { setCreateMode('generate'); createForm.resetFields(); setCreateOpen(true) }}>创建密钥</Button>
            <Button icon={<UserDeleteOutlined />} danger onClick={async () => { departureForm.resetFields(); setDepartureResult(null); try { const keys = await getAllSSHKeys(); setAllKeys(keys) } catch { setAllKeys([]) } setDepartureOpen(true) }}>离职清理</Button>
          </Space>
        </div>
        <Table
          rowKey="id"
          columns={columns}
          dataSource={data}
          loading={loading}
          size="middle"
          scroll={{ x: 1130 }}
          pagination={{
            current: page,
            pageSize,
            total,
            showSizeChanger: true,
            showTotal: (t) => `共 ${t} 条`,
            onChange: (p, ps) => { setPage(p); setPageSize(ps); fetchData(p, ps) },
          }}
        />
      </Card>

      {/* 创建 SSH Key */}
      <Modal title="创建 SSH Key" open={createOpen} onOk={handleCreate} onCancel={() => setCreateOpen(false)} destroyOnHidden width={560}>
        <Form form={createForm} layout="vertical" style={{ marginTop: 16 }}>
          <Form.Item label="创建方式">
            <Radio.Group value={createMode} onChange={(e) => setCreateMode(e.target.value)}>
              <Radio.Button value="generate">自动生成</Radio.Button>
              <Radio.Button value="import">手动导入</Radio.Button>
            </Radio.Group>
          </Form.Item>
          <Form.Item name="name" label="名称" rules={[{ required: true, message: '请输入名称' }]}>
            <Input placeholder="如 prod-deploy-key" />
          </Form.Item>
          {createMode === 'generate' && (
            <Form.Item name="key_type" label="密钥类型" initialValue="ed25519">
              <Radio.Group>
                <Radio.Button value="ed25519">ED25519（推荐）</Radio.Button>
                <Radio.Button value="rsa">RSA 4096</Radio.Button>
              </Radio.Group>
            </Form.Item>
          )}
          <Form.Item name="comment" label="备注">
            <Input placeholder="备注信息（可选）" />
          </Form.Item>
          {createMode === 'import' && (
            <>
              <Form.Item name="public_key" label="公钥" rules={[{ required: true, message: '请输入公钥' }]}>
                <Input.TextArea rows={3} placeholder="ssh-ed25519 AAAA..." />
              </Form.Item>
              <Form.Item name="private_key" label="私钥" rules={[{ required: true, message: '请输入私钥' }]}>
                <Input.TextArea rows={5} placeholder="-----BEGIN OPENSSH PRIVATE KEY-----" />
              </Form.Item>
            </>
          )}
        </Form>
      </Modal>

      {/* 部署到资产 */}
      <Modal title="部署 SSH Key" open={deployOpen} onOk={handleDeploy} onCancel={() => setDeployOpen(false)} destroyOnHidden width={860}>
        <Form form={deployForm} layout="vertical" style={{ marginTop: 16 }}>
          <Form.Item name="asset_ids" label="目标资产" rules={[{ required: true, message: '请选择资产' }]}>
            <BatchAssetSelector maxHeight={300} />
          </Form.Item>
          <Form.Item name="username" label="目标用户" rules={[{ required: true, message: '请输入用户名' }]}>
            <Input placeholder="如 root, deploy" />
          </Form.Item>
        </Form>
      </Modal>

      {/* 绑定详情（表格形式） */}
      <Modal title="绑定详情" open={bindingsOpen} onCancel={() => setBindingsOpen(false)} footer={null} width={860}>
        <Table
          rowKey="id"
          columns={bindingColumns}
          dataSource={bindings}
          loading={bindingsLoading}
          size="middle"
          scroll={{ x: 730 }}
          locale={{ emptyText: '暂无绑定记录' }}
          pagination={bindings.length > 10 ? { pageSize: 10, showSizeChanger: true, showTotal: (t) => `共 ${t} 条` } : false}
        />
      </Modal>

      {/* SSH Key 轮换 */}
      <Modal
        title={`轮换密钥: ${rotateKeyName}`}
        open={rotateOpen}
        onOk={handleRotate}
        onCancel={() => setRotateOpen(false)}
        confirmLoading={rotateLoading}
        destroyOnHidden
        width={520}
      >
        <Alert
          type="info"
          showIcon
          style={{ marginBottom: 16 }}
          message="轮换流程"
          description="选择或生成新密钥 → 部署到所有已绑定资产 → 撤销旧密钥。整个过程自动执行，撤销将在部署完成后进行。"
        />
        <Form form={rotateForm} layout="vertical" style={{ marginTop: 16 }}>
          <Form.Item label="替换方式">
            <Radio.Group value={rotateMode} onChange={(e) => { setRotateMode(e.target.value); rotateForm.resetFields() }}>
              <Radio.Button value="select">选择已有密钥</Radio.Button>
              <Radio.Button value="generate">自动生成新密钥</Radio.Button>
            </Radio.Group>
          </Form.Item>
          {rotateMode === 'select' ? (
            <Form.Item name="new_key_id" label="选择替换密钥" rules={[{ required: true, message: '请选择替换密钥' }]}>
              <Select
                placeholder="请选择一个已有密钥"
                showSearch
                optionFilterProp="label"
                options={allKeys.map(k => ({
                  value: k.id,
                  label: `${k.name}（${k.key_type}）`,
                }))}
              />
            </Form.Item>
          ) : (
            <>
              <Form.Item name="new_name" label="新密钥名称">
                <Input placeholder="留空则自动生成（原名称-rotated-日期）" />
              </Form.Item>
              <Form.Item name="new_comment" label="新密钥备注">
                <Input placeholder="备注信息" />
              </Form.Item>
            </>
          )}
        </Form>
      </Modal>

      {/* 部署历史 */}
      <Modal
        title={`操作历史: ${logsKeyName}`}
        open={logsOpen}
        onCancel={() => setLogsOpen(false)}
        footer={null}
        width={960}
      >
        <Table
          rowKey="id"
          columns={logColumns}
          dataSource={logs}
          loading={logsLoading}
          size="middle"
          scroll={{ x: 900 }}
          locale={{ emptyText: '暂无操作记录' }}
          pagination={{
            current: logsPage,
            pageSize: logsPageSize,
            total: logsTotal,
            showSizeChanger: true,
            showTotal: (t) => `共 ${t} 条`,
            onChange: (p, ps) => showDeployLogs(logsKeyId, logsKeyName, p, ps),
          }}
        />
      </Modal>

      {/* 离职清理 */}
      <Modal
        title="离职清理"
        open={departureOpen}
        onCancel={() => setDepartureOpen(false)}
        footer={departureResult ? <Button onClick={() => setDepartureOpen(false)}>关闭</Button> : undefined}
        onOk={handleDepartureCleanup}
        confirmLoading={departureLoading}
        destroyOnHidden
        width={640}
      >
        <Alert
          type="warning"
          showIcon
          style={{ marginBottom: 16 }}
          message="离职清理将撤销选中密钥在所有（或指定）资产上的部署"
          description="此操作会 SSH 到各主机并从 authorized_keys 中移除对应公钥，请确认操作对象。"
        />
        {!departureResult ? (
          <Form form={departureForm} layout="vertical" style={{ marginTop: 16 }}>
            <Form.Item name="key_ids" label="选择要清理的密钥" rules={[{ required: true, message: '请选择至少一个密钥' }]}>
              <Select
                mode="multiple"
                placeholder="选择离职人员对应的 SSH Key"
                options={allKeys.map(k => ({ label: `${k.name}（${k.fingerprint?.slice(-12) || k.key_type}）`, value: k.id }))}
                optionFilterProp="label"
                style={{ width: '100%' }}
              />
            </Form.Item>
            <Form.Item name="asset_ids" label="指定资产（可选，留空则清理所有已绑定资产）">
              <BatchAssetSelector maxHeight={240} />
            </Form.Item>
          </Form>
        ) : (
          <div>
            <Descriptions column={2} size="small" bordered style={{ marginBottom: 16 }}>
              <Descriptions.Item label="涉及密钥">{departureResult.total_keys} 个</Descriptions.Item>
              <Descriptions.Item label="涉及资产">{departureResult.total_assets} 台</Descriptions.Item>
            </Descriptions>
            {departureResult.revoke_results.map((r, idx) => {
              const { totalSuccess, totalFailed } = revokeResultSummary([r])
              return (
                <div key={idx} style={{ marginBottom: 8 }}>
                  <Space>
                    <Tag color="green">{totalSuccess} 成功</Tag>
                    {totalFailed > 0 && <Tag color="red">{totalFailed} 失败</Tag>}
                  </Space>
                  {r.results.filter(item => item.status === 'failed').map((item, i) => (
                    <div key={i} style={{ color: '#EA3636', fontSize: 12, marginTop: 4 }}>
                      {item.hostname || item.ip}: {item.error}
                    </div>
                  ))}
                </div>
              )
            })}
          </div>
        )}
      </Modal>
    </>
  )
}

export default SSHKeysPage
