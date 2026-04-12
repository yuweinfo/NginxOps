import { useState, useEffect } from 'react'
import { 
  Shield, Globe, Ban, Plus, Trash2, RefreshCw, 
  Settings, Lock, Unlock, Loader2, ToggleLeft, ToggleRight
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Switch } from '@/components/ui/switch'
import { Badge } from '@/components/ui/badge'
import { Textarea } from '@/components/ui/textarea'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
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
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { useToast } from '@/hooks/use-toast'
import { cn } from '@/lib/utils'
import {
  accessControlApi,
  AccessControlSettings,
  IPBlacklistItem,
  GeoRuleItem,
  getCountryInfo,
  COUNTRY_CODES,
} from '@/api/accessControl'

export default function AccessControl() {
  const { toast } = useToast()
  
  // 状态
  const [loading, setLoading] = useState(true)
  const [syncing, setSyncing] = useState(false)
  
  // 全局设置
  const [settings, setSettings] = useState<AccessControlSettings>({
    geoEnabled: false,
    ipBlacklistEnabled: false,
    defaultAction: 'allow',
  })
  
  // IP 黑名单
  const [ipBlacklist, setIPBlacklist] = useState<IPBlacklistItem[]>([])
  
  // Geo 规则
  const [geoRules, setGeoRules] = useState<GeoRuleItem[]>([])
  
  // 对话框状态
  const [ipDialogOpen, setIPDialogOpen] = useState(false)
  const [geoDialogOpen, setGeoDialogOpen] = useState(false)
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
  const [editingIP, setEditingIP] = useState<IPBlacklistItem | null>(null)
  const [editingGeo, setEditingGeo] = useState<GeoRuleItem | null>(null)
  const [deleteTarget, setDeleteTarget] = useState<{ type: 'ip' | 'geo'; id: number } | null>(null)
  
  // 表单数据
  const [ipForm, setIPForm] = useState({ ipAddress: '', note: '' })
  const [geoForm, setGeoForm] = useState({ countryCode: '', action: 'block' as 'allow' | 'block', note: '' })

  // 加载数据
  const loadData = async () => {
    try {
      setLoading(true)
      const [settingsRes, ipRes, geoRes] = await Promise.all([
        accessControlApi.getSettings(),
        accessControlApi.listIPBlacklist(),
        accessControlApi.listGeoRules(),
      ])
      setSettings(settingsRes.data)
      setIPBlacklist(ipRes.data || [])
      setGeoRules(geoRes.data || [])
    } catch (error) {
      toast({ title: '数据加载失败', variant: 'destructive' })
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    loadData()
  }, [])

  // 更新设置
  const updateSettings = async (key: keyof AccessControlSettings, value: boolean | string) => {
    const newSettings = { ...settings, [key]: value }
    try {
      await accessControlApi.updateSettings(newSettings)
      setSettings(newSettings)
      toast({ title: '设置已更新' })
    } catch (error) {
      toast({ title: '更新失败', variant: 'destructive' })
    }
  }

  // 同步配置
  const handleSync = async () => {
    try {
      setSyncing(true)
      await accessControlApi.syncConfig()
      toast({ title: '配置已同步到 Nginx' })
    } catch (error) {
      toast({ title: '同步失败', variant: 'destructive' })
    } finally {
      setSyncing(false)
    }
  }

  // IP 黑名单操作
  const handleCreateIP = async () => {
    if (!ipForm.ipAddress.trim()) {
      toast({ title: '请输入 IP 地址', variant: 'destructive' })
      return
    }
    try {
      await accessControlApi.createIPBlacklist({
        ipAddress: ipForm.ipAddress,
        note: ipForm.note,
        enabled: true,
      })
      toast({ title: 'IP 已添加到黑名单' })
      setIPDialogOpen(false)
      setIPForm({ ipAddress: '', note: '' })
      loadData()
    } catch (error: any) {
      toast({ title: error.userMessage || '添加失败', variant: 'destructive' })
    }
  }

  const handleToggleIP = async (id: number, enabled: boolean) => {
    try {
      await accessControlApi.toggleIPBlacklist(id, enabled)
      loadData()
    } catch (error) {
      toast({ title: '操作失败', variant: 'destructive' })
    }
  }

  const handleDeleteIP = async (id: number) => {
    try {
      await accessControlApi.deleteIPBlacklist(id)
      toast({ title: '已删除' })
      loadData()
    } catch (error) {
      toast({ title: '删除失败', variant: 'destructive' })
    }
  }

  // Geo 规则操作
  const handleCreateGeo = async () => {
    if (!geoForm.countryCode.trim()) {
      toast({ title: '请选择国家', variant: 'destructive' })
      return
    }
    try {
      await accessControlApi.createGeoRule({
        countryCode: geoForm.countryCode.toUpperCase(),
        action: geoForm.action,
        note: geoForm.note,
        enabled: true,
      })
      toast({ title: 'Geo 规则已添加' })
      setGeoDialogOpen(false)
      setGeoForm({ countryCode: '', action: 'block', note: '' })
      loadData()
    } catch (error: any) {
      toast({ title: error.userMessage || '添加失败', variant: 'destructive' })
    }
  }

  const handleToggleGeo = async (id: number, enabled: boolean) => {
    try {
      await accessControlApi.toggleGeoRule(id, enabled)
      loadData()
    } catch (error) {
      toast({ title: '操作失败', variant: 'destructive' })
    }
  }

  const handleDeleteGeo = async (id: number) => {
    try {
      await accessControlApi.deleteGeoRule(id)
      toast({ title: '已删除' })
      loadData()
    } catch (error) {
      toast({ title: '删除失败', variant: 'destructive' })
    }
  }

  // 确认删除
  const confirmDelete = () => {
    if (!deleteTarget) return
    if (deleteTarget.type === 'ip') {
      handleDeleteIP(deleteTarget.id)
    } else {
      handleDeleteGeo(deleteTarget.id)
    }
    setDeleteDialogOpen(false)
    setDeleteTarget(null)
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center py-12">
        <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
      </div>
    )
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold">访问控制</h1>
          <p className="text-sm text-muted-foreground mt-1">管理 Geo/IP 封锁规则</p>
        </div>
        <Button variant="outline" onClick={handleSync} disabled={syncing}>
          <RefreshCw className={cn("h-4 w-4 mr-2", syncing && "animate-spin")} />
          同步配置
        </Button>
      </div>

      {/* 全局设置 */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Settings className="h-5 w-5" />
            全局设置
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-3 gap-6">
            <div className="flex items-center justify-between p-4 bg-muted/50 rounded-lg">
              <div className="flex items-center gap-3">
                <Globe className="h-5 w-5 text-muted-foreground" />
                <div>
                  <div className="font-medium">Geo 封锁</div>
                  <div className="text-xs text-muted-foreground">按国家/地区控制访问</div>
                </div>
              </div>
              <Switch
                checked={settings.geoEnabled}
                onCheckedChange={(checked) => updateSettings('geoEnabled', checked)}
              />
            </div>
            
            <div className="flex items-center justify-between p-4 bg-muted/50 rounded-lg">
              <div className="flex items-center gap-3">
                <Ban className="h-5 w-5 text-muted-foreground" />
                <div>
                  <div className="font-medium">IP 黑名单</div>
                  <div className="text-xs text-muted-foreground">封锁指定 IP/网段</div>
                </div>
              </div>
              <Switch
                checked={settings.ipBlacklistEnabled}
                onCheckedChange={(checked) => updateSettings('ipBlacklistEnabled', checked)}
              />
            </div>
            
            <div className="flex items-center justify-between p-4 bg-muted/50 rounded-lg">
              <div className="flex items-center gap-3">
                {settings.defaultAction === 'allow' ? (
                  <Unlock className="h-5 w-5 text-emerald-500" />
                ) : (
                  <Lock className="h-5 w-5 text-destructive" />
                )}
                <div>
                  <div className="font-medium">默认策略</div>
                  <div className="text-xs text-muted-foreground">未匹配规则时的动作</div>
                </div>
              </div>
              <Select
                value={settings.defaultAction}
                onValueChange={(value) => updateSettings('defaultAction', value)}
              >
                <SelectTrigger className="w-28">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="allow">允许所有</SelectItem>
                  <SelectItem value="block">封锁所有</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* IP 黑名单 & Geo 规则 */}
      <div className="grid grid-cols-2 gap-6">
        {/* IP 黑名单 */}
        <Card>
          <CardHeader className="flex flex-row items-center justify-between">
            <CardTitle className="flex items-center gap-2">
              <Ban className="h-5 w-5" />
              IP 黑名单
              <Badge variant="secondary">{ipBlacklist.length}</Badge>
            </CardTitle>
            <Button size="sm" onClick={() => setIPDialogOpen(true)}>
              <Plus className="h-4 w-4 mr-1" /> 添加
            </Button>
          </CardHeader>
          <CardContent>
            {ipBlacklist.length === 0 ? (
              <div className="text-center py-8 text-muted-foreground">
                <Ban className="h-8 w-8 mx-auto mb-2 opacity-50" />
                <p>暂无 IP 黑名单</p>
              </div>
            ) : (
              <div className="space-y-2 max-h-[400px] overflow-y-auto">
                {ipBlacklist.map((item) => (
                  <div
                    key={item.id}
                    className={cn(
                      "flex items-center justify-between p-3 rounded-lg border",
                      item.enabled ? "bg-card" : "bg-muted/30 opacity-60"
                    )}
                  >
                    <div className="flex items-center gap-3">
                      <code className="text-sm font-mono">{item.ipAddress}</code>
                      {item.note && (
                        <span className="text-xs text-muted-foreground">{item.note}</span>
                      )}
                    </div>
                    <div className="flex items-center gap-2">
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => handleToggleIP(item.id, !item.enabled)}
                      >
                        {item.enabled ? (
                          <ToggleRight className="h-4 w-4 text-emerald-500" />
                        ) : (
                          <ToggleLeft className="h-4 w-4 text-muted-foreground" />
                        )}
                      </Button>
                      <Button
                        variant="ghost"
                        size="sm"
                        className="text-destructive hover:text-destructive"
                        onClick={() => {
                          setDeleteTarget({ type: 'ip', id: item.id })
                          setDeleteDialogOpen(true)
                        }}
                      >
                        <Trash2 className="h-4 w-4" />
                      </Button>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </CardContent>
        </Card>

        {/* Geo 规则 */}
        <Card>
          <CardHeader className="flex flex-row items-center justify-between">
            <CardTitle className="flex items-center gap-2">
              <Globe className="h-5 w-5" />
              Geo 规则
              <Badge variant="secondary">{geoRules.length}</Badge>
            </CardTitle>
            <Button size="sm" onClick={() => setGeoDialogOpen(true)}>
              <Plus className="h-4 w-4 mr-1" /> 添加
            </Button>
          </CardHeader>
          <CardContent>
            {geoRules.length === 0 ? (
              <div className="text-center py-8 text-muted-foreground">
                <Globe className="h-8 w-8 mx-auto mb-2 opacity-50" />
                <p>暂无 Geo 规则</p>
              </div>
            ) : (
              <div className="space-y-2 max-h-[400px] overflow-y-auto">
                {geoRules.map((item) => {
                  const country = getCountryInfo(item.countryCode)
                  return (
                    <div
                      key={item.id}
                      className={cn(
                        "flex items-center justify-between p-3 rounded-lg border",
                        item.enabled ? "bg-card" : "bg-muted/30 opacity-60"
                      )}
                    >
                      <div className="flex items-center gap-3">
                        <span className="text-lg">{country.flag}</span>
                        <span className="text-sm font-medium">{item.countryCode}</span>
                        <span className="text-xs text-muted-foreground">{country.name}</span>
                        <Badge 
                          variant={item.action === 'block' ? 'destructive' : 'default'}
                          className="text-xs"
                        >
                          {item.action === 'block' ? '封锁' : '允许'}
                        </Badge>
                      </div>
                      <div className="flex items-center gap-2">
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => handleToggleGeo(item.id, !item.enabled)}
                        >
                          {item.enabled ? (
                            <ToggleRight className="h-4 w-4 text-emerald-500" />
                          ) : (
                            <ToggleLeft className="h-4 w-4 text-muted-foreground" />
                          )}
                        </Button>
                        <Button
                          variant="ghost"
                          size="sm"
                          className="text-destructive hover:text-destructive"
                          onClick={() => {
                            setDeleteTarget({ type: 'geo', id: item.id })
                            setDeleteDialogOpen(true)
                          }}
                        >
                          <Trash2 className="h-4 w-4" />
                        </Button>
                      </div>
                    </div>
                  )
                })}
              </div>
            )}
          </CardContent>
        </Card>
      </div>

      {/* 添加 IP 对话框 */}
      <Dialog open={ipDialogOpen} onOpenChange={setIPDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>添加 IP 到黑名单</DialogTitle>
            <DialogDescription>
              支持单个 IP 或 CIDR 网段，如 192.168.1.1 或 10.0.0.0/8
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <Label>IP 地址 / 网段</Label>
              <Input
                placeholder="192.168.1.100 或 10.0.0.0/8"
                value={ipForm.ipAddress}
                onChange={(e) => setIPForm({ ...ipForm, ipAddress: e.target.value })}
              />
            </div>
            <div className="space-y-2">
              <Label>备注（可选）</Label>
              <Textarea
                placeholder="说明封禁原因..."
                value={ipForm.note}
                onChange={(e) => setIPForm({ ...ipForm, note: e.target.value })}
                rows={2}
              />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setIPDialogOpen(false)}>取消</Button>
            <Button onClick={handleCreateIP}>添加</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* 添加 Geo 规则对话框 */}
      <Dialog open={geoDialogOpen} onOpenChange={setGeoDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>添加 Geo 规则</DialogTitle>
            <DialogDescription>
              根据访问者的地理位置控制访问权限
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <Label>国家/地区</Label>
              <Select
                value={geoForm.countryCode}
                onValueChange={(value) => setGeoForm({ ...geoForm, countryCode: value })}
              >
                <SelectTrigger>
                  <SelectValue placeholder="选择国家/地区" />
                </SelectTrigger>
                <SelectContent className="max-h-[300px]">
                  {Object.entries(COUNTRY_CODES).map(([code, info]) => (
                    <SelectItem key={code} value={code}>
                      {info.flag} {info.name} ({code})
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-2">
              <Label>动作</Label>
              <Select
                value={geoForm.action}
                onValueChange={(value: 'allow' | 'block') => setGeoForm({ ...geoForm, action: value })}
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="block">封锁</SelectItem>
                  <SelectItem value="allow">允许</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-2">
              <Label>备注（可选）</Label>
              <Textarea
                placeholder="说明规则用途..."
                value={geoForm.note}
                onChange={(e) => setGeoForm({ ...geoForm, note: e.target.value })}
                rows={2}
              />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setGeoDialogOpen(false)}>取消</Button>
            <Button onClick={handleCreateGeo}>添加</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* 删除确认对话框 */}
      <AlertDialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>确认删除</AlertDialogTitle>
            <AlertDialogDescription>
              确定要删除此规则吗？此操作不可恢复。
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>取消</AlertDialogCancel>
            <AlertDialogAction onClick={confirmDelete} className="bg-destructive text-destructive-foreground hover:bg-destructive/90">
              删除
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}
