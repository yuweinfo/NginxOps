import { useState, useEffect } from 'react'
import { 
  User, 
  Mail, 
  Lock, 
  Save, 
  AlertCircle, 
  CheckCircle,
  Shield,
  Calendar
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Avatar, AvatarFallback } from '@/components/ui/avatar'
import { Separator } from '@/components/ui/separator'
import { userApi, UserInfo } from '@/api/user'
import { cn } from '@/lib/utils'

export default function Profile() {
  const [userInfo, setUserInfo] = useState<UserInfo | null>(null)
  const [email, setEmail] = useState('')
  const [oldPassword, setOldPassword] = useState('')
  const [newPassword, setNewPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [loading, setLoading] = useState(false)
  const [verifying, setVerifying] = useState(false)
  const [passwordVerified, setPasswordVerified] = useState(false)
  const [message, setMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null)

  useEffect(() => {
    loadUserInfo()
  }, [])

  const loadUserInfo = async () => {
    try {
      const res = await userApi.getCurrentUser()
      if (res.success && res.data) {
        setUserInfo(res.data)
        setEmail(res.data.email || '')
      }
    } catch (error) {
      console.error('获取用户信息失败', error)
    }
  }

  const handleVerifyOldPassword = async () => {
    if (!oldPassword.trim()) {
      setMessage({ type: 'error', text: '请输入原密码' })
      return
    }

    setVerifying(true)
    setMessage(null)
    
    try {
      const res = await userApi.verifyPassword(oldPassword)
      if (res.success) {
        setPasswordVerified(true)
        setMessage({ type: 'success', text: '原密码验证通过，请输入新密码' })
      } else {
        setMessage({ type: 'error', text: res.message || '原密码错误' })
      }
    } catch (error: any) {
      setMessage({ type: 'error', text: error.userMessage || '验证失败' })
    } finally {
      setVerifying(false)
    }
  }

  const handleCancelPasswordChange = () => {
    setOldPassword('')
    setNewPassword('')
    setConfirmPassword('')
    setPasswordVerified(false)
    setMessage(null)
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setMessage(null)

    // 如果正在修改密码
    if (newPassword || confirmPassword) {
      if (!passwordVerified) {
        setMessage({ type: 'error', text: '请先验证原密码' })
        return
      }

      if (newPassword !== confirmPassword) {
        setMessage({ type: 'error', text: '两次输入的密码不一致' })
        return
      }

      if (newPassword.length < 6) {
        setMessage({ type: 'error', text: '密码长度不能少于6位' })
        return
      }
    }

    setLoading(true)
    try {
      const res = await userApi.updateProfile({
        email: email || undefined,
        password: newPassword || undefined,
        oldPassword: newPassword ? oldPassword : undefined
      })

      if (res.success) {
        setMessage({ type: 'success', text: '个人信息更新成功' })
        setOldPassword('')
        setNewPassword('')
        setConfirmPassword('')
        setPasswordVerified(false)
        loadUserInfo()
      } else {
        setMessage({ type: 'error', text: res.message || '更新失败' })
      }
    } catch (error: any) {
      setMessage({ type: 'error', text: error.userMessage || '更新失败' })
    } finally {
      setLoading(false)
    }
  }

  const formatDate = (dateStr?: string) => {
    if (!dateStr) return '未知'
    return new Date(dateStr).toLocaleDateString('zh-CN', {
      year: 'numeric',
      month: 'long',
      day: 'numeric'
    })
  }

  return (
    <div className="max-w-4xl mx-auto space-y-6">
      {/* 用户信息概览 */}
      <Card>
        <CardContent className="p-6">
          <div className="flex flex-col sm:flex-row items-center sm:items-start gap-6">
            {/* 头像 */}
            <Avatar className="h-16 w-16">
              <AvatarFallback className="text-lg font-semibold bg-muted text-muted-foreground">
                {userInfo?.username?.charAt(0).toUpperCase() || 'A'}
              </AvatarFallback>
            </Avatar>

            {/* 用户信息 */}
            <div className="flex-1 text-center sm:text-left">
              <div className="flex flex-col sm:flex-row items-center sm:items-center gap-3 mb-2">
                <h2 className="text-xl font-semibold">{userInfo?.username || 'Loading...'}</h2>
                <Badge variant="secondary">
                  <Shield className="h-3 w-3 mr-1" />
                  {userInfo?.role === 'admin' ? '管理员' : '用户'}
                </Badge>
              </div>
              
              <div className="flex flex-wrap justify-center sm:justify-start items-center gap-x-5 gap-y-2 text-sm text-muted-foreground">
                <div className="flex items-center gap-1.5">
                  <Mail className="h-4 w-4" />
                  <span>{userInfo?.email || '未设置邮箱'}</span>
                </div>
                <div className="flex items-center gap-1.5">
                  <Calendar className="h-4 w-4" />
                  <span>加入于 {formatDate(userInfo?.createdAt)}</span>
                </div>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* 表单区域 */}
      <form onSubmit={handleSubmit} className="space-y-6">
        {/* 消息提示 */}
        {message && (
          <div
            className={cn(
              "flex items-center gap-2 p-4 rounded-lg text-sm",
              message.type === 'success'
                ? "bg-emerald-500/10 text-emerald-600 dark:text-emerald-400"
                : "bg-destructive/10 text-destructive"
            )}
          >
            {message.type === 'success' ? (
              <CheckCircle className="h-4 w-4 shrink-0" />
            ) : (
              <AlertCircle className="h-4 w-4 shrink-0" />
            )}
            <span>{message.text}</span>
          </div>
        )}

        {/* 修改邮箱 */}
        <Card>
          <CardHeader>
            <div className="flex items-center gap-2">
              <Mail className="h-4 w-4 text-muted-foreground" />
              <CardTitle className="text-base">邮箱设置</CardTitle>
            </div>
            <CardDescription>用于接收系统通知和找回密码</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label htmlFor="username" className="text-sm font-medium flex items-center gap-2">
                  <User className="h-3.5 w-3.5 text-muted-foreground" />
                  用户名
                </Label>
                <div className="relative">
                  <Input
                    id="username"
                    value={userInfo?.username || ''}
                    disabled
                    className="bg-muted cursor-not-allowed pr-10"
                  />
                  <Lock className="absolute right-3 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-muted-foreground/50" />
                </div>
                <p className="text-xs text-muted-foreground">用户名不可修改</p>
              </div>

              <div className="space-y-2">
                <Label htmlFor="email" className="text-sm font-medium flex items-center gap-2">
                  <Mail className="h-3.5 w-3.5 text-muted-foreground" />
                  邮箱地址
                </Label>
                <Input
                  id="email"
                  type="email"
                  placeholder="请输入邮箱地址"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                />
              </div>
            </div>
          </CardContent>
        </Card>

        {/* 修改密码 */}
        <Card>
          <CardHeader>
            <div className="flex items-center gap-2">
              <Lock className="h-4 w-4 text-muted-foreground" />
              <CardTitle className="text-base">密码设置</CardTitle>
            </div>
            <CardDescription>修改密码前需先验证原密码</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            {/* 原密码验证 */}
            <div className="space-y-2">
              <Label htmlFor="oldPassword" className="text-sm font-medium">
                原密码
              </Label>
              <div className="flex gap-3">
                <Input
                  id="oldPassword"
                  type="password"
                  placeholder="请输入原密码"
                  value={oldPassword}
                  onChange={(e) => {
                    setOldPassword(e.target.value)
                    if (passwordVerified) {
                      setPasswordVerified(false)
                    }
                  }}
                  disabled={passwordVerified}
                  className="flex-1"
                />
                {!passwordVerified ? (
                  <Button
                    type="button"
                    variant="outline"
                    onClick={handleVerifyOldPassword}
                    disabled={verifying || !oldPassword.trim()}
                  >
                    {verifying ? '验证中...' : '验证'}
                  </Button>
                ) : (
                  <Button
                    type="button"
                    variant="outline"
                    onClick={handleCancelPasswordChange}
                    className="text-muted-foreground"
                  >
                    取消
                  </Button>
                )}
              </div>
            </div>

            <Separator />

            {/* 新密码输入 */}
            <div className={cn(
              "space-y-4 transition-opacity",
              passwordVerified ? "opacity-100" : "opacity-50 pointer-events-none"
            )}>
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label htmlFor="newPassword" className="text-sm font-medium">
                    新密码
                  </Label>
                  <Input
                    id="newPassword"
                    type="password"
                    placeholder="请输入新密码"
                    value={newPassword}
                    onChange={(e) => setNewPassword(e.target.value)}
                    disabled={!passwordVerified}
                  />
                </div>

                <div className="space-y-2">
                  <Label htmlFor="confirmPassword" className="text-sm font-medium">
                    确认密码
                  </Label>
                  <Input
                    id="confirmPassword"
                    type="password"
                    placeholder="请再次输入新密码"
                    value={confirmPassword}
                    onChange={(e) => setConfirmPassword(e.target.value)}
                    disabled={!passwordVerified}
                  />
                </div>
              </div>

              <div className="p-3 rounded-lg bg-muted/50 border">
                <p className="text-xs text-muted-foreground">
                  密码至少 6 个字符
                </p>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* 保存按钮 */}
        <div className="flex justify-end">
          <Button type="submit" disabled={loading}>
            <Save className="h-4 w-4 mr-2" />
            {loading ? '保存中...' : '保存修改'}
          </Button>
        </div>
      </form>
    </div>
  )
}
