import { Routes, Route, Navigate } from 'react-router-dom'
import LoginPage from './pages/login'
import BasicLayout from './layouts/BasicLayout'
import DashboardPage from './pages/dashboard'
import PlatformUsersPage from './pages/users/platform'
import CloudAccountsPage from './pages/cloud-accounts'
import CloudAssetsPage from './pages/assets/cloud'
import IDCAssetsPage from './pages/assets/idc'
import AssetTagsPage from './pages/assets/tags'
import AssetOwnershipPage from './pages/assets/ownership'
import SSHKeysPage from './pages/sshkeys'
import ProbePage from './pages/probe'
import TaskExecutePage from './pages/tasks/execute'
import TaskHistoryPage from './pages/tasks/history'
import FileDistributePage from './pages/files/distribute'
import FileHistoryPage from './pages/files/history'
import LinuxUsersPage from './pages/users/linux'
import ProjectsPage from './pages/projects'
import ProjectDetailPage from './pages/projects/detail'
import AssetInitPage from './pages/assets/init'
import TerminalPage from './pages/terminal'
import TerminalSessionsPage from './pages/terminal/sessions'
import SubnetsPage from './pages/ipam/subnets'
import IPAddressesPage from './pages/ipam/addresses'
import GeneralSettingsPage from './pages/settings/general'
import SecuritySettingsPage from './pages/settings/security'
import LoginAuditPage from './pages/audit/login'
import OperationAuditPage from './pages/audit/operation'
import CommandAuditPage from './pages/audit/command'
import AccountSecurityPage from './pages/account/security'
import PlaybooksPage from './pages/playbooks'
import AssetDetailPage from './pages/assets/detail'
import NotificationsPage from './pages/notifications'
import ApprovalsPage from './pages/approvals'
import PermissionsPage from './pages/settings/permissions'

function App() {
  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route path="/" element={<BasicLayout />}>
        <Route index element={<Navigate to="/dashboard" replace />} />
        <Route path="dashboard" element={<DashboardPage />} />
        {/* 资产管理 */}
        <Route path="assets/cloud" element={<CloudAssetsPage />} />
        <Route path="assets/idc" element={<IDCAssetsPage />} />
        <Route path="assets/tags" element={<AssetTagsPage />} />
        <Route path="assets/ownership" element={<AssetOwnershipPage />} />
        <Route path="assets/:id/detail" element={<AssetDetailPage />} />
        <Route path="assets/init" element={<AssetInitPage />} />
        <Route path="cloud-accounts" element={<CloudAccountsPage />} />
        <Route path="probe" element={<ProbePage />} />
        <Route path="ipam/subnets" element={<SubnetsPage />} />
        <Route path="ipam/addresses" element={<IPAddressesPage />} />
        {/* 运维操作 */}
        <Route path="tasks/execute" element={<TaskExecutePage />} />
        <Route path="tasks/history" element={<TaskHistoryPage />} />
        <Route path="files/distribute" element={<FileDistributePage />} />
        <Route path="files/history" element={<FileHistoryPage />} />
        <Route path="terminal" element={<TerminalPage />} />
        <Route path="terminal/sessions" element={<TerminalSessionsPage />} />
        <Route path="sshkeys" element={<SSHKeysPage />} />
        <Route path="playbooks" element={<PlaybooksPage />} />
        {/* 项目管理 */}
        <Route path="projects" element={<ProjectsPage />} />
        <Route path="projects/:id" element={<ProjectDetailPage />} />
        {/* 审计中心 */}
        <Route path="audit/login" element={<LoginAuditPage />} />
        <Route path="audit/operation" element={<OperationAuditPage />} />
        <Route path="audit/command" element={<CommandAuditPage />} />
        {/* 系统管理 */}
        <Route path="users/platform" element={<PlatformUsersPage />} />
        <Route path="users/linux" element={<LinuxUsersPage />} />
        <Route path="settings/general" element={<GeneralSettingsPage />} />
        <Route path="settings/security" element={<SecuritySettingsPage />} />
        <Route path="settings/permissions" element={<PermissionsPage />} />
        {/* 通知 & 审批 */}
        <Route path="notifications" element={<NotificationsPage />} />
        <Route path="approvals" element={<ApprovalsPage />} />
        {/* 账号安全 */}
        <Route path="account/security" element={<AccountSecurityPage />} />
      </Route>
    </Routes>
  )
}

export default App
