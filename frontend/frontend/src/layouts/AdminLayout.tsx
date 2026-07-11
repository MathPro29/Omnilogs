import { useState, useMemo, useEffect } from 'react';
import { Outlet, useNavigate, useLocation } from 'react-router-dom';
import { Layout, Menu, Breadcrumb, Avatar, Dropdown, Typography } from 'antd';
import type { MenuProps } from 'antd';
import {
  ArrowLeftOnRectangleIcon,
  Bars3Icon,
  ChevronLeftIcon,
  UserCircleIcon,
} from '@heroicons/react/24/outline';
import { motion, AnimatePresence } from 'motion/react';
import { useAuthStore, useUIStore, useAppStore } from '@/store';
import { menuConfig, findMenuByPath } from '@/routes';
import { filterMenuByPermissions, getInitials } from '@/utils';
import type { MenuItem } from '@/types';

const { Header, Sider, Content } = Layout;
const { Text } = Typography;

/**
 * แปลง MenuItem ของเราเป็น antd Menu items
 */
function toAntdMenuItems(items: MenuItem[]): MenuProps['items'] {
  return items.map((item) => ({
    key: item.key,
    icon: item.icon,
    label: item.label,
    children: item.children ? toAntdMenuItems(item.children) : undefined,
  }));
}

/**
 * AdminLayout - Layout หลักของระบบ admin
 * ประกอบด้วย Sidebar, Header, Content area
 */
export function AdminLayout() {
  const navigate = useNavigate();
  const location = useLocation();

  const currentUser = useAuthStore((state) => state.currentUser);
  const clearAuth = useAuthStore((state) => state.clearAuth);
  const sidebarCollapsed = useUIStore((state) => state.sidebarCollapsed);
  const toggleSidebar = useUIStore((state) => state.toggleSidebar);
  const breadcrumbs = useAppStore((state) => state.breadcrumbs);

  const [selectedKeys, setSelectedKeys] = useState<string[]>([]);
  const [openKeys, setOpenKeys] = useState<string[]>([]);

  // Filter menu ตาม permission (menuConfig import จาก routes/routes.tsx)
  const filteredMenu = useMemo(
    () => filterMenuByPermissions(menuConfig),
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [currentUser?.permissions]
  );

  const antdMenuItems = useMemo(() => toAntdMenuItems(filteredMenu), [filteredMenu]);

  // Sync selected menu กับ current route
  useEffect(() => {
    const currentPath = location.pathname;
    const menuItem = findMenuByPath(currentPath);
    if (menuItem) {
      setSelectedKeys([menuItem.key]);
      // เปิด parent menu ถ้ามี
      const parent = filteredMenu.find((m) =>
        m.children?.some((c) => c.key === menuItem.key)
      );
      if (parent) {
        setOpenKeys((prev) => [...new Set([...prev, parent.key])]);
      }
    }
  }, [location.pathname, filteredMenu]);

  // Handle menu click -> navigate
  const handleMenuClick: MenuProps['onClick'] = ({ key }) => {
    const allItems = filteredMenu.flatMap((item) =>
      item.children ? [item, ...item.children] : [item]
    );
    const clicked = allItems.find((item) => item.key === key);
    if (clicked?.path) {
      navigate(clicked.path);
    }
  };

  // Handle logout
  const handleLogout = () => {
    clearAuth();
    navigate('/login');
  };

  // User dropdown menu
  const userMenuItems: MenuProps['items'] = [
    {
      key: 'profile',
      icon: <UserCircleIcon className="w-4 h-4" />,
      label: 'โปรไฟล์',
    },
    { type: 'divider' },
    {
      key: 'logout',
      icon: <ArrowLeftOnRectangleIcon className="w-4 h-4" />,
      label: 'ออกจากระบบ',
      danger: true,
      onClick: handleLogout,
    },
  ];

  return (
    <Layout className="min-h-screen">
      {/* Sidebar */}
      <Sider
        trigger={null}
        collapsible
        collapsed={sidebarCollapsed}
        width={260}
        collapsedWidth={80}
        className="fixed left-0 top-0 bottom-0 z-50 overflow-auto"
        style={{ background: '#100446' }}
      >
        {/* Logo */}
        <div className="sidebar-logo">
          <AnimatePresence mode="wait">
            {sidebarCollapsed ? (
              <motion.div
                key="collapsed"
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                exit={{ opacity: 0 }}
                transition={{ duration: 0.15 }}
              >
                <Text strong style={{ color: 'white', fontSize: '1.2rem' }}>
                  OMNI
                </Text>
              </motion.div>
            ) : (
              <motion.div
                key="expanded"
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                exit={{ opacity: 0 }}
                transition={{ duration: 0.15 }}
              >
                <h1 style={{ color: 'white', fontSize: '1.25rem', fontWeight: 700 }}>
                  OMNILOGS
                </h1>
              </motion.div>
            )}
          </AnimatePresence>
        </div>

        {/* Menu */}
        <Menu
          theme="dark"
          mode="inline"
          selectedKeys={selectedKeys}
          openKeys={openKeys}
          onOpenChange={(keys) => setOpenKeys(keys)}
          onClick={handleMenuClick}
          items={antdMenuItems}
          style={{ borderRight: 0, background: 'transparent' }}
        />
      </Sider>

      {/* Main Layout */}
      <Layout
        style={{
          marginLeft: sidebarCollapsed ? 80 : 260,
          transition: 'margin-left 0.2s ease',
        }}
      >
        {/* Header */}
        <Header
          className="flex items-center justify-between px-6 sticky top-0 z-40"
          style={{
            background: '#fff',
            boxShadow: '0 1px 4px rgba(0,0,0,0.06)',
            height: 64,
            lineHeight: '64px',
            padding: '0 24px',
          }}
        >
          <div className="flex items-center gap-4">
            {/* Toggle Sidebar */}
            <button
              onClick={toggleSidebar}
              className="flex items-center justify-center w-8 h-8 rounded-lg hover:bg-gray-100 transition-colors cursor-pointer"
            >
              {sidebarCollapsed ? (
                <Bars3Icon className="w-5 h-5 text-gray-600" />
              ) : (
                <ChevronLeftIcon className="w-5 h-5 text-gray-600" />
              )}
            </button>

            {/* Breadcrumb */}
            <Breadcrumb
              items={
                breadcrumbs.length > 0
                  ? breadcrumbs.map((b) => ({
                      title: b.path ? (
                        <a onClick={() => navigate(b.path!)}>{b.title}</a>
                      ) : (
                        b.title
                      ),
                    }))
                  : [{ title: 'หน้าหลัก' }]
              }
            />
          </div>

          {/* User Info */}
          <Dropdown menu={{ items: userMenuItems }} placement="bottomRight" trigger={['click']}>
            <div className="flex items-center gap-3 cursor-pointer px-3 py-1 rounded-lg hover:bg-gray-50 transition-colors">
              <Avatar
                size={36}
                style={{
                  backgroundColor: 'var(--color-primary)',
                  fontWeight: 600,
                  fontSize: '0.875rem',
                }}
              >
                {currentUser ? getInitials(currentUser.fullName) : 'U'}
              </Avatar>
              {!sidebarCollapsed && (
                <div className="hidden sm:block">
                  <div className="text-sm font-medium leading-tight" style={{ color: '#000000D9' }}>
                    {currentUser?.fullName || 'User'}
                  </div>
                  <div className="text-xs leading-tight" style={{ color: '#00000073' }}>
                    {currentUser?.position || ''}
                  </div>
                </div>
              )}
            </div>
          </Dropdown>
        </Header>

        {/* Content */}
        <Content className="admin-content">
          <AnimatePresence mode="wait">
            <Outlet />
          </AnimatePresence>
        </Content>
      </Layout>
    </Layout>
  );
}
