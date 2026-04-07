import { useState, useEffect, useCallback } from 'react'
import { 
  Plus, Pencil, Trash2, FileText, Globe, FolderOpen, ArrowRightCircle,
  ChevronRight, ChevronLeft, Server, Shield, Zap, Settings, Check,
  Layers, ArrowRight, Lock, FileCode, Loader2, RefreshCw, Code,
  Network, ChevronDown, Wifi
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import { Textarea } from '@/components/ui/textarea'
import { Switch } from '@/components/ui/switch'
import { Card, CardContent } from '@/components/ui/card'
import { Progress } from '@/components/ui/progress'
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
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover'
import { useToast } from '@/hooks/use-toast'
import { cn } from '@/lib/utils'
import { sitesApi, Site, LocationConfig, SiteUpstreamServer as UpstreamServer } from '@/api/sites'
import { certificatesApi, Certificate } from '@/api/certificates'
import { networkApi, NetworkInterface, DNSProviderInfo } from '@/api/network'
import { dnsProvidersApi, DnsProvider } from '@/api/dnsProviders'

// 步骤配置
const STEPS = [
  { id: 1, title: '基本信息', icon: Globe, description: '域名与端口' },
  { id: 2, title: '站点类型', icon: Layers, description: '选择配置模式' },
  { id: 3, title: '代理配置', icon: ArrowRight, description: '反向代理设置' },
  { id: 4, title: 'SSL证书', icon: Lock, description: 'HTTPS配置' },
  { id: 5, title: '高级设置', icon: Settings, description: '缓存与压缩' },
  { id: 6, title: '预览确认', icon: FileCode, description: '配置预览' },
]

export default function Sites() {
  const [sites, setSites] = useState<Site[]>([])
  const [certs, setCerts] = useState<Certificate[]>([])
  const [loading, setLoading] = useState(true)
  const [syncLoading, setSyncLoading] = useState(false)
  const { toast } = useToast()
  
  const [editDialogOpen, setEditDialogOpen] = useState(false)
  const [wizardOpen, setWizardOpen] = useState(false)
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)

  const [editingSite, setEditingSite] = useState<Site | null>(null)
  const [editContent, setEditContent] = useState('')

  const [deleteTarget, setDeleteTarget] = useState<Site | null>(null)

  const [currentStep, setCurrentStep] = useState(1)
  const [wizardData, setWizardData] = useState<Partial<Site>>({
    port: 80,
    siteType: 'proxy',
    upstreamServers: [{ address: '', weight: 1, backup: false }],
    locations: [{ path: '/', proxyPass: '', proxyHeaders: true, websocket: false }],
    forceHttps: true,
    gzip: true,
    cache: false,
  })
  const [isEditMode, setIsEditMode] = useState(false) // 是否为编辑模式

  // 网络相关状态
  const [serverIP, setServerIP] = useState('')
  const [networkInterfaces, setNetworkInterfaces] = useState<NetworkInterface[]>([])
  const [publicIP, setPublicIP] = useState('')
  const [dnsProviders, setDnsProviders] = useState<DnsProvider[]>([])
  const [selectedDnsProvider, setSelectedDnsProvider] = useState<number | null>(null)
  const [ipPopoverOpen, setIpPopoverOpen] = useState(false)
  const [loadingNetwork, setLoadingNetwork] = useState(false)
  const [creatingDNS, setCreatingDNS] = useState(false)
  const [dnsCreated, setDnsCreated] = useState(false)
  
  // 高级功能开关（默认关闭）
  const [showAdvancedDNS, setShowAdvancedDNS] = useState(false)

  // 加载数据
  const loadData = useCallback(async () => {
    try {
      setLoading(true)
      const [sitesRes, certsRes] = await Promise.all([
        sitesApi.list(),
        certificatesApi.list()
      ])
      setSites(sitesRes.data)
      setCerts(certsRes.data)
    } catch (error) {
      toast({ title: '数据加载失败', variant: 'destructive' })
    } finally {
      setLoading(false)
    }
  }, [toast])

  useEffect(() => {
    loadData()
  }, [loadData])

  // 加载网络信息
  const loadNetworkInfo = async () => {
    setLoadingNetwork(true)
    try {
      const [networkRes, dnsRes] = await Promise.all([
        networkApi.getInfo(),
        dnsProvidersApi.list()
      ])
      setNetworkInterfaces(networkRes.data.interfaces || [])
      setPublicIP(networkRes.data.publicIp || '')
      setDnsProviders(dnsRes.data || [])
      
      // 设置默认DNS供应商
      if (networkRes.data.defaultDns) {
        setSelectedDnsProvider(networkRes.data.defaultDns.id)
      } else if (dnsRes.data && dnsRes.data.length > 0) {
        const defaultProvider = dnsRes.data.find((p: DnsProvider) => p.isDefault)
        if (defaultProvider) {
          setSelectedDnsProvider(defaultProvider.id)
        }
      }
      
      // 默认填充公网IP
      if (networkRes.data.publicIp && !serverIP) {
        setServerIP(networkRes.data.publicIp)
      }
    } catch (error) {
      console.error('Failed to load network info:', error)
    } finally {
      setLoadingNetwork(false)
    }
  }

  // 在向导打开时加载网络信息（新建和编辑模式都需要）
  useEffect(() => {
    if (wizardOpen) {
      loadNetworkInfo()
    }
  }, [wizardOpen])

  const getCertById = (id: number | null) => certs.find(c => c.id === id)

  const openEditDialog = async (site: Site) => {
    setEditingSite(site)
    try {
      const res = await sitesApi.getConfig(site.id)
      setEditContent(res.data)
    } catch {
      setEditContent(site.config || '')
    }
    setEditDialogOpen(true)
  }

  // 打开向导编辑模式
  const openWizardEdit = (site: Site) => {
    setEditingSite(site)
    setIsEditMode(true)
    setCurrentStep(1)

    // 解析 locations 和 upstreamServers（可能是 JSON 字符串）
    let locations = site.locations
    let upstreamServers = site.upstreamServers

    if (typeof locations === 'string') {
      try {
        locations = JSON.parse(locations)
      } catch {
        locations = []
      }
    }
    if (typeof upstreamServers === 'string') {
      try {
        upstreamServers = JSON.parse(upstreamServers)
      } catch {
        upstreamServers = []
      }
    }

    // 预填充现有数据
    setWizardData({
      id: site.id,
      domain: site.domain,
      port: site.port || 80,
      siteType: site.siteType || 'proxy',
      rootDir: site.rootDir || '',
      locations: (locations as LocationConfig[])?.length > 0
        ? locations as LocationConfig[]
        : [{ path: '/', proxyPass: '', proxyHeaders: true, websocket: false }],
      upstreamServers: (upstreamServers as UpstreamServer[])?.length > 0
        ? upstreamServers as UpstreamServer[]
        : [{ address: '', weight: 1, backup: false }],
      certId: site.certId || null,
      forceHttps: site.forceHttps || false,
      gzip: site.gzip || false,
      cache: site.cache || false,
    })
    setWizardOpen(true)
  }

  const handleSaveEdit = async () => {
    if (!editingSite || !editContent.trim()) {
      toast({ title: '配置内容不能为空', variant: 'destructive' })
      return
    }
    
    try {
      await sitesApi.update(editingSite.id, { config: editContent })
      toast({ title: '配置已保存' })
      setEditDialogOpen(false)
      loadData()
    } catch (error) {
      toast({ title: '保存失败', variant: 'destructive' })
    }
  }

  const resetWizard = () => {
    setCurrentStep(1)
    setIsEditMode(false)
    setEditingSite(null)
    setWizardData({
      port: 80,
      siteType: 'proxy',
      upstreamServers: [{ address: '', weight: 1, backup: false }],
      locations: [{ path: '/', proxyPass: '', proxyHeaders: true, websocket: false }],
      forceHttps: true,
      gzip: true,
      cache: false,
    })
    // 重置网络相关状态
    setServerIP('')
    setSelectedDnsProvider(null)
    setDnsCreated(false)
    setShowAdvancedDNS(false)
  }

  const handleCreate = async () => {
    if (!wizardData.domain?.trim()) {
      toast({ title: '域名不能为空', variant: 'destructive' })
      return
    }

    const fileName = `${wizardData.domain}.conf`
    const payload = {
      fileName,
      domain: wizardData.domain,
      port: wizardData.port || 80,
      siteType: wizardData.siteType || 'proxy',
      rootDir: wizardData.rootDir || '',
      locations: JSON.stringify(wizardData.locations || []),
      upstreamServers: JSON.stringify(wizardData.upstreamServers || []),
      sslEnabled: !!wizardData.certId,
      certId: wizardData.certId || null,
      forceHttps: wizardData.forceHttps || false,
      gzip: wizardData.gzip || false,
      cache: wizardData.cache || false,
    }

    try {
      if (isEditMode && wizardData.id) {
        // 编辑模式：更新现有站点
        await sitesApi.update(wizardData.id, payload)
        toast({ title: '站点更新成功' })
      } else {
        // 新建模式
        await sitesApi.create(payload)
        toast({ title: '站点创建成功' })
      }
      setWizardOpen(false)
      resetWizard()
      loadData()
    } catch (error: any) {
      const message = error.userMessage || error.message || (isEditMode ? '更新失败' : '创建失败')
      toast({ title: message, variant: 'destructive' })
    }
  }

  const confirmDelete = (site: Site) => {
    setDeleteTarget(site)
    setDeleteDialogOpen(true)
  }

  const handleDelete = async () => {
    if (!deleteTarget) return
    
    try {
      await sitesApi.delete(deleteTarget.id)
      toast({ title: '站点已删除' })
      loadData()
    } catch (error) {
      toast({ title: '删除失败', variant: 'destructive' })
    }
    setDeleteDialogOpen(false)
    setDeleteTarget(null)
  }

  const toggleStatus = async (site: Site, enabled: boolean) => {
    try {
      await sitesApi.toggleEnabled(site.id, enabled)
      toast({ title: `站点 ${site.domain} 已${enabled ? '启用' : '禁用'}` })
      loadData()
    } catch (error) {
      toast({ title: '操作失败', variant: 'destructive' })
    }
  }

  const handleSync = async () => {
    try {
      setSyncLoading(true)
      await sitesApi.sync()
      toast({ title: '配置已同步到 Nginx' })
    } catch (error) {
      toast({ title: '同步失败', variant: 'destructive' })
    } finally {
      setSyncLoading(false)
    }
  }

  // 创建DNS解析记录
  const handleCreateDNSRecord = async () => {
    if (!wizardData.domain?.trim()) {
      toast({ title: '请先输入域名', variant: 'destructive' })
      return
    }
    if (!serverIP.trim()) {
      toast({ title: '请输入服务器IP', variant: 'destructive' })
      return
    }
    if (!selectedDnsProvider) {
      toast({ title: '请选择DNS供应商', variant: 'destructive' })
      return
    }

    setCreatingDNS(true)
    try {
      await networkApi.createDNSRecord({
        domain: wizardData.domain,
        ip: serverIP,
        recordType: 'A',
        dnsProviderId: selectedDnsProvider
      })
      toast({ title: 'DNS解析记录创建成功' })
      setDnsCreated(true)
    } catch (error: any) {
      toast({ title: error.userMessage || 'DNS解析记录创建失败', variant: 'destructive' })
    } finally {
      setCreatingDNS(false)
    }
  }

  // 获取所有可用IP
  const getAllAvailableIPs = () => {
    const ips: { ip: string; source: string; icon: React.ReactNode }[] = []
    
    if (publicIP) {
      ips.push({ ip: publicIP, source: '公网IP', icon: <Globe className="h-4 w-4" /> })
    }
    
    networkInterfaces.forEach(iface => {
      iface.ips.forEach(ip => {
        ips.push({ 
          ip, 
          source: `${iface.name}${iface.mac ? ` (${iface.mac.slice(0, 8)}...)` : ''}`,
          icon: <Wifi className="h-4 w-4" />
        })
      })
    })
    
    return ips
  }

  const canProceed = () => {
    switch (currentStep) {
      case 1:
        return wizardData.domain?.trim()
      case 2:
        return true
      case 3:
        if (wizardData.siteType === 'static') {
          return wizardData.rootDir?.trim()
        }
        if (wizardData.siteType === 'proxy') {
          return wizardData.locations?.some(l => l.proxyPass.trim())
        }
        if (wizardData.siteType === 'loadbalance') {
          return wizardData.upstreamServers?.some(s => s.address.trim())
        }
        return true
      default:
        return true
    }
  }

  const certOptions = certs.filter(c => c.status === 'valid')

  const getSiteTypeLabel = (type: string) => {
    switch (type) {
      case 'static': return '静态站点'
      case 'proxy': return '反向代理'
      case 'loadbalance': return '负载均衡'
      default: return type
    }
  }

  const getSiteTypeBadge = (type: string) => {
    const config: Record<string, { variant: 'default' | 'secondary' | 'outline', className: string }> = {
      static: { variant: 'secondary', className: '' },
      proxy: { variant: 'default', className: '' },
      loadbalance: { variant: 'outline', className: '' },
    }
    return config[type] || { variant: 'secondary', className: '' }
  }

  // 获取显示目标
  const getDisplayTarget = (site: Site) => {
    if (site.siteType === 'static' && site.rootDir) {
      return { type: 'dir', value: site.rootDir }
    }
    if (site.siteType === 'proxy' && site.locations?.length > 0) {
      return { type: 'proxy', value: site.locations[0].proxyPass }
    }
    if (site.siteType === 'loadbalance') {
      return { type: 'lb', value: `${site.upstreamServers?.length || 0} 台服务器` }
    }
    return null
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-semibold">站点配置</h1>
        <div className="flex items-center gap-3">
          <Button variant="outline" onClick={handleSync} disabled={syncLoading}>
            <RefreshCw className={cn("h-4 w-4 mr-2", syncLoading && "animate-spin")} />
            同步配置
          </Button>
          <Button onClick={() => { resetWizard(); setIsEditMode(false); setWizardOpen(true); }}>
            <Plus className="h-4 w-4 mr-2" />
            新增站点
          </Button>
        </div>
      </div>

      {/* Sites List */}
      {loading ? (
        <div className="flex items-center justify-center py-12">
          <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
        </div>
      ) : sites.length === 0 ? (
        <Card>
          <CardContent className="py-12 text-center text-muted-foreground">
            <FileText className="h-12 w-12 mx-auto mb-4 opacity-50" />
            <p>暂无站点配置</p>
            <p className="text-sm mt-1">点击"新增站点"创建第一个配置</p>
          </CardContent>
        </Card>
      ) : (
        <div className="grid gap-4">
          {sites.map((site) => {
            const cert = getCertById(site.certId || null)
            const target = getDisplayTarget(site)
            const typeBadge = getSiteTypeBadge(site.siteType)
            
            return (
              <Card key={site.id}>
                <CardContent className="p-6">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-6">
                      <div className="flex items-center gap-2">
                        <Globe className="h-4 w-4 text-primary" />
                        <span className="font-semibold">{site.domain}</span>
                      </div>
                      <Badge variant={typeBadge.variant} className={typeBadge.className}>
                        {getSiteTypeLabel(site.siteType)}
                      </Badge>
                      <div className="flex items-center gap-2 text-sm text-muted-foreground">
                        {target?.type === 'dir' && (
                          <>
                            <FolderOpen className="h-4 w-4" />
                            <span className="truncate max-w-[200px]">{target.value}</span>
                          </>
                        )}
                        {target?.type === 'proxy' && (
                          <>
                            <ArrowRightCircle className="h-4 w-4" />
                            <span className="truncate max-w-[200px]">{target.value}</span>
                          </>
                        )}
                        {target?.type === 'lb' && (
                          <>
                            <Server className="h-4 w-4" />
                            <span>{target.value}</span>
                          </>
                        )}
                        {!target && <span>-</span>}
                      </div>
                      {cert ? (
                        <div className="flex items-center gap-1.5 text-sm text-emerald-600 font-medium">
                          <Lock className="h-4 w-4" />
                          <span>{cert.issuer}</span>
                        </div>
                      ) : (
                        <span className="text-sm text-muted-foreground">HTTP</span>
                      )}
                    </div>
                    <div className="flex items-center gap-4">
                      <div className="flex items-center gap-2">
                        <Switch
                          checked={site.enabled}
                          onCheckedChange={(checked) => toggleStatus(site, checked)}
                        />
                        <span className="text-sm text-muted-foreground">{site.enabled ? '启用' : '禁用'}</span>
                      </div>
                      <Button variant="ghost" size="sm" onClick={() => openWizardEdit(site)}>
                        <Pencil className="h-4 w-4 mr-1" />编辑
                      </Button>
                      <Button variant="ghost" size="sm" onClick={() => openEditDialog(site)}>
                        <Code className="h-4 w-4 mr-1" />配置
                      </Button>
                      <Button variant="ghost" size="sm" className="text-destructive" onClick={() => confirmDelete(site)}>
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

      {/* Wizard Dialog - Apple Style */}
      <Dialog open={wizardOpen} onOpenChange={(open) => {
        setWizardOpen(open)
        if (!open) resetWizard()
      }}>
        <DialogContent className="w-[680px] max-w-none p-0 gap-0 bg-background border-border/50 shadow-2xl rounded-2xl overflow-hidden">
          {/* Content Area */}
          <div className="pt-8 pb-6 px-10">
            {/* Header - Compact */}
            <div className="mb-6">
              <h2 className="text-xl font-semibold text-foreground mb-1">
                {isEditMode ? '编辑站点' : '创建站点'}
              </h2>
              <p className="text-sm text-muted-foreground">
                {STEPS[currentStep - 1].title} · {STEPS[currentStep - 1].description}
              </p>
            </div>

            {/* Progress Bar */}
            <div className="mb-8">
              <div className="flex items-center justify-between mb-2">
                <span className="text-xs text-muted-foreground">步骤 {currentStep} / {STEPS.length}</span>
                <span className="text-xs text-muted-foreground">{Math.round((currentStep / STEPS.length) * 100)}%</span>
              </div>
              <Progress value={(currentStep / STEPS.length) * 100} className="h-1.5" />
            </div>

            {/* Form Content */}
            <div className="min-h-[320px]">
              {/* Step 1: Basic Info */}
              {currentStep === 1 && (
                <div className="space-y-5">
                  <div className="space-y-2.5">
                    <label className="text-sm font-medium text-foreground">域名</label>
                    <Input
                      placeholder="api.example.com"
                      value={wizardData.domain || ''}
                      onChange={(e) => {
                        setWizardData({ ...wizardData, domain: e.target.value })
                        setDnsCreated(false)
                      }}
                      className="h-11 px-4 rounded-xl"
                    />
                    <p className="text-xs text-muted-foreground">支持泛域名，如 *.example.com</p>
                  </div>
                  
                  <div className="space-y-2.5">
                    <label className="text-sm font-medium text-foreground">监听端口</label>
                    <Input
                      type="number"
                      value={wizardData.port || 80}
                      onChange={(e) => setWizardData({ ...wizardData, port: parseInt(e.target.value) || 80 })}
                      className="h-11 px-4 rounded-xl"
                    />
                  </div>

                  {/* 高级功能开关 */}
                  <div className="pt-2">
                    <button
                      type="button"
                      onClick={() => setShowAdvancedDNS(!showAdvancedDNS)}
                      className="flex items-center gap-2 text-sm text-muted-foreground hover:text-foreground transition-colors"
                    >
                      <ChevronRight className={cn(
                        "h-4 w-4 transition-transform duration-200",
                        showAdvancedDNS && "rotate-90"
                      )} />
                      <span>DNS解析配置</span>
                      <Badge variant="outline" className="text-xs">可选</Badge>
                    </button>
                  </div>

                  {/* 可选的DNS解析功能 */}
                  {showAdvancedDNS && (
                    <div className="pl-6 space-y-4 border-l-2 border-muted ml-1.5">
                      <div className="space-y-2.5">
                        <label className="text-sm font-medium text-foreground flex items-center gap-2">
                          <Network className="h-4 w-4" />
                          服务器IP
                          <span className="text-xs text-muted-foreground font-normal">（自动解析目标）</span>
                        </label>
                        <Popover open={ipPopoverOpen} onOpenChange={setIpPopoverOpen}>
                          <PopoverTrigger asChild>
                            <div className="relative">
                              <Input
                                placeholder="输入或选择IP地址"
                                value={serverIP}
                                onChange={(e) => setServerIP(e.target.value)}
                                className="h-11 px-4 rounded-xl pr-10"
                              />
                              <button 
                                type="button"
                                className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
                                onClick={() => setIpPopoverOpen(!ipPopoverOpen)}
                              >
                                <ChevronDown className="h-4 w-4" />
                              </button>
                            </div>
                          </PopoverTrigger>
                          <PopoverContent className="w-72 p-2" align="start">
                            <div className="space-y-1">
                              <div className="text-xs text-muted-foreground px-2 py-1">选择IP地址</div>
                              {loadingNetwork ? (
                                <div className="flex items-center justify-center py-4">
                                  <Loader2 className="h-4 w-4 animate-spin" />
                                </div>
                              ) : (
                                getAllAvailableIPs().map((item, idx) => (
                                  <button
                                    type="button"
                                    key={idx}
                                    className={cn(
                                      "w-full flex items-center gap-3 px-3 py-2 rounded-lg text-left hover:bg-muted transition-colors",
                                      serverIP === item.ip && "bg-muted"
                                    )}
                                    onClick={() => {
                                      setServerIP(item.ip)
                                      setIpPopoverOpen(false)
                                    }}
                                  >
                                    {item.icon}
                                    <div className="flex-1 min-w-0">
                                      <div className="text-sm font-medium">{item.ip}</div>
                                      <div className="text-xs text-muted-foreground truncate">{item.source}</div>
                                    </div>
                                    {serverIP === item.ip && <Check className="h-4 w-4 text-primary" />}
                                  </button>
                                ))
                              )}
                            </div>
                          </PopoverContent>
                        </Popover>
                      </div>
                      
                      <div className="space-y-2.5">
                        <label className="text-sm font-medium text-foreground flex items-center gap-2">
                          <Globe className="h-4 w-4" />
                          DNS供应商
                          <span className="text-xs text-muted-foreground font-normal">（解析服务提供商）</span>
                        </label>
                        <Select 
                          value={selectedDnsProvider?.toString() || ''} 
                          onValueChange={(v) => setSelectedDnsProvider(v ? parseInt(v) : null)}
                        >
                          <SelectTrigger className="h-11 px-4 rounded-xl">
                            <SelectValue placeholder="选择DNS供应商" />
                          </SelectTrigger>
                          <SelectContent className="rounded-xl">
                            {dnsProviders.length === 0 ? (
                              <div className="px-2 py-4 text-center text-sm text-muted-foreground">
                                暂无DNS供应商配置
                              </div>
                            ) : (
                              dnsProviders.map((provider) => (
                                <SelectItem key={provider.id} value={provider.id.toString()}>
                                  {provider.name} ({provider.providerType})
                                </SelectItem>
                              ))
                            )}
                          </SelectContent>
                        </Select>
                      </div>
                      
                      {/* DNS解析操作区 */}
                      {wizardData.domain && showAdvancedDNS && (
                        selectedDnsProvider && serverIP ? (
                          <div className="flex items-center justify-between p-3 bg-muted/30 rounded-lg border">
                            <div className="flex-1 min-w-0">
                              <div className="text-sm font-medium text-foreground flex items-center gap-2">
                                {dnsCreated ? (
                                  <>
                                    <Check className="h-4 w-4 text-emerald-500" />
                                    DNS解析已创建
                                  </>
                                ) : (
                                  '自动配置DNS解析'
                                )}
                              </div>
                              <div className="text-xs text-muted-foreground truncate mt-0.5">
                                {wizardData.domain} → {serverIP}
                              </div>
                            </div>
                            {!dnsCreated && (
                              <Button
                                size="sm"
                                variant="secondary"
                                onClick={handleCreateDNSRecord}
                                disabled={creatingDNS || !wizardData.domain}
                                className="rounded-lg shrink-0 ml-3"
                              >
                                {creatingDNS ? (
                                  <>
                                    <Loader2 className="h-3.5 w-3.5 mr-1.5 animate-spin" />
                                    创建中
                                  </>
                                ) : '创建解析'}
                              </Button>
                            )}
                          </div>
                        ) : (selectedDnsProvider || serverIP) ? (
                          <div className="flex items-center gap-2 p-3 bg-amber-50 dark:bg-amber-950/20 rounded-lg border border-amber-200 dark:border-amber-800">
                            <span className="text-xs text-amber-600 dark:text-amber-400">
                              请同时填写服务器IP和选择DNS供应商以启用自动DNS解析
                            </span>
                          </div>
                        ) : null
                      )}
                    </div>
                  )}
                </div>
              )}

              {/* Step 2: Site Type */}
              {currentStep === 2 && (
                <div className="grid grid-cols-3 gap-3">
                  {[
                    { type: 'static', icon: FolderOpen, title: '静态站点', desc: '托管静态文件' },
                    { type: 'proxy', icon: ArrowRightCircle, title: '反向代理', desc: '转发请求到后端' },
                    { type: 'loadbalance', icon: Server, title: '负载均衡', desc: '多服务器分发' },
                  ].map((item) => {
                    const Icon = item.icon
                    const selected = wizardData.siteType === item.type
                    return (
                      <button
                        key={item.type}
                        onClick={() => setWizardData({ ...wizardData, siteType: item.type as Site['siteType'] })}
                        className={cn(
                          "p-5 rounded-xl text-center transition-all duration-200",
                          "border hover:shadow-md",
                          selected
                            ? "bg-primary text-primary-foreground border-primary shadow-lg"
                            : "bg-card border-border hover:border-primary/50"
                        )}
                      >
                        <Icon className={cn("h-6 w-6 mx-auto mb-3", selected ? "text-primary-foreground" : "text-muted-foreground")} />
                        <div className={cn("text-sm font-medium mb-1", selected ? "text-primary-foreground" : "text-foreground")}>
                          {item.title}
                        </div>
                        <div className={cn("text-xs", selected ? "text-primary-foreground/70" : "text-muted-foreground")}>
                          {item.desc}
                        </div>
                      </button>
                    )
                  })}
                </div>
              )}

              {/* Step 3: Proxy Config */}
              {currentStep === 3 && (
                <div className="space-y-5">
                  {wizardData.siteType === 'static' && (
                    <>
                      <div className="space-y-2.5">
                        <label className="text-sm font-medium text-foreground">网站根目录</label>
                        <Input
                          placeholder="/var/www/html"
                          value={wizardData.rootDir || ''}
                          onChange={(e) => setWizardData({ ...wizardData, rootDir: e.target.value })}
                          className="h-11 px-4 rounded-xl"
                        />
                      </div>
                      <div className="space-y-2.5">
                        <label className="text-sm font-medium text-foreground">默认首页</label>
                        <Input placeholder="index.html" defaultValue="index.html" className="h-11 px-4 rounded-xl" />
                      </div>
                    </>
                  )}

                  {wizardData.siteType === 'proxy' && (
                    <div className="space-y-5">
                      {(wizardData.locations || []).map((loc, idx) => (
                        <div key={idx} className="p-5 bg-muted/50 rounded-xl border space-y-4">
                          <div className="grid grid-cols-2 gap-4">
                            <div className="space-y-2">
                              <label className="text-xs font-medium text-muted-foreground">路径</label>
                              <Input
                                placeholder="/"
                                value={loc.path}
                                onChange={(e) => {
                                  const newLocs = [...(wizardData.locations || [])]
                                  newLocs[idx].path = e.target.value
                                  setWizardData({ ...wizardData, locations: newLocs })
                                }}
                                className="h-11 px-3 rounded-lg text-sm"
                              />
                            </div>
                            <div className="space-y-2">
                              <label className="text-xs font-medium text-muted-foreground">代理目标</label>
                              <Input
                                placeholder="http://127.0.0.1:8080"
                                value={loc.proxyPass}
                                onChange={(e) => {
                                  const newLocs = [...(wizardData.locations || [])]
                                  newLocs[idx].proxyPass = e.target.value
                                  setWizardData({ ...wizardData, locations: newLocs })
                                }}
                                className="h-11 px-3 rounded-lg text-sm"
                              />
                            </div>
                          </div>
                          <div className="flex items-center gap-5">
                            <label className="flex items-center gap-2 text-xs text-muted-foreground cursor-pointer">
                              <input type="checkbox" checked={loc.proxyHeaders} onChange={(e) => {
                                const newLocs = [...(wizardData.locations || [])]
                                newLocs[idx].proxyHeaders = e.target.checked
                                setWizardData({ ...wizardData, locations: newLocs })
                              }} className="w-4 h-4 rounded" />
                              传递客户端信息
                            </label>
                            <label className="flex items-center gap-2 text-xs text-muted-foreground cursor-pointer">
                              <input type="checkbox" checked={loc.websocket} onChange={(e) => {
                                const newLocs = [...(wizardData.locations || [])]
                                newLocs[idx].websocket = e.target.checked
                                setWizardData({ ...wizardData, locations: newLocs })
                              }} className="w-4 h-4 rounded" />
                              WebSocket
                            </label>
                          </div>
                        </div>
                      ))}
                      <Button variant="ghost" size="sm" onClick={() => setWizardData({ ...wizardData, locations: [...(wizardData.locations || []), { path: '/', proxyPass: '', proxyHeaders: true, websocket: false }] })} className="text-muted-foreground">
                        <Plus className="h-4 w-4 mr-1" /> 添加路径
                      </Button>
                    </div>
                  )}

                  {wizardData.siteType === 'loadbalance' && (
                    <div className="space-y-5">
                      {(wizardData.upstreamServers || []).map((server, idx) => (
                        <div key={idx} className="p-5 bg-muted/50 rounded-xl border">
                          <div className="grid grid-cols-3 gap-4">
                            <div className="col-span-2 space-y-2">
                              <label className="text-xs font-medium text-muted-foreground">地址</label>
                              <Input placeholder="192.168.1.10:8080" value={server.address} onChange={(e) => {
                                const newServers = [...(wizardData.upstreamServers || [])]
                                newServers[idx].address = e.target.value
                                setWizardData({ ...wizardData, upstreamServers: newServers })
                              }} className="h-11 px-3 rounded-lg text-sm" />
                            </div>
                            <div className="space-y-2">
                              <label className="text-xs font-medium text-muted-foreground">权重</label>
                              <Input type="number" min="1" value={server.weight} onChange={(e) => {
                                const newServers = [...(wizardData.upstreamServers || [])]
                                newServers[idx].weight = parseInt(e.target.value) || 1
                                setWizardData({ ...wizardData, upstreamServers: newServers })
                              }} className="h-11 px-3 rounded-lg text-sm" />
                            </div>
                          </div>
                        </div>
                      ))}
                      <Button variant="ghost" size="sm" onClick={() => setWizardData({ ...wizardData, upstreamServers: [...(wizardData.upstreamServers || []), { address: '', weight: 1, backup: false }] })} className="text-muted-foreground">
                        <Plus className="h-4 w-4 mr-1" /> 添加服务器
                      </Button>
                    </div>
                  )}
                </div>
              )}

              {/* Step 4: SSL */}
              {currentStep === 4 && (
                <div className="space-y-6">
                  <div className="space-y-2.5">
                    <label className="text-sm font-medium text-foreground">SSL 证书</label>
                    <Select value={wizardData.certId?.toString() || ''} onValueChange={(v) => setWizardData({ ...wizardData, certId: v ? parseInt(v) : null })}>
                      <SelectTrigger className="h-11 px-4 rounded-xl">
                        <SelectValue placeholder="不使用 HTTPS" />
                      </SelectTrigger>
                      <SelectContent className="rounded-xl">
                        {certOptions.map((cert) => (
                          <SelectItem key={cert.id} value={cert.id.toString()}>{cert.domain} ({cert.issuer})</SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>
                  {wizardData.certId && (
                    <label className="flex items-center gap-4 p-4 bg-muted/50 rounded-xl border cursor-pointer">
                      <input type="checkbox" checked={wizardData.forceHttps} onChange={(e) => setWizardData({ ...wizardData, forceHttps: e.target.checked })} className="w-5 h-5 rounded" />
                      <div>
                        <div className="text-sm font-medium text-foreground">强制 HTTPS</div>
                        <div className="text-xs text-muted-foreground">自动将 HTTP 重定向到 HTTPS</div>
                      </div>
                    </label>
                  )}
                </div>
              )}

              {/* Step 5: Advanced */}
              {currentStep === 5 && (
                <div className="space-y-4">
                  <label className="flex items-center gap-4 p-5 bg-muted/50 rounded-xl border cursor-pointer hover:bg-muted transition-colors">
                    <Zap className="h-5 w-5 text-muted-foreground" />
                    <div className="flex-1">
                      <div className="text-sm font-medium text-foreground">Gzip 压缩</div>
                      <div className="text-xs text-muted-foreground">压缩文本资源</div>
                    </div>
                    <input type="checkbox" checked={wizardData.gzip} onChange={(e) => setWizardData({ ...wizardData, gzip: e.target.checked })} className="w-5 h-5 rounded" />
                  </label>
                  <label className="flex items-center gap-4 p-5 bg-muted/50 rounded-xl border cursor-pointer hover:bg-muted transition-colors">
                    <Shield className="h-5 w-5 text-muted-foreground" />
                    <div className="flex-1">
                      <div className="text-sm font-medium text-foreground">静态缓存</div>
                      <div className="text-xs text-muted-foreground">浏览器缓存策略</div>
                    </div>
                    <input type="checkbox" checked={wizardData.cache} onChange={(e) => setWizardData({ ...wizardData, cache: e.target.checked })} className="w-5 h-5 rounded" />
                  </label>
                </div>
              )}

              {/* Step 6: Preview */}
              {currentStep === 6 && (
                <div className="space-y-6">
                  <div className="bg-muted rounded-xl p-5 text-muted-foreground font-mono text-sm leading-relaxed">
                    <div className="text-muted-foreground/50"># {wizardData.domain}</div>
                    <div className="mt-2">类型: {getSiteTypeLabel(wizardData.siteType || 'proxy')}</div>
                    <div>端口: {wizardData.port || 80}</div>
                    <div>SSL: {wizardData.certId ? '已配置' : '未配置'}</div>
                    {wizardData.gzip && <div className="text-emerald-500">✓ Gzip</div>}
                    {wizardData.cache && <div className="text-emerald-500">✓ 缓存</div>}
                  </div>
                  <div className="flex items-center justify-center gap-4 text-sm text-muted-foreground">
                    <span className="px-3 py-1 bg-muted rounded-full">{wizardData.domain}</span>
                    {wizardData.certId && <Badge variant="secondary" className="text-emerald-500">HTTPS</Badge>}
                  </div>
                </div>
              )}
            </div>
          </div>

          {/* Footer */}
          <div className="px-10 py-5 border-t flex items-center justify-between bg-muted/30">
            <Button
              variant="ghost"
              onClick={() => currentStep > 1 ? setCurrentStep(currentStep - 1) : setWizardOpen(false)}
              className="text-muted-foreground"
            >
              {currentStep > 1 ? '上一步' : '取消'}
            </Button>
            <Button
              onClick={() => {
                if (currentStep < STEPS.length) {
                  setCurrentStep(currentStep + 1)
                } else {
                  handleCreate()
                }
              }}
              disabled={!canProceed()}
              className="rounded-xl h-11 px-8"
            >
              {currentStep < STEPS.length ? '继续' : (isEditMode ? '保存' : '创建')}
            </Button>
          </div>
        </DialogContent>
      </Dialog>

      {/* Edit Dialog */}
      <Dialog open={editDialogOpen} onOpenChange={setEditDialogOpen}>
        <DialogContent className="max-w-3xl">
          <DialogHeader>
            <DialogTitle>编辑配置 - {editingSite?.fileName}</DialogTitle>
            <DialogDescription>直接修改 Nginx 配置文件内容</DialogDescription>
          </DialogHeader>
          <div className="py-4">
            <Label>配置内容</Label>
            <Textarea
              placeholder="Nginx 配置内容"
              value={editContent}
              onChange={(e) => setEditContent(e.target.value)}
              rows={20}
              className="font-mono text-sm mt-2 bg-muted text-muted-foreground"
            />
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setEditDialogOpen(false)}>取消</Button>
            <Button onClick={handleSaveEdit}>保存</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Delete Dialog */}
      <AlertDialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>确认删除</AlertDialogTitle>
            <AlertDialogDescription>
              确定要删除站点 <strong>{deleteTarget?.domain}</strong> 吗？此操作不可恢复。
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>取消</AlertDialogCancel>
            <AlertDialogAction onClick={handleDelete} className="bg-destructive text-destructive-foreground hover:bg-destructive/90">
              删除
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}
