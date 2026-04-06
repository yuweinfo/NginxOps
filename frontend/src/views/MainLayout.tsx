import { useEffect } from 'react'
import { Outlet, useNavigate, useLocation } from 'react-router-dom'
import {
  LayoutDashboard,
  Server,
  Network,
  Lock,
  FileText,
  Settings,
  Moon,
  Sun,
  LogOut,
  User,
  Shield,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { Avatar, AvatarFallback } from '@/components/ui/avatar'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { useAuth } from '@/contexts/AuthContext'

const menuGroups = [
  {
    title: '概览',
    items: [
      { value: '/dashboard', label: '仪表盘', icon: LayoutDashboard },
    ],
  },
  {
    title: '配置管理',
    items: [
      { value: '/sites', label: '站点配置', icon: Server },
      { value: '/loadbalancer', label: '负载均衡', icon: Network },
      { value: '/certificates', label: '证书管理', icon: Lock },
    ],
  },
  {
    title: '运维',
    items: [
      { value: '/logs', label: '日志查看', icon: FileText },
      { value: '/control', label: '服务控制', icon: Settings },
    ],
  },
  {
    title: '系统',
    items: [
      { value: '/profile', label: '个人中心', icon: User },
      { value: '/audit', label: '操作审计', icon: Shield },
    ],
  },
]

interface MainLayoutProps {
  dark: boolean
  toggleTheme: () => void
}

export default function MainLayout({ dark, toggleTheme }: MainLayoutProps) {
  const navigate = useNavigate()
  const location = useLocation()
  const { user, logout } = useAuth()

  // 获取当前菜单标题
  const currentMenu = menuGroups
    .flatMap(g => g.items)
    .find(item => location.pathname.startsWith(item.value))

  // 更新浏览器标签页标题
  useEffect(() => {
    document.title = currentMenu 
      ? `${currentMenu.label} - NginxOps` 
      : 'NginxOps'
  }, [currentMenu])

  const handleLogout = () => {
    logout()
    navigate('/login')
  }

  return (
    <div className="flex h-screen bg-background">
      {/* Sidebar */}
      <aside className="relative flex flex-col w-64 bg-background">
        {/* 分割线 - 中间实线，上下渐变淡化 */}
        <div
          className="absolute right-0 top-0 bottom-0 w-px"
          style={{
            background: `linear-gradient(to bottom, transparent 0%, hsl(var(--border)) 33%, hsl(var(--border)) 66%, transparent 100%)`
          }}
        />
        {/* Logo */}
        <div className="h-14 flex items-center px-6">
          <svg
            viewBox="0 0 32 32"
            className="h-6 w-6 text-foreground"
            fill="currentColor"
          >
            <path d="M16 2L4 8v16l12 6 12-6V8L16 2zm0 2.5l9.5 4.75v11.5L16 25.5l-9.5-4.75V9.25L16 4.5z" />
            <path d="M16 8l-6 3v6l6 3 6-3v-6l-6-3zm0 2l3.5 1.75v3.5L16 17l-3.5-1.75v-3.5L16 10z" />
          </svg>
        </div>

        {/* Navigation */}
        <nav className="flex-1 overflow-y-auto px-4 py-4">
          {menuGroups.map((group) => (
            <div key={group.title} className="mb-6">
              <h3 className="mb-2 px-2 text-xs font-semibold text-muted-foreground uppercase tracking-wider">
                {group.title}
              </h3>
              <div className="space-y-1">
                {group.items.map((item) => {
                  const Icon = item.icon
                  const isActive = location.pathname.startsWith(item.value)

                  return (
                    <button
                      key={item.value}
                      onClick={() => navigate(item.value)}
                      className={cn(
                        "w-full flex items-center gap-3 px-3 py-2 rounded-md text-sm font-medium transition-colors",
                        isActive
                          ? "bg-accent text-accent-foreground"
                          : "text-muted-foreground hover:bg-accent/50 hover:text-foreground"
                      )}
                    >
                      <Icon className="h-4 w-4" />
                      <span>{item.label}</span>
                    </button>
                  )
                })}
              </div>
            </div>
          ))}
        </nav>

        {/* User section */}
        <div className="p-4">
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <button className="w-full flex items-center gap-3 rounded-md p-2 hover:bg-accent transition-colors">
                <Avatar className="h-8 w-8">
                  <AvatarFallback className="bg-muted text-foreground text-xs font-medium">
                    {user?.username?.charAt(0).toUpperCase() || 'A'}
                  </AvatarFallback>
                </Avatar>
                <div className="flex-1 text-left">
                  <p className="text-sm font-medium">{user?.username || 'Admin'}</p>
                  <p className="text-xs text-muted-foreground">{dark ? '深色模式' : '浅色模式'}</p>
                </div>
              </button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" className="w-48">
              <DropdownMenuLabel>我的账户</DropdownMenuLabel>
              <DropdownMenuSeparator />
              <DropdownMenuItem onClick={toggleTheme}>
                {dark ? (
                  <>
                    <Sun className="mr-2 h-4 w-4" />
                    浅色模式
                  </>
                ) : (
                  <>
                    <Moon className="mr-2 h-4 w-4" />
                    深色模式
                  </>
                )}
              </DropdownMenuItem>
              <DropdownMenuSeparator />
              <DropdownMenuItem onClick={handleLogout} className="text-destructive">
                <LogOut className="mr-2 h-4 w-4" />
                退出登录
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </aside>

      {/* Main content */}
      <div className="flex-1 flex flex-col overflow-hidden">
        {/* Header */}
        <header className="h-14 flex items-center px-8">
          {currentMenu && (
            <div className="flex items-center gap-3">
              <currentMenu.icon className="h-5 w-5 text-muted-foreground" />
              <h1 className="text-lg font-semibold">{currentMenu.label}</h1>
            </div>
          )}
        </header>
        {/* Content */}
        <main className="flex-1 overflow-auto p-8">
          <Outlet />
        </main>
      </div>
    </div>
  )
}
