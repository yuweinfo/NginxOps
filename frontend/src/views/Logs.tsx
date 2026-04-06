import { useState, useEffect, useRef } from 'react'
import { Search, Download, Trash2, ChevronDown, ChevronRight, Loader2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { useToast } from '@/hooks/use-toast'
import { statsApi, LogEntry } from '@/api/stats'

const methodColor = (m: string) => {
  const map: Record<string, string> = { GET: '#3b82f6', POST: '#22c55e', PUT: '#f59e0b', DELETE: '#ef4444', PATCH: '#ec4899', HEAD: '#6b7280' }
  return map[m] || '#6b7280'
}

const statusVariant = (s: number) => {
  if (s >= 200 && s < 300) return 'success'
  if (s >= 300 && s < 400) return 'secondary'
  if (s >= 400 && s < 500) return 'warning'
  return 'destructive'
}

export default function Logs() {
  const [logs, setLogs] = useState<LogEntry[]>([])
  const [loading, setLoading] = useState(false)
  const [searchIp, setSearchIp] = useState('')
  const { toast } = useToast()

  const [streamLines, setStreamLines] = useState<string[]>([])
  const [autoScroll, setAutoScroll] = useState(true)
  const [wsConnected, setWsConnected] = useState(false)
  const streamRef = useRef<HTMLPreElement>(null)
  const wsRef = useRef<WebSocket | null>(null)

  useEffect(() => { handleSearch() }, [])

  useEffect(() => {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const wsUrl = `${protocol}//${window.location.host}/ws/logs`
    try {
      const ws = new WebSocket(wsUrl)
      wsRef.current = ws
      ws.onopen = () => setWsConnected(true)
      ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data)
          if (data.type === 'log' && data.line) {
            setStreamLines((prev) => [...prev.slice(-499), data.line])
          } else if (data.type === 'clear') {
            setStreamLines([])
          }
        } catch { setStreamLines((prev) => [...prev.slice(-499), event.data]) }
      }
      ws.onclose = () => setWsConnected(false)
      ws.onerror = () => setWsConnected(false)
    } catch { console.error('Failed to create WebSocket') }
    return () => { wsRef.current?.close() }
  }, [])

  useEffect(() => {
    if (autoScroll && streamRef.current) {
      streamRef.current.scrollTop = streamRef.current.scrollHeight
    }
  }, [streamLines, autoScroll])

  const handleSearch = async () => {
    setLoading(true)
    try {
      const end = new Date().toISOString()
      const start = new Date(Date.now() - 24 * 60 * 60 * 1000).toISOString()
      const res = await statsApi.queryLogs({ start, end, ip: searchIp || undefined, page: 1, size: 50 })
      if (res.success) setLogs(res.data.records || [])
    } catch { toast({ title: '获取日志失败', variant: 'destructive' }) }
    finally { setLoading(false) }
  }

  const handleClearStream = () => {
    setStreamLines([])
    if (wsRef.current?.readyState === WebSocket.OPEN) wsRef.current.send(JSON.stringify({ action: 'clear' }))
    toast({ title: '实时日志已清空' })
  }

  const handleExport = () => {
    const blob = new Blob([streamLines.join('\n')], { type: 'text/plain' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `nginx-logs-${new Date().toISOString().slice(0, 10)}.txt`
    a.click()
    URL.revokeObjectURL(url)
    toast({ title: '日志已下载' })
  }

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-semibold">日志查看</h1>
      
      <Tabs defaultValue="table" className="space-y-4">
        <TabsList>
          <TabsTrigger value="table">日志查询</TabsTrigger>
          <TabsTrigger value="stream">实时日志</TabsTrigger>
        </TabsList>

        <TabsContent value="table">
          <Card>
            <CardHeader className="pb-4">
              <div className="flex gap-3">
                <Input placeholder="筛选 IP 地址" value={searchIp} onChange={(e) => setSearchIp(e.target.value)} className="w-48" />
                <Button onClick={handleSearch}><Search className="h-4 w-4 mr-2" />查询</Button>
                <Button variant="outline" onClick={() => { setSearchIp(''); handleSearch() }}>重置</Button>
              </div>
            </CardHeader>
            <CardContent>
              {loading ? (
                <div className="flex justify-center py-8"><Loader2 className="h-6 w-6 animate-spin" /></div>
              ) : (
                <div className="space-y-2">
                  {logs.map((log) => (
                    <div key={log.id} className="flex items-center gap-4 p-3 bg-muted rounded-lg text-sm">
                      <span className="text-muted-foreground w-40">{log.time}</span>
                      <span className="font-mono w-32">{log.ip}</span>
                      <Badge style={{ color: methodColor(log.method), borderColor: methodColor(log.method) }} variant="outline">{log.method}</Badge>
                      <span className="flex-1 truncate">{log.url}</span>
                      <Badge variant={statusVariant(log.status) as any}>{log.status}</Badge>
                      <span className="w-16 text-right">{log.size > 1024 ? `${(log.size / 1024).toFixed(1)}KB` : `${log.size}B`}</span>
                      <span className="w-16 text-right text-muted-foreground">{log.duration}ms</span>
                    </div>
                  ))}
                  {logs.length === 0 && <div className="text-center py-8 text-muted-foreground">暂无日志数据</div>}
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="stream">
          <Card>
            <CardHeader className="pb-4">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-4">
                  <Badge variant={wsConnected ? 'success' : 'secondary'}>{wsConnected ? `${streamLines.length} 条日志` : '未连接'}</Badge>
                  <Button variant="ghost" size="sm" onClick={() => setAutoScroll(!autoScroll)}>
                    {autoScroll ? <ChevronDown className="h-4 w-4 mr-1" /> : <ChevronRight className="h-4 w-4 mr-1" />}
                    {autoScroll ? '自动滚动' : '已暂停'}
                  </Button>
                </div>
                <div className="flex gap-2">
                  <Button variant="outline" size="sm" onClick={handleClearStream}><Trash2 className="h-4 w-4 mr-1" />清空</Button>
                  <Button variant="outline" size="sm" onClick={handleExport}><Download className="h-4 w-4 mr-1" />导出</Button>
                </div>
              </div>
            </CardHeader>
            <CardContent>
              <pre ref={streamRef} className="bg-muted p-4 rounded-lg text-sm font-mono h-96 overflow-auto">
                {streamLines.length === 0 ? (
                  <span className="text-muted-foreground">{wsConnected ? '等待实时日志流...' : '连接中...'}</span>
                ) : (
                  streamLines.map((line, idx) => {
                    const parts = line.split(' | ')
                    const status = parts[3]
                    const isError = status && parseInt(status) >= 400
                    return <div key={idx} className={isError ? 'text-destructive' : ''}>{line}</div>
                  })
                )}
              </pre>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  )
}
