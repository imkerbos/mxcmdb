import { useEffect, useState } from 'react'
import { Card, Form, Input, Button, message, Spin, Descriptions, Space, Typography } from 'antd'
import { SaveOutlined, ReloadOutlined } from '@ant-design/icons'
import { getConfigs, batchUpdateConfigs, type SystemConfig } from '../../../services/settings'

const categoryLabels: Record<string, string> = {
  ssh: 'SSH 配置',
  terminal: '终端配置',
  file: '文件配置',
}

const GeneralSettingsPage = () => {
  const [configs, setConfigs] = useState<SystemConfig[]>([])
  const [loading, setLoading] = useState(false)
  const [saving, setSaving] = useState(false)
  const [form] = Form.useForm()

  const fetchConfigs = async () => {
    setLoading(true)
    try {
      const data = await getConfigs()
      // 只显示非安全类的配置
      const filtered = data.filter((c) => c.category !== 'security')
      setConfigs(filtered)
      const values: Record<string, string> = {}
      filtered.forEach((c) => {
        values[c.key] = c.value
      })
      form.setFieldsValue(values)
    } catch {
      // error handled by interceptor
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchConfigs()
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

  // 按 category 分组
  const grouped = configs.reduce<Record<string, SystemConfig[]>>((acc, c) => {
    if (!acc[c.category]) acc[c.category] = []
    acc[c.category].push(c)
    return acc
  }, {})

  return (
    <Spin spinning={loading}>
      <Form form={form} layout="vertical">
        {Object.entries(grouped).map(([category, items]) => (
          <Card
            key={category}
            title={categoryLabels[category] || category}
            style={{ marginBottom: 16, border: '1px solid #E7E9EF' }}
          >
            <Descriptions column={1} bordered size="small">
              {items.map((item) => (
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
                    <Input style={{ width: 300 }} />
                  </Form.Item>
                </Descriptions.Item>
              ))}
            </Descriptions>
          </Card>
        ))}

        <Card style={{ border: '1px solid #E7E9EF' }}>
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
    </Spin>
  )
}

export default GeneralSettingsPage
