import { useState, useEffect } from 'react'
import { RefreshCw, Power, Play, CheckCircle, XCircle, Clock, AlertTriangle, Globe, ArrowUp, ArrowDown, Server, Activity } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Switch } from '@/components/ui/switch'
import { Label } from '@/components/ui/label'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import { useToast } from '@/hooks/use-toast'

interface NginxStatus {
  running: boolean
  pid: number
  version: string
  uptime: string
  workerProcesses: number
  activeConnections: number
  accepts: number
  handled: number
  requests: number
  reading: number
  writing: number
  waiting: number
}

interface ConfigHistory {
  id: number
  configName: string
  operator: string
  remark: string
  createdAt: string
  testResult: boolean
}

const mockStatus: NginxStatus = {
  running: true, pid: 1234, version: 'nginx/1.24.0', uptime: '15天 8小时 32分钟',
  workerProcesses: 4, activeConnections: 287, accepts: 1854320, handled: 1854320, requests: 6823451,
  reading: 12, writing: 45, waiting: 230,
}

const mockHistory: ConfigHistory[] = [
  { id: 1, configName: 'nginx.conf', operator: 'admin', remark: '添加 rate limiting 配置', createdAt: '2026-03-23 10:30:00', testResult: true },
  { id: 2, configName: 'example.com.conf', operator: 'admin', remark: '更新 SSL 证书路径', createdAt: '2026-03-22 15:20:00', testResult: true },
  { id: 3, configName: 'nginx.conf', operator: 'admin', remark: '调整 worker_connections', createdAt: '2026-03-21 09:10:00', testResult: true },
  { id: 4, configName: 'api.example.com.conf', operator: 'admin', remark: '修改反向代理目标端口', createdAt: '2026-03-20 14:45:00', testResult: false },
]

export default function Control() {
  const [status, setStatus] = useState<NginxStatus>(mockStatus)
  const [loading, setLoading] = useState<string | null>(null)
  const [confirmDialogOpen, setConfirmDialogOpen] = useState(false)
  const [confirmAction, setConfirmAction] = useState<'reload' | 'restart' | 'stop'>('reload')
  const [liveUpdate, setLiveUpdate] = useState(true)
  const { toast } = useToast()

  useEffect(() => {
    if (!liveUpdate || !status.running) return
    const timer = setInterval(() => {
      setStatus((prev) => ({
        ...prev,
        activeConnections: prev.activeConnections + Math.floor(Math.random() * 10 - 4),
        requests: prev.requests + Math.floor(Math.random() * 50),
        reading: Math.floor(Math.random() * 20),
        writing: Math.floor(Math.random() * 60),
        waiting: Math.floor(Math.random() * 250 + 150),
      }))
    }, 3000)
    return () => clearInterval(timer)
  }, [liveUpdate, status.running])

  const handleConfirmOpen = (action: 'reload' | 'restart' | 'stop') => {
    setConfirmAction(action)
    setConfirmDialogOpen(true)
  }

  const handleAction = async () => {
    setLoading(confirmAction)
    setTimeout(() => {
      if (confirmAction === 'stop') {
        setStatus({ ...status, running: false, activeConnections: 0, reading: 0, writing: 0, waiting: 0 })
        toast({ title: 'Nginx 已停止', variant: 'destructive' })
      } else {
        toast({ title: `Nginx ${confirmAction === 'reload' ? '重载配置' : '重启'}成功` })
      }
      setLoading(null)
      setConfirmDialogOpen(false)
    }, 1500)
  }

  const handleStart = () => {
    setLoading('start')
    setTimeout(() => {
      setStatus({ ...mockStatus, pid: Math.floor(Math.random() * 9000 + 1000) })
      toast({ title: 'Nginx 已启动' })
      setLoading(null)
    }, 1500)
  }

  const StatCard = ({ icon: Icon, label, value, color }: { icon: any; label: string; value: number | string; color?: string }) => (
    <Card>
      <CardContent className="p-4 text-center">
        <Icon className={`h-6 w-6 mx-auto mb-2 ${color || 'text-primary'}`} />
        <div className="text-2xl font-bold">{typeof value === 'number' ? value.toLocaleString() : value}</div>
        <div className="text-sm text-muted-foreground">{label}</div>
      </CardContent>
    </Card>
  )

  const actionBtns = status.running ? [
    { key: 'reload' as const, label: '重载配置', icon: RefreshCw, desc: '重新加载配置文件，不中断现有连接' },
    { key: 'restart' as const, label: '重启服务', icon: Play, desc: '完全重启 Nginx 服务，会短暂中断连接' },
    { key: 'stop' as const, label: '停止服务', icon: Power, desc: '立即停止 Nginx 服务' },
  ] : [
    { key: 'start' as const, label: '启动服务', icon: Play, desc: '启动 Nginx 服务' },
  ]

  const actionLabels = { reload: '重载配置', restart: '重启', stop: '停止' }

  return (
    <div className="space-y-6">
      {/* Status Overview */}
      <Card>
        <CardContent className="p-6">
          <div className="flex items-center justify-between mb-6">
            <div className="flex items-center gap-4">
              {status.running ? (
                <Badge variant="success" className="gap-1 text-base px-3 py-1"><CheckCircle className="h-4 w-4" />运行中</Badge>
              ) : (
                <Badge variant="destructive" className="gap-1 text-base px-3 py-1"><XCircle className="h-4 w-4" />已停止</Badge>
              )}
              <span className="text-muted-foreground">PID: {status.pid} | {status.version}</span>
            </div>
            <div className="flex items-center gap-2">
              <Label className="text-sm text-muted-foreground">实时更新</Label>
              <Switch checked={liveUpdate} onCheckedChange={setLiveUpdate} />
            </div>
          </div>

          <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-4">
            <StatCard icon={Globe} label="活跃连接" value={status.activeConnections} />
            <StatCard icon={Activity} label="总请求数" value={status.requests} color="text-green-500" />
            <StatCard icon={ArrowUp} label="Reading" value={status.reading} color="text-blue-500" />
            <StatCard icon={ArrowDown} label="Writing" value={status.writing} color="text-orange-500" />
            <StatCard icon={Clock} label="Waiting" value={status.waiting} color="text-gray-500" />
            <StatCard icon={Server} label="Worker 进程" value={status.workerProcesses} color="text-red-500" />
          </div>

          <div className="mt-4 p-3 bg-muted rounded-lg text-sm text-muted-foreground">
            运行时长: {status.uptime} | Accepts: {status.accepts.toLocaleString()} | Handled: {status.handled.toLocaleString()}
          </div>
        </CardContent>
      </Card>

      {/* Control Buttons */}
      <Card>
        <CardHeader><CardTitle>服务控制</CardTitle></CardHeader>
        <CardContent>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            {actionBtns.map((btn) => {
              const Icon = btn.icon
              return (
                <Card key={btn.key} className="cursor-pointer hover:border-primary transition-colors" onClick={() => btn.key === 'start' ? handleStart() : handleConfirmOpen(btn.key)}>
                  <CardContent className="p-6 text-center">
                    <Icon className={`h-10 w-10 mx-auto mb-3 ${btn.key === 'stop' ? 'text-destructive' : btn.key === 'restart' ? 'text-orange-500' : 'text-primary'}`} />
                    <div className="font-semibold mb-1">{btn.label}</div>
                    <div className="text-sm text-muted-foreground">{btn.desc}</div>
                  </CardContent>
                </Card>
              )
            })}
          </div>
        </CardContent>
      </Card>

      {/* Config History */}
      <Card>
        <CardHeader><CardTitle>配置变更历史</CardTitle></CardHeader>
        <CardContent>
          <div className="space-y-2">
            {mockHistory.map((item) => (
              <div key={item.id} className="flex items-center gap-4 p-3 bg-muted rounded-lg">
                <Badge variant="outline">{item.configName}</Badge>
                <span className="flex-1 truncate">{item.remark}</span>
                <span className="text-sm text-muted-foreground">{item.operator}</span>
                <span className="text-sm text-muted-foreground">{item.createdAt}</span>
                {item.testResult ? (
                  <Badge variant="success" className="gap-1"><CheckCircle className="h-3 w-3" />通过</Badge>
                ) : (
                  <Badge variant="destructive" className="gap-1"><XCircle className="h-3 w-3" />失败</Badge>
                )}
              </div>
            ))}
          </div>
        </CardContent>
      </Card>

      {/* Confirm Dialog */}
      <AlertDialog open={confirmDialogOpen} onOpenChange={setConfirmDialogOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle className="flex items-center gap-2">
              {confirmAction === 'stop' ? <AlertTriangle className="text-destructive" /> : <Clock />}
              确认操作
            </AlertDialogTitle>
            <AlertDialogDescription className="space-y-2">
              <p>您即将执行 <strong>{actionLabels[confirmAction]}</strong> 操作。</p>
              <p className="text-sm text-muted-foreground">
                {confirmAction === 'reload' && '重新加载配置文件不会中断现有连接，但新配置可能影响后续请求。'}
                {confirmAction === 'restart' && '重启服务将导致所有连接短暂中断，正在处理的请求会被中断。'}
                {confirmAction === 'stop' && '停止服务后所有网站将无法访问，请确认后再操作。'}
              </p>
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>取消</AlertDialogCancel>
            <AlertDialogAction
              onClick={handleAction}
              className={confirmAction === 'stop' ? 'bg-destructive text-destructive-foreground hover:bg-destructive/90' : ''}
            >
              {loading ? '执行中...' : '确认'}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}
