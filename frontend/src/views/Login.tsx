import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Loader2, LogIn } from 'lucide-react'
import { useToast } from '@/hooks/use-toast'
import { useAuth } from '@/contexts/AuthContext'

export default function Login() {
  const navigate = useNavigate()
  const { login } = useAuth()
  const { toast } = useToast()
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [loading, setLoading] = useState(false)

  const handleSubmit = async () => {
    if (!username.trim()) {
      toast({ title: '请输入用户名', variant: 'destructive' })
      return
    }
    if (!password.trim()) {
      toast({ title: '请输入密码', variant: 'destructive' })
      return
    }

    setLoading(true)
    try {
      const result = await login(username, password)
      if (result.success) {
        toast({ title: '登录成功' })
        navigate('/dashboard')
      } else {
        const errorMsg = result.message || '用户名或密码错误'
        toast({ title: errorMsg, variant: 'destructive' })
      }
    } catch (error) {
      toast({ title: '登录失败，请稍后重试', variant: 'destructive' })
    } finally {
      setLoading(false)
    }
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      handleSubmit()
    }
  }

  return (
    <div className="min-h-screen bg-neutral-50 flex flex-col items-center justify-center px-4 py-8">
      {/* Logo */}
      <div className="flex items-center gap-3.5 mb-10">
        <img src="/favicon.svg" alt="NginxOps" className="w-14 h-14" />
        <span className="text-neutral-900 font-bold text-3xl tracking-tight">NginxOps</span>
      </div>

      {/* 主卡片 */}
      <div 
        className="w-full bg-white rounded-3xl shadow-lg border border-neutral-200 overflow-hidden flex flex-col"
        style={{ 
          maxWidth: '480px',
          minHeight: '400px'
        }}
      >
        {/* 卡片头部 */}
        <div className="pt-6 pb-4 px-8 text-center shrink-0">
          <div className="w-12 h-12 rounded-xl bg-neutral-100 flex items-center justify-center mx-auto mb-3">
            <LogIn className="w-6 h-6 text-neutral-600" />
          </div>
          <h2 className="text-lg font-semibold text-neutral-900">
            欢迎回来
          </h2>
          <p className="text-neutral-500 mt-1 text-sm">
            请输入您的账户信息登录
          </p>
        </div>

        {/* 卡片内容 */}
        <div className="flex-1 px-8 overflow-auto">
          <div className="space-y-4">
            {/* Username */}
            <div>
              <label className="block text-sm font-medium text-neutral-700 mb-1.5">
                用户名
              </label>
              <input
                id="username"
                type="text"
                placeholder="请输入用户名"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                onKeyDown={handleKeyDown}
                className="w-full h-10 px-3 text-sm border border-neutral-200 rounded-lg focus:border-neutral-400 focus:ring-2 focus:ring-neutral-100 outline-none"
              />
            </div>

            {/* Password */}
            <div>
              <label className="block text-sm font-medium text-neutral-700 mb-1.5">
                密码
              </label>
              <input
                id="password"
                type="password"
                placeholder="请输入密码"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                onKeyDown={handleKeyDown}
                className="w-full h-10 px-3 text-sm border border-neutral-200 rounded-lg focus:border-neutral-400 focus:ring-2 focus:ring-neutral-100 outline-none"
              />
            </div>
          </div>
        </div>

        {/* 底部按钮 */}
        <div className="px-8 pb-6 flex gap-3 shrink-0">
          <button
            onClick={handleSubmit}
            disabled={loading}
            className="flex-1 flex items-center justify-center gap-2 h-10 text-sm font-medium text-white bg-neutral-900 hover:bg-neutral-800 rounded-xl transition-colors disabled:opacity-60"
          >
            {loading ? (
              <>
                <Loader2 className="w-4 h-4 animate-spin" />
                登录中...
              </>
            ) : (
              '登 录'
            )}
          </button>
        </div>
      </div>

      {/* 底部提示 */}
      <p className="mt-6 text-xs text-neutral-400">
        © 2024 NginxOps. All rights reserved.
      </p>
    </div>
  )
}
