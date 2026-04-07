import { useState, useEffect } from 'react'
import { Plus, Pencil, Trash2, Network, CheckCircle, XCircle, RefreshCw, Server, Code, Activity, HeartPulse } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Switch } from '@/components/ui/switch'
import { Badge } from '@/components/ui/badge'
import { Textarea } from '@/components/ui/textarea'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { useToast } from '@/hooks/use-toast'
import { upstreamsApi, Upstream, UpstreamServer, HealthCheckResult } from '@/api/upstreams'

interface BackendServer extends UpstreamServer {
  id: number
}

const algorithmOptions = [
  { label: '轮询 (Round Robin)', value: 'round_robin' },
  { label: '权重 (Weight)', value: 'weight' },
  { label: 'IP Hash', value: 'ip_hash' },
  { label: '最少连接 (Least Conn)', value: 'least_conn' },
]

// 将后端数据转换为前端格式
const transformFromBackend = (data: any): Upstream => ({
  id: data.id,
  name: data.name,
  lbMode: data.lbMode,
  healthCheck: data.healthCheck ?? false,
  checkInterval: data.checkInterval,
  checkPath: data.checkPath,
  checkTimeout: data.checkTimeout,
  servers: (data.servers || []).map((s: any, index: number) => ({
    id: s.id ?? index + 1,
    host: s.host,
    port: s.port,
    weight: s.weight ?? 1,
    maxFails: s.maxFails ?? 3,
    failTimeout: s.failTimeout ?? 10,
    status: s.status ?? 'up',
    backup: s.backup ?? false,
  })),
  createdAt: data.createdAt,
  updatedAt: data.updatedAt,
})

// 将前端数据转换为后端格式
const transformToBackend = (data: Upstream): any => ({
  name: data.name,
  lbMode: data.lbMode,
  healthCheck: data.healthCheck,
  checkInterval: data.checkInterval,
  checkPath: data.checkPath,
  checkTimeout: data.checkTimeout,
  servers: data.servers.map(s => ({
    host: s.host,
    port: s.port,
    weight: s.weight ?? 1,
    maxFails: s.maxFails ?? 3,
    failTimeout: s.failTimeout ?? 10,
    status: s.status ?? 'up',
    backup: s.backup ?? false,
  })),
})

let nextServerId = 100

const generateUpstreamConfig = (upstream: Upstream): string => {
  let config = `upstream ${upstream.name} {`
  if (upstream.lbMode === 'ip_hash') config += `\n    ip_hash;`
  else if (upstream.lbMode === 'least_conn') config += `\n    least_conn;`
  
  upstream.servers.forEach(server => {
    let serverLine = `    server ${server.host}:${server.port}`
    if (upstream.lbMode === 'weight' && server.weight !== 1) serverLine += ` weight=${server.weight}`
    serverLine += ` max_fails=${server.maxFails ?? 3} fail_timeout=${server.failTimeout ?? 10}s`
    if (server.status === 'down') serverLine += ' down'
    if (server.backup) serverLine += ' backup'
    config += `\n${serverLine};`
  })
  
  if (upstream.healthCheck && upstream.checkInterval) {
    config += `\n    # 健康检查配置`
    config += `\n    # check interval=${upstream.checkInterval}s timeout=${upstream.checkTimeout ?? 3}s`
    if (upstream.checkPath) config += ` type=http`
  }
  config += `\n}`
  return config
}

export default function LoadBalancer() {
  const [upstreams, setUpstreams] = useState<Upstream[]>([])
  const { toast } = useToast()
  
  const [createDialogOpen, setCreateDialogOpen] = useState(false)
  const [editDialogOpen, setEditDialogOpen] = useState(false)
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
  const [serversDialogOpen, setServersDialogOpen] = useState(false)
  const [configDialogOpen, setConfigDialogOpen] = useState(false)
  const [loading, setLoading] = useState(true)
  
  const [editingUpstream, setEditingUpstream] = useState<Upstream | null>(null)
  const [deleteTarget, setDeleteTarget] = useState<Upstream | null>(null)
  const [viewingUpstream, setViewingUpstream] = useState<Upstream | null>(null)
  const [configContent, setConfigContent] = useState('')

  const [newName, setNewName] = useState('')
  const [newAlgorithm, setNewAlgorithm] = useState<'round_robin' | 'weight' | 'ip_hash' | 'least_conn'>('round_robin')
  const [newHealthEnabled, setNewHealthEnabled] = useState(true)
  const [newHealthInterval, setNewHealthInterval] = useState('5')
  const [newHealthPath, setNewHealthPath] = useState('/health')
  const [newHealthTimeout, setNewHealthTimeout] = useState('3')

  const [newServerAddress, setNewServerAddress] = useState('')
  const [newServerPort, setNewServerPort] = useState('8080')
  const [newServerWeight, setNewServerWeight] = useState('1')
  const [newServerMaxFails, setNewServerMaxFails] = useState('3')
  const [newServerFailTimeout, setNewServerFailTimeout] = useState('10')

  // 健康检查状态
  const [healthChecking, setHealthChecking] = useState<number | null>(null) // 正在检查的 upstream ID
  const [healthResults, setHealthResults] = useState<Record<number, HealthCheckResult[]>>({})

  // 加载数据
  useEffect(() => {
    loadUpstreams()
  }, [])

  const loadUpstreams = async () => {
    try {
      setLoading(true)
      const res = await upstreamsApi.list()
      if (res.data) {
        setUpstreams(res.data.map(transformFromBackend))
      }
    } catch (error) {
      toast({ title: '加载失败', variant: 'destructive' })
    } finally {
      setLoading(false)
    }
  }

  const getAlgorithmLabel = (algo: string) => algorithmOptions.find(a => a.value === algo)?.label || algo

  const getStatusBadge = (status: 'up' | 'down' | 'checking') => {
    const config = {
      up: { variant: 'success' as const, icon: CheckCircle, text: '运行中' },
      down: { variant: 'destructive' as const, icon: XCircle, text: '已下线' },
      checking: { variant: 'warning' as const, icon: RefreshCw, text: '检测中' },
    }
    const { variant, icon: Icon, text } = config[status]
    return <Badge variant={variant} className="gap-1"><Icon className="h-3 w-3" />{text}</Badge>
  }

  const openCreate = () => {
    setNewName('')
    setNewAlgorithm('round_robin')
    setNewHealthEnabled(true)
    setNewHealthInterval('5')
    setNewHealthPath('/health')
    setNewHealthTimeout('3')
    setCreateDialogOpen(true)
  }

  const handleCreate = async () => {
    if (!newName.trim()) {
      toast({ title: '请输入 Upstream 名称', variant: 'destructive' })
      return
    }
    if (!/^[a-zA-Z][a-zA-Z0-9_]*$/.test(newName)) {
      toast({ title: '名称必须以字母开头，只能包含字母、数字和下划线', variant: 'destructive' })
      return
    }
    if (upstreams.some(u => u.name === newName)) {
      toast({ title: '该名称已存在', variant: 'destructive' })
      return
    }

    try {
      const data: Upstream = {
        name: newName,
        lbMode: newAlgorithm,
        healthCheck: newHealthEnabled,
        checkInterval: parseInt(newHealthInterval) || 5,
        checkPath: newHealthPath,
        checkTimeout: parseInt(newHealthTimeout) || 3,
        servers: [],
      }
      await upstreamsApi.create(transformToBackend(data))
      toast({ title: 'Upstream 创建成功' })
      setCreateDialogOpen(false)
      loadUpstreams()
    } catch (error: any) {
      toast({ title: error?.userMessage || error?.message || '创建失败', variant: 'destructive' })
    }
  }

  const openEdit = (upstream: Upstream) => {
    setEditingUpstream(upstream)
    setNewAlgorithm(upstream.lbMode)
    setNewHealthEnabled(upstream.healthCheck)
    setNewHealthInterval(String(upstream.checkInterval || 5))
    setNewHealthPath(upstream.checkPath || '/health')
    setNewHealthTimeout(String(upstream.checkTimeout || 3))
    setEditDialogOpen(true)
  }

  const handleSaveEdit = async () => {
    if (!editingUpstream) return
    try {
      const data: Upstream = {
        ...editingUpstream,
        lbMode: newAlgorithm,
        healthCheck: newHealthEnabled,
        checkInterval: parseInt(newHealthInterval) || 5,
        checkPath: newHealthPath,
        checkTimeout: parseInt(newHealthTimeout) || 3,
      }
      await upstreamsApi.update(editingUpstream.id!, transformToBackend(data))
      toast({ title: '配置已保存' })
      setEditDialogOpen(false)
      loadUpstreams()
    } catch (error: any) {
      toast({ title: error?.userMessage || error?.message || '保存失败', variant: 'destructive' })
    }
  }

  const confirmDelete = (upstream: Upstream) => {
    setDeleteTarget(upstream)
    setDeleteDialogOpen(true)
  }

  const handleDelete = async () => {
    if (!deleteTarget?.id) return
    try {
      await upstreamsApi.delete(deleteTarget.id)
      toast({ title: '已删除' })
      setDeleteDialogOpen(false)
      setDeleteTarget(null)
      loadUpstreams()
    } catch (error: any) {
      toast({ title: error?.userMessage || error?.message || '删除失败', variant: 'destructive' })
    }
  }

  const toggleStatus = async (upstream: Upstream, enabled: boolean) => {
    if (!upstream.id) return
    try {
      await upstreamsApi.update(upstream.id, { healthCheck: enabled } as any)
      toast({ title: `Upstream ${upstream.name} 已${enabled ? '启用' : '禁用'}` })
      loadUpstreams()
    } catch (error: any) {
      toast({ title: error?.userMessage || error?.message || '操作失败', variant: 'destructive' })
    }
  }

  const openServers = (upstream: Upstream) => {
    setViewingUpstream(upstream)
    setNewServerAddress('')
    setNewServerPort('8080')
    setNewServerWeight('1')
    setNewServerMaxFails('3')
    setNewServerFailTimeout('10')
    setServersDialogOpen(true)
  }

  const handleAddServer = async () => {
    if (!viewingUpstream?.id || !newServerAddress.trim()) {
      toast({ title: '请输入服务器地址', variant: 'destructive' })
      return
    }
    const newServer: UpstreamServer = {
      host: newServerAddress,
      port: parseInt(newServerPort) || 8080,
      weight: parseInt(newServerWeight) || 1,
      maxFails: parseInt(newServerMaxFails) || 3,
      failTimeout: parseInt(newServerFailTimeout) || 10,
      status: 'up',
      backup: false,
    }
    try {
      const updatedUpstream: Upstream = {
        ...viewingUpstream,
        servers: [...viewingUpstream.servers, { ...newServer, id: nextServerId++ }],
      }
      await upstreamsApi.update(viewingUpstream.id, transformToBackend(updatedUpstream))
      toast({ title: '服务器已添加' })
      // 重置表单
      setNewServerAddress('')
      setNewServerPort('8080')
      setNewServerWeight('1')
      setNewServerMaxFails('3')
      setNewServerFailTimeout('10')
      // 同步更新 viewingUpstream
      setViewingUpstream(updatedUpstream)
      loadUpstreams()
    } catch (error: any) {
      toast({ title: error?.userMessage || error?.message || '添加失败', variant: 'destructive' })
    }
  }

  const handleRemoveServer = async (serverId: number) => {
    if (!viewingUpstream?.id) return
    try {
      const updatedUpstream: Upstream = {
        ...viewingUpstream,
        servers: viewingUpstream.servers.filter(s => s.id !== serverId),
      }
      await upstreamsApi.update(viewingUpstream.id, transformToBackend(updatedUpstream))
      toast({ title: '服务器已移除' })
      // 同步更新 viewingUpstream
      setViewingUpstream(updatedUpstream)
      loadUpstreams()
    } catch (error: any) {
      toast({ title: error?.userMessage || error?.message || '移除失败', variant: 'destructive' })
    }
  }

  const toggleServerStatus = async (serverId: number) => {
    if (!viewingUpstream?.id) return
    const server = viewingUpstream.servers.find(s => s.id === serverId)
    if (!server) return
    
    const newStatus = server.status === 'up' ? 'down' : 'up'
    const updatedUpstream: Upstream = {
      ...viewingUpstream,
      servers: viewingUpstream.servers.map(s => 
        s.id === serverId ? { ...s, status: newStatus } : s
      ),
    }
    try {
      await upstreamsApi.update(viewingUpstream.id, transformToBackend(updatedUpstream))
      toast({ title: `服务器已${newStatus === 'up' ? '上线' : '下线'}` })
      setViewingUpstream(updatedUpstream)
      loadUpstreams()
    } catch (error: any) {
      toast({ title: error?.userMessage || error?.message || '操作失败', variant: 'destructive' })
    }
  }

  const toggleServerBackup = async (serverId: number) => {
    if (!viewingUpstream?.id) return
    const server = viewingUpstream.servers.find(s => s.id === serverId)
    if (!server) return
    
    const updatedUpstream: Upstream = {
      ...viewingUpstream,
      servers: viewingUpstream.servers.map(s => 
        s.id === serverId ? { ...s, backup: !s.backup } : s
      ),
    }
    try {
      await upstreamsApi.update(viewingUpstream.id, transformToBackend(updatedUpstream))
      toast({ title: `已${!server.backup ? '设为' : '取消'}备用服务器` })
      setViewingUpstream(updatedUpstream)
      loadUpstreams()
    } catch (error: any) {
      toast({ title: error?.userMessage || error?.message || '操作失败', variant: 'destructive' })
    }
  }

  // 健康检查
  const handleHealthCheck = async (upstream: Upstream) => {
    if (!upstream.id || upstream.servers.length === 0) {
      toast({ title: '没有服务器可供检查', variant: 'destructive' })
      return
    }
    setHealthChecking(upstream.id)
    try {
      const res = await upstreamsApi.healthCheck(upstream.id)
      if (res.data) {
        setHealthResults(prev => ({ ...prev, [upstream.id!]: res.data! }))
        const healthyCount = res.data!.filter(r => r.healthy).length
        toast({ title: `健康检查完成: ${healthyCount}/${res.data!.length} 健康` })
        // 刷新列表以更新服务器状态
        loadUpstreams()
      }
    } catch (error: any) {
      toast({ title: error?.userMessage || error?.message || '健康检查失败', variant: 'destructive' })
    } finally {
      setHealthChecking(null)
    }
  }

  const openConfig = (upstream: Upstream) => {
    setConfigContent(generateUpstreamConfig(upstream))
    setConfigDialogOpen(true)
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-semibold">负载均衡</h1>
        <Button onClick={openCreate}><Plus className="h-4 w-4 mr-2" />新建 Upstream</Button>
      </div>

      <div className="grid gap-4">
        {loading ? (
          <div className="text-center py-8 text-muted-foreground">加载中...</div>
        ) : upstreams.length === 0 ? (
          <div className="text-center py-12 text-muted-foreground border-2 border-dashed rounded-xl">
            <Network className="h-10 w-10 mx-auto mb-3 opacity-50" />
            <p>暂无 Upstream 配置</p>
            <p className="text-sm mt-1">点击上方按钮创建负载均衡组</p>
          </div>
        ) : (
          upstreams.map((upstream) => (
            <Card key={upstream.id} className="overflow-hidden">
              <CardContent className="p-0">
                <div className="flex items-center justify-between p-5 border-b">
                  <div className="flex items-center gap-4">
                    <div className="w-10 h-10 rounded-lg bg-primary/10 flex items-center justify-center">
                      <Network className="h-5 w-5 text-primary" />
                    </div>
                    <div>
                      <code className="font-mono font-semibold text-lg">{upstream.name}</code>
                      <div className="flex items-center gap-2 mt-1">
                        <Badge variant="secondary" className="text-xs">{getAlgorithmLabel(upstream.lbMode)}</Badge>
                        <Badge variant={upstream.healthCheck ? 'success' : 'outline'} className="text-xs">
                          {upstream.healthCheck ? '健康检查已启用' : '无健康检查'}
                        </Badge>
                      </div>
                    </div>
                  </div>
                  <div className="flex items-center gap-2">
                    <Button 
                      variant="outline" 
                      size="sm" 
                      onClick={() => handleHealthCheck(upstream)}
                      disabled={healthChecking === upstream.id || upstream.servers.length === 0}
                    >
                      {healthChecking === upstream.id ? (
                        <RefreshCw className="h-4 w-4 mr-1 animate-spin" />
                      ) : (
                        <HeartPulse className="h-4 w-4 mr-1" />
                      )}
                      检测
                    </Button>
                    <Button variant="outline" size="sm" onClick={() => openServers(upstream)}>
                      <Server className="h-4 w-4 mr-1" />服务器 ({upstream.servers.length})
                    </Button>
                    <Button variant="ghost" size="sm" onClick={() => openConfig(upstream)}>
                      <Code className="h-4 w-4 mr-1" />配置
                    </Button>
                    <Button variant="ghost" size="sm" onClick={() => openEdit(upstream)}>
                      <Pencil className="h-4 w-4 mr-1" />编辑
                    </Button>
                    <Button variant="ghost" size="sm" className="text-destructive hover:text-destructive" onClick={() => confirmDelete(upstream)}>
                      <Trash2 className="h-4 w-4" />
                    </Button>
                  </div>
                </div>
                
                {/* 服务器状态栏 */}
                {upstream.servers.length > 0 && (
                  <div className="p-4 bg-muted/30">
                    <div className="flex items-center gap-4 text-sm">
                      <div className="flex items-center gap-2">
                        <span className="text-muted-foreground">服务器状态:</span>
                        <Badge variant="success" className="gap-1">
                          <CheckCircle className="h-3 w-3" />
                          {upstream.servers.filter(s => s.status === 'up').length} 在线
                        </Badge>
                        {upstream.servers.filter(s => s.status === 'down').length > 0 && (
                          <Badge variant="destructive" className="gap-1">
                            <XCircle className="h-3 w-3" />
                            {upstream.servers.filter(s => s.status === 'down').length} 离线
                          </Badge>
                        )}
                        {upstream.servers.filter(s => s.backup).length > 0 && (
                          <Badge variant="outline" className="gap-1">
                            {upstream.servers.filter(s => s.backup).length} 备用
                          </Badge>
                        )}
                      </div>
                      <div className="flex-1" />
                      <div className="flex items-center gap-1 text-muted-foreground">
                        {upstream.servers.slice(0, 5).map((server, idx) => (
                          <div
                            key={idx}
                            className={`w-2.5 h-2.5 rounded-full ${
                              server.status === 'up'
                                ? server.backup ? 'bg-blue-400' : 'bg-green-400'
                                : 'bg-red-400'
                            }`}
                            title={`${server.host}:${server.port} - ${server.status}${server.backup ? ' (备用)' : ''}`}
                          />
                        ))}
                        {upstream.servers.length > 5 && (
                          <span className="text-xs ml-1">+{upstream.servers.length - 5}</span>
                        )}
                      </div>
                    </div>

                    {/* 健康检查结果 */}
                    {healthResults[upstream.id!] && (
                      <div className="mt-3 pt-3 border-t border-muted">
                        <div className="text-xs text-muted-foreground mb-2 flex items-center gap-1">
                          <Activity className="h-3 w-3" />
                          最近检测结果:
                        </div>
                        <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 gap-2">
                          {healthResults[upstream.id!].map((result, idx) => (
                            <div key={idx} className="flex items-center gap-2 p-2 bg-background rounded-lg text-xs">
                              <div className={`w-2 h-2 rounded-full ${result.healthy ? 'bg-green-500' : 'bg-red-500'}`} />
                              <div className="flex-1 min-w-0">
                                <div className="font-mono truncate">{result.serverHost}:{result.serverPort}</div>
                                <div className="text-muted-foreground">
                                  {result.healthy ? `${result.responseTime}ms` : result.error || '失败'}
                                </div>
                              </div>
                            </div>
                          ))}
                        </div>
                      </div>
                    )}
                  </div>
                )}
              </CardContent>
            </Card>
          ))
        )}
      </div>

      {/* Create Dialog */}
      <Dialog open={createDialogOpen} onOpenChange={setCreateDialogOpen}>
        <DialogContent>
          <DialogHeader><DialogTitle>新建 Upstream</DialogTitle></DialogHeader>
          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <Label>名称 <span className="text-destructive">*</span></Label>
              <Input placeholder="例如: api_backend" value={newName} onChange={(e) => setNewName(e.target.value)} />
            </div>
            <div className="space-y-2">
              <Label>负载均衡算法</Label>
              <Select value={newAlgorithm} onValueChange={(v: any) => setNewAlgorithm(v)}>
                <SelectTrigger><SelectValue /></SelectTrigger>
                <SelectContent>{algorithmOptions.map(opt => <SelectItem key={opt.value} value={opt.value}>{opt.label}</SelectItem>)}</SelectContent>
              </Select>
            </div>
            <div className="flex items-center gap-2">
              <Switch checked={newHealthEnabled} onCheckedChange={setNewHealthEnabled} />
              <Label>启用健康检查</Label>
            </div>
            {newHealthEnabled && (
              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label>检查间隔 (秒)</Label>
                  <Input type="number" value={newHealthInterval} onChange={(e) => setNewHealthInterval(e.target.value)} />
                </div>
                <div className="space-y-2">
                  <Label>超时时间 (秒)</Label>
                  <Input type="number" value={newHealthTimeout} onChange={(e) => setNewHealthTimeout(e.target.value)} />
                </div>
                <div className="col-span-2 space-y-2">
                  <Label>检查路径</Label>
                  <Input placeholder="/health" value={newHealthPath} onChange={(e) => setNewHealthPath(e.target.value)} />
                </div>
              </div>
            )}
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setCreateDialogOpen(false)}>取消</Button>
            <Button onClick={handleCreate}>创建</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Edit Dialog */}
      <Dialog open={editDialogOpen} onOpenChange={setEditDialogOpen}>
        <DialogContent>
          <DialogHeader><DialogTitle>编辑 Upstream - {editingUpstream?.name}</DialogTitle></DialogHeader>
          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <Label>负载均衡算法</Label>
              <Select value={newAlgorithm} onValueChange={(v: any) => setNewAlgorithm(v)}>
                <SelectTrigger><SelectValue /></SelectTrigger>
                <SelectContent>{algorithmOptions.map(opt => <SelectItem key={opt.value} value={opt.value}>{opt.label}</SelectItem>)}</SelectContent>
              </Select>
            </div>
            <div className="flex items-center gap-2">
              <Switch checked={newHealthEnabled} onCheckedChange={setNewHealthEnabled} />
              <Label>启用健康检查</Label>
            </div>
            {newHealthEnabled && (
              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label>检查间隔 (秒)</Label>
                  <Input type="number" value={newHealthInterval} onChange={(e) => setNewHealthInterval(e.target.value)} />
                </div>
                <div className="space-y-2">
                  <Label>超时时间 (秒)</Label>
                  <Input type="number" value={newHealthTimeout} onChange={(e) => setNewHealthTimeout(e.target.value)} />
                </div>
                <div className="col-span-2 space-y-2">
                  <Label>检查路径</Label>
                  <Input placeholder="/health" value={newHealthPath} onChange={(e) => setNewHealthPath(e.target.value)} />
                </div>
              </div>
            )}
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setEditDialogOpen(false)}>取消</Button>
            <Button onClick={handleSaveEdit}>保存</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Servers Dialog */}
      <Dialog open={serversDialogOpen} onOpenChange={setServersDialogOpen}>
        <DialogContent className="max-w-4xl">
          <DialogHeader><DialogTitle>服务器管理 - {viewingUpstream?.name}</DialogTitle></DialogHeader>
          <div className="space-y-4 py-4">
            {/* 添加服务器表单 */}
            <Card>
              <CardContent className="p-4">
                <div className="grid grid-cols-6 gap-3 items-end">
                  <div className="col-span-2 space-y-1">
                    <Label className="text-xs text-muted-foreground">服务器地址 *</Label>
                    <Input placeholder="192.168.1.10" value={newServerAddress} onChange={(e) => setNewServerAddress(e.target.value)} />
                  </div>
                  <div className="space-y-1">
                    <Label className="text-xs text-muted-foreground">端口</Label>
                    <Input type="number" value={newServerPort} onChange={(e) => setNewServerPort(e.target.value)} />
                  </div>
                  {viewingUpstream?.lbMode === 'weight' && (
                    <div className="space-y-1">
                      <Label className="text-xs text-muted-foreground">权重</Label>
                      <Input type="number" value={newServerWeight} onChange={(e) => setNewServerWeight(e.target.value)} />
                    </div>
                  )}
                  <div className="space-y-1">
                    <Label className="text-xs text-muted-foreground">最大失败</Label>
                    <Input type="number" value={newServerMaxFails} onChange={(e) => setNewServerMaxFails(e.target.value)} />
                  </div>
                  <Button onClick={handleAddServer} className="self-end"><Plus className="h-4 w-4 mr-1" />添加</Button>
                </div>
              </CardContent>
            </Card>
            
            {/* 服务器列表 */}
            <div className="space-y-2">
              {viewingUpstream?.servers.map((server) => (
                <div key={server.id} className="flex items-center justify-between p-4 bg-muted/50 rounded-xl border">
                  <div className="flex items-center gap-4">
                    <div className="flex items-center gap-2">
                      <div className={`w-2 h-2 rounded-full ${server.status === 'up' ? 'bg-green-500' : 'bg-red-500'}`} />
                      <code className="font-mono text-sm font-medium">{server.host}:{server.port}</code>
                    </div>
                    {viewingUpstream.lbMode === 'weight' && (
                      <Badge variant="outline" className="text-xs">weight={server.weight}</Badge>
                    )}
                    {server.backup && (
                      <Badge variant="secondary" className="text-xs">备用</Badge>
                    )}
                    <span className="text-xs text-muted-foreground">
                      max_fails={server.maxFails}, timeout={server.failTimeout}s
                    </span>
                  </div>
                  <div className="flex items-center gap-2">
                    <Button 
                      variant="ghost" 
                      size="sm" 
                      onClick={() => toggleServerStatus(server.id!)}
                      className={server.status === 'up' ? 'text-orange-600' : 'text-green-600'}
                    >
                      {server.status === 'up' ? '下线' : '上线'}
                    </Button>
                    <Button 
                      variant="ghost" 
                      size="sm" 
                      onClick={() => toggleServerBackup(server.id!)}
                      className={server.backup ? 'text-blue-600' : ''}
                    >
                      {server.backup ? '取消备用' : '设为备用'}
                    </Button>
                    <Button variant="ghost" size="sm" className="text-destructive hover:text-destructive" onClick={() => handleRemoveServer(server.id!)}>
                      <Trash2 className="h-4 w-4" />
                    </Button>
                  </div>
                </div>
              ))}
              {viewingUpstream?.servers.length === 0 && (
                <div className="text-center py-12 text-muted-foreground border-2 border-dashed rounded-xl">
                  <Server className="h-8 w-8 mx-auto mb-2 opacity-50" />
                  <p>暂无服务器，请添加后端服务器</p>
                </div>
              )}
            </div>
          </div>
        </DialogContent>
      </Dialog>

      {/* Config Dialog */}
      <Dialog open={configDialogOpen} onOpenChange={setConfigDialogOpen}>
        <DialogContent className="max-w-2xl">
          <DialogHeader><DialogTitle>配置预览</DialogTitle></DialogHeader>
          <pre className="bg-muted p-4 rounded-lg text-sm font-mono overflow-auto">{configContent}</pre>
        </DialogContent>
      </Dialog>

      {/* Delete Dialog */}
      <AlertDialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>确认删除</AlertDialogTitle>
            <AlertDialogDescription>确定要删除 Upstream <strong>{deleteTarget?.name}</strong> 吗？此操作不可恢复。</AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>取消</AlertDialogCancel>
            <AlertDialogAction onClick={handleDelete} className="bg-destructive text-destructive-foreground hover:bg-destructive/90">删除</AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}
