import { useState, useEffect } from 'react'
import { 
  Shield, 
  User, 
  Clock, 
  CheckCircle, 
  XCircle,
  Search,
  RefreshCw,
  ChevronLeft,
  ChevronRight,
  Eye,
  Copy,
  Check
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { auditApi, AuditLog } from '@/api/audit'
import { cn } from '@/lib/utils'

const moduleLabels: Record<string, string> = {
  AUTH: '认证',
  SITE: '站点',
  CERTIFICATE: '证书',
  UPSTREAM: '负载均衡',
  USER: '用户',
  CONFIG: '配置'
}

const actionLabels: Record<string, string> = {
  LOGIN: '登录',
  LOGOUT: '登出',
  PASSWORD_CHANGE: '修改密码',
  CREATE: '创建',
  UPDATE: '更新',
  DELETE: '删除',
  ENABLE: '启用',
  DISABLE: '禁用',
  RELOAD: '重载',
  TEST: '测试',
  SYNC: '同步',
  RENEW: '续签',
  IMPORT: '导入'
}

export default function Audit() {
  const [logs, setLogs] = useState<AuditLog[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [size] = useState(20)
  const [loading, setLoading] = useState(true)
  
  const [module, setModule] = useState<string>('all')
  const [action, setAction] = useState<string>('all')
  const [status, setStatus] = useState<string>('all')
  const [username, setUsername] = useState('')
  
  const [selectedLog, setSelectedLog] = useState<AuditLog | null>(null)
  const [detailOpen, setDetailOpen] = useState(false)
  const [copied, setCopied] = useState(false)

  const handleCopy = async (text: string) => {
    try {
      await navigator.clipboard.writeText(text)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    } catch (err) {
      console.error('复制失败', err)
    }
  }

  const formatJson = (jsonStr: string) => {
    try {
      return JSON.stringify(JSON.parse(jsonStr), null, 2)
    } catch {
      return jsonStr
    }
  }

  const loadData = async () => {
    try {
      setLoading(true)
      const res = await auditApi.list({
        page,
        size,
        module: module === 'all' ? undefined : module,
        action: action === 'all' ? undefined : action,
        status: status === 'all' ? undefined : status,
      })
      if (res.success && res.data) {
        // 如果有用户名筛选，前端过滤
        let filtered = res.data.list
        if (username.trim()) {
          filtered = filtered.filter(log => 
            log.username?.toLowerCase().includes(username.toLowerCase())
          )
        }
        setLogs(filtered)
        setTotal(res.data.total)
      }
    } catch (error) {
      console.error('获取审计日志失败', error)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    loadData()
  }, [page, module, action, status])

  const handleSearch = () => {
    setPage(1)
    loadData()
  }

  const handleRefresh = () => {
    loadData()
  }

  const handleViewDetail = (log: AuditLog) => {
    setSelectedLog(log)
    setDetailOpen(true)
  }

  const formatDate = (dateStr: string) => {
    return new Date(dateStr).toLocaleString('zh-CN', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit'
    })
  }

  const totalPages = Math.ceil(total / size)

  return (
    <div className="space-y-6">
      {/* 筛选区域 */}
      <Card>
        <CardHeader>
          <CardTitle className="text-base flex items-center gap-2">
            <Shield className="h-4 w-4" />
            操作审计
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-5 gap-4">
            <div className="space-y-2">
              <Label className="text-sm">模块</Label>
              <Select value={module} onValueChange={setModule}>
                <SelectTrigger>
                  <SelectValue placeholder="全部模块" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">全部</SelectItem>
                  {Object.entries(moduleLabels).map(([key, label]) => (
                    <SelectItem key={key} value={key}>{label}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <div className="space-y-2">
              <Label className="text-sm">操作</Label>
              <Select value={action} onValueChange={setAction}>
                <SelectTrigger>
                  <SelectValue placeholder="全部操作" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">全部</SelectItem>
                  {Object.entries(actionLabels).map(([key, label]) => (
                    <SelectItem key={key} value={key}>{label}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <div className="space-y-2">
              <Label className="text-sm">状态</Label>
              <Select value={status} onValueChange={setStatus}>
                <SelectTrigger>
                  <SelectValue placeholder="全部状态" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">全部</SelectItem>
                  <SelectItem value="SUCCESS">成功</SelectItem>
                  <SelectItem value="FAILURE">失败</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div className="space-y-2">
              <Label className="text-sm">用户名</Label>
              <Input
                placeholder="输入用户名"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
              />
            </div>

            <div className="flex items-end gap-2">
              <Button onClick={handleSearch} className="flex-1">
                <Search className="h-4 w-4 mr-2" />
                查询
              </Button>
              <Button variant="outline" onClick={handleRefresh}>
                <RefreshCw className={cn("h-4 w-4", loading && "animate-spin")} />
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* 日志列表 */}
      <Card>
        <CardContent className="p-0">
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="border-b bg-muted/50">
                  <th className="px-4 py-3 text-left text-sm font-medium text-muted-foreground">时间</th>
                  <th className="px-4 py-3 text-left text-sm font-medium text-muted-foreground">用户</th>
                  <th className="px-4 py-3 text-left text-sm font-medium text-muted-foreground">模块</th>
                  <th className="px-4 py-3 text-left text-sm font-medium text-muted-foreground">操作</th>
                  <th className="px-4 py-3 text-left text-sm font-medium text-muted-foreground">目标</th>
                  <th className="px-4 py-3 text-left text-sm font-medium text-muted-foreground">IP地址</th>
                  <th className="px-4 py-3 text-left text-sm font-medium text-muted-foreground">状态</th>
                  <th className="px-4 py-3 text-left text-sm font-medium text-muted-foreground">操作</th>
                </tr>
              </thead>
              <tbody>
                {loading ? (
                  <tr>
                    <td colSpan={8} className="px-4 py-12 text-center text-muted-foreground">
                      加载中...
                    </td>
                  </tr>
                ) : logs.length === 0 ? (
                  <tr>
                    <td colSpan={8} className="px-4 py-12 text-center text-muted-foreground">
                      暂无审计日志
                    </td>
                  </tr>
                ) : (
                  logs.map((log) => (
                    <tr key={log.id} className="border-b last:border-0 hover:bg-muted/30 transition-colors">
                      <td className="px-4 py-3 text-sm">
                        <div className="flex items-center gap-1.5">
                          <Clock className="h-3.5 w-3.5 text-muted-foreground" />
                          {formatDate(log.createdAt)}
                        </div>
                      </td>
                      <td className="px-4 py-3 text-sm">
                        <div className="flex items-center gap-1.5">
                          <User className="h-3.5 w-3.5 text-muted-foreground" />
                          {log.username || '-'}
                        </div>
                      </td>
                      <td className="px-4 py-3">
                        <Badge variant="secondary">
                          {moduleLabels[log.module] || log.module}
                        </Badge>
                      </td>
                      <td className="px-4 py-3 text-sm">
                        {actionLabels[log.action] || log.action}
                      </td>
                      <td className="px-4 py-3 text-sm max-w-[200px] truncate">
                        {log.targetName || log.targetType || '-'}
                      </td>
                      <td className="px-4 py-3 text-sm text-muted-foreground">
                        {log.ipAddress || '-'}
                      </td>
                      <td className="px-4 py-3">
                        {log.status === 'SUCCESS' ? (
                          <div className="flex items-center gap-1 text-emerald-600">
                            <CheckCircle className="h-4 w-4" />
                            <span className="text-sm">成功</span>
                          </div>
                        ) : (
                          <div className="flex items-center gap-1 text-destructive">
                            <XCircle className="h-4 w-4" />
                            <span className="text-sm">失败</span>
                          </div>
                        )}
                      </td>
                      <td className="px-4 py-3">
                        <Button 
                          variant="ghost" 
                          size="sm"
                          onClick={() => handleViewDetail(log)}
                        >
                          <Eye className="h-4 w-4 mr-1" />
                          详情
                        </Button>
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>

          {/* 分页 */}
          {total > 0 && (
            <div className="flex items-center justify-between px-4 py-3 border-t">
              <div className="text-sm text-muted-foreground">
                共 {total} 条记录
              </div>
              <div className="flex items-center gap-2">
                <Button
                  variant="outline"
                  size="sm"
                  disabled={page <= 1}
                  onClick={() => setPage(page - 1)}
                >
                  <ChevronLeft className="h-4 w-4" />
                </Button>
                <span className="text-sm">
                  {page} / {totalPages}
                </span>
                <Button
                  variant="outline"
                  size="sm"
                  disabled={page >= totalPages}
                  onClick={() => setPage(page + 1)}
                >
                  <ChevronRight className="h-4 w-4" />
                </Button>
              </div>
            </div>
          )}
        </CardContent>
      </Card>

      {/* 详情弹窗 */}
      <Dialog open={detailOpen} onOpenChange={setDetailOpen}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>操作详情</DialogTitle>
          </DialogHeader>
          {selectedLog && (
            <div className="space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <Label className="text-sm text-muted-foreground">时间</Label>
                  <p className="text-sm mt-1">{formatDate(selectedLog.createdAt)}</p>
                </div>
                <div>
                  <Label className="text-sm text-muted-foreground">用户</Label>
                  <p className="text-sm mt-1">{selectedLog.username || '-'}</p>
                </div>
                <div>
                  <Label className="text-sm text-muted-foreground">模块</Label>
                  <p className="text-sm mt-1">{moduleLabels[selectedLog.module] || selectedLog.module}</p>
                </div>
                <div>
                  <Label className="text-sm text-muted-foreground">操作</Label>
                  <p className="text-sm mt-1">{actionLabels[selectedLog.action] || selectedLog.action}</p>
                </div>
                <div>
                  <Label className="text-sm text-muted-foreground">IP地址</Label>
                  <p className="text-sm mt-1">{selectedLog.ipAddress || '-'}</p>
                </div>
                <div>
                  <Label className="text-sm text-muted-foreground">状态</Label>
                  <p className="text-sm mt-1">{selectedLog.status === 'SUCCESS' ? '成功' : '失败'}</p>
                </div>
              </div>

              {selectedLog.targetName && (
                <div>
                  <Label className="text-sm text-muted-foreground">目标</Label>
                  <p className="text-sm mt-1">{selectedLog.targetName}</p>
                </div>
              )}

              {selectedLog.detail && (
                <div>
                  <div className="flex items-center justify-between mb-1">
                    <Label className="text-sm text-muted-foreground">详情</Label>
                    <Button
                      variant="ghost"
                      size="sm"
                      className="h-6 px-2"
                      onClick={() => handleCopy(formatJson(selectedLog.detail || '{}'))}
                    >
                      {copied ? (
                        <Check className="h-3.5 w-3.5 text-emerald-600" />
                      ) : (
                        <Copy className="h-3.5 w-3.5" />
                      )}
                      <span className="ml-1 text-xs">{copied ? '已复制' : '复制'}</span>
                    </Button>
                  </div>
                  <div className="relative bg-muted rounded-lg overflow-hidden">
                    <pre className="text-xs p-3 overflow-x-auto max-h-60 whitespace-pre-wrap break-all">
                      {formatJson(selectedLog.detail || '{}')}
                    </pre>
                  </div>
                </div>
              )}

              {selectedLog.errorMsg && (
                <div>
                  <Label className="text-sm text-muted-foreground">错误信息</Label>
                  <p className="text-sm mt-1 text-destructive">{selectedLog.errorMsg}</p>
                </div>
              )}

              {selectedLog.userAgent && (
                <div>
                  <Label className="text-sm text-muted-foreground">User Agent</Label>
                  <p className="text-xs mt-1 text-muted-foreground break-all">{selectedLog.userAgent}</p>
                </div>
              )}
            </div>
          )}
        </DialogContent>
      </Dialog>
    </div>
  )
}
