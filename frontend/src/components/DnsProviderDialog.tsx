import { useState, useEffect } from 'react'
import { Plus, Trash2, Star, StarOff } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
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
import { dnsProvidersApi, DnsProvider } from '@/api/dnsProviders'

interface DnsProviderDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onRefresh: () => void
}

export default function DnsProviderDialog({ open, onOpenChange, onRefresh }: DnsProviderDialogProps) {
  const { toast } = useToast()
  const [providers, setProviders] = useState<DnsProvider[]>([])
  const [editDialogOpen, setEditDialogOpen] = useState(false)
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
  const [editingProvider, setEditingProvider] = useState<DnsProvider | null>(null)
  const [deleteTarget, setDeleteTarget] = useState<DnsProvider | null>(null)
  
  // 表单状态
  const [formName, setFormName] = useState('')
  const [formType, setFormType] = useState<'aliyun' | 'tencent' | 'cloudflare'>('aliyun')
  const [formKeyId, setFormKeyId] = useState('')
  const [formKeySecret, setFormKeySecret] = useState('')

  const loadProviders = async () => {
    try {
      const res = await dnsProvidersApi.list()
      setProviders(res.data)
    } catch (error) {
      toast({ title: '加载 DNS 服务商列表失败', variant: 'destructive' })
    }
  }

  useEffect(() => {
    if (open) {
      loadProviders()
    }
  }, [open])

  const resetForm = () => {
    setFormName('')
    setFormType('aliyun')
    setFormKeyId('')
    setFormKeySecret('')
    setEditingProvider(null)
  }

  const openAddDialog = () => {
    resetForm()
    setEditDialogOpen(true)
  }

  const openEditDialog = (provider: DnsProvider) => {
    setEditingProvider(provider)
    setFormName(provider.name)
    setFormType(provider.providerType)
    setFormKeyId(provider.accessKeyId)
    setFormKeySecret('')
    setEditDialogOpen(true)
  }

  const handleSave = async () => {
    if (!formName.trim()) {
      toast({ title: '请输入配置名称', variant: 'destructive' })
      return
    }
    if (!formKeyId.trim()) {
      toast({ title: '请输入 Access Key ID', variant: 'destructive' })
      return
    }
    if (!editingProvider && !formKeySecret.trim()) {
      toast({ title: '请输入 Access Key Secret', variant: 'destructive' })
      return
    }

    try {
      const data: Partial<DnsProvider> = {
        name: formName,
        providerType: formType,
        accessKeyId: formKeyId,
        accessKeySecret: formKeySecret || undefined,
      }

      if (editingProvider) {
        await dnsProvidersApi.update(editingProvider.id, data)
        toast({ title: '更新成功' })
      } else {
        await dnsProvidersApi.create(data)
        toast({ title: '添加成功' })
      }

      setEditDialogOpen(false)
      resetForm()
      loadProviders()
      onRefresh()
    } catch (error: unknown) {
      const message = error instanceof Error ? error.message : '操作失败'
      toast({ title: message, variant: 'destructive' })
    }
  }

  const confirmDelete = (provider: DnsProvider) => {
    setDeleteTarget(provider)
    setDeleteDialogOpen(true)
  }

  const handleDelete = async () => {
    if (deleteTarget) {
      try {
        await dnsProvidersApi.delete(deleteTarget.id)
        toast({ title: '删除成功' })
        loadProviders()
        onRefresh()
      } catch (error) {
        toast({ title: '删除失败', variant: 'destructive' })
      }
    }
    setDeleteDialogOpen(false)
    setDeleteTarget(null)
  }

  const setDefault = async (provider: DnsProvider) => {
    try {
      await dnsProvidersApi.setDefault(provider.id)
      toast({ title: '已设为默认' })
      loadProviders()
      onRefresh()
    } catch (error) {
      toast({ title: '设置失败', variant: 'destructive' })
    }
  }

  return (
    <>
      <Dialog open={open} onOpenChange={onOpenChange}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>DNS 服务商配置</DialogTitle>
            <DialogDescription>
              配置 DNS 服务商用于 ACME DNS-01 验证，支持阿里云、腾讯云和 Cloudflare DNS
            </DialogDescription>
          </DialogHeader>
          
          <div className="py-4">
            <Button onClick={openAddDialog} className="mb-4">
              <Plus className="h-4 w-4 mr-2" />添加配置
            </Button>
            
            {providers.length === 0 ? (
              <div className="text-center py-8 text-muted-foreground">
                暂无 DNS 服务商配置，点击上方按钮添加
              </div>
            ) : (
              <div className="space-y-3">
                {providers.map((provider) => (
                  <div
                    key={provider.id}
                    className="flex items-center justify-between p-4 border rounded-lg"
                  >
                    <div className="flex items-center gap-3">
                      <span className="font-medium">{provider.name}</span>
                      <Badge variant={provider.providerType === 'aliyun' ? 'default' : 'secondary'}>
                        {provider.providerType === 'aliyun' ? '阿里云' : provider.providerType === 'tencent' ? '腾讯云' : 'Cloudflare'}
                      </Badge>
                      {provider.isDefault && (
                        <Badge variant="outline" className="text-yellow-600 border-yellow-500">
                          <Star className="h-3 w-3 mr-1" />默认
                        </Badge>
                      )}
                    </div>
                    <div className="flex items-center gap-2">
                      {!provider.isDefault && (
                        <Button variant="ghost" size="sm" onClick={() => setDefault(provider)}>
                          <StarOff className="h-4 w-4 mr-1" />设为默认
                        </Button>
                      )}
                      <Button variant="outline" size="sm" onClick={() => openEditDialog(provider)}>
                        编辑
                      </Button>
                      <Button
                        variant="ghost"
                        size="sm"
                        className="text-destructive"
                        onClick={() => confirmDelete(provider)}
                      >
                        <Trash2 className="h-4 w-4" />
                      </Button>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        </DialogContent>
      </Dialog>

      {/* 添加/编辑对话框 */}
      <Dialog open={editDialogOpen} onOpenChange={setEditDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{editingProvider ? '编辑配置' : '添加配置'}</DialogTitle>
          </DialogHeader>
          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <Label>配置名称 <span className="text-destructive">*</span></Label>
              <Input
                placeholder="例如: 阿里云DNS"
                value={formName}
                onChange={(e) => setFormName(e.target.value)}
              />
            </div>
            <div className="space-y-2">
              <Label>服务商类型 <span className="text-destructive">*</span></Label>
              <Select value={formType} onValueChange={(v) => setFormType(v as 'aliyun' | 'tencent' | 'cloudflare')}>
                <SelectTrigger><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="aliyun">阿里云 DNS</SelectItem>
                  <SelectItem value="tencent">腾讯云 DNSPod</SelectItem>
                  <SelectItem value="cloudflare">Cloudflare DNS</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-2">
              <Label>
                {formType === 'aliyun' ? 'AccessKey ID' : formType === 'cloudflare' ? 'API Token' : 'SecretId'} <span className="text-destructive">*</span>
              </Label>
              <Input
                placeholder={formType === 'aliyun' ? 'LTAI...' : formType === 'cloudflare' ? 'API Token...' : 'AKID...'}
                value={formKeyId}
                onChange={(e) => setFormKeyId(e.target.value)}
              />
              {formType === 'cloudflare' && (
                <p className="text-xs text-muted-foreground">
                  在 Cloudflare 控制台 "我的个人资料" &gt; "API 令牌" 中创建，需要 Zone.DNS 编辑权限
                </p>
              )}
            </div>
            {formType !== 'cloudflare' && (
              <div className="space-y-2">
                <Label>
                  {formType === 'aliyun' ? 'AccessKey Secret' : 'SecretKey'}
                  {!editingProvider && <span className="text-destructive"> *</span>}
                </Label>
                <Input
                  type="password"
                  placeholder={editingProvider ? '留空保持不变' : '请输入密钥'}
                  value={formKeySecret}
                  onChange={(e) => setFormKeySecret(e.target.value)}
                />
                {editingProvider && (
                  <p className="text-xs text-muted-foreground">如需修改密钥请填写，否则留空</p>
                )}
              </div>
            )}
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setEditDialogOpen(false)}>取消</Button>
            <Button onClick={handleSave}>保存</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* 删除确认对话框 */}
      <AlertDialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>确认删除</AlertDialogTitle>
            <AlertDialogDescription>
              确定要删除 DNS 服务商配置 <strong>{deleteTarget?.name}</strong> 吗？
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>取消</AlertDialogCancel>
            <AlertDialogAction
              onClick={handleDelete}
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
            >
              删除
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  )
}
