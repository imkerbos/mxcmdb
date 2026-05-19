import { useState } from 'react'
import { Card, Button, Form, Input, Upload, Space, message } from 'antd'
import { UploadOutlined, SendOutlined } from '@ant-design/icons'
import type { UploadFile } from 'antd'
import { distributeFile } from '../../../services/file-distribution'
import BatchAssetSelector from '../../../components/BatchAssetSelector'

const FileDistributePage = () => {
  const [form] = Form.useForm()
  const [fileList, setFileList] = useState<UploadFile[]>([])
  const [submitting, setSubmitting] = useState(false)

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields()
      if (fileList.length === 0) {
        message.warning('请选择要分发的文件')
        return
      }

      setSubmitting(true)
      const formData = new FormData()
      formData.append('file', fileList[0].originFileObj as File)
      formData.append('data', JSON.stringify({
        name: values.name,
        remote_path: values.remote_path,
        file_mode: values.file_mode || '0644',
        asset_ids: values.asset_ids,
      }))

      await distributeFile(formData)
      message.success('文件分发任务已提交')
      form.resetFields()
      setFileList([])
    } catch { /* validation */ } finally { setSubmitting(false) }
  }

  return (
    <Card title="文件下发" style={{ border: '1px solid #E7E9EF' }}>
      <Form form={form} layout="vertical" style={{ maxWidth: 600 }}>
        <Form.Item name="name" label="任务名称" rules={[{ required: true, message: '请输入任务名称' }]}>
          <Input placeholder="如：同步 Nginx 配置" />
        </Form.Item>
        <Form.Item label="选择文件" required>
          <Upload
            fileList={fileList}
            beforeUpload={() => false}
            onChange={({ fileList: fl }) => setFileList(fl.slice(-1))}
            maxCount={1}
          >
            <Button icon={<UploadOutlined />}>选择文件</Button>
          </Upload>
        </Form.Item>
        <Form.Item name="remote_path" label="远程路径" rules={[{ required: true, message: '请输入远程路径' }]}>
          <Input placeholder="如：/etc/nginx/nginx.conf" />
        </Form.Item>
        <Form.Item name="file_mode" label="文件权限">
          <Input placeholder="0644" />
        </Form.Item>
        <Form.Item name="asset_ids" label="目标资产" rules={[{ required: true, message: '请选择资产' }]}>
          <BatchAssetSelector />
        </Form.Item>
        <Form.Item>
          <Space>
            <Button type="primary" icon={<SendOutlined />} onClick={handleSubmit} loading={submitting}>
              开始分发
            </Button>
          </Space>
        </Form.Item>
      </Form>
    </Card>
  )
}

export default FileDistributePage
