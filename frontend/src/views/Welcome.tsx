import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { Loader2, Database, Shield, User, Plug, Copy, CheckCheck, Check, ArrowLeft, ArrowRight, Server } from 'lucide-react'
import { useToast } from '@/hooks/use-toast'
import { setupApi, type SetupRequest } from '@/api/setup'
import { Progress } from '@/components/ui/progress'

type Step = 'dbType' | 'dbConfig' | 'jwt' | 'admin' | 'complete'

const stepConfig: { key: Step; title: string; description: string; icon: typeof Database }[] = [
  { key: 'dbType', title: '选择数据库', description: '选择您的数据库类型', icon: Database },
  { key: 'dbConfig', title: '数据库配置', description: '配置数据库连接', icon: Server },
  { key: 'jwt', title: '安全设置', description: '配置 JWT 密钥', icon: Shield },
  { key: 'admin', title: '管理员账户', description: '创建初始管理员', icon: User },
  { key: 'complete', title: '配置完成', description: '正在初始化系统', icon: Check },
]

// 检测操作系统
function detectOS(): 'windows' | 'mac' | 'linux' {
  const ua = navigator.userAgent.toLowerCase()
  if (ua.includes('win')) return 'windows'
  if (ua.includes('mac')) return 'mac'
  return 'linux'
}

export default function Welcome() {
  const navigate = useNavigate()
  const { toast } = useToast()
  const [currentStep, setCurrentStep] = useState<Step>('dbType')
  const [loading, setLoading] = useState(false)
  const [testingDB, setTestingDB] = useState(false)
  const [dbTestResult, setDbTestResult] = useState<{ success: boolean; message: string } | null>(null)
  const [animating, setAnimating] = useState(false)
  const [direction, setDirection] = useState<'left' | 'right'>('right')
  const [checkingStatus, setCheckingStatus] = useState(true)

  // 检查系统是否已配置，防止已初始化用户重新进入
  useEffect(() => {
    const checkStatus = async () => {
      try {
        const response = await setupApi.getStatus()
        if (response.data?.configured) {
          // 已配置，重定向到登录页
          navigate('/login', { replace: true })
          return
        }
      } catch (error) {
        console.error('Failed to check setup status:', error)
      } finally {
        setCheckingStatus(false)
      }
    }
    checkStatus()
  }, [navigate])

  // 正在检查状态时显示加载
  if (checkingStatus) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-neutral-50">
        <div className="text-center">
          <Loader2 className="w-8 h-8 animate-spin mx-auto text-neutral-400" />
          <p className="mt-4 text-sm text-neutral-500">检查系统状态...</p>
        </div>
      </div>
    )
  }

  // 表单数据
  const [useExternalDB, setUseExternalDB] = useState(false)
  const [dbHost, setDbHost] = useState('')
  const [dbPort, setDbPort] = useState('5432')
  const [dbName, setDbName] = useState('nginxops')
  const [dbUser, setDbUser] = useState('postgres')
  const [dbPassword, setDbPassword] = useState('')
  const [jwtSecret, setJwtSecret] = useState('')
  const [adminUsername, setAdminUsername] = useState('admin')
  const [adminEmail, setAdminEmail] = useState('')
  const [adminPassword, setAdminPassword] = useState('')

  const currentStepIndex = stepConfig.findIndex(s => s.key === currentStep)
  const progressValue = ((currentStepIndex) / (stepConfig.length - 1)) * 100
  const detectedOS = detectOS()
  const [copiedCommand, setCopiedCommand] = useState<string | null>(null)

  const jwtCommands: Record<string, { label: string; cmd: string }> = {
    mac: { label: 'macOS / Linux', cmd: 'openssl rand -base64 64' },
    windows: { label: 'Windows (PowerShell)', cmd: '[Convert]::ToBase64String((1..64 | ForEach-Object { Get-Random -Maximum 256 }))' },
  }

  const generateRandomSecret = () => {
    const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789!@#$%^&*'
    let secret = ''
    for (let i = 0; i < 64; i++) {
      secret += chars.charAt(Math.floor(Math.random() * chars.length))
    }
    setJwtSecret(secret)
  }

  const copyCommand = (cmd: string, key: string) => {
    navigator.clipboard.writeText(cmd)
    setCopiedCommand(key)
    setTimeout(() => setCopiedCommand(null), 2000)
    toast({ title: '已复制到剪贴板' })
  }

  const testDBConnection = async () => {
    if (!dbHost.trim() || !dbPort.trim() || !dbName.trim() || !dbUser.trim() || !dbPassword.trim()) {
      toast({ title: '请填写完整的数据库连接信息', variant: 'destructive' })
      return
    }

    setTestingDB(true)
    setDbTestResult(null)
    try {
      const response = await setupApi.testConnection({
        host: dbHost,
        port: parseInt(dbPort),
        name: dbName,
        user: dbUser,
        password: dbPassword,
      })
      
      if (response.data?.connected) {
        setDbTestResult({ success: true, message: '连接成功' })
      } else {
        setDbTestResult({ success: false, message: response.message || '连接失败' })
      }
    } catch (error: any) {
      setDbTestResult({ success: false, message: error.userMessage || '连接失败' })
    } finally {
      setTestingDB(false)
    }
  }

  const validateStep = (step: Step): boolean => {
    switch (step) {
      case 'dbConfig':
        if (useExternalDB) {
          if (!dbHost.trim()) { toast({ title: '请输入数据库主机地址', variant: 'destructive' }); return false }
          if (!dbPort.trim()) { toast({ title: '请输入数据库端口', variant: 'destructive' }); return false }
          if (!dbName.trim()) { toast({ title: '请输入数据库名称', variant: 'destructive' }); return false }
          if (!dbUser.trim()) { toast({ title: '请输入数据库用户名', variant: 'destructive' }); return false }
          if (!dbPassword.trim()) { toast({ title: '请输入数据库密码', variant: 'destructive' }); return false }
        }
        return true
      case 'jwt':
        if (!jwtSecret.trim() || jwtSecret.length < 32) {
          toast({ title: 'JWT 密钥必须至少32个字符', variant: 'destructive' })
          return false
        }
        return true
      case 'admin':
        if (!adminUsername.trim() || adminUsername.length < 3) {
          toast({ title: '用户名必须至少3个字符', variant: 'destructive' })
          return false
        }
        if (!adminEmail.trim() || !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(adminEmail)) {
          toast({ title: '请输入有效的邮箱地址', variant: 'destructive' })
          return false
        }
        if (!adminPassword.trim() || adminPassword.length < 6) {
          toast({ title: '密码必须至少6个字符', variant: 'destructive' })
          return false
        }
        return true
      default:
        return true
    }
  }

  const goToStep = (step: Step, dir: 'left' | 'right') => {
    if (animating) return
    setDirection(dir)
    setAnimating(true)
    setTimeout(() => {
      setCurrentStep(step)
      setTimeout(() => setAnimating(false), 50)
    }, 300)
  }

  const handleNext = () => {
    if (!validateStep(currentStep)) return
    
    // 特殊处理：如果选择内部数据库，跳过 dbConfig 步骤
    if (currentStep === 'dbType' && !useExternalDB) {
      goToStep('jwt', 'right')
      return
    }
    
    const nextIndex = currentStepIndex + 1
    if (nextIndex < stepConfig.length) {
      goToStep(stepConfig[nextIndex].key, 'right')
    }
  }

  const handlePrev = () => {
    // 特殊处理：如果之前选择的是内部数据库，从 jwt 返回到 dbType
    if (currentStep === 'jwt' && !useExternalDB) {
      goToStep('dbType', 'left')
      return
    }
    
    const prevIndex = currentStepIndex - 1
    if (prevIndex >= 0) {
      goToStep(stepConfig[prevIndex].key, 'left')
    }
  }

  const handleComplete = async () => {
    if (!validateStep('admin')) return
    setLoading(true)
    try {
      const req: SetupRequest = {
        useExternalDB,
        dbHost: useExternalDB ? dbHost : '127.0.0.1',
        dbPort: useExternalDB ? parseInt(dbPort) : 5432,
        dbName: useExternalDB ? dbName : 'nginxops',
        dbUser: useExternalDB ? dbUser : 'postgres',
        dbPassword: useExternalDB ? dbPassword : 'postgres',
        jwtSecret,
        adminUsername,
        adminEmail,
        adminPassword,
      }
      const response = await setupApi.initialize(req)
      if (response.success) {
        goToStep('complete', 'right')
        pollForRestart()
      }
    } catch (error: any) {
      toast({ title: error.userMessage || '初始化失败，请重试', variant: 'destructive' })
      setLoading(false)
    }
  }

  const pollForRestart = async () => {
    let attempts = 0
    const poll = async () => {
      attempts++
      try {
        const healthResponse = await fetch('/api/health')
        if (healthResponse.ok) {
          const statusResponse = await setupApi.getStatus()
          if (statusResponse.data?.configured) { navigate('/login'); return }
        }
      } catch {}
      if (attempts < 60) setTimeout(poll, 1000)
      else { toast({ title: '等待超时，请手动刷新页面', variant: 'destructive' }); setLoading(false) }
    }
    setTimeout(poll, 2000)
  }

  // 渲染步骤内容
  const renderStepContent = () => {
    switch (currentStep) {
      case 'dbType':
        return (
          <div className="space-y-4">
            <div className="grid grid-cols-1 gap-4">
              <button
                onClick={() => { setUseExternalDB(false); setDbTestResult(null) }}
                className={`w-full p-6 rounded-2xl border-2 transition-all duration-200 text-left ${
                  !useExternalDB 
                    ? 'border-neutral-900 bg-neutral-50 shadow-sm' 
                    : 'border-neutral-200 hover:border-neutral-300'
                }`}
              >
                <div className="flex items-start gap-4">
                  <div className="w-12 h-12 rounded-xl bg-neutral-100 flex items-center justify-center shrink-0">
                    <Database className="w-6 h-6 text-neutral-600" />
                  </div>
                  <div className="flex-1">
                    <div className="font-medium text-neutral-900 text-lg">内部数据库</div>
                    <p className="text-sm text-neutral-500 mt-1">
                      系统将自动配置内置 PostgreSQL 数据库
                    </p>
                  </div>
                  {!useExternalDB && (
                    <Check className="w-5 h-5 text-neutral-900 shrink-0 mt-1" />
                  )}
                </div>
              </button>
              
              <button
                onClick={() => { setUseExternalDB(true); setDbTestResult(null) }}
                className={`w-full p-6 rounded-2xl border-2 transition-all duration-200 text-left ${
                  useExternalDB 
                    ? 'border-neutral-900 bg-neutral-50 shadow-sm' 
                    : 'border-neutral-200 hover:border-neutral-300'
                }`}
              >
                <div className="flex items-start gap-4">
                  <div className="w-12 h-12 rounded-xl bg-neutral-100 flex items-center justify-center shrink-0">
                    <Server className="w-6 h-6 text-neutral-600" />
                  </div>
                  <div className="flex-1">
                    <div className="font-medium text-neutral-900 text-lg">外部数据库</div>
                    <p className="text-sm text-neutral-500 mt-1">
                      连接已有的 PostgreSQL 数据库实例
                    </p>
                  </div>
                  {useExternalDB && (
                    <Check className="w-5 h-5 text-neutral-900 shrink-0 mt-1" />
                  )}
                </div>
              </button>
            </div>
          </div>
        )

      case 'dbConfig':
        return (
          <div className="space-y-4">
            {!useExternalDB ? (
              <div className="p-4 bg-neutral-50 rounded-xl border border-neutral-100 text-sm text-neutral-600 text-center">
                使用内置数据库，无需额外配置
              </div>
            ) : (
              <>
                <div className="grid grid-cols-2 gap-3">
                  <div>
                    <label className="block text-sm font-medium text-neutral-700 mb-1.5">主机地址</label>
                    <input
                      type="text"
                      value={dbHost}
                      onChange={(e) => setDbHost(e.target.value)}
                      placeholder="数据库主机"
                      className="w-full h-10 px-3 text-sm border border-neutral-200 rounded-lg focus:border-neutral-400 focus:ring-2 focus:ring-neutral-100 outline-none"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-neutral-700 mb-1.5">端口</label>
                    <input
                      type="number"
                      value={dbPort}
                      onChange={(e) => setDbPort(e.target.value)}
                      placeholder="5432"
                      className="w-full h-10 px-3 text-sm border border-neutral-200 rounded-lg focus:border-neutral-400 focus:ring-2 focus:ring-neutral-100 outline-none"
                    />
                  </div>
                </div>
                <div>
                  <label className="block text-sm font-medium text-neutral-700 mb-1.5">数据库名称</label>
                  <input
                    type="text"
                    value={dbName}
                    onChange={(e) => setDbName(e.target.value)}
                    placeholder="nginxops"
                    className="w-full h-10 px-3 text-sm border border-neutral-200 rounded-lg focus:border-neutral-400 focus:ring-2 focus:ring-neutral-100 outline-none"
                  />
                </div>
                <div className="grid grid-cols-2 gap-3">
                  <div>
                    <label className="block text-sm font-medium text-neutral-700 mb-1.5">用户名</label>
                    <input
                      type="text"
                      value={dbUser}
                      onChange={(e) => setDbUser(e.target.value)}
                      placeholder="用户名"
                      className="w-full h-10 px-3 text-sm border border-neutral-200 rounded-lg focus:border-neutral-400 focus:ring-2 focus:ring-neutral-100 outline-none"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-neutral-700 mb-1.5">密码</label>
                    <input
                      type="password"
                      value={dbPassword}
                      onChange={(e) => setDbPassword(e.target.value)}
                      placeholder="密码"
                      className="w-full h-10 px-3 text-sm border border-neutral-200 rounded-lg focus:border-neutral-400 focus:ring-2 focus:ring-neutral-100 outline-none"
                    />
                  </div>
                </div>
                
                <button
                  onClick={testDBConnection}
                  disabled={testingDB}
                  className="flex items-center justify-center gap-2 w-full h-10 text-sm font-medium text-neutral-700 bg-neutral-100 hover:bg-neutral-200 rounded-lg transition-colors disabled:opacity-60"
                >
                  {testingDB ? <Loader2 className="w-4 h-4 animate-spin" /> : <Plug className="w-4 h-4" />}
                  {testingDB ? '测试中...' : '测试连接'}
                </button>
                
                {dbTestResult && (
                  <div className={`p-3 rounded-lg text-sm flex items-center gap-2 ${
                    dbTestResult.success ? 'bg-green-50 text-green-700 border border-green-200' : 'bg-red-50 text-red-700 border border-red-200'
                  }`}>
                    {dbTestResult.success ? <Check className="w-4 h-4" /> : <span className="text-base">×</span>}
                    {dbTestResult.message}
                  </div>
                )}
              </>
            )}
          </div>
        )

      case 'jwt':
        return (
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-neutral-700 mb-1.5">
                JWT 密钥 <span className="text-neutral-400">(至少32个字符)</span>
              </label>
              <div className="flex gap-2">
                <input
                  type="text"
                  value={jwtSecret}
                  onChange={(e) => setJwtSecret(e.target.value)}
                  placeholder="请输入或生成 JWT 密钥"
                  className="flex-1 h-10 px-3 text-sm border border-neutral-200 rounded-lg focus:border-neutral-400 focus:ring-2 focus:ring-neutral-100 outline-none font-mono"
                />
                <button
                  onClick={generateRandomSecret}
                  className="px-4 h-10 text-sm font-medium text-neutral-700 bg-neutral-100 hover:bg-neutral-200 rounded-lg transition-colors whitespace-nowrap"
                >
                  随机生成
                </button>
              </div>
            </div>

            <div className="p-4 bg-blue-50 rounded-lg border border-blue-100">
              <p className="text-sm font-medium text-blue-800 mb-3">如何生成 JWT 密钥？</p>
              <div className="space-y-2">
                {Object.entries(jwtCommands).map(([key, { label, cmd }]) => (
                  <div 
                    key={key} 
                    className={`flex items-center gap-2 p-2.5 rounded-lg transition-colors ${
                      detectedOS === key 
                        ? 'bg-blue-100 border-2 border-blue-400' 
                        : 'bg-blue-50 border border-transparent'
                    }`}
                  >
                    <span className={`text-xs font-medium shrink-0 w-28 ${
                      detectedOS === key ? 'text-blue-800' : 'text-blue-600'
                    }`}>
                      {label}:
                    </span>
                    <code className={`text-xs px-2 py-1 rounded font-mono flex-1 truncate ${
                      detectedOS === key ? 'bg-white/80' : 'bg-white/50'
                    }`}>
                      {cmd}
                    </code>
                    <button
                      onClick={() => copyCommand(cmd, key)}
                      className="p-1.5 hover:bg-white/50 rounded transition-colors shrink-0"
                    >
                      {copiedCommand === key ? (
                        <CheckCheck className="w-4 h-4 text-green-600" />
                      ) : (
                        <Copy className={`w-4 h-4 ${detectedOS === key ? 'text-blue-600' : 'text-blue-400'}`} />
                      )}
                    </button>
                  </div>
                ))}
              </div>
            </div>
          </div>
        )

      case 'admin':
        return (
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-neutral-700 mb-1.5">
                用户名 <span className="text-neutral-400">(至少3个字符)</span>
              </label>
              <input
                type="text"
                value={adminUsername}
                onChange={(e) => setAdminUsername(e.target.value)}
                placeholder="请输入管理员用户名"
                className="w-full h-10 px-3 text-sm border border-neutral-200 rounded-lg focus:border-neutral-400 focus:ring-2 focus:ring-neutral-100 outline-none"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-neutral-700 mb-1.5">
                邮箱
              </label>
              <input
                type="email"
                value={adminEmail}
                onChange={(e) => setAdminEmail(e.target.value)}
                placeholder="请输入管理员邮箱"
                className="w-full h-10 px-3 text-sm border border-neutral-200 rounded-lg focus:border-neutral-400 focus:ring-2 focus:ring-neutral-100 outline-none"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-neutral-700 mb-1.5">
                密码 <span className="text-neutral-400">(至少6个字符)</span>
              </label>
              <input
                type="password"
                value={adminPassword}
                onChange={(e) => setAdminPassword(e.target.value)}
                placeholder="请输入管理员密码"
                className="w-full h-10 px-3 text-sm border border-neutral-200 rounded-lg focus:border-neutral-400 focus:ring-2 focus:ring-neutral-100 outline-none"
              />
            </div>
          </div>
        )

      case 'complete':
        return (
          <div className="py-6 text-center space-y-4">
            <div className="w-14 h-14 rounded-full bg-green-100 flex items-center justify-center mx-auto">
              <Check className="w-7 h-7 text-green-600" />
            </div>
            <div>
              <h3 className="text-lg font-semibold text-neutral-900">配置完成</h3>
              <p className="text-neutral-500 mt-1 text-sm">正在初始化数据库...</p>
            </div>
            <div className="flex items-center justify-center gap-2 text-sm text-neutral-400">
              <Loader2 className="w-4 h-4 animate-spin" />
              <span>即将跳转到登录页面</span>
            </div>
          </div>
        )
    }
  }

  const CurrentIcon = stepConfig[currentStepIndex]?.icon || Database

  // 计算实际步骤（隐藏不需要显示的步骤）
  const getDisplaySteps = () => {
    if (!useExternalDB) {
      return stepConfig.filter(s => s.key !== 'dbConfig')
    }
    return stepConfig
  }
  
  const displaySteps = getDisplaySteps()
  const displayStepIndex = displaySteps.findIndex(s => s.key === currentStep)
  const displayProgress = ((displayStepIndex) / (displaySteps.length - 1)) * 100

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
          maxWidth: '680px',
          minHeight: '530px',
          maxHeight: '80vh'
        }}
      >
        {/* 卡片头部 */}
        {currentStep !== 'complete' && (
          <div className="pt-6 pb-4 px-8 text-center shrink-0">
            <div className="w-12 h-12 rounded-xl bg-neutral-100 flex items-center justify-center mx-auto mb-3">
              <CurrentIcon className="w-6 h-6 text-neutral-600" />
            </div>
            <h2 className="text-lg font-semibold text-neutral-900">
              {stepConfig[currentStepIndex]?.title}
            </h2>
            <p className="text-neutral-500 mt-1 text-sm">
              {stepConfig[currentStepIndex]?.description}
            </p>
          </div>
        )}

        {/* 卡片内容 */}
        <div className={`flex-1 px-8 overflow-auto transition-all duration-300 ease-out ${
          animating 
            ? direction === 'right' 
              ? 'opacity-0 -translate-x-6' 
              : 'opacity-0 translate-x-6'
            : 'opacity-100 translate-x-0'
        }`}>
          {renderStepContent()}
        </div>

        {/* 进度条 */}
        {currentStep !== 'complete' && (
          <div className="px-8 py-3 shrink-0">
            <Progress value={displayProgress} className="h-1" />
          </div>
        )}

        {/* 底部按钮 */}
        {currentStep !== 'complete' && (
          <div className="px-8 pb-6 flex gap-3 shrink-0">
            {!(currentStep === 'dbType') && (
              <button
                onClick={handlePrev}
                disabled={animating}
                className="flex items-center justify-center gap-2 px-4 h-10 text-sm font-medium text-neutral-700 bg-neutral-100 hover:bg-neutral-200 rounded-xl transition-colors disabled:opacity-50"
              >
                <ArrowLeft className="w-4 h-4" />
                上一步
              </button>
            )}
            <button
              onClick={currentStep === 'admin' ? handleComplete : handleNext}
              disabled={loading || animating}
              className="flex-1 flex items-center justify-center gap-2 h-10 text-sm font-medium text-white bg-neutral-900 hover:bg-neutral-800 rounded-xl transition-colors disabled:opacity-60"
            >
              {loading ? (
                <>
                  <Loader2 className="w-4 h-4 animate-spin" />
                  初始化中...
                </>
              ) : currentStep === 'admin' ? (
                '完成配置'
              ) : (
                <>
                  下一步
                  <ArrowRight className="w-4 h-4" />
                </>
              )}
            </button>
          </div>
        )}
      </div>

      {/* 底部提示 */}
      <p className="mt-6 text-xs text-neutral-400">
        © 2024 NginxOps. All rights reserved.
      </p>
    </div>
  )
}
