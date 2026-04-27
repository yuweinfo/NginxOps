import { useState, useEffect, useMemo, useCallback, useRef } from 'react'
import { DateRange } from 'react-day-picker'
import * as echarts from 'echarts'
import ReactECharts from 'echarts-for-react'
import { Card, CardContent } from '@/components/ui/card'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { metricsApi, MetricsOverview, TrendPoint, DistributionItem, ErrorPathItem, ClientAnalysis } from '@/api/metrics'
import { useThemeColors } from '@/hooks/useThemeColor'
import DateRangePicker from '@/components/DateRangePicker'
import { cn } from '@/lib/utils'
import { Loader2, AlertCircle, TrendingUp, Activity, Zap, Clock, ChevronRight, Lightbulb, ArrowUpRight, ArrowDownRight, Minus, Calendar, BarChart3, AlertTriangle, Monitor, Hash } from 'lucide-react'

type Granularity = '1m' | '5m' | '1h' | '1d'
type TabKey = 'traffic' | 'response' | 'error' | 'client' | 'status'

const granularityOptions: { value: Granularity; label: string }[] = [
  { value: '1m', label: '1分钟' },
  { value: '5m', label: '5分钟' },
  { value: '1h', label: '1小时' },
  { value: '1d', label: '1天' },
]

const tabs: { key: TabKey; label: string; icon: React.ElementType }[] = [
  { key: 'traffic', label: '流量分析', icon: BarChart3 },
  { key: 'response', label: '响应性能', icon: Zap },
  { key: 'error', label: '错误分析', icon: AlertTriangle },
  { key: 'client', label: '客户端', icon: Monitor },
  { key: 'status', label: '状态码', icon: Hash },
]

const timePresets: { label: string; range: () => DateRange }[] = [
  { label: '最近1小时', range: () => ({ from: new Date(Date.now() - 60 * 60 * 1000), to: new Date() }) },
  { label: '最近24小时', range: () => ({ from: new Date(Date.now() - 24 * 60 * 60 * 1000), to: new Date() }) },
  { label: '最近7天', range: () => ({ from: new Date(Date.now() - 7 * 24 * 60 * 60 * 1000), to: new Date() }) },
  { label: '今天', range: () => { const now = new Date(); return { from: new Date(now.getFullYear(), now.getMonth(), now.getDate()), to: now } } },
]

function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(1024))
  return `${(bytes / Math.pow(1024, i)).toFixed(2)} ${units[i]}`
}

function formatRT(rt: number): string {
  if (rt === 0) return '0 ms'
  if (rt < 1) return `${(rt * 1000).toFixed(0)} ms`
  return `${rt.toFixed(3)} s`
}

function MetricCardWithSparkline({
  icon,
  label,
  value,
  sparklineData,
  trend,
  onClick,
  active,
}: {
  icon: React.ReactNode
  label: string
  value: string
  sparklineData?: number[]
  trend?: number
  onClick?: () => void
  active?: boolean
}) {
  const colors = useThemeColors()
  const fg = colors.foreground || '#0a0a0a'
  const muted = colors.muted || '#e5e5e5'

  const sparklineOption = useMemo(() => {
    if (!sparklineData || sparklineData.length === 0) return null
    return {
      grid: { left: 0, right: 0, top: 5, bottom: 0 },
      xAxis: { type: 'category' as const, show: false, data: sparklineData.map((_, i) => i) },
      yAxis: { type: 'value' as const, show: false },
      series: [{
        type: 'line' as const,
        data: sparklineData,
        smooth: true,
        showSymbol: false,
        lineStyle: { width: 1.5, color: fg },
        areaStyle: {
          color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
            { offset: 0, color: fg + '20' },
            { offset: 1, color: fg + '05' },
          ]),
        },
      }],
    }
  }, [sparklineData, fg])

  const TrendIcon = trend !== undefined ? (trend > 0 ? ArrowUpRight : trend < 0 ? ArrowDownRight : Minus) : null
  const trendColor = trend !== undefined ? (trend > 0 ? 'text-foreground' : trend < 0 ? 'text-muted-foreground' : 'text-muted-foreground') : ''

  return (
    <Card
      className={cn(
        'cursor-pointer transition-all duration-200 border',
        active ? 'border-foreground/30 shadow-sm' : 'hover:border-foreground/20 hover:shadow-sm'
      )}
      onClick={onClick}
    >
      <CardContent className="p-4">
        <div className="flex items-start justify-between mb-2">
          <div className="flex items-center gap-2">
            <div className="text-muted-foreground">{icon}</div>
            <span className="text-xs font-medium text-muted-foreground">{label}</span>
          </div>
          {TrendIcon && trend !== undefined && (
            <div className={cn('flex items-center gap-0.5 text-xs font-medium', trendColor)}>
              <TrendIcon className="h-3 w-3" />
              <span>{Math.abs(trend).toFixed(1)}%</span>
            </div>
          )}
        </div>
        <p className="text-xl font-bold tracking-tight text-foreground">{value}</p>
        {sparklineOption && (
          <div className="mt-2 h-10">
            <ReactECharts option={sparklineOption} style={{ height: '100%' }} opts={{ renderer: 'canvas' }} />
          </div>
        )}
      </CardContent>
    </Card>
  )
}

function InsightBadge({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex items-start gap-2 p-3 rounded-lg bg-muted/50 border border-border/50">
      <Lightbulb className="h-4 w-4 text-muted-foreground mt-0.5 shrink-0" />
      <span className="text-sm text-muted-foreground">{children}</span>
    </div>
  )
}

function ChartCard({ title, children, className }: { title: string; children: React.ReactNode; className?: string }) {
  return (
    <Card className={cn('transition-all duration-200 hover:shadow-md', className)}>
      <CardContent className="pt-4">
        <h4 className="text-sm font-semibold tracking-wide text-foreground mb-3">{title}</h4>
        {children}
      </CardContent>
    </Card>
  )
}

export default function MetricsAnalysis() {
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [activeTab, setActiveTab] = useState<TabKey>('traffic')
  const [granularity, setGranularity] = useState<Granularity>('5m')
  const [dateRange, setDateRange] = useState<DateRange | undefined>(() => {
    const now = new Date()
    const start = new Date(now.getTime() - 24 * 60 * 60 * 1000)
    return { from: start, to: now }
  })
  const [showPresets, setShowPresets] = useState(false)
  const [overview, setOverview] = useState<MetricsOverview | null>(null)
  const [trafficTrend, setTrafficTrend] = useState<TrendPoint[]>([])
  const [responseTrend, setResponseTrend] = useState<TrendPoint[]>([])
  const [slowRequestTrend, setSlowRequestTrend] = useState<TrendPoint[]>([])
  const [methodDistribution, setMethodDistribution] = useState<DistributionItem[]>([])
  const [statusDistribution, setStatusDistribution] = useState<DistributionItem[]>([])
  const [errorRateTrend, setErrorRateTrend] = useState<TrendPoint[]>([])
  const [errorPaths, setErrorPaths] = useState<ErrorPathItem[]>([])
  const [clientAnalysis, setClientAnalysis] = useState<ClientAnalysis | null>(null)
  const [tabTransition, setTabTransition] = useState<'enter' | 'exit' | 'idle'>('idle')

  const colors = useThemeColors()
  const presetRef = useRef<HTMLDivElement>(null)

  const getApiParams = useCallback(() => {
    return {
      start: dateRange?.from?.toISOString() ?? '',
      end: dateRange?.to?.toISOString() ?? '',
    }
  }, [dateRange])

  const fetchData = useCallback(async () => {
    if (!dateRange?.from) return
    try {
      setLoading(true)
      setError(null)
      const params = getApiParams()

      const [
        overviewRes,
        trafficRes,
        responseRes,
        slowReqRes,
        methodRes,
        statusRes,
        errorRateRes,
        errorPathsRes,
        clientRes,
      ] = await Promise.all([
        metricsApi.getOverview(params),
        metricsApi.getTrafficTrend({ ...params, granularity }),
        metricsApi.getResponseTrend({ ...params, granularity }),
        metricsApi.getSlowRequestTrend({ ...params, granularity }),
        metricsApi.getMethodDistribution(params),
        metricsApi.getStatusDistribution(params),
        metricsApi.getErrorRateTrend({ ...params, granularity }),
        metricsApi.getErrorPaths(params),
        metricsApi.getClientAnalysis(params),
      ])

      if (overviewRes.success) setOverview(overviewRes.data)
      if (trafficRes.success) setTrafficTrend(trafficRes.data)
      if (responseRes.success) setResponseTrend(responseRes.data)
      if (slowReqRes.success) setSlowRequestTrend(slowReqRes.data)
      if (methodRes.success) setMethodDistribution(methodRes.data)
      if (statusRes.success) setStatusDistribution(statusRes.data)
      if (errorRateRes.success) setErrorRateTrend(errorRateRes.data)
      if (errorPathsRes.success) setErrorPaths(errorPathsRes.data)
      if (clientRes.success) setClientAnalysis(clientRes.data)
    } catch (e: any) {
      console.error('Failed to fetch metrics:', e)
      setError(e.message || '获取数据失败')
    } finally {
      setLoading(false)
    }
  }, [dateRange, granularity, getApiParams])

  useEffect(() => {
    fetchData()
  }, [fetchData])

  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (presetRef.current && !presetRef.current.contains(e.target as Node)) {
        setShowPresets(false)
      }
    }
    document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [])

  const handleDateRangeChange = useCallback((range: DateRange) => {
    setDateRange(range)
    setShowPresets(false)
  }, [])

  const handleGranularityChange = useCallback((value: string) => {
    setGranularity(value as Granularity)
  }, [])

  const handleTabChange = useCallback((key: TabKey) => {
    if (key === activeTab) return
    setTabTransition('exit')
    setTimeout(() => {
      setActiveTab(key)
      setTabTransition('enter')
      setTimeout(() => setTabTransition('idle'), 200)
    }, 150)
  }, [activeTab])

  const handlePresetSelect = useCallback((range: DateRange) => {
    setDateRange(range)
    setShowPresets(false)
  }, [])

  const fg = colors.foreground || '#0a0a0a'
  const muted = colors.muted || '#e5e5e5'
  const border = colors.border || '#e5e5e5'
  const isDark = document.documentElement.classList.contains('dark')

  const lineColors = ['#171717', '#525252', '#a3a3a3', '#d4d4d4']
  const pieColors = ['#171717', '#404040', '#737373', '#a3a3a3', '#d4d4d4', '#e5e5e5']

  const sparklineData = useMemo(() => {
    return {
      requests: trafficTrend.map(d => d.requests ?? 0),
      bytes: trafficTrend.map(d => d.bytes ?? 0),
      p50: responseTrend.map(d => d.p50 ?? 0),
      errorRate: errorRateTrend.map(d => d.errorRate ?? 0),
    }
  }, [trafficTrend, responseTrend, errorRateTrend])

  const trafficChartOption = useMemo(() => ({
    tooltip: {
      trigger: 'axis' as const,
      backgroundColor: colors.card || '#fff',
      borderColor: border,
      borderWidth: 1,
      textStyle: { color: fg },
    },
    legend: {
      data: ['请求数'],
      textStyle: { color: fg },
      top: 0,
    },
    grid: { left: 50, right: 20, top: 30, bottom: 30 },
    xAxis: {
      type: 'category' as const,
      data: trafficTrend.map(d => d.time),
      axisLabel: { color: muted, fontSize: 10 },
      axisLine: { lineStyle: { color: border } },
    },
    yAxis: {
      type: 'value' as const,
      axisLabel: { color: muted, fontSize: 10 },
      splitLine: { lineStyle: { color: border, type: 'dashed' as const } },
    },
    series: [{
      name: '请求数',
      type: 'line' as const,
      data: trafficTrend.map(d => d.requests ?? 0),
      smooth: true,
      areaStyle: { opacity: 0.1, color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
        { offset: 0, color: isDark ? 'rgba(255,255,255,0.15)' : 'rgba(0,0,0,0.1)' },
        { offset: 1, color: isDark ? 'rgba(255,255,255,0.02)' : 'rgba(0,0,0,0.02)' },
      ]) },
      lineStyle: { width: 2, color: lineColors[0] },
      itemStyle: { color: lineColors[0] },
      markPoint: {
        data: [{ type: 'max', name: '峰值' }],
        label: { color: fg },
      },
    }],
  }), [trafficTrend, colors, fg, muted, border, isDark])

  const bandwidthChartOption = useMemo(() => ({
    tooltip: {
      trigger: 'axis' as const,
      backgroundColor: colors.card || '#fff',
      borderColor: border,
      borderWidth: 1,
      textStyle: { color: fg },
      formatter: (params: any[]) => {
        const p = params[0]
        return `${p.axisValue}<br/>${p.seriesName}: ${formatBytes(p.value)}`
      },
    },
    legend: {
      data: ['带宽'],
      textStyle: { color: fg },
      top: 0,
    },
    grid: { left: 50, right: 20, top: 30, bottom: 30 },
    xAxis: {
      type: 'category' as const,
      data: trafficTrend.map(d => d.time),
      axisLabel: { color: muted, fontSize: 10 },
      axisLine: { lineStyle: { color: border } },
    },
    yAxis: {
      type: 'value' as const,
      axisLabel: { color: muted, fontSize: 10 },
      splitLine: { lineStyle: { color: border, type: 'dashed' as const } },
    },
    series: [{
      name: '带宽',
      type: 'line' as const,
      data: trafficTrend.map(d => d.bytes ?? 0),
      smooth: true,
      areaStyle: { opacity: 0.15, color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
        { offset: 0, color: isDark ? 'rgba(255,255,255,0.12)' : 'rgba(0,0,0,0.08)' },
        { offset: 1, color: isDark ? 'rgba(255,255,255,0.02)' : 'rgba(0,0,0,0.02)' },
      ]) },
      lineStyle: { width: 2, color: lineColors[1] },
      itemStyle: { color: lineColors[1] },
      markPoint: {
        data: [{ type: 'max', name: '峰值' }],
        label: { color: fg },
      },
    }],
  }), [trafficTrend, colors, fg, muted, border, isDark])

  const responseChartOption = useMemo(() => ({
    tooltip: {
      trigger: 'axis' as const,
      backgroundColor: colors.card || '#fff',
      borderColor: border,
      borderWidth: 1,
      textStyle: { color: fg },
      formatter: (params: any[]) => {
        let result = params[0].axisValue + '<br/>'
        params.forEach(p => {
          result += `${p.seriesName}: ${formatRT(p.value)}<br/>`
        })
        return result
      },
    },
    legend: {
      data: ['P50', 'P90', 'P99'],
      textStyle: { color: fg },
      top: 0,
    },
    grid: { left: 50, right: 20, top: 30, bottom: 30 },
    xAxis: {
      type: 'category' as const,
      data: responseTrend.map(d => d.time),
      axisLabel: { color: muted, fontSize: 10 },
      axisLine: { lineStyle: { color: border } },
    },
    yAxis: {
      type: 'value' as const,
      axisLabel: { color: muted, fontSize: 10, formatter: (v: number) => formatRT(v) },
      splitLine: { lineStyle: { color: border, type: 'dashed' as const } },
    },
    series: [
      { name: 'P50', type: 'line' as const, data: responseTrend.map(d => d.p50 ?? 0), smooth: true, lineStyle: { width: 2 }, itemStyle: { color: lineColors[0] } },
      { name: 'P90', type: 'line' as const, data: responseTrend.map(d => d.p90 ?? 0), smooth: true, lineStyle: { width: 2 }, itemStyle: { color: lineColors[1] } },
      { name: 'P99', type: 'line' as const, data: responseTrend.map(d => d.p99 ?? 0), smooth: true, lineStyle: { width: 2 }, itemStyle: { color: lineColors[2] } },
    ],
  }), [responseTrend, colors, fg, muted, border])

  const slowRequestChartOption = useMemo(() => ({
    tooltip: {
      trigger: 'axis' as const,
      backgroundColor: colors.card || '#fff',
      borderColor: border,
      borderWidth: 1,
      textStyle: { color: fg },
    },
    grid: { left: 50, right: 20, top: 20, bottom: 30 },
    xAxis: {
      type: 'category' as const,
      data: slowRequestTrend.map(d => d.time),
      axisLabel: { color: muted, fontSize: 10 },
      axisLine: { lineStyle: { color: border } },
    },
    yAxis: {
      type: 'value' as const,
      axisLabel: { color: muted, fontSize: 10 },
      splitLine: { lineStyle: { color: border, type: 'dashed' as const } },
    },
    series: [{
      name: '慢请求数',
      type: 'line' as const,
      data: slowRequestTrend.map(d => d.count ?? 0),
      smooth: true,
      areaStyle: { color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
        { offset: 0, color: isDark ? 'rgba(255,255,255,0.2)' : 'rgba(0,0,0,0.15)' },
        { offset: 1, color: isDark ? 'rgba(255,255,255,0.02)' : 'rgba(0,0,0,0.02)' },
      ]) },
      lineStyle: { width: 2, color: lineColors[0] },
      itemStyle: { color: lineColors[0] },
    }],
  }), [slowRequestTrend, colors, fg, muted, border, isDark])

  const methodPieOption = useMemo(() => ({
    tooltip: {
      trigger: 'item' as const,
      backgroundColor: colors.card || '#fff',
      borderColor: border,
      borderWidth: 1,
      textStyle: { color: fg },
    },
    series: [{
      type: 'pie' as const,
      radius: ['40%', '70%'],
      center: ['50%', '50%'],
      data: methodDistribution.map((d, i) => ({
        name: d.name,
        value: d.count,
        itemStyle: { color: pieColors[i % pieColors.length] },
      })),
      label: { color: fg, fontSize: 11 },
      emphasis: { label: { fontSize: 13, fontWeight: 'bold' } },
    }],
  }), [methodDistribution, colors, fg, border])

  const statusPieOption = useMemo(() => ({
    tooltip: {
      trigger: 'item' as const,
      backgroundColor: colors.card || '#fff',
      borderColor: border,
      borderWidth: 1,
      textStyle: { color: fg },
    },
    series: [{
      type: 'pie' as const,
      radius: ['40%', '70%'],
      center: ['50%', '50%'],
      data: statusDistribution.map((d, i) => ({
        name: d.name,
        value: d.count,
        itemStyle: { color: pieColors[i % pieColors.length] },
      })),
      label: { color: fg, fontSize: 11 },
      emphasis: { label: { fontSize: 13, fontWeight: 'bold' } },
    }],
  }), [statusDistribution, colors, fg, border])

  const errorRateChartOption = useMemo(() => ({
    tooltip: {
      trigger: 'axis' as const,
      backgroundColor: colors.card || '#fff',
      borderColor: border,
      borderWidth: 1,
      textStyle: { color: fg },
      formatter: (params: any[]) => {
        const p = params[0]
        return `${p.axisValue}<br/>错误率: ${p.value.toFixed(2)}%`
      },
    },
    grid: { left: 50, right: 20, top: 20, bottom: 30 },
    xAxis: {
      type: 'category' as const,
      data: errorRateTrend.map(d => d.time),
      axisLabel: { color: muted, fontSize: 10 },
      axisLine: { lineStyle: { color: border } },
    },
    yAxis: {
      type: 'value' as const,
      axisLabel: { color: muted, fontSize: 10, formatter: (v: number) => `${v.toFixed(1)}%` },
      splitLine: { lineStyle: { color: border, type: 'dashed' as const } },
    },
    series: [{
      name: '错误率',
      type: 'line' as const,
      data: errorRateTrend.map(d => d.errorRate ?? 0),
      smooth: true,
      lineStyle: { width: 2, color: lineColors[0] },
      itemStyle: { color: lineColors[0] },
      areaStyle: { color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
        { offset: 0, color: isDark ? 'rgba(255,255,255,0.15)' : 'rgba(0,0,0,0.1)' },
        { offset: 1, color: isDark ? 'rgba(255,255,255,0.02)' : 'rgba(0,0,0,0.02)' },
      ]) },
    }],
  }), [errorRateTrend, colors, fg, muted, border, isDark])

  const errorPathsBarOption = useMemo(() => ({
    tooltip: {
      trigger: 'axis' as const,
      backgroundColor: colors.card || '#fff',
      borderColor: border,
      borderWidth: 1,
      textStyle: { color: fg },
    },
    grid: { left: 120, right: 20, top: 10, bottom: 30 },
    xAxis: {
      type: 'value' as const,
      axisLabel: { color: muted, fontSize: 10 },
      splitLine: { lineStyle: { color: border, type: 'dashed' as const } },
    },
    yAxis: {
      type: 'category' as const,
      data: errorPaths.slice(0, 10).map(d => d.path),
      axisLabel: { color: muted, fontSize: 10, width: 100, overflow: 'truncate' as const },
      axisLine: { lineStyle: { color: border } },
    },
    series: [{
      type: 'bar' as const,
      data: errorPaths.slice(0, 10).map(d => d.count),
      itemStyle: { color: lineColors[0], borderRadius: [0, 4, 4, 0] },
      barWidth: '60%',
    }],
  }), [errorPaths, colors, fg, muted, border])

  const deviceTypePieOption = useMemo(() => ({
    tooltip: {
      trigger: 'item' as const,
      backgroundColor: colors.card || '#fff',
      borderColor: border,
      borderWidth: 1,
      textStyle: { color: fg },
    },
    series: [{
      type: 'pie' as const,
      radius: ['40%', '70%'],
      data: (clientAnalysis?.deviceTypeRank || []).map((d, i) => ({
        name: d.name,
        value: d.count,
        itemStyle: { color: pieColors[i % pieColors.length] },
      })),
      label: { color: fg, fontSize: 11 },
    }],
  }), [clientAnalysis, colors, fg, border])

  const browserPieOption = useMemo(() => ({
    tooltip: {
      trigger: 'item' as const,
      backgroundColor: colors.card || '#fff',
      borderColor: border,
      borderWidth: 1,
      textStyle: { color: fg },
    },
    series: [{
      type: 'pie' as const,
      radius: ['40%', '70%'],
      data: (clientAnalysis?.browserRank || []).slice(0, 6).map((d, i) => ({
        name: d.name,
        value: d.count,
        itemStyle: { color: pieColors[i % pieColors.length] },
      })),
      label: { color: fg, fontSize: 11 },
    }],
  }), [clientAnalysis, colors, fg, border])

  const osPieOption = useMemo(() => ({
    tooltip: {
      trigger: 'item' as const,
      backgroundColor: colors.card || '#fff',
      borderColor: border,
      borderWidth: 1,
      textStyle: { color: fg },
    },
    series: [{
      type: 'pie' as const,
      radius: ['40%', '70%'],
      data: (clientAnalysis?.osRank || []).slice(0, 6).map((d, i) => ({
        name: d.name,
        value: d.count,
        itemStyle: { color: pieColors[i % pieColors.length] },
      })),
      label: { color: fg, fontSize: 11 },
    }],
  }), [clientAnalysis, colors, fg, border])

  const insights = useMemo(() => {
    const results: string[] = []

    if (trafficTrend.length > 0) {
      const maxTraffic = trafficTrend.reduce((max, d) => Math.max(max, d.requests ?? 0), 0)
      const maxPoint = trafficTrend.find(d => d.requests === maxTraffic)
      if (maxPoint) {
        results.push(`流量峰值出现在 ${maxPoint.time}，达到 ${maxTraffic.toLocaleString()} 次请求`)
      }
    }

    if (overview && overview.avgRT > 1) {
      results.push(`平均响应时间为 ${formatRT(overview.avgRT)}，存在性能优化空间`)
    }

    if (errorRateTrend.length > 0) {
      const avgErrorRate = errorRateTrend.reduce((sum, d) => sum + (d.errorRate ?? 0), 0) / errorRateTrend.length
      if (avgErrorRate > 5) {
        results.push(`平均错误率为 ${avgErrorRate.toFixed(2)}%，建议排查高频错误路径`)
      }
    }

    if (errorPaths.length > 0) {
      const topPath = errorPaths[0]
      results.push(`${topPath.path} 是错误最多的路径，共 ${topPath.count} 次错误`)
    }

    if (clientAnalysis?.deviceTypeRank && clientAnalysis.deviceTypeRank.length > 0) {
      const topDevice = clientAnalysis.deviceTypeRank[0]
      results.push(`${topDevice.name} 设备占比最高，达到 ${topDevice.percent.toFixed(1)}%`)
    }

    return results
  }, [trafficTrend, overview, errorRateTrend, errorPaths, clientAnalysis])

  const relatedTabs = useMemo(() => {
    const tabMap: Record<TabKey, TabKey[]> = {
      traffic: ['response', 'status'],
      response: ['traffic', 'error'],
      error: ['response', 'status'],
      client: ['traffic', 'response'],
      status: ['error', 'traffic'],
    }
    return tabMap[activeTab] || []
  }, [activeTab])

  if (loading && !overview) {
    return (
      <div className="flex items-center justify-center h-96">
        <Loader2 className="h-8 w-8 animate-spin text-primary" />
      </div>
    )
  }

  if (error) {
    return (
      <div className="flex flex-col items-center justify-center h-96 text-center">
        <AlertCircle className="h-12 w-12 text-destructive mb-4" />
        <h3 className="text-lg font-semibold mb-2">数据加载失败</h3>
        <p className="text-muted-foreground mb-4">{error}</p>
        <button onClick={() => window.location.reload()} className="px-4 py-2 bg-primary text-primary-foreground rounded-md hover:bg-primary/90">
          重新加载
        </button>
      </div>
    )
  }

  const renderTabContent = () => {
    switch (activeTab) {
      case 'traffic':
        return (
          <div className="grid gap-4 md:grid-cols-2">
            <ChartCard title="请求量趋势">
              <ReactECharts option={trafficChartOption} style={{ height: '300px' }} opts={{ renderer: 'canvas' }} />
            </ChartCard>
            <ChartCard title="带宽趋势">
              <ReactECharts option={bandwidthChartOption} style={{ height: '300px' }} opts={{ renderer: 'canvas' }} />
            </ChartCard>
            <ChartCard title="请求方法分布">
              <ReactECharts option={methodPieOption} style={{ height: '280px' }} opts={{ renderer: 'canvas' }} />
            </ChartCard>
            <div className="space-y-4">
              <ChartCard title="慢请求趋势">
                <ReactECharts option={slowRequestChartOption} style={{ height: '200px' }} opts={{ renderer: 'canvas' }} />
              </ChartCard>
            </div>
          </div>
        )
      case 'response':
        return (
          <div className="grid gap-4 md:grid-cols-2">
            <ChartCard title="响应时间分布 (P50/P90/P99)">
              <ReactECharts option={responseChartOption} style={{ height: '300px' }} opts={{ renderer: 'canvas' }} />
            </ChartCard>
            <ChartCard title="慢请求趋势 (RT > 1s)">
              <ReactECharts option={slowRequestChartOption} style={{ height: '300px' }} opts={{ renderer: 'canvas' }} />
            </ChartCard>
          </div>
        )
      case 'error':
        return (
          <div className="grid gap-4 md:grid-cols-2">
            <ChartCard title="错误率趋势">
              <ReactECharts option={errorRateChartOption} style={{ height: '280px' }} opts={{ renderer: 'canvas' }} />
            </ChartCard>
            <ChartCard title="状态码分布">
              <ReactECharts option={statusPieOption} style={{ height: '280px' }} opts={{ renderer: 'canvas' }} />
            </ChartCard>
            <div className="md:col-span-2">
              <ChartCard title="错误路径 TOP 10">
                <ReactECharts option={errorPathsBarOption} style={{ height: '350px' }} opts={{ renderer: 'canvas' }} />
              </ChartCard>
            </div>
          </div>
        )
      case 'client':
        return (
          <div className="grid gap-4 md:grid-cols-3">
            <ChartCard title="设备类型分布">
              <ReactECharts option={deviceTypePieOption} style={{ height: '280px' }} opts={{ renderer: 'canvas' }} />
            </ChartCard>
            <ChartCard title="浏览器分布">
              <ReactECharts option={browserPieOption} style={{ height: '280px' }} opts={{ renderer: 'canvas' }} />
            </ChartCard>
            <ChartCard title="操作系统分布">
              <ReactECharts option={osPieOption} style={{ height: '280px' }} opts={{ renderer: 'canvas' }} />
            </ChartCard>
          </div>
        )
      case 'status':
        return (
          <div className="grid gap-4 md:grid-cols-2">
            <ChartCard title="状态码分布">
              <ReactECharts option={statusPieOption} style={{ height: '300px' }} opts={{ renderer: 'canvas' }} />
            </ChartCard>
            <ChartCard title="错误率趋势">
              <ReactECharts option={errorRateChartOption} style={{ height: '300px' }} opts={{ renderer: 'canvas' }} />
            </ChartCard>
            <div className="md:col-span-2">
              <ChartCard title="错误路径 TOP 10">
                <ReactECharts option={errorPathsBarOption} style={{ height: '350px' }} opts={{ renderer: 'canvas' }} />
              </ChartCard>
            </div>
          </div>
        )
      default:
        return null
    }
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h2 className="text-lg font-semibold">指标分析</h2>
      </div>

      <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
        <div className="flex flex-wrap items-center gap-2">
          <div className="relative" ref={presetRef}>
            <button
              onClick={() => setShowPresets(!showPresets)}
              className="flex items-center gap-2 px-3 py-2 text-sm border rounded-lg hover:bg-muted transition-colors"
            >
              <Calendar className="h-4 w-4 text-muted-foreground" />
              <span>{dateRange?.from?.toLocaleDateString('zh-CN')} ~ {dateRange?.to?.toLocaleDateString('zh-CN')}</span>
            </button>
            {showPresets && (
              <div className="absolute top-full left-0 mt-2 p-2 bg-card border rounded-lg shadow-lg z-50 min-w-[180px]">
                {timePresets.map((preset) => (
                  <button
                    key={preset.label}
                    onClick={() => handlePresetSelect(preset.range())}
                    className="w-full text-left px-3 py-2 text-sm rounded-md hover:bg-muted transition-colors"
                  >
                    {preset.label}
                  </button>
                ))}
                <div className="mt-2 pt-2 border-t">
                  <DateRangePicker
                    value={dateRange}
                    onChange={handleDateRangeChange}
                    loading={loading}
                  />
                </div>
              </div>
            )}
          </div>
          <Select value={granularity} onValueChange={handleGranularityChange}>
            <SelectTrigger className="w-[100px]">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {granularityOptions.map(opt => (
                <SelectItem key={opt.value} value={opt.value}>{opt.label}</SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
      </div>

      <div className="grid gap-3 grid-cols-2 lg:grid-cols-4">
        <MetricCardWithSparkline
          icon={<Activity className="h-4 w-4" />}
          label="请求总数"
          value={overview?.totalRequests.toLocaleString() ?? '0'}
          sparklineData={sparklineData.requests}
          onClick={() => handleTabChange('traffic')}
          active={activeTab === 'traffic'}
        />
        <MetricCardWithSparkline
          icon={<TrendingUp className="h-4 w-4" />}
          label="总流量"
          value={formatBytes(overview?.totalBytes ?? 0)}
          sparklineData={sparklineData.bytes}
          onClick={() => handleTabChange('traffic')}
          active={activeTab === 'traffic'}
        />
        <MetricCardWithSparkline
          icon={<Zap className="h-4 w-4" />}
          label="峰值 QPS"
          value={overview?.peakQPS.toFixed(2) ?? '0'}
          onClick={() => handleTabChange('response')}
          active={activeTab === 'response'}
        />
        <MetricCardWithSparkline
          icon={<Clock className="h-4 w-4" />}
          label="平均响应"
          value={formatRT(overview?.avgRT ?? 0)}
          sparklineData={sparklineData.p50}
          onClick={() => handleTabChange('response')}
          active={activeTab === 'response'}
        />
      </div>

      <div>
        <div className="flex items-center gap-1 border-b border-border">
          {tabs.map((tab) => (
            <button
              key={tab.key}
              onClick={() => handleTabChange(tab.key)}
              className={cn(
                'flex items-center gap-1.5 px-4 py-2.5 text-sm font-medium border-b-2 transition-colors',
                activeTab === tab.key
                  ? 'border-foreground text-foreground'
                  : 'border-transparent text-muted-foreground hover:text-foreground'
              )}
            >
              <tab.icon className="h-4 w-4" />
              <span>{tab.label}</span>
            </button>
          ))}
        </div>
      </div>

      <div className={cn(
        'transition-all duration-200',
        tabTransition === 'exit' ? 'opacity-0 translate-y-1' : tabTransition === 'enter' ? 'opacity-100 translate-y-0' : 'opacity-100 translate-y-0'
      )}>
        {renderTabContent()}
      </div>

      {insights.length > 0 && (
        <div className="space-y-2">
          <h4 className="text-sm font-medium text-muted-foreground">智能洞察</h4>
          <div className="grid gap-2 md:grid-cols-2">
            {insights.map((insight, index) => (
              <InsightBadge key={index}>{insight}</InsightBadge>
            ))}
          </div>
        </div>
      )}

      {relatedTabs.length > 0 && (
        <div className="flex items-center gap-2 flex-wrap">
          <span className="text-xs text-muted-foreground">相关指标:</span>
          {relatedTabs.map(tabKey => {
            const tab = tabs.find(t => t.key === tabKey)
            if (!tab) return null
            return (
              <button
                key={tabKey}
                onClick={() => handleTabChange(tabKey)}
                className="flex items-center gap-1 px-3 py-1.5 text-xs rounded-full border border-border hover:bg-muted transition-colors"
              >
                <tab.icon className="h-3 w-3" />
                <span>{tab.label}</span>
                <ChevronRight className="h-3 w-3" />
              </button>
            )
          })}
        </div>
      )}
    </div>
  )
}
