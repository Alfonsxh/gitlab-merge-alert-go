<template>
  <div ref="chartRef" class="line-chart"></div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch, computed } from 'vue'
import * as echarts from 'echarts/core'
import { LineChart } from 'echarts/charts'
import {
  TitleComponent,
  TooltipComponent,
  GridComponent,
  LegendComponent,
  ToolboxComponent,
  DataZoomComponent
} from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'
import type { EChartsOption } from 'echarts'

// 注册必要的 ECharts 组件
echarts.use([
  LineChart,
  TitleComponent,
  TooltipComponent,
  GridComponent,
  LegendComponent,
  ToolboxComponent,
  DataZoomComponent,
  CanvasRenderer
])

interface Props {
  title?: string
  data: Array<{
    name: string
    data: Array<{
      date: string
      count: number
    }>
  }>
  height?: string
  showDataZoom?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  height: '400px',
  showDataZoom: true
})

const chartRef = ref<HTMLDivElement>()
let chartInstance: echarts.ECharts | null = null

// 计算图表选项
const chartOption = computed<EChartsOption>(() => {
  // 获取所有日期
  const allDates = new Set<string>()
  props.data.forEach(item => {
    item.data.forEach(d => allDates.add(d.date))
  })
  const dates = Array.from(allDates).sort()

  // 构建系列数据
  const series = props.data.map((item) => {
    const dataMap = new Map(item.data.map(d => [d.date, d.count]))
    const seriesData = dates.map(date => dataMap.get(date) || 0)

    return {
      name: item.name,
      type: 'line' as const,
      smooth: true,
      symbol: 'circle',
      symbolSize: 6,
      data: seriesData,
      animationDuration: 300,
      animationEasing: 'cubicOut' as const
    }
  })

  // 配置颜色方案
  const colors = [
    '#409EFF', '#67C23A', '#E6A23C', '#F56C6C', '#909399',
    '#00B7FF', '#FF6B6B', '#4ECDC4', '#45B7D1', '#F9CA24'
  ]

  const option: EChartsOption = {
    color: colors,
    title: props.title ? {
      text: props.title,
      left: 'center',
      textStyle: {
        fontSize: 18,
        fontWeight: 600,
        color: '#303133'
      }
    } : undefined,
    tooltip: {
      trigger: 'axis',
      position: 'top',
      backgroundColor: 'rgba(255, 255, 255, 0.95)',
      borderColor: '#e6e8eb',
      borderWidth: 1,
      textStyle: {
        color: '#303133'
      },
      confine: true, // 限制tooltip在图表区域内
      formatter: (params: any) => {
        if (!Array.isArray(params)) return ''
        
        let html = `<div style="font-weight: 600; margin-bottom: 8px;">${params[0].axisValue}</div>`
        params.forEach((item: any) => {
          html += `
            <div style="display: flex; align-items: center; margin: 4px 0;">
              <span style="display: inline-block; width: 10px; height: 10px; background: ${item.color}; border-radius: 50%; margin-right: 8px;"></span>
              <span style="flex: 1;">${item.seriesName}</span>
              <span style="font-weight: 600; margin-left: 20px;">${item.value}</span>
            </div>
          `
        })
        return html
      }
    },
    legend: {
      type: 'scroll',
      bottom: 10,
      data: props.data.map(item => item.name),
      textStyle: {
        fontSize: 12
      },
      itemWidth: 20,
      itemHeight: 10,
      itemGap: 8,
      width: '90%',
      left: 'center',
      orient: 'horizontal',
      // 设置图例最大高度，超过时显示滚动条
      height: 60,
      formatter: (name: string) => {
        // 限制显示长度，超过则显示省略号
        return name.length > 15 ? name.substring(0, 15) + '...' : name
      },
      tooltip: {
        show: true
      },
      pageButtonItemGap: 5,
      pageButtonGap: 10,
      pageIconSize: 12,
      pageTextStyle: {
        fontSize: 12
      },
      scrollDataIndex: 0,
      animationDurationUpdate: 800
    },
    grid: {
      left: '3%',
      right: '4%',
      bottom: props.showDataZoom ? '25%' : '22%',
      top: props.title ? '15%' : '10%',
      containLabel: true
    },
    xAxis: {
      type: 'category',
      boundaryGap: false,
      data: dates,
      axisLine: {
        lineStyle: {
          color: '#e6e8eb'
        }
      },
      axisLabel: {
        color: '#909399',
        rotate: 45,
        formatter: (value: string) => {
          // 格式化日期显示
          const date = new Date(value)
          return `${date.getMonth() + 1}/${date.getDate()}`
        }
      }
    },
    yAxis: {
      type: 'value',
      minInterval: 1,
      axisLine: {
        show: false
      },
      axisTick: {
        show: false
      },
      axisLabel: {
        color: '#909399'
      },
      splitLine: {
        lineStyle: {
          color: '#f5f7fa',
          type: 'dashed'
        }
      }
    },
    dataZoom: props.showDataZoom ? [
      {
        type: 'inside',
        start: 0,
        end: 100
      },
      {
        type: 'slider',
        start: 0,
        end: 100,
        height: 20,
        bottom: 40,
        borderColor: '#e6e8eb',
        textStyle: {
          color: '#909399'
        }
      }
    ] : undefined,
    series
  }

  return option
})

// 初始化图表
const initChart = () => {
  if (!chartRef.value) return
  
  // 使用 echarts 的 init 方法
  chartInstance = echarts.init(chartRef.value)
  chartInstance?.setOption(chartOption.value)
}

// 响应式调整
const handleResize = () => {
  chartInstance?.resize()
}

// 监听数据变化
watch(() => props.data, () => {
  if (chartInstance) {
    chartInstance?.setOption(chartOption.value)
  }
}, { deep: true })

onMounted(() => {
  // 延迟初始化以确保 DOM 已渲染
  setTimeout(() => {
    initChart()
    window.addEventListener('resize', handleResize)
  }, 100)
})

onUnmounted(() => {
  window.removeEventListener('resize', handleResize)
  chartInstance?.dispose()
})
</script>

<style scoped lang="less">
.line-chart {
  width: 100%;
  height: v-bind(height);
  min-height: 300px;
}
</style>