import { useEffect, useState, useCallback } from 'react'
import { LockOutlined, UserOutlined, SafetyOutlined, KeyOutlined } from '@ant-design/icons'
import { Button, Form, Input, message } from 'antd'
import { useNavigate } from 'react-router-dom'
import { login, getCaptcha, type CaptchaResult } from '../../services/auth'
import logo from '../../assets/logo.png'

interface LoginFormValues {
  username: string
  password: string
  captcha: string
  mfa_code?: string
}

const LoginPage = () => {
  const navigate = useNavigate()
  const [form] = Form.useForm()
  const [captcha, setCaptcha] = useState<CaptchaResult | null>(null)
  const [requireMFA, setRequireMFA] = useState(false)
  const [pendingValues, setPendingValues] = useState<LoginFormValues | null>(null)
  const [loading, setLoading] = useState(false)

  const refreshCaptcha = useCallback(async () => {
    try {
      const result = await getCaptcha()
      setCaptcha(result)
    } catch {
      // handled by interceptor
    }
  }, [])

  useEffect(() => {
    refreshCaptcha()
  }, [refreshCaptcha])

  const handleSubmit = async (values: LoginFormValues) => {
    if (!captcha) return
    setLoading(true)

    try {
      const result = await login({
        username: pendingValues?.username || values.username,
        password: pendingValues?.password || values.password,
        captcha_id: captcha.captcha_id,
        captcha: values.captcha,
        mfa_code: values.mfa_code,
      })

      if (result.require_mfa && !values.mfa_code) {
        setRequireMFA(true)
        setPendingValues(values)
        refreshCaptcha()
        message.info('请输入 MFA 验证码')
        setLoading(false)
        return
      }

      localStorage.setItem('access_token', result.access_token)
      localStorage.setItem('refresh_token', result.refresh_token)

      if (result.require_mfa_setup) {
        message.warning('系统要求启用 MFA 两步验证，请先完成绑定')
        navigate('/account/security')
        return
      }

      message.success('登录成功')
      navigate('/')
    } catch {
      refreshCaptcha()
      form.setFieldValue('captcha', '')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div
      style={{
        display: 'flex',
        justifyContent: 'center',
        alignItems: 'center',
        minHeight: '100vh',
        background: 'linear-gradient(160deg, #0a1628 0%, #1a2a4a 40%, #1e3a5f 70%, #0d2137 100%)',
        position: 'relative',
        overflow: 'hidden',
      }}
    >
      {/* 背景装饰 */}
      <div
        style={{
          position: 'absolute',
          width: 600,
          height: 600,
          borderRadius: '50%',
          background: 'radial-gradient(circle, rgba(30,120,200,0.08) 0%, transparent 70%)',
          top: -200,
          right: -100,
        }}
      />
      <div
        style={{
          position: 'absolute',
          width: 400,
          height: 400,
          borderRadius: '50%',
          background: 'radial-gradient(circle, rgba(0,180,220,0.06) 0%, transparent 70%)',
          bottom: -100,
          left: -50,
        }}
      />

      <div
        style={{
          width: 420,
          background: 'rgba(255, 255, 255, 0.97)',
          borderRadius: 4,
          boxShadow: '0 20px 60px rgba(0, 0, 0, 0.3), 0 0 40px rgba(30, 120, 200, 0.1)',
          padding: '48px 40px 36px',
          position: 'relative',
          zIndex: 1,
          animation: 'loginCardIn 0.5s ease-out',
        }}
      >
        {/* Logo */}
        <div style={{ textAlign: 'center', marginBottom: 12 }}>
          <img src={logo} alt="MXCMDB" style={{ height: 52 }} />
        </div>
        <div
          style={{
            textAlign: 'center',
            marginBottom: 36,
            color: '#8c8c8c',
            fontSize: 14,
            letterSpacing: 1,
          }}
        >
          基础设施管理与运维控制平台
        </div>

        <Form form={form} onFinish={handleSubmit} size="large" layout="vertical">
          <Form.Item name="username" rules={[{ required: true, message: '请输入用户名' }]}>
            <Input
              prefix={<UserOutlined style={{ color: '#1e78c8' }} />}
              placeholder="用户名"
              style={{ borderRadius: 8, height: 44 }}
            />
          </Form.Item>

          <Form.Item name="password" rules={[{ required: true, message: '请输入密码' }]}>
            <Input.Password
              prefix={<LockOutlined style={{ color: '#1e78c8' }} />}
              placeholder="密码"
              style={{ borderRadius: 8, height: 44 }}
            />
          </Form.Item>

          {/* 验证码 */}
          <Form.Item>
            <div style={{ display: 'flex', gap: 12 }}>
              <Form.Item
                name="captcha"
                noStyle
                rules={[{ required: true, message: '请输入验证码' }]}
              >
                <Input
                  prefix={<SafetyOutlined style={{ color: '#1e78c8' }} />}
                  placeholder="验证码"
                  style={{ borderRadius: 8, height: 44 }}
                />
              </Form.Item>
              {captcha && (
                <img
                  src={captcha.captcha_image}
                  alt="验证码"
                  onClick={refreshCaptcha}
                  style={{
                    height: 44,
                    cursor: 'pointer',
                    borderRadius: 8,
                    border: '1px solid #d9d9d9',
                    flexShrink: 0,
                  }}
                  title="点击刷新验证码"
                />
              )}
            </div>
          </Form.Item>

          {/* MFA */}
          {requireMFA && (
            <Form.Item name="mfa_code" rules={[{ required: true, message: '请输入 MFA 验证码' }]}>
              <Input
                prefix={<KeyOutlined style={{ color: '#1e78c8' }} />}
                placeholder="MFA 验证码（6位数字）"
                maxLength={6}
                style={{ borderRadius: 8, height: 44 }}
              />
            </Form.Item>
          )}

          <Form.Item style={{ marginBottom: 16 }}>
            <Button
              type="primary"
              htmlType="submit"
              block
              loading={loading}
              style={{
                height: 46,
                borderRadius: 8,
                fontSize: 16,
                fontWeight: 500,
                background: 'linear-gradient(135deg, #1e78c8 0%, #0ea5c0 100%)',
                border: 'none',
                boxShadow: '0 4px 12px rgba(30, 120, 200, 0.35)',
              }}
            >
              登 录
            </Button>
          </Form.Item>
        </Form>

        <div
          style={{
            textAlign: 'center',
            color: '#bfbfbf',
            fontSize: 12,
          }}
        >
          MXCMDB v1.0 &copy; {new Date().getFullYear()}
        </div>
      </div>
    </div>
  )
}

export default LoginPage
