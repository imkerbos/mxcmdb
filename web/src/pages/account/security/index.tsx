import { useState, useEffect } from 'react'
import { Card, Button, Steps, Input, Alert, Tag, Space, message, Modal } from 'antd'
import { SafetyCertificateOutlined, CheckCircleOutlined, QrcodeOutlined } from '@ant-design/icons'
import { getUserInfo, setupMFA, bindMFA } from '../../../services/auth'
import QRCode from 'qrcode'

const AccountSecurityPage = () => {
  const [mfaEnabled, setMfaEnabled] = useState(false)
  const [loading, setLoading] = useState(true)
  const [setupOpen, setSetupOpen] = useState(false)
  const [step, setStep] = useState(0)
  const [secret, setSecret] = useState('')
  const [qrImage, setQrImage] = useState('')
  const [code, setCode] = useState('')
  const [binding, setBinding] = useState(false)

  const fetchStatus = async () => {
    setLoading(true)
    try {
      const info = await getUserInfo()
      setMfaEnabled(info.mfa_enabled)
    } catch { /* handled */ } finally { setLoading(false) }
  }

  useEffect(() => { fetchStatus() }, [])

  const handleSetup = async () => {
    try {
      const result = await setupMFA()
      setSecret(result.secret)
      const dataUrl = await QRCode.toDataURL(result.url, { width: 200, margin: 2 })
      setQrImage(dataUrl)
      setStep(0)
      setCode('')
      setSetupOpen(true)
    } catch {
      message.error('生成 MFA 密钥失败')
    }
  }

  const handleBind = async () => {
    if (code.length !== 6) {
      message.warning('请输入 6 位验证码')
      return
    }
    setBinding(true)
    try {
      await bindMFA(secret, code)
      message.success('MFA 绑定成功')
      setSetupOpen(false)
      setMfaEnabled(true)
    } catch { /* handled by interceptor */ } finally { setBinding(false) }
  }

  return (
    <>
      <Card title="账号安全" style={{ border: '1px solid #E7E9EF' }} loading={loading}>
        <div style={{ maxWidth: 600 }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', padding: '16px 0', borderBottom: '1px solid #f0f0f0' }}>
            <div>
              <Space>
                <SafetyCertificateOutlined style={{ fontSize: 20, color: '#3A84FF' }} />
                <span style={{ fontSize: 16, fontWeight: 500 }}>MFA 两步验证</span>
                {mfaEnabled
                  ? <Tag icon={<CheckCircleOutlined />} color="success">已启用</Tag>
                  : <Tag color="default">未启用</Tag>
                }
              </Space>
              <div style={{ color: '#8c8c8c', fontSize: 13, marginTop: 4, marginLeft: 28 }}>
                启用后，登录时需输入身份验证器应用生成的动态验证码
              </div>
            </div>
            {!mfaEnabled && (
              <Button type="primary" onClick={handleSetup}>启用 MFA</Button>
            )}
          </div>
        </div>
      </Card>

      <Modal
        title="启用 MFA 两步验证"
        open={setupOpen}
        onCancel={() => setSetupOpen(false)}
        footer={null}
        width={480}
        destroyOnHidden
      >
        <Steps current={step} size="small" style={{ marginBottom: 24, marginTop: 16 }} items={[
          { title: '扫描二维码' },
          { title: '验证绑定' },
        ]} />

        {step === 0 && (
          <div style={{ textAlign: 'center' }}>
            <p style={{ color: '#595959' }}>使用 Google Authenticator、Microsoft Authenticator 或其他 TOTP 应用扫描下方二维码：</p>
            {qrImage && <img src={qrImage} alt="MFA QR Code" style={{ margin: '16px 0' }} />}
            <Alert
              type="info"
              showIcon
              icon={<QrcodeOutlined />}
              message="无法扫码？"
              description={<>手动输入密钥：<code style={{ userSelect: 'all', fontWeight: 600 }}>{secret}</code></>}
              style={{ textAlign: 'left', marginTop: 8 }}
            />
            <Button type="primary" style={{ marginTop: 20 }} onClick={() => setStep(1)}>下一步</Button>
          </div>
        )}

        {step === 1 && (
          <div style={{ textAlign: 'center' }}>
            <p style={{ color: '#595959' }}>请输入身份验证器应用显示的 6 位验证码：</p>
            <Input
              value={code}
              onChange={(e) => setCode(e.target.value.replace(/\D/g, '').slice(0, 6))}
              placeholder="000000"
              maxLength={6}
              style={{ width: 200, fontSize: 24, textAlign: 'center', letterSpacing: 8, margin: '16px 0' }}
              onPressEnter={handleBind}
            />
            <div>
              <Space>
                <Button onClick={() => setStep(0)}>上一步</Button>
                <Button type="primary" loading={binding} onClick={handleBind} disabled={code.length !== 6}>确认绑定</Button>
              </Space>
            </div>
          </div>
        )}
      </Modal>
    </>
  )
}

export default AccountSecurityPage
