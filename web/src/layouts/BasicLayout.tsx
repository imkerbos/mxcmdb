import { useEffect, useState, useCallback, useRef } from 'react'
import { Outlet, useNavigate, useLocation } from 'react-router-dom'
import { ProLayout } from '@ant-design/pro-components'
import type { MenuDataItem } from '@ant-design/pro-components'
import {
  DashboardOutlined,
  CloudServerOutlined,
  CloudOutlined,
  DatabaseOutlined,
  TagsOutlined,
  ApartmentOutlined,
  KeyOutlined,
  SafetyCertificateOutlined,
  UserOutlined,
  LinuxOutlined,
  CodeOutlined,
  FileOutlined,
  SettingOutlined,
  CodeSandboxOutlined,
  AuditOutlined,
  LoginOutlined,
  FileSearchOutlined,
  HistoryOutlined,
  GatewayOutlined,
  NodeIndexOutlined,
  RadarChartOutlined,
  DesktopOutlined,
  LogoutOutlined,
  ToolOutlined,
  BookOutlined,
  VideoCameraOutlined,
  BellOutlined,
  CheckOutlined,
  SolutionOutlined,
  InfoCircleOutlined,
  CheckCircleOutlined,
  WarningOutlined,
  CloseCircleOutlined,
} from '@ant-design/icons'
import { Dropdown, Badge, List, Button, Empty, Spin, Tag } from 'antd'
import { getUserInfo, type UserInfo } from '../services/auth'
import { getNotifications, getUnreadCount, markRead, markAllRead, type NotificationItem } from '../services/notification'
import logo from '../assets/logo.png'

const menuData: MenuDataItem[] = [
  {
    path: '/dashboard',
    name: '仪表盘',
    icon: <DashboardOutlined />,
  },
  {
    path: '/assets',
    name: '资产管理',
    icon: <CloudServerOutlined />,
    children: [
      { path: '/assets/cloud', name: '云资源', icon: <CloudOutlined /> },
      { path: '/assets/idc', name: 'IDC 资产', icon: <DatabaseOutlined /> },
      { path: '/assets/tags', name: '资产标签', icon: <TagsOutlined /> },
      { path: '/assets/ownership', name: '资产归属', icon: <ApartmentOutlined /> },
      { path: '/assets/init', name: '资产初始化', icon: <SettingOutlined /> },
      { path: '/cloud-accounts', name: '云账号托管', icon: <SafetyCertificateOutlined /> },
      { path: '/probe', name: '资产探针', icon: <RadarChartOutlined /> },
      { path: '/ipam/subnets', name: '网段管理', icon: <GatewayOutlined /> },
      { path: '/ipam/addresses', name: 'IP 分配', icon: <NodeIndexOutlined /> },
    ],
  },
  {
    path: '/ops',
    name: '运维操作',
    icon: <ToolOutlined />,
    children: [
      { path: '/tasks/execute', name: '命令执行', icon: <CodeSandboxOutlined /> },
      { path: '/tasks/history', name: '执行历史', icon: <HistoryOutlined /> },
      { path: '/files/distribute', name: '文件下发', icon: <FileOutlined /> },
      { path: '/files/history', name: '分发记录', icon: <FileSearchOutlined /> },
      { path: '/terminal', name: 'Web Terminal', icon: <DesktopOutlined /> },
      { path: '/terminal/sessions', name: '终端会话', icon: <VideoCameraOutlined /> },
      { path: '/sshkeys', name: 'SSH Key 管理', icon: <KeyOutlined /> },
      { path: '/playbooks', name: '运维剧本', icon: <BookOutlined /> },
    ],
  },
  {
    path: '/projects',
    name: '项目管理',
    icon: <ApartmentOutlined />,
  },
  {
    path: '/approvals',
    name: '审批管理',
    icon: <SolutionOutlined />,
  },
  {
    path: '/audit',
    name: '审计中心',
    icon: <AuditOutlined />,
    children: [
      { path: '/audit/login', name: '登录审计', icon: <LoginOutlined /> },
      { path: '/audit/operation', name: '操作审计', icon: <FileSearchOutlined /> },
      { path: '/audit/command', name: '命令审计', icon: <CodeOutlined /> },
    ],
  },
  {
    path: '/system',
    name: '系统管理',
    icon: <SettingOutlined />,
    children: [
      { path: '/users/platform', name: '平台用户', icon: <UserOutlined /> },
      { path: '/users/linux', name: 'Linux 用户', icon: <LinuxOutlined /> },
      { path: '/settings/general', name: '基础配置', icon: <ToolOutlined /> },
      { path: '/settings/security', name: '安全设置', icon: <SafetyCertificateOutlined /> },
      { path: '/settings/permissions', name: '权限管理', icon: <KeyOutlined /> },
    ],
  },
]

const typeIcons: Record<string, React.ReactNode> = {
  info: <InfoCircleOutlined style={{ color: '#3A84FF' }} />,
  success: <CheckCircleOutlined style={{ color: '#2DCB56' }} />,
  warning: <WarningOutlined style={{ color: '#FF9C01' }} />,
  error: <CloseCircleOutlined style={{ color: '#EA3636' }} />,
}

const typeColors: Record<string, string> = {
  info: 'blue', success: 'green', warning: 'orange', error: 'red',
}

const BasicLayout = () => {
  const navigate = useNavigate()
  const location = useLocation()
  const [userInfo, setUserInfo] = useState<UserInfo | null>(null)

  // 通知状态
  const [unreadCount, setUnreadCount] = useState(0)
  const [notifyOpen, setNotifyOpen] = useState(false)
  const [notifications, setNotifications] = useState<NotificationItem[]>([])
  const [notifyLoading, setNotifyLoading] = useState(false)
  const pollRef = useRef<ReturnType<typeof setInterval> | null>(null)

  const fetchUnreadCount = useCallback(() => {
    getUnreadCount().then(r => setUnreadCount(r.count)).catch(() => {})
  }, [])

  const fetchNotifications = useCallback(async () => {
    setNotifyLoading(true)
    try {
      const r = await getNotifications({ page: 1, page_size: 10 })
      setNotifications(r.list || [])
    } catch { /* handled */ } finally { setNotifyLoading(false) }
  }, [])

  useEffect(() => {
    const token = localStorage.getItem('access_token')
    if (!token) {
      navigate('/login')
      return
    }
    getUserInfo()
      .then(setUserInfo)
      .catch(() => navigate('/login'))

    // 轮询未读数
    fetchUnreadCount()
    pollRef.current = setInterval(fetchUnreadCount, 30000)
    return () => {
      if (pollRef.current) clearInterval(pollRef.current)
    }
  }, [navigate, fetchUnreadCount])

  const handleLogout = () => {
    localStorage.removeItem('access_token')
    localStorage.removeItem('refresh_token')
    if (pollRef.current) clearInterval(pollRef.current)
    navigate('/login')
  }

  const handleNotifyOpen = (open: boolean) => {
    setNotifyOpen(open)
    if (open) fetchNotifications()
  }

  const handleMarkRead = async (id: number) => {
    await markRead(id)
    setNotifications(prev => prev.map(n => n.id === id ? { ...n, read: true } : n))
    setUnreadCount(prev => Math.max(0, prev - 1))
  }

  const handleMarkAllRead = async () => {
    await markAllRead()
    setNotifications(prev => prev.map(n => ({ ...n, read: true })))
    setUnreadCount(0)
  }

  const notificationDropdown = (
    <div style={{
      width: 380,
      background: '#fff',
      borderRadius: 2,
      boxShadow: '0 4px 12px rgba(0,0,0,0.15)',
    }}>
      <div style={{
        padding: '12px 16px',
        borderBottom: '1px solid #E7E9EF',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'space-between',
      }}>
        <span style={{ fontSize: 14, fontWeight: 600, color: '#313238' }}>
          通知 {unreadCount > 0 && <Tag color="blue" style={{ marginLeft: 6 }}>{unreadCount} 条未读</Tag>}
        </span>
        {unreadCount > 0 && (
          <Button type="link" size="small" icon={<CheckOutlined />} onClick={handleMarkAllRead}>
            全部已读
          </Button>
        )}
      </div>
      <div style={{ maxHeight: 400, overflow: 'auto' }}>
        {notifyLoading ? (
          <div style={{ padding: '40px 0', textAlign: 'center' }}><Spin /></div>
        ) : notifications.length === 0 ? (
          <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无通知" style={{ padding: '40px 0' }} />
        ) : (
          <List
            dataSource={notifications}
            renderItem={item => (
              <div
                key={item.id}
                onClick={() => { if (!item.read) handleMarkRead(item.id) }}
                style={{
                  padding: '12px 16px',
                  cursor: item.read ? 'default' : 'pointer',
                  borderBottom: '1px solid #F0F1F5',
                  background: item.read ? 'transparent' : '#F0F5FF',
                  transition: 'background 0.15s ease',
                }}
              >
                <div style={{ display: 'flex', gap: 10, alignItems: 'flex-start' }}>
                  <span style={{ marginTop: 2 }}>{typeIcons[item.type] || typeIcons.info}</span>
                  <div style={{ flex: 1, minWidth: 0 }}>
                    <div style={{
                      fontSize: 13,
                      fontWeight: item.read ? 400 : 500,
                      color: '#313238',
                      marginBottom: 4,
                    }}>
                      {item.title}
                      {item.source && (
                        <Tag color={typeColors[item.type] || 'default'} style={{ marginLeft: 6, fontSize: 11 }}>
                          {item.source}
                        </Tag>
                      )}
                    </div>
                    {item.content && (
                      <div style={{
                        fontSize: 12, color: '#979BA5',
                        overflow: 'hidden', textOverflow: 'ellipsis',
                        display: '-webkit-box', WebkitLineClamp: 2, WebkitBoxOrient: 'vertical',
                      }}>
                        {item.content}
                      </div>
                    )}
                    <div style={{ fontSize: 11, color: '#C4C6CC', marginTop: 4 }}>
                      {item.created_at}
                    </div>
                  </div>
                  {!item.read && (
                    <span style={{
                      width: 8, height: 8, borderRadius: '50%',
                      background: '#3A84FF', flexShrink: 0, marginTop: 6,
                    }} />
                  )}
                </div>
              </div>
            )}
          />
        )}
      </div>
      <div style={{
        padding: '8px 16px',
        borderTop: '1px solid #E7E9EF',
        textAlign: 'center',
      }}>
        <Button type="link" size="small" onClick={() => { setNotifyOpen(false); navigate('/notifications') }}>
          查看全部通知
        </Button>
      </div>
    </div>
  )

  return (
    <ProLayout
      title={false}
      logo={<img src={logo} alt="MXCMDB" style={{ height: 32 }} />}
      layout="mix"
      splitMenus={false}
      fixedHeader
      fixSiderbar
      siderWidth={220}
      menuDataRender={() => menuData}
      location={{ pathname: location.pathname }}
      menuItemRender={(item, dom) => (
        <a
          href={item.path || '/'}
          onClick={(e) => {
            // Shift+Click / Cmd+Click / Ctrl+Click / 中键：浏览器原生处理（新窗口/新Tab）
            if (e.metaKey || e.ctrlKey || e.shiftKey || e.button === 1) return
            e.preventDefault()
            navigate(item.path || '/')
          }}
          style={{ color: 'inherit', textDecoration: 'none' }}
        >
          {dom}
        </a>
      )}
      subMenuItemRender={(_item, dom) => <div>{dom}</div>}
      token={{
        header: {
          colorBgHeader: '#182132',
          colorHeaderTitle: '#fff',
          colorTextMenu: 'rgba(255,255,255,0.65)',
          colorTextMenuSelected: '#fff',
          colorBgMenuItemSelected: 'transparent',
          colorTextMenuActive: '#fff',
          colorTextRightActionsItem: 'rgba(255,255,255,0.75)',
          heightLayoutHeader: 52,
        },
        sider: {
          colorMenuBackground: '#fff',
          colorTextMenu: '#63656E',
          colorTextMenuSelected: '#3A84FF',
          colorBgMenuItemSelected: '#E1ECFF',
          colorTextMenuItemHover: '#3A84FF',
          colorTextMenuActive: '#3A84FF',
          colorBgMenuItemHover: '#F0F5FF',
        },
        pageContainer: {
          paddingBlockPageContainerContent: 20,
          paddingInlinePageContainerContent: 24,
        },
      }}
      bgLayoutImgList={[]}
      actionsRender={() => [
        <Dropdown
          key="notifications"
          trigger={['click']}
          open={notifyOpen}
          onOpenChange={handleNotifyOpen}
          popupRender={() => notificationDropdown}
          placement="bottomRight"
        >
          <div style={{
            cursor: 'pointer',
            padding: '0 8px',
            display: 'flex',
            alignItems: 'center',
            height: 52,
          }}>
            <Badge count={unreadCount} size="small" offset={[2, -2]}>
              <BellOutlined style={{ fontSize: 18, color: 'rgba(255,255,255,0.75)' }} />
            </Badge>
          </div>
        </Dropdown>,
      ]}
      avatarProps={{
        title: userInfo?.nickname || userInfo?.username || '',
        size: 'small',
        style: { backgroundColor: '#3A84FF' },
        render: (_props, dom) => (
          <Dropdown
            menu={{
              items: [
                {
                  key: 'role',
                  label: `角色：${userInfo?.role === 'admin' ? '管理员' : userInfo?.role === 'operator' ? '运维员' : '访客'}`,
                  disabled: true,
                },
                { type: 'divider' },
                {
                  key: 'security',
                  icon: <SafetyCertificateOutlined />,
                  label: '安全设置',
                  onClick: () => navigate('/account/security'),
                },
                {
                  key: 'logout',
                  icon: <LogoutOutlined />,
                  label: '退出登录',
                  danger: true,
                  onClick: handleLogout,
                },
              ],
            }}
          >
            {dom}
          </Dropdown>
        ),
      }}
      contentStyle={{
        background: '#F5F7FA',
        minHeight: 'calc(100vh - 52px)',
      }}
    >
      <Outlet />
    </ProLayout>
  )
}

export default BasicLayout
