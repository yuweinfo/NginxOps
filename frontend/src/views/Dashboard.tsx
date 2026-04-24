import { useState, useEffect, useMemo, useRef } from 'react'
import {
  Activity,
  BarChart3,
  Users,
  Server,
  Loader2,
  AlertCircle,
  Moon,
  Sun,
  Globe,
  Link,
  FileText,
  FileType,
  CheckCircle,
  Monitor,
  Smartphone,
  Laptop,
  Terminal,
  Network,
} from 'lucide-react'
import ReactECharts from 'echarts-for-react'
import * as echarts from 'echarts'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Switch } from '@/components/ui/switch'
import { statsApi, DashboardData } from '@/api/stats'
import { useThemeColors } from '@/hooks/useThemeColor'
import { cn } from '@/lib/utils'
// @ts-ignore
import worldMapGeojson from 'world-map-geojson'

const emptyData: DashboardData = {
  qps: 0,
  bandwidth: 0,
  pvToday: 0,
  activeSites: 0,
  hourlyTrend: [],
  statusDistribution: { '2xx': 0, '3xx': 0, '4xx': 0, '5xx': 0 },
  ipLocations: [],
  ipRegionRank: [],
  ipTopRank: [],
  hostRank: [],
  refererRank: [],
  pathRank: [],
  resourceTypeRank: [],
  statusRank: [],
  browserRank: [],
  deviceTypeRank: [],
  osRank: [],
  userAgentRank: [],
}

interface RankingItem {
  name: string
  count: number
  percent: number
}

interface RankingCardProps {
  title: string
  icon: React.ReactNode
  data: RankingItem[]
  colors: ReturnType<typeof useThemeColors>
}

function RankingCard({ title, icon, data, colors }: RankingCardProps) {
  const fg = colors.foreground || '#0a0a0a'
  const bg = colors.background || '#fafafa'
  const muted = colors.muted || '#e5e5e5'

  const displayData = data.slice(0, 10)

  return (
    <Card>
      <CardHeader className="pb-2">
        <CardTitle className="text-sm font-medium flex items-center gap-2">
          {icon}
          {title}
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="space-y-2 max-h-[320px] overflow-y-auto custom-scrollbar">
          {displayData.map((item, index) => (
            <div key={index} className="flex items-center gap-2 text-xs">
              <span
                className={cn(
                  'w-4 h-4 flex items-center justify-center rounded text-[10px] font-medium flex-shrink-0',
                  index < 3 ? 'bg-primary text-primary-foreground' : 'bg-muted text-muted-foreground'
                )}
              >
                {index + 1}
              </span>
              <span className="flex-1 truncate text-muted-foreground" title={item.name}>
                {item.name}
              </span>
              <div className="w-20 h-1.5 bg-muted rounded-full overflow-hidden flex-shrink-0">
                <div
                  className="h-full rounded-full transition-all"
                  style={{
                    width: `${Math.min(item.percent, 100)}%`,
                    backgroundColor: index === 0 ? fg : muted,
                  }}
                />
              </div>
              <span className="text-muted-foreground w-14 text-right flex-shrink-0">
                {item.count.toLocaleString()}
              </span>
              <span className="text-muted-foreground w-10 text-right flex-shrink-0">
                {item.percent.toFixed(1)}%
              </span>
            </div>
          ))}
          {data.length === 0 && (
            <div className="text-center text-muted-foreground py-4 text-xs">暂无数据</div>
          )}
        </div>
      </CardContent>
    </Card>
  )
}

export default function Dashboard() {
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [data, setData] = useState<DashboardData>(emptyData)
  const [mapReady, setMapReady] = useState(false)
  const [darkMode, setDarkMode] = useState(() => {
    if (typeof document !== 'undefined') {
      return document.documentElement.classList.contains('dark')
    }
    return false
  })
  const chartRef = useRef<any>(null)
  const colors = useThemeColors()

  useEffect(() => {
    if (darkMode) {
      document.documentElement.classList.add('dark')
    } else {
      document.documentElement.classList.remove('dark')
    }
  }, [darkMode])

  useEffect(() => {
    echarts.registerMap('world', worldMapGeojson)
    setMapReady(true)
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
            ipTopRank: res.data.ipTopRank || [],
            hostRank: res.data.hostRank || [],
            refererRank: res.data.refererRank || [],
            pathRank: res.data.pathRank || [],
            resourceTypeRank: res.data.resourceTypeRank || [],
            statusRank: res.data.statusRank || [],
            browserRank: res.data.browserRank || [],
            deviceTypeRank: res.data.deviceTypeRank || [],
            osRank: res.data.osRank || [],
            userAgentRank: res.data.userAgentRank || [],
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
          setData((prev) => ({
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
              itemStyle: { color: fg },
            },
            {
              value: status['3xx'] || 0,
              name: '3xx',
              itemStyle: { color: muted },
            },
            {
              value: status['4xx'],
              name: '4xx',
              itemStyle: { color: destructive },
            },
            {
              value: status['5xx'],
              name: '5xx',
              itemStyle: { color: destructive, opacity: 0.6 },
            },
          ],
        },
      ],
    }
  }, [data?.statusDistribution, colors])

  const ipTopRankConverted = useMemo(() => {
    const total = (data?.ipTopRank || []).reduce((sum, item) => sum + item.requests, 0)
    return (data?.ipTopRank || []).map((item) => ({
      name: `${item.ip}${item.region && item.region !== 'Unknown' ? ` (${item.region})` : ''}`,
      count: item.requests,
      percent: total > 0 ? (item.requests / total) * 100 : 0,
    }))
  }, [data?.ipTopRank])

  const countryNameMap: Record<string, string> = {
    China: '中国',
    'United States': '美国',
    Japan: '日本',
    Singapore: '新加坡',
    Germany: '德国',
    'United Kingdom': '英国',
    Australia: '澳大利亚',
    France: '法国',
    Russia: '俄罗斯',
    India: '印度',
    Canada: '加拿大',
    Brazil: '巴西',
    'South Korea': '韩国',
    Indonesia: '印度尼西亚',
    Vietnam: '越南',
    Thailand: '泰国',
    Philippines: '菲律宾',
    Malaysia: '马来西亚',
    Mexico: '墨西哥',
    Turkey: '土耳其',
    Italy: '意大利',
    Spain: '西班牙',
    Netherlands: '荷兰',
    Poland: '波兰',
    Sweden: '瑞典',
    Switzerland: '瑞士',
    Belgium: '比利时',
    Austria: '奥地利',
    Portugal: '葡萄牙',
    'Czech Republic': '捷克',
    Greece: '希腊',
    Ukraine: '乌克兰',
    Romania: '罗马尼亚',
    Hungary: '匈牙利',
    Israel: '以色列',
    'United Arab Emirates': '阿联酋',
    'New Zealand': '新西兰',
    Ireland: '爱尔兰',
    Denmark: '丹麦',
    Norway: '挪威',
    Finland: '芬兰',
    Kazakhstan: '哈萨克斯坦',
    Pakistan: '巴基斯坦',
    Bangladesh: '孟加拉国',
    Egypt: '埃及',
    'South Africa': '南非',
    Nigeria: '尼日利亚',
    Kenya: '肯尼亚',
    Argentina: '阿根廷',
    Chile: '智利',
    Colombia: '哥伦比亚',
    Peru: '秘鲁',
    Venezuela: '委内瑞拉',
    Ecuador: '厄瓜多尔',
    Uruguay: '乌拉圭',
    Paraguay: '巴拉圭',
    Bolivia: '玻利维亚',
    'Saudi Arabia': '沙特阿拉伯',
    Iran: '伊朗',
    Iraq: '伊拉克',
    Syria: '叙利亚',
    Jordan: '约旦',
    Lebanon: '黎巴嫩',
    Kuwait: '科威特',
    Qatar: '卡塔尔',
    Bahrain: '巴林',
    Oman: '阿曼',
    Yemen: '也门',
    Azerbaijan: '阿塞拜疆',
    Armenia: '亚美尼亚',
    Georgia: '格鲁吉亚',
    Mongolia: '蒙古',
    'North Korea': '朝鲜',
    Nepal: '尼泊尔',
    'Sri Lanka': '斯里兰卡',
    Myanmar: '缅甸',
    Cambodia: '柬埔寨',
    Laos: '老挝',
    Brunei: '文莱',
    'Timor-Leste': '东帝汶',
    Maldives: '马尔代夫',
    Bhutan: '不丹',
    Iceland: '冰岛',
    Luxembourg: '卢森堡',
    Malta: '马耳他',
    Cyprus: '塞浦路斯',
    Estonia: '爱沙尼亚',
    Latvia: '拉脱维亚',
    Lithuania: '立陶宛',
    Slovenia: '斯洛文尼亚',
    Croatia: '克罗地亚',
    'Bosnia and Herzegovina': '波黑',
    Serbia: '塞尔维亚',
    Montenegro: '黑山',
    'North Macedonia': '北马其顿',
    Albania: '阿尔巴尼亚',
    Bulgaria: '保加利亚',
    Slovakia: '斯洛伐克',
    Belarus: '白俄罗斯',
    Moldova: '摩尔多瓦',
    Tajikistan: '塔吉克斯坦',
    Turkmenistan: '土库曼斯坦',
    Uzbekistan: '乌兹别克斯坦',
    Kyrgyzstan: '吉尔吉斯斯坦',
    Afghanistan: '阿富汗',
    Morocco: '摩洛哥',
    Algeria: '阿尔及利亚',
    Tunisia: '突尼斯',
    Libya: '利比亚',
    Sudan: '苏丹',
    Ethiopia: '埃塞俄比亚',
    Somalia: '索马里',
    Tanzania: '坦桑尼亚',
    Uganda: '乌干达',
    Rwanda: '卢旺达',
    Burundi: '布隆迪',
    Zambia: '赞比亚',
    Zimbabwe: '津巴布韦',
    Botswana: '博茨瓦纳',
    Namibia: '纳米比亚',
    Angola: '安哥拉',
    Mozambique: '莫桑比克',
    Madagascar: '马达加斯加',
    Mauritius: '毛里求斯',
    Seychelles: '塞舌尔',
    Comoros: '科摩罗',
    'Cape Verde': '佛得角',
    'Sao Tome and Principe': '圣多美和普林西比',
    Gabon: '加蓬',
    'Equatorial Guinea': '赤道几内亚',
    Cameroon: '喀麦隆',
    'Central African Republic': '中非',
    Chad: '乍得',
    Congo: '刚果（布）',
    'Democratic Republic of the Congo': '刚果（金）',
    'Republic of the Congo': '刚果（布）',
    Djibouti: '吉布提',
    Eritrea: '厄立特里亚',
    'The Gambia': '冈比亚',
    Ghana: '加纳',
    Guinea: '几内亚',
    'Guinea-Bissau': '几内亚比绍',
    'Ivory Coast': '科特迪瓦',
    Liberia: '利比里亚',
    Mali: '马里',
    Mauritania: '毛里塔尼亚',
    Niger: '尼日尔',
    Senegal: '塞内加尔',
    'Sierra Leone': '塞拉利昂',
    Togo: '多哥',
    'Burkina Faso': '布基纳法索',
    Benin: '贝宁',
    Lesotho: '莱索托',
    Eswatini: '斯威士兰',
    Malawi: '马拉维',
    'Western Sahara': '西撒哈拉',
    'South Sudan': '南苏丹',
    'Papua New Guinea': '巴布亚新几内亚',
    Fiji: '斐济',
    'Solomon Islands': '所罗门群岛',
    Vanuatu: '瓦努阿图',
    Samoa: '萨摩亚',
    Tonga: '汤加',
    Kiribati: '基里巴斯',
    Palau: '帕劳',
    Nauru: '瑙鲁',
    Tuvalu: '图瓦卢',
    'Marshall Islands': '马绍尔群岛',
    Micronesia: '密克罗尼西亚',
    Cuba: '古巴',
    'Dominican Republic': '多米尼加',
    Haiti: '海地',
    Jamaica: '牙买加',
    'Trinidad and Tobago': '特立尼达和多巴哥',
    Barbados: '巴巴多斯',
    'Saint Lucia': '圣卢西亚',
    Grenada: '格林纳达',
    'Saint Vincent and the Grenadines': '圣文森特和格林纳丁斯',
    'Antigua and Barbuda': '安提瓜和巴布达',
    Dominica: '多米尼克',
    'Saint Kitts and Nevis': '圣基茨和尼维斯',
    Bahamas: '巴哈马',
    Belize: '伯利兹',
    'Costa Rica': '哥斯达黎加',
    'El Salvador': '萨尔瓦多',
    Guatemala: '危地马拉',
    Honduras: '洪都拉斯',
    Nicaragua: '尼加拉瓜',
    Panama: '巴拿马',
    Guyana: '圭亚那',
    Suriname: '苏里南',
    'French Guiana': '法属圭亚那',
    Greenland: '格陵兰',
  }

  const mapChartOption = useMemo(() => {
    const ipLocations = data?.ipLocations || []
    const fg = colors.foreground || '#0a0a0a'
    const border = colors.border || '#e5e5e5'
    const accent = colors.accent || '#f5f5f5'
    const muted = colors.muted || '#e5e5e5'

    const mapData = ipLocations
      .filter((loc) => loc.name && Array.isArray(loc.value) && loc.value.length > 2)
      .map((loc) => ({
        name: loc.name,
        value: Number(loc.value[2]) || 0,
      }))

    const maxValue = Math.max(...mapData.map((d) => d.value), 1000)

    const mapColors = darkMode
      ? [colors.border || '#262626', colors.mutedForeground || '#a3a3a3', colors.foreground || '#e5e5e5']
      : [colors.muted || '#f5f5f5', colors.mutedForeground || '#a3a3a3', colors.foreground || '#171717']

    return {
      tooltip: {
        trigger: 'item',
        backgroundColor: colors.card || '#fff',
        borderColor: border,
        borderWidth: 1,
        borderRadius: 8,
        textStyle: { color: fg, fontSize: 13 },
        formatter: (params: any) => {
          const displayName = countryNameMap[params.name] || params.name
          const val = Number(params.value)
          if (!isNaN(val) && params.value != null) {
            return `${displayName}<br/>访问量: ${val.toLocaleString()}`
          }
          return displayName
        },
      },
      visualMap: {
        show: true,
        min: 0,
        max: maxValue,
        calculable: true,
        inRange: {
          color: mapColors,
        },
        text: ['高', '低'],
        textStyle: { color: fg },
        bottom: 20,
        left: 20,
      },
      series: [
        {
          name: '访客分布',
          type: 'map',
          map: 'world',
          roam: true,
          zoom: 1.2,
          center: [30, 30],
          scaleLimit: { min: 1, max: 20 },
          label: {
            show: false,
          },
          emphasis: {
            label: {
              show: true,
              color: fg,
              fontSize: 11,
              formatter: (params: any) => countryNameMap[params.name] || params.name,
            },
            itemStyle: {
              areaColor: accent,
              shadowBlur: 10,
              shadowColor: 'rgba(0, 0, 0, 0.2)',
            },
          },
          select: {
            itemStyle: {
              areaColor: accent,
            },
          },
          itemStyle: {
            areaColor: muted,
            borderColor: border,
            borderWidth: 0.5,
          },
          data: mapData,
        },
      ],
    }
  }, [data?.ipLocations, colors, darkMode])

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
      {/* Header with Dark Mode Toggle */}
      <div className="flex items-center justify-between">
        <h2 className="text-lg font-semibold">仪表盘</h2>
        <div className="flex items-center gap-2">
          <Sun className="h-4 w-4 text-muted-foreground" />
          <Switch checked={darkMode} onCheckedChange={setDarkMode} />
          <Moon className="h-4 w-4 text-muted-foreground" />
        </div>
      </div>

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
        <CardHeader>
          <CardTitle className="text-base font-medium">访客分布</CardTitle>
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
                />
              )}
            </div>
            <div className="space-y-2">
              <h4 className="font-medium text-sm">国家排名</h4>
              <div className="space-y-2">
                {(data?.ipRegionRank || []).map((item, index) => (
                  <div key={index} className="flex items-center gap-2 text-sm">
                    <span
                      className={cn(
                        'w-5 h-5 flex items-center justify-center rounded text-xs font-medium',
                        index < 3 ? 'bg-primary text-primary-foreground' : 'bg-muted'
                      )}
                    >
                      {index + 1}
                    </span>
                    <span className="flex-1 truncate">
                      {countryNameMap[item.country]
                        ? `${countryNameMap[item.country]}(${item.country})`
                        : item.country}
                    </span>
                    <div className="w-16 h-2 bg-muted rounded-full overflow-hidden">
                      <div
                        className="h-full bg-primary rounded-full"
                        style={{ width: `${item.percent}%` }}
                      />
                    </div>
                    <span className="text-muted-foreground w-12 text-right">
                      {item.percent.toFixed(2)}%
                    </span>
                  </div>
                ))}
              </div>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Rankings Grid */}
      <div className="grid gap-6 md:grid-cols-2">
        <RankingCard
          title="Host 排行"
          icon={<Globe className="h-4 w-4" />}
          data={data?.hostRank || []}
          colors={colors}
        />
        <RankingCard
          title="IP 访问排行"
          icon={<Network className="h-4 w-4" />}
          data={ipTopRankConverted}
          colors={colors}
        />
        <RankingCard
          title="Referer 排行"
          icon={<Link className="h-4 w-4" />}
          data={data?.refererRank || []}
          colors={colors}
        />
        <RankingCard
          title="URL Path 排行"
          icon={<FileText className="h-4 w-4" />}
          data={data?.pathRank || []}
          colors={colors}
        />
        <RankingCard
          title="资源类型排行"
          icon={<FileType className="h-4 w-4" />}
          data={data?.resourceTypeRank || []}
          colors={colors}
        />
        <RankingCard
          title="状态码排行"
          icon={<CheckCircle className="h-4 w-4" />}
          data={data?.statusRank || []}
          colors={colors}
        />
        <RankingCard
          title="客户端浏览器排行"
          icon={<Monitor className="h-4 w-4" />}
          data={data?.browserRank || []}
          colors={colors}
        />
        <RankingCard
          title="客户端设备类型排行"
          icon={<Smartphone className="h-4 w-4" />}
          data={data?.deviceTypeRank || []}
          colors={colors}
        />
        <RankingCard
          title="客户端操作系统排行"
          icon={<Laptop className="h-4 w-4" />}
          data={data?.osRank || []}
          colors={colors}
        />
        <RankingCard
          title="User-Agent 排行"
          icon={<Terminal className="h-4 w-4" />}
          data={data?.userAgentRank || []}
          colors={colors}
        />
      </div>
    </div>
  )
}
