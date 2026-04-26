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

function formatRangeLabel(range: DateRange | undefined): string {
  if (!range?.from) return '选择时间范围'
  if (range.to) {
    return `${format(range.from, 'yyyy-MM-dd HH:mm')} ~ ${format(range.to, 'yyyy-MM-dd HH:mm')}`
  }
  return format(range.from, 'yyyy-MM-dd HH:mm')
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

export default function DateRangePicker({
  value,
  onChange,
  presets = defaultPresets,
  className,
  loading = false,
}: DateRangePickerProps) {
  const [open, setOpen] = React.useState(false)
  const [showCustom, setShowCustom] = React.useState(false)

  const activePresetIndex = findPresetIndex(value, presets)
  const isCustom = activePresetIndex === -1 && value?.from !== undefined
  const isDefault = activePresetIndex === -1 && !value?.from

  const handlePresetSelect = React.useCallback(
    (preset: PresetRange) => {
      const range = preset.getDateRange()
      onChange?.(range)
      setShowCustom(false)
      setOpen(false)
    },
    [onChange]
  )

  const handleCustomSelect = React.useCallback(
    (range: DateRange | undefined) => {
      if (range?.from) {
        onChange?.(range)
      }
    },
    [onChange]
  )

  const handleToggleCustom = React.useCallback(() => {
    setShowCustom((prev) => !prev)
  }, [])

  const handleClear = React.useCallback(() => {
    onChange?.({ from: undefined, to: undefined })
    setShowCustom(false)
  }, [onChange])

  return (
    <Popover open={open} onOpenChange={setOpen}>
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
        className="w-auto p-0"
        align="start"
        sideOffset={8}
      >
        {!showCustom ? (
          <div className="flex flex-col md:flex-row">
            <div className="flex flex-row md:flex-col gap-1 p-2 border-b md:border-b-0 md:border-r border-border min-w-[120px]">
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
              <Button
                variant={isCustom ? 'secondary' : 'ghost'}
                size="sm"
                className={cn(
                  'justify-start text-sm px-3 py-1.5 h-auto',
                  isCustom && 'font-medium'
                )}
                onClick={handleToggleCustom}
              >
                自定义
              </Button>
            </div>
            {value?.from && (
              <div className="p-3 flex items-center justify-center min-h-[200px] md:min-h-[300px]">
                <div className="text-center space-y-2">
                  <CalendarIcon className="h-8 w-8 mx-auto text-muted-foreground" />
                  <p className="text-sm text-muted-foreground">
                    已选择: {formatRangeLabel(value)}
                  </p>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={handleClear}
                    className="mt-2"
                  >
                    清除选择
                  </Button>
                </div>
              </div>
            )}
            {!value?.from && (
              <div className="p-3 flex items-center justify-center min-h-[200px] md:min-h-[300px]">
                <div className="text-center space-y-2">
                  <CalendarIcon className="h-8 w-8 mx-auto text-muted-foreground" />
                  <p className="text-sm text-muted-foreground">
                    请选择一个预设时间范围<br />或点击"自定义"选择具体时间
                  </p>
                </div>
              </div>
            )}
          </div>
        ) : (
          <div className="flex flex-col md:flex-row">
            <div className="p-2 border-b md:border-b-0 md:border-r border-border">
              <Button
                variant="ghost"
                size="sm"
                className="w-full justify-start text-sm px-3 py-1.5 h-auto mb-1"
                onClick={handleToggleCustom}
              >
                ← 返回预设
              </Button>
            </div>
            <div className="p-3">
              <Calendar
                mode="range"
                selected={value}
                onSelect={handleCustomSelect}
                locale={zhCN}
                numberOfMonths={1}
                defaultMonth={value?.from}
                className="rounded-md"
              />
              {value?.from && (
                <div className="mt-3 flex items-center justify-between gap-2">
                  <span className="text-xs text-muted-foreground truncate">
                    {formatRangeLabel(value)}
                  </span>
                  <Button
                    variant="outline"
                    size="sm"
                    className="h-7 px-2 text-xs"
                    onClick={() => {
                      onChange?.(value)
                      setShowCustom(false)
                      setOpen(false)
                    }}
                  >
                    确认
                  </Button>
                </div>
              )}
            </div>
          </div>
        )}
      </PopoverContent>
    </Popover>
  )
}
