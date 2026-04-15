import { useState, useEffect } from 'react'
import {
  Shield, Plus, Trash2, RefreshCw,
  Settings, Globe, Ban, Loader2, ToggleLeft, ToggleRight,
  Pencil, ChevronDown, ChevronUp, X
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
  accessRuleApi,
  AccessRuleSummary,
  AccessRule,
  AccessRuleItem,
  AccessRuleDto,
} from '@/api/accessRules'
import { COUNTRY_CODES, getCountryInfo } from '@/api/accessControl'

export default function AccessControl() {
  const { toast } = useToast()

  // 状态
  const [loading, setLoading] = useState(true)
  const [syncing, setSyncing] = useState(false)
  const [rules, setRules] = useState<AccessRuleSummary[]>([])

  // 对话框状态
  const [editDialogOpen, setEditDialogOpen] = useState(false)
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
  const [editingRule, setEditingRule] = useState<AccessRule | null>(null)
  const [deleteTarget, setDeleteTarget] = useState<AccessRuleSummary | null>(null)
  const [expandedRuleId, setExpandedRuleId] = useState<number | null>(null)

  // 表单数据
  const [form, setForm] = useState<AccessRuleDto>({
    name: '',
    description: '',
    enabled: true,
    items: [],
  })

  // 加载数据
  const loadData = async () => {
    try {
      setLoading(true)
      const res = await accessRuleApi.list()
      setRules(res.data || [])
    } catch (error) {
      toast({ title: '数据加载失败', variant: 'destructive' })
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    loadData()
  }, [])

  // 同步配置
  const handleSync = async () => {
    try {
      setSyncing(true)
      await accessRuleApi.syncConfig()
      toast({ title: '配置已同步到 Nginx' })
    } catch (error) {
      toast({ title: '同步失败', variant: 'destructive' })
    } finally {
      setSyncing(false)
    }
  }

  // 打开新建对话框
  const openCreateDialog = () => {
    setEditingRule(null)
    setForm({
      name: '',
      description: '',
      enabled: true,
      items: [],
    })
    setEditDialogOpen(true)
  }

  // 打开编辑对话框
  const openEditDialog = async (rule: AccessRuleSummary) => {
    try {
      const res = await accessRuleApi.getById(rule.id)
      const detail = res.data
      setEditingRule(detail)
      setForm({
        name: detail.name,
        description: detail.description,
        enabled: detail.enabled,
        items: detail.items.map(item => ({ ...item })),
      })
      setEditDialogOpen(true)
    } catch (error) {
      toast({ title: '获取规则详情失败', variant: 'destructive' })
    }
  }

  // 保存规则（创建或更新）
  const handleSave = async () => {
    if (!form.name.trim()) {
      toast({ title: '规则名称不能为空', variant: 'destructive' })
      return
    }
    if (form.items.length === 0) {
      toast({ title: '请至少添加一条规则条目', variant: 'destructive' })
      return
    }

    try {
      if (editingRule) {
        await accessRuleApi.update(editingRule.id, form)
        toast({ title: '规则已更新' })
      } else {
        await accessRuleApi.create(form)
        toast({ title: '规则已创建' })
      }
      setEditDialogOpen(false)
      loadData()
    } catch (error: any) {
      toast({ title: error.userMessage || '保存失败', variant: 'destructive' })
    }
  }

  // 删除规则
  const handleDelete = async () => {
    if (!deleteTarget) return
    try {
      await accessRuleApi.delete(deleteTarget.id)
      toast({ title: '规则已删除' })
      loadData()
    } catch (error: any) {
      toast({ title: error.userMessage || '删除失败', variant: 'destructive' })
    }
    setDeleteDialogOpen(false)
    setDeleteTarget(null)
  }

  // 切换启用状态
  const handleToggle = async (id: number, enabled: boolean) => {
    try {
      await accessRuleApi.toggle(id, enabled)
      loadData()
    } catch (error) {
      toast({ title: '操作失败', variant: 'destructive' })
    }
  }

  // ==================== 规则条目操作 ====================

  const addIPItem = () => {
    setForm({
      ...form,
      items: [...form.items, {
        id: 0,
        ruleId: 0,
        itemType: 'ip',
        ipAddress: '',
        countryCode: '',
        action: 'block',
        note: '',
      }],
    })
  }

  const addGeoItem = () => {
    setForm({
      ...form,
      items: [...form.items, {
        id: 0,
        ruleId: 0,
        itemType: 'geo',
        ipAddress: '',
        countryCode: '',
        action: 'block',
        note: '',
      }],
    })
  }

  const removeItem = (index: number) => {
    const newItems = form.items.filter((_, i) => i !== index)
    setForm({ ...form, items: newItems })
  }

  const updateItem = (index: number, updates: Partial<AccessRuleItem>) => {
    const newItems = [...form.items]
    newItems[index] = { ...newItems[index], ...updates }
    setForm({ ...form, items: newItems })
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
          <p className="text-sm text-muted-foreground mt-1">管理访问规则，站点可多选规则进行访问控制</p>
        </div>
        <div className="flex items-center gap-3">
          <Button variant="outline" onClick={handleSync} disabled={syncing}>
            <RefreshCw className={cn("h-4 w-4 mr-2", syncing && "animate-spin")} />
            同步配置
          </Button>
          <Button onClick={openCreateDialog}>
            <Plus className="h-4 w-4 mr-2" />
            新建规则
          </Button>
        </div>
      </div>

      {/* 规则列表 */}
      {rules.length === 0 ? (
        <Card>
          <CardContent className="py-12 text-center text-muted-foreground">
            <Shield className="h-12 w-12 mx-auto mb-4 opacity-50" />
            <p>暂无访问规则</p>
            <p className="text-sm mt-1">点击"新建规则"创建第一条访问控制规则</p>
          </CardContent>
        </Card>
      ) : (
        <div className="grid gap-4">
          {rules.map((rule) => (
            <Card key={rule.id} className={cn(!rule.enabled && "opacity-60")}>
              <CardContent className="p-5">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-4">
                    <div className={cn(
                      "flex items-center justify-center h-10 w-10 rounded-lg",
                      rule.enabled ? "bg-primary/10 text-primary" : "bg-muted text-muted-foreground"
                    )}>
                      <Shield className="h-5 w-5" />
                    </div>
                    <div>
                      <div className="flex items-center gap-2">
                        <span className="font-medium">{rule.name}</span>
                        {!rule.enabled && (
                          <Badge variant="secondary" className="text-xs">已禁用</Badge>
                        )}
                      </div>
                      {rule.description && (
                        <p className="text-sm text-muted-foreground">{rule.description}</p>
                      )}
                    </div>
                  </div>

                  <div className="flex items-center gap-3">
                    {/* 规则统计 */}
                    <div className="flex items-center gap-2 text-sm text-muted-foreground">
                      {rule.ipCount > 0 && (
                        <Badge variant="outline" className="gap-1">
                          <Ban className="h-3 w-3" />
                          {rule.ipCount} IP
                        </Badge>
                      )}
                      {rule.geoCount > 0 && (
                        <Badge variant="outline" className="gap-1">
                          <Globe className="h-3 w-3" />
                          {rule.geoCount} 地区
                        </Badge>
                      )}
                      {rule.siteCount > 0 && (
                        <Badge variant="secondary" className="gap-1">
                          引用 {rule.siteCount} 站点
                        </Badge>
                      )}
                    </div>

                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => handleToggle(rule.id, !rule.enabled)}
                    >
                      {rule.enabled ? (
                        <ToggleRight className="h-4 w-4 text-emerald-500" />
                      ) : (
                        <ToggleLeft className="h-4 w-4 text-muted-foreground" />
                      )}
                    </Button>
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => setExpandedRuleId(expandedRuleId === rule.id ? null : rule.id)}
                    >
                      {expandedRuleId === rule.id ? (
                        <ChevronUp className="h-4 w-4" />
                      ) : (
                        <ChevronDown className="h-4 w-4" />
                      )}
                    </Button>
                    <Button variant="ghost" size="sm" onClick={() => openEditDialog(rule)}>
                      <Pencil className="h-4 w-4" />
                    </Button>
                    <Button
                      variant="ghost"
                      size="sm"
                      className="text-destructive hover:text-destructive"
                      onClick={() => {
                        setDeleteTarget(rule)
                        setDeleteDialogOpen(true)
                      }}
                    >
                      <Trash2 className="h-4 w-4" />
                    </Button>
                  </div>
                </div>

                {/* 展开的详情 */}
                {expandedRuleId === rule.id && (
                  <div className="mt-4 pt-4 border-t">
                    <RuleDetail ruleId={rule.id} />
                  </div>
                )}
              </CardContent>
            </Card>
          ))}
        </div>
      )}

      {/* 编辑/创建规则对话框 */}
      <Dialog open={editDialogOpen} onOpenChange={setEditDialogOpen}>
        <DialogContent className="max-w-2xl max-h-[85vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>{editingRule ? '编辑规则' : '新建规则'}</DialogTitle>
            <DialogDescription>
              创建独立的访问控制规则，每条规则可包含 IP 黑名单和 Geo 地区规则
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-5 py-4">
            {/* 基本信息 */}
            <div className="space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label>规则名称 *</Label>
                  <Input
                    placeholder="如：封锁恶意IP"
                    value={form.name}
                    onChange={(e) => setForm({ ...form, name: e.target.value })}
                  />
                </div>
                <div className="space-y-2">
                  <Label>启用状态</Label>
                  <div className="flex items-center h-9">
                    <Switch
                      checked={form.enabled}
                      onCheckedChange={(checked) => setForm({ ...form, enabled: checked })}
                    />
                    <span className="ml-2 text-sm text-muted-foreground">
                      {form.enabled ? '启用' : '禁用'}
                    </span>
                  </div>
                </div>
              </div>
              <div className="space-y-2">
                <Label>规则描述</Label>
                <Textarea
                  placeholder="描述规则的用途..."
                  value={form.description}
                  onChange={(e) => setForm({ ...form, description: e.target.value })}
                  rows={2}
                />
              </div>
            </div>

            {/* 规则条目 */}
            <div className="space-y-3">
              <div className="flex items-center justify-between">
                <Label className="text-base font-medium">规则条目</Label>
                <div className="flex gap-2">
                  <Button size="sm" variant="outline" onClick={addIPItem}>
                    <Plus className="h-3 w-3 mr-1" /> IP
                  </Button>
                  <Button size="sm" variant="outline" onClick={addGeoItem}>
                    <Plus className="h-3 w-3 mr-1" /> 地区
                  </Button>
                </div>
              </div>

              {form.items.length === 0 ? (
                <div className="text-center py-6 text-muted-foreground border rounded-lg border-dashed">
                  <p className="text-sm">暂无条目，点击上方按钮添加 IP 或地区规则</p>
                </div>
              ) : (
                <div className="space-y-2">
                  {form.items.map((item, index) => (
                    <div key={index} className="flex items-center gap-3 p-3 rounded-lg border bg-muted/30">
                      <Badge variant={item.itemType === 'ip' ? (item.action === 'block' ? 'destructive' : 'default') : 'default'} className="text-xs shrink-0">
                        {item.itemType === 'ip' ? (item.action === 'block' ? 'IP拒绝' : 'IP允许') : '地区'}
                      </Badge>

                      {item.itemType === 'ip' ? (
                        <>
                          <Input
                            placeholder="192.168.1.0/24"
                            value={item.ipAddress}
                            onChange={(e) => updateItem(index, { ipAddress: e.target.value })}
                            className="h-8 text-sm flex-1"
                          />
                          <Select
                            value={item.action}
                            onValueChange={(value: 'allow' | 'block') => updateItem(index, { action: value })}
                          >
                            <SelectTrigger className="h-8 text-sm w-24">
                              <SelectValue />
                            </SelectTrigger>
                            <SelectContent>
                              <SelectItem value="block">拒绝</SelectItem>
                              <SelectItem value="allow">允许</SelectItem>
                            </SelectContent>
                          </Select>
                        </>
                      ) : (
                        <>
                          <Select
                            value={item.countryCode}
                            onValueChange={(value) => updateItem(index, { countryCode: value })}
                          >
                            <SelectTrigger className="h-8 text-sm w-44">
                              <SelectValue placeholder="选择国家" />
                            </SelectTrigger>
                            <SelectContent className="max-h-[200px]">
                              {Object.entries(COUNTRY_CODES).map(([code, info]) => (
                                <SelectItem key={code} value={code}>
                                  {info.flag} {info.name}
                                </SelectItem>
                              ))}
                            </SelectContent>
                          </Select>
                          <Select
                            value={item.action}
                            onValueChange={(value: 'allow' | 'block') => updateItem(index, { action: value })}
                          >
                            <SelectTrigger className="h-8 text-sm w-24">
                              <SelectValue />
                            </SelectTrigger>
                            <SelectContent>
                              <SelectItem value="block">封锁</SelectItem>
                              <SelectItem value="allow">允许</SelectItem>
                            </SelectContent>
                          </Select>
                        </>
                      )}

                      <Input
                        placeholder="备注"
                        value={item.note}
                        onChange={(e) => updateItem(index, { note: e.target.value })}
                        className="h-8 text-sm w-28"
                      />

                      <Button
                        variant="ghost"
                        size="sm"
                        className="text-destructive hover:text-destructive shrink-0 h-8 w-8 p-0"
                        onClick={() => removeItem(index)}
                      >
                        <X className="h-4 w-4" />
                      </Button>
                    </div>
                  ))}
                </div>
              )}
            </div>
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => setEditDialogOpen(false)}>取消</Button>
            <Button onClick={handleSave}>{editingRule ? '保存' : '创建'}</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* 删除确认对话框 */}
      <AlertDialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>确认删除</AlertDialogTitle>
            <AlertDialogDescription>
              确定要删除规则 <strong>{deleteTarget?.name}</strong> 吗？
              {deleteTarget && deleteTarget.siteCount > 0 && (
                <span className="text-destructive block mt-1">
                  该规则正在被 {deleteTarget.siteCount} 个站点引用，无法删除。
                </span>
              )}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>取消</AlertDialogCancel>
            <AlertDialogAction
              onClick={handleDelete}
              disabled={deleteTarget ? deleteTarget.siteCount > 0 : false}
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
            >
              删除
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}

// 规则详情子组件
function RuleDetail({ ruleId }: { ruleId: number }) {
  const [rule, setRule] = useState<AccessRule | null>(null)

  useEffect(() => {
    accessRuleApi.getById(ruleId).then(res => setRule(res.data)).catch(() => {})
  }, [ruleId])

  if (!rule) {
    return <div className="flex justify-center py-4"><Loader2 className="h-4 w-4 animate-spin" /></div>
  }

  const ipBlockItems = rule.items.filter(i => i.itemType === 'ip' && i.action === 'block')
  const ipAllowItems = rule.items.filter(i => i.itemType === 'ip' && i.action !== 'block')
  const geoItems = rule.items.filter(i => i.itemType === 'geo')

  return (
    <div className="grid grid-cols-2 gap-4">
      {ipBlockItems.length > 0 && (
        <div>
          <h4 className="text-sm font-medium mb-2 flex items-center gap-2">
            <Ban className="h-4 w-4" /> IP 黑名单
          </h4>
          <div className="space-y-1">
            {ipBlockItems.map(item => (
              <div key={item.id} className="flex items-center gap-2 text-sm p-1.5 bg-muted/50 rounded">
                <code className="font-mono text-xs">{item.ipAddress}</code>
                {item.note && <span className="text-xs text-muted-foreground">- {item.note}</span>}
              </div>
            ))}
          </div>
        </div>
      )}
      {ipAllowItems.length > 0 && (
        <div>
          <h4 className="text-sm font-medium mb-2 flex items-center gap-2">
            <Shield className="h-4 w-4 text-emerald-500" /> IP 白名单
          </h4>
          <div className="space-y-1">
            {ipAllowItems.map(item => (
              <div key={item.id} className="flex items-center gap-2 text-sm p-1.5 bg-muted/50 rounded">
                <code className="font-mono text-xs">{item.ipAddress}</code>
                {item.note && <span className="text-xs text-muted-foreground">- {item.note}</span>}
              </div>
            ))}
          </div>
        </div>
      )}
      {geoItems.length > 0 && (
        <div>
          <h4 className="text-sm font-medium mb-2 flex items-center gap-2">
            <Globe className="h-4 w-4" /> Geo 规则
          </h4>
          <div className="space-y-1">
            {geoItems.map(item => {
              const country = getCountryInfo(item.countryCode)
              return (
                <div key={item.id} className="flex items-center gap-2 text-sm p-1.5 bg-muted/50 rounded">
                  <span>{country.flag}</span>
                  <span className="font-medium">{item.countryCode}</span>
                  <span className="text-xs text-muted-foreground">{country.name}</span>
                  <Badge variant={item.action === 'block' ? 'destructive' : 'default'} className="text-xs">
                    {item.action === 'block' ? '封锁' : '允许'}
                  </Badge>
                </div>
              )
            })}
          </div>
        </div>
      )}
      {ipBlockItems.length === 0 && ipAllowItems.length === 0 && geoItems.length === 0 && (
        <div className="col-span-2 text-center py-4 text-sm text-muted-foreground">该规则没有条目</div>
      )}
    </div>
  )
}
