import { useState, useEffect } from 'react'
import { Plus, Lock, Trash2, RefreshCw, Upload, Shield, FileText, Key, Settings, Star } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Switch } from '@/components/ui/switch'
import { Badge } from '@/components/ui/badge'
import { Textarea } from '@/components/ui/textarea'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
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
import { certificatesApi, Certificate, CertificateImportData } from '@/api/certificates'
import { dnsProvidersApi, DnsProvider } from '@/api/dnsProviders'
import DnsProviderDialog from '@/components/DnsProviderDialog'

export default function Certificates() {
  const [certs, setCerts] = useState<Certificate[]>([])
  const [loading, setLoading] = useState(true)
  const { toast } = useToast()
  
  // 对话框状态
  const [applyDialogOpen, setApplyDialogOpen] = useState(false)
  const [importDialogOpen, setImportDialogOpen] = useState(false)
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
  const [deleteTarget, setDeleteTarget] = useState<Certificate | null>(null)
  const [dnsProviderDialogOpen, setDnsProviderDialogOpen] = useState(false)
  
  // 申请证书表单
  const [certDomain, setCertDomain] = useState('')
  const [certIssuer, setCertIssuer] = useState('letsencrypt')
  const [certDnsProviderId, setCertDnsProviderId] = useState<number | undefined>()
  
  // DNS 服务商列表
  const [dnsProviders, setDnsProviders] = useState<DnsProvider[]>([])
  
  // 导入证书表单
  const [importDomain, setImportDomain] = useState('')
  const [importCert, setImportCert] = useState('')
  const [importKey, setImportKey] = useState('')
  const [importChain, setImportChain] = useState('')
  const [activeTab, setActiveTab] = useState('cert')

  // 文件上传处理
  const handleFileUpload = (e: React.ChangeEvent<HTMLInputElement>, type: 'cert' | 'key' | 'chain') => {
    const file = e.target.files?.[0]
    if (!file) return

    const reader = new FileReader()
    reader.onload = (event) => {
      const content = event.target?.result as string
      switch (type) {
        case 'cert':
          setImportCert(content)
          break
        case 'key':
          setImportKey(content)
          break
        case 'chain':
          setImportChain(content)
          break
      }
    }
    reader.onerror = () => {
      toast({ title: '文件读取失败', variant: 'destructive' })
    }
    reader.readAsText(file)
    // 清空 input 以便重复选择同一文件
    e.target.value = ''
  }

  // 加载证书列表
  const loadCertificates = async () => {
    try {
      setLoading(true)
      const res = await certificatesApi.list()
      setCerts(res.data)
    } catch (error) {
      toast({ title: '加载证书列表失败', variant: 'destructive' })
    } finally {
      setLoading(false)
    }
  }

  // 加载 DNS 服务商列表
  const loadDnsProviders = async () => {
    try {
      const res = await dnsProvidersApi.list()
      setDnsProviders(res.data)
    } catch (error) {
      console.error('加载 DNS 服务商列表失败', error)
    }
  }

  useEffect(() => {
    loadCertificates()
    loadDnsProviders()
  }, [])

  // 申请证书
  const handleApplyCert = async () => {
    if (!certDomain.trim()) {
      toast({ title: '域名不能为空', variant: 'destructive' })
      return
    }

    if (dnsProviders.length === 0) {
      toast({ title: '请先配置 DNS 服务商', description: '点击右上角"DNS配置"按钮添加', variant: 'destructive' })
      return
    }

    try {
      // 先创建证书记录
      const certRes = await certificatesApi.create({ domain: certDomain, autoRenew: true })
      // 然后申请证书
      await certificatesApi.request(certRes.data.id, certIssuer, certDnsProviderId)
      
      toast({ title: '证书申请已提交，正在进行 DNS 验证...' })
      setApplyDialogOpen(false)
      setCertDomain('')
      setCertIssuer('letsencrypt')
      setCertDnsProviderId(undefined)
      loadCertificates()
    } catch (error) {
      toast({ title: '证书申请失败', variant: 'destructive' })
    }
  }

  // 导入证书
  const handleImportCert = async () => {
    if (!importCert.trim()) {
      toast({ title: '证书内容不能为空', variant: 'destructive' })
      return
    }
    if (!importKey.trim()) {
      toast({ title: '私钥内容不能为空', variant: 'destructive' })
      return
    }

    try {
      const data: CertificateImportData = {
        certificate: importCert,
        privateKey: importKey,
        certificateChain: importChain || undefined,
        domain: importDomain || undefined,
      }
      
      await certificatesApi.import(data)
      toast({ title: '证书导入成功' })
      setImportDialogOpen(false)
      resetImportForm()
      loadCertificates()
    } catch (error: unknown) {
      const message = error instanceof Error ? error.message : '证书导入失败'
      toast({ title: message, variant: 'destructive' })
    }
  }

  // 重置导入表单
  const resetImportForm = () => {
    setImportDomain('')
    setImportCert('')
    setImportKey('')
    setImportChain('')
    setActiveTab('cert')
  }

  // 切换自动续期
  const toggleAutoRenew = async (cert: Certificate) => {
    try {
      await certificatesApi.toggleAutoRenew(cert.id, !cert.autoRenew)
      toast({ title: `证书 ${cert.domain} 自动续期已${cert.autoRenew ? '关闭' : '开启'}` })
      loadCertificates()
    } catch (error) {
      toast({ title: '操作失败', variant: 'destructive' })
    }
  }

  // 确认删除
  const confirmDelete = (cert: Certificate) => {
    setDeleteTarget(cert)
    setDeleteDialogOpen(true)
  }

  // 删除证书
  const handleDelete = async () => {
    if (deleteTarget) {
      try {
        await certificatesApi.delete(deleteTarget.id)
        toast({ title: '证书已删除' })
        loadCertificates()
      } catch (error) {
        toast({ title: '删除失败', variant: 'destructive' })
      }
    }
    setDeleteDialogOpen(false)
    setDeleteTarget(null)
  }

  // 续期证书
  const renewCert = async (cert: Certificate) => {
    try {
      toast({ title: `正在续期证书 ${cert.domain}...` })
      await certificatesApi.renew(cert.id)
      toast({ title: `证书 ${cert.domain} 续期成功！` })
      loadCertificates()
    } catch (error) {
      toast({ title: '续期失败', variant: 'destructive' })
    }
  }

  // 状态标签
  const getStatusBadge = (status: string) => {
    const config: Record<string, { variant: 'default' | 'destructive' | 'secondary' | 'outline', text: string }> = {
      valid: { variant: 'default', text: '有效' },
      expired: { variant: 'destructive', text: '已过期' },
      pending: { variant: 'secondary', text: '申请中' },
    }
    const c = config[status] || { variant: 'outline', text: status }
    return <Badge variant={c.variant}>{c.text}</Badge>
  }

  // 计算剩余天数
  const getDaysLeft = (expiresAt?: string) => {
    if (!expiresAt) return null
    return Math.ceil((new Date(expiresAt).getTime() - Date.now()) / (1000 * 60 * 60 * 24))
  }

  // 格式化日期
  const formatDate = (dateStr?: string) => {
    if (!dateStr) return '-'
    return new Date(dateStr).toLocaleDateString('zh-CN')
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-semibold">SSL证书管理</h1>
        <div className="flex gap-2">
          <Button variant="outline" onClick={() => setDnsProviderDialogOpen(true)}>
            <Settings className="h-4 w-4 mr-2" />DNS配置
          </Button>
          <Button variant="outline" onClick={() => setImportDialogOpen(true)}>
            <Upload className="h-4 w-4 mr-2" />导入证书
          </Button>
          <Button onClick={() => setApplyDialogOpen(true)}>
            <Plus className="h-4 w-4 mr-2" />申请证书
          </Button>
        </div>
      </div>

      {loading ? (
        <div className="text-center py-8 text-muted-foreground">加载中...</div>
      ) : certs.length === 0 ? (
        <Card>
          <CardContent className="py-12 text-center text-muted-foreground">
            <Shield className="h-12 w-12 mx-auto mb-4 opacity-50" />
            <p>暂无证书，点击上方按钮申请或导入证书</p>
          </CardContent>
        </Card>
      ) : (
        <div className="grid gap-4">
          {certs.map((cert) => {
            const daysLeft = getDaysLeft(cert.expiresAt)
            const isExpiringSoon = daysLeft !== null && daysLeft > 0 && daysLeft <= 14
            
            return (
              <Card key={cert.id}>
                <CardContent className="p-6">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-6">
                      <div className="flex items-center gap-2">
                        <Lock className="h-4 w-4 text-primary" />
                        <span className="font-semibold">{cert.domain}</span>
                      </div>
                      {cert.issuer && (
                        <Badge variant={cert.issuer.includes("Let's") ? 'default' : 'secondary'}>
                          {cert.issuer}
                        </Badge>
                      )}
                      <span className="text-sm text-muted-foreground">
                        {formatDate(cert.issuedAt)} ~ {formatDate(cert.expiresAt)}
                      </span>
                      {getStatusBadge(cert.status)}
                      {isExpiringSoon && cert.status === 'valid' && (
                        <Badge variant="outline" className="gap-1 border-yellow-500 text-yellow-600">
                          <RefreshCw className="h-3 w-3" />即将到期 ({daysLeft}天)
                        </Badge>
                      )}
                    </div>
                    <div className="flex items-center gap-4">
                      <div className="flex items-center gap-2">
                        <Label className="text-sm">自动续期</Label>
                        <Switch 
                          checked={cert.autoRenew} 
                          onCheckedChange={() => toggleAutoRenew(cert)} 
                          disabled={cert.status !== 'valid'} 
                        />
                      </div>
                      {(cert.status === 'valid' || cert.status === 'expired') && (
                        <Button variant="outline" size="sm" onClick={() => renewCert(cert)}>
                          <RefreshCw className="h-4 w-4 mr-1" />续期
                        </Button>
                      )}
                      <Button variant="ghost" size="sm" className="text-destructive" onClick={() => confirmDelete(cert)}>
                        <Trash2 className="h-4 w-4 mr-1" />删除
                      </Button>
                    </div>
                  </div>
                </CardContent>
              </Card>
            )
          })}
        </div>
      )}

      {/* 申请证书对话框 */}
      <Dialog open={applyDialogOpen} onOpenChange={setApplyDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>申请SSL证书</DialogTitle>
          </DialogHeader>
          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <Label>域名 <span className="text-destructive">*</span></Label>
              <Input placeholder="例如: example.com 或 *.example.com" value={certDomain} onChange={(e) => setCertDomain(e.target.value)} />
            </div>
            <div className="space-y-2">
              <Label>DNS 服务商 <span className="text-destructive">*</span></Label>
              {dnsProviders.length > 0 ? (
                <Select value={certDnsProviderId?.toString() || ''} onValueChange={(v) => setCertDnsProviderId(v ? Number(v) : undefined)}>
                  <SelectTrigger><SelectValue placeholder="选择 DNS 服务商" /></SelectTrigger>
                  <SelectContent>
                    {dnsProviders.map((p) => (
                      <SelectItem key={p.id} value={p.id.toString()}>
                        {p.name} ({p.providerType === 'aliyun' ? '阿里云' : p.providerType === 'tencent' ? '腾讯云' : 'Cloudflare'})
                        {p.isDefault && <span className="ml-1 text-muted-foreground">(默认)</span>}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              ) : (
                <div className="text-sm text-muted-foreground p-3 bg-muted rounded-lg">
                  暂无 DNS 服务商配置，请先点击右上角"DNS配置"按钮添加
                </div>
              )}
            </div>
            <div className="space-y-2">
              <Label>证书颁发机构</Label>
              <Select value={certIssuer} onValueChange={setCertIssuer}>
                <SelectTrigger><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="letsencrypt">Let's Encrypt (推荐)</SelectItem>
                  <SelectItem value="zerossl">ZeroSSL</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div className="bg-muted p-4 rounded-lg text-sm text-muted-foreground">
              <p className="font-medium text-foreground mb-2">申请说明：</p>
              <ul className="list-disc list-inside space-y-1">
                <li>证书有效期：90天</li>
                <li>验证方式：DNS TXT 记录验证</li>
                <li>支持通配符域名（如 *.example.com）</li>
                <li>开启自动续期可自动续签</li>
              </ul>
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setApplyDialogOpen(false)}>取消</Button>
            <Button onClick={handleApplyCert} disabled={dnsProviders.length === 0}>申请</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* 导入证书对话框 */}
      <Dialog open={importDialogOpen} onOpenChange={setImportDialogOpen}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>导入SSL证书</DialogTitle>
            <DialogDescription>
              粘贴PEM格式的证书和私钥内容，系统将自动解析证书信息并保存到服务器
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <Label>域名（可选）</Label>
              <Input 
                placeholder="留空将自动从证书中解析域名" 
                value={importDomain} 
                onChange={(e) => setImportDomain(e.target.value)} 
              />
            </div>
            
            <Tabs value={activeTab} onValueChange={setActiveTab}>
              <TabsList className="grid w-full grid-cols-3">
                <TabsTrigger value="cert"><FileText className="h-4 w-4 mr-1" />证书</TabsTrigger>
                <TabsTrigger value="key"><Key className="h-4 w-4 mr-1" />私钥</TabsTrigger>
                <TabsTrigger value="chain">证书链</TabsTrigger>
              </TabsList>
              <TabsContent value="cert" className="mt-4">
                <div className="space-y-2">
                  <div className="flex items-center justify-between">
                    <Label>证书内容 <span className="text-destructive">*</span></Label>
                    <label className="cursor-pointer">
                      <input
                        type="file"
                        accept=".crt,.pem,.cer,.cert"
                        className="hidden"
                        onChange={(e) => handleFileUpload(e, 'cert')}
                      />
                      <Button variant="outline" size="sm" type="button" asChild>
                        <span><Upload className="h-4 w-4 mr-1" />选择文件</span>
                      </Button>
                    </label>
                  </div>
                  <Textarea
                    placeholder={'-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----'}
                    value={importCert}
                    onChange={(e) => setImportCert(e.target.value)}
                    className="font-mono text-xs min-h-[200px]"
                  />
                  <p className="text-xs text-muted-foreground">支持 .crt、.pem、.cer、.cert 格式</p>
                </div>
              </TabsContent>
              <TabsContent value="key" className="mt-4">
                <div className="space-y-2">
                  <div className="flex items-center justify-between">
                    <Label>私钥内容 <span className="text-destructive">*</span></Label>
                    <label className="cursor-pointer">
                      <input
                        type="file"
                        accept=".key,.pem"
                        className="hidden"
                        onChange={(e) => handleFileUpload(e, 'key')}
                      />
                      <Button variant="outline" size="sm" type="button" asChild>
                        <span><Upload className="h-4 w-4 mr-1" />选择文件</span>
                      </Button>
                    </label>
                  </div>
                  <Textarea
                    placeholder={'-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----'}
                    value={importKey}
                    onChange={(e) => setImportKey(e.target.value)}
                    className="font-mono text-xs min-h-[200px]"
                  />
                  <p className="text-xs text-muted-foreground">支持 .key、.pem 格式</p>
                </div>
              </TabsContent>
              <TabsContent value="chain" className="mt-4">
                <div className="space-y-2">
                  <div className="flex items-center justify-between">
                    <Label>证书链（可选）</Label>
                    <label className="cursor-pointer">
                      <input
                        type="file"
                        accept=".crt,.pem,.cer,.cert"
                        className="hidden"
                        onChange={(e) => handleFileUpload(e, 'chain')}
                      />
                      <Button variant="outline" size="sm" type="button" asChild>
                        <span><Upload className="h-4 w-4 mr-1" />选择文件</span>
                      </Button>
                    </label>
                  </div>
                  <Textarea
                    placeholder={'-----BEGIN CERTIFICATE-----\n(Intermediate CA)\n-----END CERTIFICATE-----'}
                    value={importChain}
                    onChange={(e) => setImportChain(e.target.value)}
                    className="font-mono text-xs min-h-[200px]"
                  />
                  <p className="text-xs text-muted-foreground">如有中间证书，请粘贴或选择文件，将自动合并到证书文件中</p>
                </div>
              </TabsContent>
            </Tabs>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setImportDialogOpen(false)}>取消</Button>
            <Button onClick={handleImportCert}>导入</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* 删除确认对话框 */}
      <AlertDialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>确认删除</AlertDialogTitle>
            <AlertDialogDescription>
              确定要删除证书 <strong>{deleteTarget?.domain}</strong> 吗？
              {deleteTarget?.status === 'valid' && (
                <span className="block mt-2 text-destructive">注意：删除证书后，使用该证书的站点将无法使用 HTTPS。</span>
              )}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>取消</AlertDialogCancel>
            <AlertDialogAction onClick={handleDelete} className="bg-destructive text-destructive-foreground hover:bg-destructive/90">删除</AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      {/* DNS 服务商配置对话框 */}
      <DnsProviderDialog
        open={dnsProviderDialogOpen}
        onOpenChange={setDnsProviderDialogOpen}
        onRefresh={loadDnsProviders}
      />
    </div>
  )
}
