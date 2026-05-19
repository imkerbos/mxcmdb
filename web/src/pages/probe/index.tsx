import { useState } from 'react'
import { Card, Button, Space, message } from 'antd'
import { ThunderboltOutlined } from '@ant-design/icons'
import { executeProbe } from '../../services/probe'
import BatchAssetSelector from '../../components/BatchAssetSelector'

const ProbePage = () => {
  const [selectedAssets, setSelectedAssets] = useState<number[]>([])
  const [executing, setExecuting] = useState(false)

  const handleExecute = async () => {
    if (selectedAssets.length === 0) {
      message.warning('请选择要探测的资产')
      return
    }
    setExecuting(true)
    try {
      await executeProbe(selectedAssets)
      message.success('探针任务已提交，请稍后查看结果')
    } catch { /* handled */ } finally { setExecuting(false) }
  }

  return (
    <Card title="资产探针" style={{ border: '1px solid #E7E9EF' }}>
      <Space direction="vertical" style={{ width: '100%' }} size="middle">
        <BatchAssetSelector value={selectedAssets} onChange={setSelectedAssets} maxHeight={460} />
        <Button type="primary" icon={<ThunderboltOutlined />} onClick={handleExecute} loading={executing} disabled={selectedAssets.length === 0}>
          执行探测（{selectedAssets.length} 台）
        </Button>
      </Space>
    </Card>
  )
}

export default ProbePage
