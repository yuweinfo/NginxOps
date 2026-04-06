import { useState, useEffect, useMemo, useRef } from 'react'
import { 
  Activity, 
  BarChart3, 
  Users, 
  Server,
  Loader2,
  AlertCircle
} from 'lucide-react'
import ReactECharts from 'echarts-for-react'
import * as echarts from 'echarts'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { statsApi, DashboardData } from '@/api/stats'
import { useThemeColors } from '@/hooks/useThemeColor'
import { cn } from '@/lib/utils'
// @ts-ignore
import worldMapGeojson from 'world-map-geojson'

// 默认空数据
const emptyData: DashboardData = {
  qps: 0,
  bandwidth: 0,
  pvToday: 0,
  activeSites: 0,
  hourlyTrend: [],
  statusDistribution: { '2xx': 0, '3xx': 0, '4xx': 0, '5xx': 0 },
  ipLocations: [],
  ipRegionRank: [],
}

export default function Dashboard() {
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [data, setData] = useState<DashboardData>(emptyData)
  const [mapReady, setMapReady] = useState(false)
  const [isChinaMap, setIsChinaMap] = useState(false)
  const chartRef = useRef<any>(null)
  const colors = useThemeColors()

  useEffect(() => {
    fetch('https://geo.datav.aliyun.com/areas_v3/bound/100000_full.json')
      .then(r => r.json())
      .then(chinaMap => {
        echarts.registerMap('world', worldMapGeojson)
        echarts.registerMap('china', chinaMap)
        setMapReady(true)
      })
      .catch(() => {
        echarts.registerMap('world', worldMapGeojson)
        setMapReady(true)
      })
  }, [])

  useEffect(() => {
    const fetchData = async () => {
      try {
        setLoading(true)
        setError(null)
        const res = await statsApi.getDashboard()
        if (res.success) {
          setData({
            qps: res.data.qps || 0,
            bandwidth: res.data.bandwidth || 0,
            pvToday: res.data.pvToday || 0,
            activeSites: res.data.activeSites || 0,
            hourlyTrend: res.data.hourlyTrend || [],
            statusDistribution: res.data.statusDistribution || { '2xx': 0, '3xx': 0, '4xx': 0, '5xx': 0 },
            ipLocations: res.data.ipLocations || [],
            ipRegionRank: res.data.ipRegionRank || [],
          })
        } else {
          setError(res.message || '获取数据失败')
        }
      } catch (e: any) {
        console.error('Failed to fetch dashboard data:', e)
        setError(e.message || '网络请求失败，请检查后端服务是否启动')
      } finally {
        setLoading(false)
      }
    }
    fetchData()
  }, [])

  useEffect(() => {
    const timer = setInterval(async () => {
      try {
        const res = await statsApi.getDashboard()
        if (res.success) {
          setData(prev => ({
            ...prev,
            qps: res.data.qps || 0,
            bandwidth: res.data.bandwidth || 0,
          }))
        }
      } catch (e) {
        // ignore periodic update errors
      }
    }, 5000)
    return () => clearInterval(timer)
  }, [])

  const lineChartOption = useMemo(() => {
    const fg = colors.foreground || '#0a0a0a'
    return {
      tooltip: {
        trigger: 'axis',
        backgroundColor: colors.card || '#fff',
        borderColor: colors.border || '#e5e5e5',
        borderWidth: 1,
        borderRadius: 8,
        textStyle: { color: fg, fontSize: 13 },
      },
      grid: {
        left: '3%',
        right: '4%',
        bottom: '3%',
        containLabel: true,
      },
      xAxis: {
        type: 'category',
        boundaryGap: false,
        data: data?.hourlyTrend?.map((item) => item.hour) || [],
        axisLine: { lineStyle: { color: colors.border || '#e5e5e5' } },
        axisLabel: { color: colors.mutedForeground || '#737373', fontSize: 11 },
      },
      yAxis: {
        type: 'value',
        axisLine: { show: false },
        axisLabel: { color: colors.mutedForeground || '#737373', fontSize: 11 },
        splitLine: { lineStyle: { color: colors.border || '#e5e5e5', opacity: 0.5 } },
      },
      series: [
        {
          name: '请求数',
          type: 'line',
          smooth: true,
          data: data?.hourlyTrend?.map((item) => item.requests) || [],
          itemStyle: { color: fg },
          lineStyle: { color: fg, width: 2 },
          areaStyle: {
            color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
              { offset: 0, color: `${fg}33` },
              { offset: 1, color: `${fg}05` },
            ]),
          },
          symbol: 'none',
          emphasis: {
            focus: 'series',
          },
        },
      ],
    }
  }, [data?.hourlyTrend, colors])

  const pieChartOption = useMemo(() => {
    const status = data?.statusDistribution || { '2xx': 0, '4xx': 0, '5xx': 0 }
    const fg = colors.foreground || '#0a0a0a'
    const bg = colors.background || '#fafafa'
    const muted = colors.mutedForeground || '#737373'
    const destructive = colors.destructive || '#ef4444'
    
    return {
      tooltip: {
        trigger: 'item',
        formatter: '{b}: {c} ({d}%)',
        backgroundColor: colors.card || '#fff',
        borderColor: colors.border || '#e5e5e5',
        borderWidth: 1,
        borderRadius: 8,
        textStyle: { color: fg, fontSize: 13 },
      },
      legend: {
        orient: 'horizontal',
        bottom: '5%',
        left: 'center',
        textStyle: { color: muted, fontSize: 12 },
      },
      series: [
        {
          name: 'HTTP Status',
          type: 'pie',
          radius: ['50%', '70%'],
          center: ['50%', '42%'],
          avoidLabelOverlap: true,
          itemStyle: {
            borderRadius: 6,
            borderColor: bg,
            borderWidth: 2,
          },
          label: {
            show: true,
            position: 'outside',
            formatter: '{b}\n{d}%',
            color: fg,
            fontSize: 12,
            fontWeight: 500,
          },
          labelLine: {
            show: true,
            length: 15,
            length2: 10,
            lineStyle: { color: colors.border || '#e5e5e5' },
          },
          data: [
            { 
              value: status['2xx'], 
              name: '2xx', 
              itemStyle: { color: fg } 
            },
            { 
              value: status['4xx'], 
              name: '4xx', 
              itemStyle: { color: destructive } 
            },
            { 
              value: status['5xx'], 
              name: '5xx', 
              itemStyle: { color: destructive, opacity: 0.6 } 
            },
          ],
        },
      ],
    }
  }, [data?.statusDistribution, colors])

  const handleMapClick = (params: any) => {
    if (!isChinaMap && (params.name === 'China')) {
      setIsChinaMap(true)
    }
  }

  const handleBackToWorld = () => {
    setIsChinaMap(false)
  }

  const mapChartOption = useMemo(() => {
    const ipLocations = data?.ipLocations || []
    const fg = colors.foreground || '#0a0a0a'
    const border = colors.border || '#e5e5e5'
    const accent = colors.accent || '#f5f5f5'
    const muted = colors.muted || '#e5e5e5'
    const destructive = colors.destructive || '#ef4444'

    // 热力图数据转换
    const heatData = ipLocations.map(loc => [loc.value[0], loc.value[1], loc.value[2]])

    if (isChinaMap) {
      const chinaData = ipLocations.filter(loc => loc.country === 'China')
      const chinaHeatData = chinaData.map(loc => [loc.value[0], loc.value[1], loc.value[2]])

      return {
        tooltip: {
          trigger: 'item',
          backgroundColor: colors.card || '#fff',
          borderColor: border,
          borderWidth: 1,
          borderRadius: 8,
          textStyle: { color: fg, fontSize: 13 },
          formatter: (params: any) => {
            if (params.seriesType === 'scatter' || params.seriesType === 'effectScatter') {
              return `${params.name}<br/>访问量: ${params.value[2]}`
            }
            return params.name
          },
        },
        visualMap: {
          show: true,
          min: 0,
          max: Math.max(...chinaData.map(d => d.value[2]), 1000),
          calculable: true,
          inRange: {
            color: ['#e5e5e5', '#888888', '#333333', '#000000']
          },
          text: ['高', '低'],
          textStyle: { color: fg },
          bottom: 20,
          left: 20,
        },
        geo: {
          map: 'china',
          roam: true,
          zoom: 1.2,
          center: [105, 36],
          scaleLimit: { min: 1, max: 10 },
          itemStyle: {
            areaColor: muted,
            borderColor: border,
            borderWidth: 0.5,
          },
          emphasis: {
            itemStyle: {
              areaColor: accent,
            },
          },
        },
        series: [
          // 热力图层
          {
            name: '访问热力',
            type: 'heatmap',
            coordinateSystem: 'geo',
            data: chinaHeatData,
            pointSize: 25,
            blurSize: 15,
          },
          // 动态热点标记
          {
            name: '活跃区域',
            type: 'effectScatter',
            coordinateSystem: 'geo',
            data: chinaData.slice(0, 3),
            symbolSize: (val: number[]) => Math.max(Math.sqrt(val[2]) / 5, 12),
            showEffectOn: 'render',
            rippleEffect: {
              brushType: 'stroke',
              scale: 4,
              period: 3,
            },
            itemStyle: {
              color: destructive,
              shadowBlur: 15,
              shadowColor: `${destructive}66`,
            },
            label: {
              show: true,
              position: 'right',
              formatter: '{b}',
              color: fg,
              fontSize: 11,
            },
          },
        ],
      }
    }

    return {
      tooltip: {
        trigger: 'item',
        backgroundColor: colors.card || '#fff',
        borderColor: border,
        borderWidth: 1,
        borderRadius: 8,
        textStyle: { color: fg, fontSize: 13 },
        formatter: (params: any) => {
          if (params.seriesType === 'scatter' || params.seriesType === 'effectScatter') {
            return `${params.name}<br/>访问量: ${params.value[2]}`
          }
          return params.name
        },
      },
      visualMap: {
        show: true,
        min: 0,
        max: Math.max(...ipLocations.map(d => d.value[2]), 1000),
        calculable: true,
        inRange: {
          color: ['#e5e5e5', '#888888', '#333333', '#000000']
        },
        text: ['高', '低'],
        textStyle: { color: fg },
        bottom: 20,
        left: 20,
      },
      geo: {
        map: 'world',
        roam: true,
        zoom: 1.2,
        center: [30, 30],
        scaleLimit: { min: 1, max: 20 },
        nameMap: {
          'China': '中国',
          'United States': '美国',
          'Japan': '日本',
          'Singapore': '新加坡',
          'Germany': '德国',
          'United Kingdom': '英国',
          'Australia': '澳大利亚',
          'France': '法国',
          'Russia': '俄罗斯',
          'India': '印度',
        },
        itemStyle: {
          areaColor: muted,
          borderColor: border,
          borderWidth: 0.5,
        },
        emphasis: {
          itemStyle: {
            areaColor: accent,
          },
        },
      },
      series: [
        // 热力图层
        {
          name: '访问热力',
          type: 'heatmap',
          coordinateSystem: 'geo',
          data: heatData,
          pointSize: 20,
          blurSize: 12,
        },
        // 动态热点标记（Top 5）
        {
          name: '活跃区域',
          type: 'effectScatter',
          coordinateSystem: 'geo',
          data: ipLocations.slice(0, 5),
          symbolSize: (val: number[]) => Math.max(Math.sqrt(val[2]) / 6, 10),
          showEffectOn: 'render',
          rippleEffect: {
            brushType: 'stroke',
            scale: 4,
            period: 3,
          },
          itemStyle: {
            color: destructive,
            shadowBlur: 15,
            shadowColor: `${destructive}66`,
          },
          label: {
            show: true,
            position: 'right',
            formatter: '{b}',
            color: fg,
            fontSize: 10,
          },
        },
      ],
    }
  }, [data?.ipLocations, isChinaMap, colors])

  const statsCards = [
    { title: '当前 QPS', value: data?.qps || 0, suffix: 'req/s', icon: Activity },
    { title: '带宽', value: data?.bandwidth || 0, suffix: 'MB/s', icon: BarChart3 },
    { title: '今日 PV', value: data?.pvToday || 0, suffix: '', icon: Users },
    { title: '活跃站点', value: data?.activeSites || 0, suffix: '', icon: Server },
  ]

  if (loading) {
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
        <button
          onClick={() => window.location.reload()}
          className="px-4 py-2 bg-primary text-primary-foreground rounded-md hover:bg-primary/90"
        >
          重新加载
        </button>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      {/* Stats Cards */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        {statsCards.map((stat, index) => {
          const Icon = stat.icon
          return (
            <Card key={index}>
              <CardContent className="p-6">
                <div className="flex items-center gap-4">
                  <div className="p-3 rounded-lg bg-muted text-foreground">
                    <Icon className="h-6 w-6" />
                  </div>
                  <div>
                    <p className="text-sm text-muted-foreground">{stat.title}</p>
                    <p className="text-2xl font-bold">
                      {stat.value.toLocaleString()}
                      {stat.suffix && (
                        <span className="text-sm font-normal text-muted-foreground ml-1">
                          {stat.suffix}
                        </span>
                      )}
                    </p>
                  </div>
                </div>
              </CardContent>
            </Card>
          )
        })}
      </div>

      {/* Charts */}
      <div className="grid gap-6 md:grid-cols-3">
        <Card className="md:col-span-2">
          <CardHeader>
            <CardTitle className="text-base font-medium">请求趋势 (24小时)</CardTitle>
          </CardHeader>
          <CardContent>
            <ReactECharts option={lineChartOption} style={{ height: 300 }} />
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle className="text-base font-medium">HTTP 状态分布</CardTitle>
          </CardHeader>
          <CardContent>
            <ReactECharts option={pieChartOption} style={{ height: 300 }} />
          </CardContent>
        </Card>
      </div>

      {/* Map */}
      <Card>
        <CardHeader className="flex flex-row items-center justify-between">
          <CardTitle className="text-base font-medium">访客分布</CardTitle>
          {isChinaMap && (
            <button
              onClick={handleBackToWorld}
              className="text-sm text-primary hover:underline flex items-center gap-1"
            >
              ← 返回世界地图
            </button>
          )}
        </CardHeader>
        <CardContent>
          <div className="grid gap-6 lg:grid-cols-4">
            <div className="lg:col-span-3 h-96">
              {mapReady && (
                <ReactECharts
                  ref={chartRef}
                  option={mapChartOption}
                  style={{ height: '100%', width: '100%' }}
                  opts={{ renderer: 'canvas' }}
                  notMerge={true}
                  lazyUpdate={true}
                  onEvents={{ click: handleMapClick }}
                />
              )}
            </div>
            <div className="space-y-2">
              <h4 className="font-medium text-sm">
                {isChinaMap ? '中国地区排名' : 'IP 地区排名'}
              </h4>
              <div className="space-y-2">
                {(data?.ipRegionRank || []).map((item, index) => (
                  <div key={index} className="flex items-center gap-2 text-sm">
                    <span
                      className={cn(
                        "w-5 h-5 flex items-center justify-center rounded text-xs font-medium",
                        index < 3 ? "bg-primary text-primary-foreground" : "bg-muted"
                      )}
                    >
                      {index + 1}
                    </span>
                    <span className="flex-1 truncate">{item.city}</span>
                    <div className="w-16 h-2 bg-muted rounded-full overflow-hidden">
                      <div
                        className="h-full bg-primary rounded-full"
                        style={{ width: `${item.percent}%` }}
                      />
                    </div>
                    <span className="text-muted-foreground w-12 text-right">{item.percent.toFixed(2)}%</span>
                  </div>
                ))}
              </div>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
