import { useState, useRef, useEffect } from 'react'
import { Card, Button, Form, Input, Space, Tag, Table, message, Collapse, Alert } from 'antd'
import { ThunderboltOutlined } from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import { executeTask, type TaskResultItem } from '../../../services/task'
import { connectTaskWS, type ManagedWS, type WSState } from '../../../utils/websocket'
import BatchAssetSelector from '../../../components/BatchAssetSelector'

const TaskExecutePage = () => {
  const [form] = Form.useForm()
  const [executing, setExecuting] = useState(false)
  const [results, setResults] = useState<TaskResultItem[]>([])
  const [taskStatus, setTaskStatus] = useState<string>('')
  const [wsState, setWsState] = useState<WSState | null>(null)
  const wsRef = useRef<ManagedWS | null>(null)

  useEffect(() => {
    return () => { wsRef.current?.close() }
  }, [])

  const handleExecute = async () => {
    try {
      const values = await form.validateFields()
      setExecuting(true)
      setResults([])
      setTaskStatus('running')

      const task = await executeTask({
        name: values.name,
        command: values.command,
        asset_ids: values.asset_ids,
      })

      // 连接 WebSocket 接收实时结果
      const ws = connectTaskWS(task.id, (msg) => {
        if (msg.type === 'task_result') {
          const result = msg.data as TaskResultItem
          setResults((prev) => [...prev, result])
        } else if (msg.type === 'task_done') {
          const data = msg.data as { status: string; success_count: number; fail_count: number }
          setTaskStatus(data.status)
          setExecuting(false)
          message.success(`执行完成：成功 ${data.success_count}，失败 ${data.fail_count}`)
          ws?.close()
        }
      }, setWsState)
      if (!ws) {
        message.error('未登录或连接失败')
        setExecuting(false)
        return
      }
      wsRef.current = ws
    } catch {
      setExecuting(false)
    }
  }

  const statusColors: Record<string, string> = { success: 'green', failed: 'red', pending: 'gold', running: 'processing' }

  const columns: ColumnsType<TaskResultItem> = [
    { title: '主机名', dataIndex: 'hostname', width: 140, ellipsis: true },
    { title: 'IP', dataIndex: 'ip', width: 130 },
    {
      title: '状态', dataIndex: 'status', width: 80,
      render: (v) => <Tag color={statusColors[v] || 'default'}>{v}</Tag>,
    },
    { title: '退出码', dataIndex: 'exit_code', width: 80 },
    { title: '耗时(ms)', dataIndex: 'duration', width: 90 },
  ]

  return (
    <Space direction="vertical" style={{ width: '100%' }} size="middle">
      <Card title="命令执行" style={{ border: '1px solid #E7E9EF' }}>
        <Form form={form} layout="vertical">
          <Form.Item name="name" label="任务名称" rules={[{ required: true, message: '请输入任务名称' }]}>
            <Input placeholder="如：检查磁盘使用率" />
          </Form.Item>
          <Form.Item name="command" label="命令" rules={[{ required: true, message: '请输入命令' }]}>
            <Input.TextArea rows={3} placeholder="输入要执行的 Shell 命令" />
          </Form.Item>
          <Form.Item name="asset_ids" label="目标资产" rules={[{ required: true, message: '请选择资产' }]}>
            <BatchAssetSelector />
          </Form.Item>
          <Form.Item>
            <Button type="primary" icon={<ThunderboltOutlined />} onClick={handleExecute} loading={executing}>
              执行
            </Button>
          </Form.Item>
        </Form>
      </Card>

      {(results.length > 0 || taskStatus) && (
        <Card
          title={
            <Space>
              执行结果
              {taskStatus && <Tag color={statusColors[taskStatus] || 'default'}>{taskStatus}</Tag>}
            </Space>
          }
          style={{ border: '1px solid #E7E9EF' }}
        >
          {wsState === 'reconnecting' && (
            <Alert type="warning" message="连接中断，正在尝试重连..." showIcon style={{ marginBottom: 16 }} />
          )}
          {wsState === 'disconnected' && executing && (
            <Alert type="error" message="连接已断开，无法接收实时结果" showIcon style={{ marginBottom: 16 }} />
          )}
          <Table
            rowKey="id"
            columns={columns}
            dataSource={results}
            pagination={false}
            size="middle"
            expandable={{
              expandedRowRender: (record) => (
                <Collapse
                  size="small"
                  items={[
                    ...(record.stdout ? [{ key: 'stdout', label: 'Stdout', children: <pre style={{ maxHeight: 200, overflow: 'auto', fontSize: 12, margin: 0 }}>{record.stdout}</pre> }] : []),
                    ...(record.stderr ? [{ key: 'stderr', label: 'Stderr', children: <pre style={{ maxHeight: 200, overflow: 'auto', fontSize: 12, margin: 0, color: '#EA3636' }}>{record.stderr}</pre> }] : []),
                  ]}
                  defaultActiveKey={['stdout']}
                />
              ),
            }}
          />
        </Card>
      )}
    </Space>
  )
}

export default TaskExecutePage
