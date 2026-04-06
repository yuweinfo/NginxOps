import { useState, useEffect } from 'react'
import { Plus, Pencil, Trash2, Network, CheckCircle, XCircle, RefreshCw, Server, Code } from 'lucide-react'
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
import { upstreamsApi, Upstream, UpstreamServer } from '@/api/upstreams'

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
    if (!/^[a-z][a-z0-9_]*$/.test(newName)) {
      toast({ title: '名称只能包含小写字母、数字和下划线，且以字母开头', variant: 'destructive' })
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
      toast({ title: error?.message || '创建失败', variant: 'destructive' })
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
      toast({ title: error?.message || '保存失败', variant: 'destructive' })
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
      toast({ title: error?.message || '删除失败', variant: 'destructive' })
    }
  }

  const toggleStatus = async (upstream: Upstream, enabled: boolean) => {
    if (!upstream.id) return
    try {
      await upstreamsApi.update(upstream.id, { healthCheck: enabled } as any)
      toast({ title: `Upstream ${upstream.name} 已${enabled ? '启用' : '禁用'}` })
      loadUpstreams()
    } catch (error: any) {
      toast({ title: error?.message || '操作失败', variant: 'destructive' })
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
    }
    try {
      const updatedUpstream: Upstream = {
        ...viewingUpstream,
        servers: [...viewingUpstream.servers, { ...newServer, id: nextServerId++ }],
      }
      await upstreamsApi.update(viewingUpstream.id, transformToBackend(updatedUpstream))
      toast({ title: '服务器已添加' })
      setNewServerAddress('')
      loadUpstreams()
    } catch (error: any) {
      toast({ title: error?.message || '添加失败', variant: 'destructive' })
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
      loadUpstreams()
    } catch (error: any) {
      toast({ title: error?.message || '移除失败', variant: 'destructive' })
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
          <div className="text-center py-8 text-muted-foreground">
            暂无 Upstream 配置，点击上方按钮创建
          </div>
        ) : (
          upstreams.map((upstream) => (
            <Card key={upstream.id}>
              <CardContent className="p-6">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-6">
                    <div className="flex items-center gap-2">
                      <Network className="h-4 w-4 text-primary" />
                      <code className="font-mono font-semibold">{upstream.name}</code>
                    </div>
                    <Badge variant="secondary">{getAlgorithmLabel(upstream.lbMode)}</Badge>
                    <div className="flex items-center gap-2 text-sm text-muted-foreground">
                      <Server className="h-4 w-4" />
                      <span>{upstream.servers.length} 台</span>
                      <Badge variant="outline" className="text-green-600 border-green-600">
                        {upstream.servers.filter(s => s.status === 'up').length} 在线
                      </Badge>
                    </div>
                    <Badge variant={upstream.healthCheck ? 'success' : 'secondary'}>
                      {upstream.healthCheck ? '健康检查' : '无健康检查'}
                    </Badge>
                  </div>
                  <div className="flex items-center gap-4">
                    <div className="flex items-center gap-2">
                      <Switch checked={upstream.healthCheck} onCheckedChange={(checked) => toggleStatus(upstream, checked)} />
                      <span className="text-sm text-muted-foreground">{upstream.healthCheck ? '启用' : '禁用'}</span>
                    </div>
                    <Button variant="ghost" size="sm" onClick={() => openServers(upstream)}>服务器</Button>
                    <Button variant="ghost" size="sm" onClick={() => openConfig(upstream)}><Code className="h-4 w-4 mr-1" />配置</Button>
                    <Button variant="ghost" size="sm" onClick={() => openEdit(upstream)}><Pencil className="h-4 w-4 mr-1" />编辑</Button>
                    <Button variant="ghost" size="sm" className="text-destructive" onClick={() => confirmDelete(upstream)}><Trash2 className="h-4 w-4 mr-1" />删除</Button>
                  </div>
                </div>
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

      {/* Servers Dialog */}
      <Dialog open={serversDialogOpen} onOpenChange={setServersDialogOpen}>
        <DialogContent className="max-w-4xl">
          <DialogHeader><DialogTitle>服务器管理 - {viewingUpstream?.name}</DialogTitle></DialogHeader>
          <div className="space-y-4 py-4">
            <Card>
              <CardContent className="p-4">
                <div className="flex flex-wrap gap-3 items-end">
                  <div className="space-y-1">
                    <Label className="text-xs">地址</Label>
                    <Input className="w-32" placeholder="192.168.1.10" value={newServerAddress} onChange={(e) => setNewServerAddress(e.target.value)} />
                  </div>
                  <div className="space-y-1">
                    <Label className="text-xs">端口</Label>
                    <Input className="w-20" type="number" value={newServerPort} onChange={(e) => setNewServerPort(e.target.value)} />
                  </div>
                  {viewingUpstream?.lbMode === 'weight' && (
                    <div className="space-y-1">
                      <Label className="text-xs">权重</Label>
                      <Input className="w-16" type="number" value={newServerWeight} onChange={(e) => setNewServerWeight(e.target.value)} />
                    </div>
                  )}
                  <Button onClick={handleAddServer}><Plus className="h-4 w-4 mr-1" />添加</Button>
                </div>
              </CardContent>
            </Card>
            <div className="space-y-2">
              {viewingUpstream?.servers.map((server) => (
                <div key={server.id} className="flex items-center justify-between p-3 bg-muted rounded-lg">
                  <div className="flex items-center gap-4">
                    <code className="font-mono text-sm">{server.host}:{server.port}</code>
                    {viewingUpstream.lbMode === 'weight' && <Badge variant="outline">weight={server.weight}</Badge>}
                    <span className="text-sm text-muted-foreground">max_fails={server.maxFails}, fail_timeout={server.failTimeout}s</span>
                  </div>
                  <div className="flex items-center gap-2">
                    {getStatusBadge(server.status as 'up' | 'down' | 'checking')}
                    <Button variant="ghost" size="sm" className="text-destructive" onClick={() => handleRemoveServer(server.id!)}><Trash2 className="h-4 w-4" /></Button>
                  </div>
                </div>
              ))}
              {viewingUpstream?.servers.length === 0 && (
                <div className="text-center py-8 text-muted-foreground">暂无服务器，请添加后端服务器</div>
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
