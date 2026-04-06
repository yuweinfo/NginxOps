import { useState } from 'react'
import { Outlet, useNavigate, useLocation } from 'react-router-dom'
import {
  LayoutDashboard, Server, GitBranch, Shield, FileText, Settings,
  Moon, Sun, ChevronLeft, ChevronRight, LogOut, Home
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
  DropdownMenuSeparator,
} from '@/components/ui/dropdown-menu'
import { useAuth } from '@/contexts/AuthContext'
import { cn } from '@/lib/utils'

const menuItems = [
  { value: '/dashboard', label: '仪表盘', icon: LayoutDashboard },
  { value: '/sites', label: '站点配置', icon: Server },
  { value: '/loadbalancer', label: '负载均衡', icon: GitBranch },
  { value: '/certificates', label: '证书管理', icon: Shield },
  { value: '/logs', label: '日志查看', icon: FileText },
  { value: '/control', label: '服务控制', icon: Settings },
]

interface MainLayoutProps {
  dark: boolean
  toggleTheme: () => void
}

export default function MainLayout({ dark, toggleTheme }: MainLayoutProps) {
  const navigate = useNavigate()
  const location = useLocation()
  const { user, logout } = useAuth()
  const [collapsed, setCollapsed] = useState(false)

  const currentMenu = menuItems.find((item) => location.pathname.startsWith(item.value))

  const handleLogout = () => {
    logout()
    navigate('/login')
  }

  return (
    <div className="flex h-screen bg-background">
      {/* 侧边栏 */}
      <aside
        className={cn(
          "flex flex-col border-r bg-card transition-all duration-300",
          collapsed ? "w-[72px]" : "w-60"
        )}
      >
        {/* Logo */}
        <div className="h-16 flex items-center border-b px-4">
          <div className="w-8 h-8 rounded-md bg-primary flex items-center justify-center text-primary-foreground font-bold text-sm">
            N
          </div>
          {!collapsed && (
            <div className="ml-3 flex flex-col">
              <span className="text-sm font-semibold">NginxOps</span>
              <span className="text-[10px] text-muted-foreground">管理控制台</span>
            </div>
          )}
        </div>

        {/* 菜单 */}
        <nav className="flex-1 overflow-y-auto py-2">
          {menuItems.map((item) => {
            const Icon = item.icon
            const isActive = currentMenu?.value === item.value
            return (
              <button
                key={item.value}
                onClick={() => navigate(item.value)}
                className={cn(
                  "w-full flex items-center gap-3 px-4 py-2.5 text-sm transition-colors",
                  isActive 
                    ? "bg-primary/10 text-primary font-medium" 
                    : "text-muted-foreground hover:bg-muted hover:text-foreground",
                  collapsed && "justify-center"
                )}
              >
                <Icon className="h-5 w-5 flex-shrink-0" />
                {!collapsed && <span>{item.label}</span>}
              </button>
            )
          })}
        </nav>

        {/* 折叠按钮 */}
        <button
          onClick={() => setCollapsed(!collapsed)}
          className="m-4 p-2 rounded-md border text-muted-foreground hover:bg-muted transition-colors flex items-center justify-center"
        >
          {collapsed ? <ChevronRight className="h-4 w-4" /> : <ChevronLeft className="h-4 w-4" />}
        </button>
      </aside>

      {/* 主内容区 */}
      <div className="flex-1 flex flex-col overflow-hidden">
        {/* 顶部导航 */}
        <header className="h-16 flex items-center justify-between border-b bg-card/95 backdrop-blur px-6">
          {/* 面包屑 */}
          <div className="flex items-center gap-2 text-sm">
            <Home className="h-4 w-4 text-muted-foreground" />
            <span className="text-muted-foreground">/</span>
            <span className="font-medium">{currentMenu?.label || '仪表盘'}</span>
          </div>

          {/* 用户菜单 */}
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" className="flex items-center gap-2 px-2">
                <div className="w-8 h-8 rounded-full bg-primary flex items-center justify-center text-primary-foreground font-semibold text-sm">
                  {user?.username?.charAt(0).toUpperCase() || 'A'}
                </div>
                <span className="font-medium">{user?.username || 'Admin'}</span>
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" className="w-40">
              <DropdownMenuItem onClick={toggleTheme}>
                {dark ? <Sun className="h-4 w-4 mr-2" /> : <Moon className="h-4 w-4 mr-2" />}
                {dark ? '浅色模式' : '深色模式'}
              </DropdownMenuItem>
              <DropdownMenuSeparator />
              <DropdownMenuItem onClick={handleLogout} className="text-destructive">
                <LogOut className="h-4 w-4 mr-2" />
                退出登录
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </header>

        {/* 内容区 */}
        <main className="flex-1 overflow-auto p-6">
          <Outlet />
        </main>
      </div>
    </div>
  )
}
