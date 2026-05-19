import { useEffect, useState } from 'react'
import { Select, Tag } from 'antd'
import { getAllAssets, type AssetSimple } from '../../services/asset'

interface AssetSelectProps {
  value?: number[]
  onChange?: (value: number[]) => void
  mode?: 'multiple' | 'tags'
  placeholder?: string
  style?: React.CSSProperties
}

const AssetSelect = ({ value, onChange, mode = 'multiple', placeholder = '选择资产', style }: AssetSelectProps) => {
  const [assets, setAssets] = useState<AssetSimple[]>([])
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    setLoading(true)
    getAllAssets()
      .then(setAssets)
      .catch(() => {})
      .finally(() => setLoading(false))
  }, [])

  return (
    <Select
      mode={mode}
      value={value}
      onChange={onChange}
      placeholder={placeholder}
      style={style}
      loading={loading}
      showSearch
      filterOption={(input, option) =>
        (option?.label as string)?.toLowerCase().includes(input.toLowerCase()) ?? false
      }
      optionFilterProp="label"
      options={assets.map((a) => ({
        value: a.id,
        label: `${a.hostname || a.ip} (${a.ip})`,
      }))}
      tagRender={({ label, closable, onClose }) => (
        <Tag closable={closable} onClose={onClose} style={{ marginInlineEnd: 4 }}>
          {label}
        </Tag>
      )}
    />
  )
}

export default AssetSelect
