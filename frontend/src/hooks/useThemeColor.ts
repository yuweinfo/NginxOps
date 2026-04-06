import { useEffect, useState } from 'react'

/**
 * 将 HSL 空格分隔格式转换为标准逗号分隔格式
 * 输入: "0 0% 3.9%" -> 输出: "hsl(0, 0%, 3.9%)"
 */
function formatHsl(value: string): string {
  if (!value) return ''
  // 将空格分隔的值转换为逗号分隔
  const parts = value.split(/\s+/)
  if (parts.length >= 3) {
    return `hsl(${parts[0]}, ${parts[1]}, ${parts[2]})`
  }
  return `hsl(${value})`
}

/**
 * 将 HSL 颜色转换为十六进制格式，确保 ECharts 兼容
 * 输入: "0 0% 3.9%" -> 输出: "#0a0a0a"
 */
function hslToHex(hslValue: string): string {
  if (!hslValue) return ''

  const parts = hslValue.split(/\s+/)
  if (parts.length < 3) return ''

  let h = parseFloat(parts[0])
  let s = parseFloat(parts[1]) / 100
  let l = parseFloat(parts[2]) / 100

  const a = s * Math.min(l, 1 - l)
  const f = (n: number) => {
    const k = (n + h / 30) % 12
    const color = l - a * Math.max(Math.min(k - 3, 9 - k, 1), -1)
    return Math.round(255 * color).toString(16).padStart(2, '0')
  }

  return `#${f(0)}${f(8)}${f(4)}`
}

export function useThemeColor(variableName: string): string {
  const [color, setColor] = useState('')

  useEffect(() => {
    const updateColor = () => {
      const style = getComputedStyle(document.documentElement)
      const value = style.getPropertyValue(variableName).trim()
      // 使用十六进制格式，确保 ECharts 兼容
      setColor(value ? hslToHex(value) : '')
    }

    updateColor()

    // 监听主题变化
    const observer = new MutationObserver((mutations) => {
      mutations.forEach((mutation) => {
        if (mutation.attributeName === 'class') {
          updateColor()
        }
      })
    })

    observer.observe(document.documentElement, {
      attributes: true,
      attributeFilter: ['class'],
    })

    return () => observer.disconnect()
  }, [variableName])

  return color
}

export function useThemeColors() {
  const foreground = useThemeColor('--foreground')
  const background = useThemeColor('--background')
  const card = useThemeColor('--card')
  const muted = useThemeColor('--muted')
  const mutedForeground = useThemeColor('--muted-foreground')
  const border = useThemeColor('--border')
  const accent = useThemeColor('--accent')
  const destructive = useThemeColor('--destructive')

  return {
    foreground,
    background,
    card,
    muted,
    mutedForeground,
    border,
    accent,
    destructive,
  }
}
