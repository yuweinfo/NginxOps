import * as React from 'react'
import { format } from 'date-fns'
import { zhCN } from 'date-fns/locale'
import { CalendarIcon, ChevronDownIcon } from 'lucide-react'
import { DateRange } from 'react-day-picker'

import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { Calendar } from '@/components/ui/calendar'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover'

export interface PresetRange {
  label: string
  getDateRange: () => DateRange
}

export interface DateRangePickerProps {
  value?: DateRange
  onChange?: (range: DateRange) => void
  presets?: PresetRange[]
  className?: string
  loading?: boolean
}

const defaultPresets: PresetRange[] = [
  {
    label: '今天',
    getDateRange: () => {
      const now = new Date()
      const start = new Date(now.getFullYear(), now.getMonth(), now.getDate())
      return { from: start, to: now }
    },
  },
  {
    label: '昨天',
    getDateRange: () => {
      const now = new Date()
      const start = new Date(now.getFullYear(), now.getMonth(), now.getDate() - 1)
      const end = new Date(now.getFullYear(), now.getMonth(), now.getDate(), 23, 59, 59, 999)
      return { from: start, to: end }
    },
  },
  {
    label: '近 24 小时',
    getDateRange: () => {
      const now = new Date()
      const start = new Date(now.getTime() - 24 * 60 * 60 * 1000)
      return { from: start, to: now }
    },
  },
  {
    label: '近 7 天',
    getDateRange: () => {
      const now = new Date()
      const start = new Date(now.getFullYear(), now.getMonth(), now.getDate() - 6)
      const end = new Date(now.getFullYear(), now.getMonth(), now.getDate(), 23, 59, 59, 999)
      return { from: start, to: end }
    },
  },
  {
    label: '近 30 天',
    getDateRange: () => {
      const now = new Date()
      const start = new Date(now.getFullYear(), now.getMonth(), now.getDate() - 29)
      const end = new Date(now.getFullYear(), now.getMonth(), now.getDate(), 23, 59, 59, 999)
      return { from: start, to: end }
    },
  },
  {
    label: '本月',
    getDateRange: () => {
      const now = new Date()
      const start = new Date(now.getFullYear(), now.getMonth(), 1)
      return { from: start, to: now }
    },
  },
  {
    label: '上月',
    getDateRange: () => {
      const now = new Date()
      const start = new Date(now.getFullYear(), now.getMonth() - 1, 1)
      const end = new Date(now.getFullYear(), now.getMonth(), 0, 23, 59, 59, 999)
      return { from: start, to: end }
    },
  },
]

const hours = Array.from({ length: 24 }, (_, i) => i)
const minutes = Array.from({ length: 60 }, (_, i) => i)
const seconds = Array.from({ length: 60 }, (_, i) => i)

function padZero(n: number): string {
  return n.toString().padStart(2, '0')
}

function formatRangeLabel(range: DateRange | undefined): string {
  if (!range?.from) return '选择时间范围'
  if (range.to) {
    return `${format(range.from, 'yyyy-MM-dd HH:mm:ss')} ~ ${format(range.to, 'yyyy-MM-dd HH:mm:ss')}`
  }
  return format(range.from, 'yyyy-MM-dd HH:mm:ss')
}

function isSameRange(a: DateRange | undefined, b: DateRange | undefined): boolean {
  if (!a && !b) return true
  if (!a || !b) return false
  const fromMatch = a.from?.getTime() === b.from?.getTime()
  const toMatch = a.to?.getTime() === b.to?.getTime()
  return fromMatch && toMatch
}

function findPresetIndex(range: DateRange | undefined, presets: PresetRange[]): number {
  if (!range) return -1
  return presets.findIndex((preset) => {
    const presetRange = preset.getDateRange()
    return isSameRange(range, presetRange)
  })
}

interface TimePickerProps {
  hours: number
  minutes: number
  seconds: number
  onHoursChange: (h: number) => void
  onMinutesChange: (m: number) => void
  onSecondsChange: (s: number) => void
  label: string
}

function TimePicker({ hours: h, minutes: m, seconds: s, onHoursChange, onMinutesChange, onSecondsChange, label }: TimePickerProps) {
  return (
    <div className="flex items-center gap-2">
      <span className="text-xs text-muted-foreground whitespace-nowrap">{label}</span>
      <div className="flex items-center gap-1">
        <select
          value={h}
          onChange={(e) => onHoursChange(Number(e.target.value))}
          className="h-8 w-14 rounded-md border border-input bg-background px-1 text-xs text-center focus:outline-none focus:ring-2 focus:ring-ring"
          aria-label={`${label}-小时`}
        >
          {hours.map((hour) => (
            <option key={hour} value={hour}>{padZero(hour)}</option>
          ))}
        </select>
        <span className="text-muted-foreground">:</span>
        <select
          value={m}
          onChange={(e) => onMinutesChange(Number(e.target.value))}
          className="h-8 w-14 rounded-md border border-input bg-background px-1 text-xs text-center focus:outline-none focus:ring-2 focus:ring-ring"
          aria-label={`${label}-分钟`}
        >
          {minutes.map((minute) => (
            <option key={minute} value={minute}>{padZero(minute)}</option>
          ))}
        </select>
        <span className="text-muted-foreground">:</span>
        <select
          value={s}
          onChange={(e) => onSecondsChange(Number(e.target.value))}
          className="h-8 w-14 rounded-md border border-input bg-background px-1 text-xs text-center focus:outline-none focus:ring-2 focus:ring-ring"
          aria-label={`${label}-秒`}
        >
          {seconds.map((second) => (
            <option key={second} value={second}>{padZero(second)}</option>
          ))}
        </select>
      </div>
    </div>
  )
}

export default function DateRangePicker({
  value,
  onChange,
  presets = defaultPresets,
  className,
  loading = false,
}: DateRangePickerProps) {
  const [open, setOpen] = React.useState(false)

  const activePresetIndex = findPresetIndex(value, presets)
  const isCustom = activePresetIndex === -1 && value?.from !== undefined

  const [localRange, setLocalRange] = React.useState<DateRange | undefined>(value)
  const [startTime, setStartTime] = React.useState({
    hours: value?.from?.getHours() ?? 0,
    minutes: value?.from?.getMinutes() ?? 0,
    seconds: value?.from?.getSeconds() ?? 0,
  })
  const [endTime, setEndTime] = React.useState({
    hours: value?.to?.getHours() ?? 23,
    minutes: value?.to?.getMinutes() ?? 59,
    seconds: value?.to?.getSeconds() ?? 59,
  })

  React.useEffect(() => {
    if (open) {
      setLocalRange(value)
      if (value?.from) {
        setStartTime({
          hours: value.from.getHours(),
          minutes: value.from.getMinutes(),
          seconds: value.from.getSeconds(),
        })
      }
      if (value?.to) {
        setEndTime({
          hours: value.to.getHours(),
          minutes: value.to.getMinutes(),
          seconds: value.to.getSeconds(),
        })
      }
    }
  }, [open, value])

  const applyTimeToDate = React.useCallback((date: Date | undefined, time: { hours: number; minutes: number; seconds: number }): Date | undefined => {
    if (!date) return undefined
    const newDate = new Date(date)
    newDate.setHours(time.hours, time.minutes, time.seconds, 0)
    return newDate
  }, [])

  const getCombinedRange = React.useCallback((): DateRange => {
    const from = applyTimeToDate(localRange?.from, startTime)
    const to = applyTimeToDate(localRange?.to, endTime)
    return { from, to }
  }, [localRange?.from, localRange?.to, startTime, endTime, applyTimeToDate])

  const handlePresetSelect = React.useCallback(
    (preset: PresetRange) => {
      const range = preset.getDateRange()
      onChange?.(range)
      setOpen(false)
    },
    [onChange]
  )

  const handleCalendarSelect = React.useCallback(
    (range: DateRange | undefined) => {
      if (range?.from) {
        const newFrom = applyTimeToDate(range.from, startTime)
        const newTo = range.to ? applyTimeToDate(range.to, endTime) : undefined
        setLocalRange({ from: newFrom, to: newTo })
      } else if (range === undefined) {
        setLocalRange({ from: undefined, to: undefined })
      }
    },
    [startTime, endTime, applyTimeToDate]
  )

  const handleConfirm = React.useCallback(() => {
    if (localRange?.from) {
      const range = getCombinedRange()
      onChange?.(range)
      setOpen(false)
    }
  }, [localRange?.from, getCombinedRange, onChange])

  return (
    <Popover open={open} onOpenChange={setOpen} modal={false}>
      <PopoverTrigger asChild>
        <Button
          variant="outline"
          className={cn(
            'w-full justify-start gap-2 text-left font-normal md:w-auto',
            !value?.from && 'text-muted-foreground',
            loading && 'opacity-70',
            className
          )}
          disabled={loading}
          aria-label="选择时间范围"
        >
          <CalendarIcon className="h-4 w-4 shrink-0" />
          <span className="truncate">
            {loading ? '加载中...' : formatRangeLabel(value)}
          </span>
          <ChevronDownIcon className="h-4 w-4 shrink-0 opacity-50" />
        </Button>
      </PopoverTrigger>
      <PopoverContent
        className="w-auto p-0 mx-4"
        align="start"
        sideOffset={8}
        avoidCollisions={false}
      >
        <div className="flex flex-col md:flex-row">
          <div className="flex flex-row md:flex-col gap-1 p-2 border-b md:border-b-0 md:border-r border-border min-w-[120px] max-w-[160px]">
            {presets.map((preset, index) => {
              const isActive = index === activePresetIndex
              return (
                <Button
                  key={preset.label}
                  variant={isActive ? 'secondary' : 'ghost'}
                  size="sm"
                  className={cn(
                    'justify-start text-sm px-3 py-1.5 h-auto whitespace-normal break-all',
                    isActive && 'font-medium'
                  )}
                  onClick={() => handlePresetSelect(preset)}
                >
                  {preset.label}
                </Button>
              )
            })}
          </div>
          <div className="flex-1">
            <Calendar
              mode="range"
              selected={localRange}
              onSelect={handleCalendarSelect}
              locale={zhCN}
              numberOfMonths={2}
              defaultMonth={localRange?.from}
            />
            <div className="px-3 pb-3 space-y-2 border-t border-border pt-3">
              <TimePicker
                hours={startTime.hours}
                minutes={startTime.minutes}
                seconds={startTime.seconds}
                onHoursChange={(h) => setStartTime((prev) => ({ ...prev, hours: h }))}
                onMinutesChange={(m) => setStartTime((prev) => ({ ...prev, minutes: m }))}
                onSecondsChange={(s) => setStartTime((prev) => ({ ...prev, seconds: s }))}
                label="开始"
              />
              <TimePicker
                hours={endTime.hours}
                minutes={endTime.minutes}
                seconds={endTime.seconds}
                onHoursChange={(h) => setEndTime((prev) => ({ ...prev, hours: h }))}
                onMinutesChange={(m) => setEndTime((prev) => ({ ...prev, minutes: m }))}
                onSecondsChange={(s) => setEndTime((prev) => ({ ...prev, seconds: s }))}
                label="结束"
              />
              {localRange?.from && (
                <div className="flex items-center justify-between gap-2 pt-1">
                  <span className="text-xs text-muted-foreground truncate">
                    {formatRangeLabel(getCombinedRange())}
                  </span>
                  <Button
                    variant="default"
                    size="sm"
                    className="h-7 px-3 text-xs"
                    onClick={handleConfirm}
                  >
                    确认
                  </Button>
                </div>
              )}
              {!localRange?.from && (
                <div className="text-center text-xs text-muted-foreground py-2">
                  请在上方日历中选择日期范围
                </div>
              )}
            </div>
          </div>
        </div>
      </PopoverContent>
    </Popover>
  )
}
