import { useState, useEffect, useMemo, useCallback } from 'react'
import { DateRange } from 'react-day-picker'
import * as echarts from 'echarts'
import ReactECharts from 'echarts-for-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { metricsApi, MetricsOverview, TrendPoint, DistributionItem, ErrorPathItem, ClientAnalysis } from '@/api/metrics'
import { useThemeColors } from '@/hooks/useThemeColor'
import DateRangePicker from '@/components/DateRangePicker'
import { cn } from '@/lib/utils'
import { Loader2, AlertCircle, TrendingUp, Activity, Zap, Clock } from 'lucide-react'

type Granularity = '1m' | '5m' | '1h' | '1d'

const granularityOptions: { value: Granularity; label: string }[] = [
  { value: '1m', label: '1分钟' },
  { value: '5m', label: '5分钟' },
  { value: '1h', label: '1小时' },
  { value: '1d', label: '1天' },
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

function MetricCard({ icon, label, value, className }: { icon: React.ReactNode; label: string; value: string; className?: string }) {
  return (
    <Card className={className}>
      <CardContent className="p-4">
        <div className="flex items-center gap-3">
          <div className="p-2 rounded-lg bg-primary/10 text-primary">
            {icon}
          </div>
          <div>
            <p className="text-xs text-muted-foreground">{label}</p>
            <p className="text-lg font-semibold">{value}</p>
          </div>
        </div>
      </CardContent>
    </Card>
  )
}

function ChartCard({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <Card>
      <CardHeader className="pb-2">
        <CardTitle className="text-sm font-medium">{title}</CardTitle>
      </CardHeader>
      <CardContent>
        {children}
      </CardContent>
    </Card>
  )
}

export default function MetricsAnalysis() {
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [granularity, setGranularity] = useState<Granularity>('5m')
  const [dateRange, setDateRange] = useState<DateRange | undefined>(() => {
    const now = new Date()
    const start = new Date(now.getTime() - 24 * 60 * 60 * 1000)
    return { from: start, to: now }
  })
  const [overview, setOverview] = useState<MetricsOverview | null>(null)
  const [trafficTrend, setTrafficTrend] = useState<TrendPoint[]>([])
  const [responseTrend, setResponseTrend] = useState<TrendPoint[]>([])
  const [slowRequestTrend, setSlowRequestTrend] = useState<TrendPoint[]>([])
  const [methodDistribution, setMethodDistribution] = useState<DistributionItem[]>([])
  const [statusDistribution, setStatusDistribution] = useState<DistributionItem[]>([])
  const [errorRateTrend, setErrorRateTrend] = useState<TrendPoint[]>([])
  const [errorPaths, setErrorPaths] = useState<ErrorPathItem[]>([])
  const [clientAnalysis, setClientAnalysis] = useState<ClientAnalysis | null>(null)

  const colors = useThemeColors()

  const getApiParams = useCallback(() => {
    const params: { start: string; end: string; granularity?: string } = {
      start: dateRange?.from?.toISOString() ?? '',
      end: dateRange?.to?.toISOString() ?? '',
    }
    return params
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

  const handleDateRangeChange = useCallback((range: DateRange) => {
    setDateRange(range)
  }, [])

  const handleGranularityChange = useCallback((value: string) => {
    setGranularity(value as Granularity)
  }, [])

  const fg = colors.foreground || '#0a0a0a'
  const muted = colors.muted || '#e5e5e5'
  const border = colors.border || '#e5e5e5'

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
      areaStyle: { opacity: 0.1 },
      lineStyle: { width: 2 },
      itemStyle: { color: colors.border || '#3b82f6' },
      markPoint: {
        data: [
          { type: 'max', name: '峰值' },
        ],
      },
    }],
  }), [trafficTrend, colors, fg, muted, border])

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
      areaStyle: { opacity: 0.2 },
      lineStyle: { width: 2 },
      itemStyle: { color: colors.mutedForeground || '#8b5cf6' },
      markPoint: {
        data: [{ type: 'max', name: '峰值' }],
      },
    }],
  }), [trafficTrend, colors, fg, muted, border])

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
      { name: 'P50', type: 'line' as const, data: responseTrend.map(d => d.p50 ?? 0), smooth: true, lineStyle: { width: 2 }, itemStyle: { color: '#22c55e' } },
      { name: 'P90', type: 'line' as const, data: responseTrend.map(d => d.p90 ?? 0), smooth: true, lineStyle: { width: 2 }, itemStyle: { color: '#f59e0b' } },
      { name: 'P99', type: 'line' as const, data: responseTrend.map(d => d.p99 ?? 0), smooth: true, lineStyle: { width: 2 }, itemStyle: { color: '#ef4444' } },
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
        { offset: 0, color: 'rgba(239,68,68,0.3)' },
        { offset: 1, color: 'rgba(239,68,68,0.05)' },
      ]) },
      lineStyle: { width: 2, color: '#ef4444' },
      itemStyle: { color: '#ef4444' },
    }],
  }), [slowRequestTrend, colors, fg, muted, border])

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
        itemStyle: { color: ['#3b82f6', '#22c55e', '#f59e0b', '#ef4444', '#8b5cf6'][i % 5] },
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
        itemStyle: { color: ['#22c55e', '#3b82f6', '#f59e0b', '#ef4444'][i % 4] },
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
      lineStyle: { width: 2, color: '#ef4444' },
      itemStyle: { color: '#ef4444' },
      areaStyle: { color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
        { offset: 0, color: 'rgba(239,68,68,0.3)' },
        { offset: 1, color: 'rgba(239,68,68,0.05)' },
      ]) },
    }],
  }), [errorRateTrend, colors, fg, muted, border])

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
      itemStyle: { color: '#ef4444', borderRadius: [0, 4, 4, 0] },
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
        itemStyle: { color: ['#3b82f6', '#22c55e', '#f59e0b', '#ef4444', '#8b5cf6'][i % 5] },
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
        itemStyle: { color: ['#3b82f6', '#f59e0b', '#ef4444', '#22c55e', '#8b5cf6', '#ec4899'][i % 6] },
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
        itemStyle: { color: ['#3b82f6', '#22c55e', '#f59e0b', '#8b5cf6', '#ec4899', '#06b6d4'][i % 6] },
      })),
      label: { color: fg, fontSize: 11 },
    }],
  }), [clientAnalysis, colors, fg, border])

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

  return (
    <div className="space-y-6">
      {/* Global Filter Bar */}
      <div className="sticky top-0 z-10 bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60 border-b pb-4">
        <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <h2 className="text-lg font-semibold">指标分析</h2>
          <div className="flex flex-wrap items-center gap-3">
            <DateRangePicker
              value={dateRange}
              onChange={handleDateRangeChange}
              loading={loading}
              className="min-w-[260px]"
            />
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
      </div>

      {/* Module 1: Traffic Overview */}
      <div className="space-y-4">
        <h3 className="text-base font-semibold">流量概览</h3>
        <div className="grid gap-4 grid-cols-2 lg:grid-cols-4">
          <MetricCard icon={<Activity className="h-5 w-5" />} label="请求总数" value={overview?.totalRequests.toLocaleString() ?? '0'} />
          <MetricCard icon={<TrendingUp className="h-5 w-5" />} label="总流量" value={formatBytes(overview?.totalBytes ?? 0)} />
          <MetricCard icon={<Zap className="h-5 w-5" />} label="峰值 QPS" value={overview?.peakQPS.toFixed(2) ?? '0'} />
          <MetricCard icon={<Clock className="h-5 w-5" />} label="平均响应时间" value={formatRT(overview?.avgRT ?? 0)} />
        </div>
        <div className="grid gap-4 md:grid-cols-2">
          <ChartCard title="请求量趋势">
            <ReactECharts option={trafficChartOption} style={{ height: '300px' }} opts={{ renderer: 'canvas' }} />
          </ChartCard>
          <ChartCard title="带宽趋势">
            <ReactECharts option={bandwidthChartOption} style={{ height: '300px' }} opts={{ renderer: 'canvas' }} />
          </ChartCard>
        </div>
      </div>

      {/* Module 2: Response Performance */}
      <div className="space-y-4">
        <h3 className="text-base font-semibold">响应性能</h3>
        <div className="grid gap-4 md:grid-cols-2">
          <ChartCard title="响应时间分布 (P50/P90/P99)">
            <ReactECharts option={responseChartOption} style={{ height: '300px' }} opts={{ renderer: 'canvas' }} />
          </ChartCard>
          <ChartCard title="慢请求趋势 (RT > 1s)">
            <ReactECharts option={slowRequestChartOption} style={{ height: '300px' }} opts={{ renderer: 'canvas' }} />
          </ChartCard>
        </div>
        <div className="grid gap-4 md:grid-cols-3">
          <ChartCard title="请求方法分布">
            <ReactECharts option={methodPieOption} style={{ height: '280px' }} opts={{ renderer: 'canvas' }} />
          </ChartCard>
          <ChartCard title="状态码分布">
            <ReactECharts option={statusPieOption} style={{ height: '280px' }} opts={{ renderer: 'canvas' }} />
          </ChartCard>
          <ChartCard title="错误率趋势">
            <ReactECharts option={errorRateChartOption} style={{ height: '280px' }} opts={{ renderer: 'canvas' }} />
          </ChartCard>
        </div>
      </div>

      {/* Module 3: Status Code Analysis */}
      <div className="space-y-4">
        <h3 className="text-base font-semibold">错误路径 TOP</h3>
        <ChartCard title="">
          <ReactECharts option={errorPathsBarOption} style={{ height: '350px' }} opts={{ renderer: 'canvas' }} />
        </ChartCard>
      </div>

      {/* Module 4: Client Analysis */}
      <div className="space-y-4">
        <h3 className="text-base font-semibold">客户端分析</h3>
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
      </div>
    </div>
  )
}
