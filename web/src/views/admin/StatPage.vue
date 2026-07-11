<script setup lang="ts">
import {defineAsyncComponent, onMounted, onUnmounted, ref, computed} from "vue";
import {
  PhChartBar, PhArrowBendRightDown, PhArrowBendLeftUp, PhCalendar, PhCalendarDot, PhCalendarDots, PhUsers,
  PhUserCirclePlus, PhColumnsPlusRight, PhColumns
} from "@phosphor-icons/vue";
import {formatNumber} from "@/utils";
import {GetStatDashboard} from "@/api";
import CountTo from "@/components/CountTo.vue";

// 异步按需加载图表组件
const BaseChart = defineAsyncComponent(() => import('@/components/BaseChart.vue'));

const data = ref({
  userCount: {count: 0, change: 0},
  userMonthRegisterCount: {count: 0, change: 0},
  apiAppCount: {count: 0, change: 0},
  apiAppMonthCount: {count: 0, change: 0},
  apiMonthCount: {count: 0, change: 0},
  apiAllCount: {count: 0, change: 0},
  apiTodayCount: {count: 0, change: 0},
  apiWeekCount: {count: 0, change: 0},
  array: {
    apiWeekCount: {date: [] as string[], count: [] as number[]},
    apiTodayCount: {name: [] as string[], count: [] as number[]}
  }
});

const getCardData = async () => {
  const res = await GetStatDashboard();
  if (res.data.code !== 200) return;

  const dashboard = res.data.data;
  data.value.apiAppMonthCount = dashboard.apiAppMonthCount;
  data.value.apiAppCount = dashboard.apiAppCount;
  data.value.userMonthRegisterCount = dashboard.userMonthRegisterCount;
  data.value.userCount = dashboard.userCount;
  data.value.apiMonthCount = dashboard.apiMonthCount;
  data.value.apiAllCount = dashboard.apiAllCount;
  data.value.apiTodayCount = dashboard.apiTodayCount;
  data.value.apiWeekCount = dashboard.apiWeekCount;
  data.value.array.apiWeekCount.date = dashboard.apiWeekCountArray.map(item => item.date.slice(5, 10));
  data.value.array.apiWeekCount.count = dashboard.apiWeekCountArray.map(item => item.count);
  data.value.array.apiTodayCount.name = dashboard.apiTodayCountArray.map(item => item.name);
  data.value.array.apiTodayCount.count = dashboard.apiTodayCountArray.map(item => item.count);
};

const chart1Options = computed(() => ({
  tooltip: {
    trigger: 'axis',
    backgroundColor: 'rgba(255, 255, 255, 0.9)',
    padding: [10, 15],
    textStyle: {color: '#333'},
    extraCssText: 'box-shadow: 0 0 10px rgba(0, 0, 0, 0.1); border-radius: 4px;'
  },
  grid: {left: '3%', right: '4%', bottom: '3%', top: '15%', containLabel: true},
  xAxis: {
    type: 'category',
    boundaryGap: false,
    data: data.value.array.apiWeekCount.date,
    axisLine: {show: false},
    axisTick: {show: false},
    axisLabel: {color: '#909399', margin: 15}
  },
  yAxis: {
    type: 'value',
    axisLabel: {color: '#909399'},
    splitLine: {lineStyle: {color: '#ebeef5', type: 'dashed'}}
  },
  series: [
    {
      name: '调用量',
      type: 'line',
      smooth: true,
      symbol: 'circle',
      symbolSize: 8,
      showSymbol: true,
      data: data.value.array.apiWeekCount.count,
      itemStyle: {color: '#409EFF', borderWidth: 2, borderColor: '#fff'},
      lineStyle: {width: 3, color: '#409EFF'},
      areaStyle: {
        color: {
          type: 'linear', x: 0, y: 0, x2: 0, y2: 1,
          colorStops: [
            {offset: 0, color: 'rgba(64, 158, 255, 0.3)'},
            {offset: 1, color: 'rgba(64, 158, 255, 0.05)'}
          ]
        }
      },
      label: {
        show: true,
        position: 'top',
        color: '#1f2d3d',
        fontWeight: 'bold',
        formatter: (params: any) => params.value.toLocaleString()
      }
    }
  ]
}));

const chart2Options = computed(() => ({
  tooltip: {
    trigger: 'axis',
    axisPointer: {type: 'shadow'},
    backgroundColor: 'rgba(255, 255, 255, 0.9)',
    padding: [10, 15],
    textStyle: {color: '#333'},
    extraCssText: 'box-shadow: 0 0 10px rgba(0, 0, 0, 0.1); border-radius: 4px;'
  },
  grid: {left: '3%', right: '4%', bottom: '3%', top: '15%', containLabel: true},
  xAxis: {
    type: 'category',
    data: data.value.array.apiTodayCount.name,
    axisLine: {show: false},
    axisTick: {show: false},
    axisLabel: {color: '#909399', margin: 15, interval: 0, rotate: 15}
  },
  yAxis: {
    type: 'value',
    axisLabel: {color: '#909399'},
    splitLine: {lineStyle: {color: '#ebeef5', type: 'dashed'}}
  },
  series: [
    {
      name: '请求次数',
      type: 'bar',
      barWidth: '35%',
      data: data.value.array.apiTodayCount.count,
      itemStyle: {
        borderRadius: [4, 4, 0, 0],
        color: {
          type: 'linear', x: 0, y: 0, x2: 0, y2: 1,
          colorStops: [
            {offset: 0, color: '#38D677'},
            {offset: 1, color: 'rgba(56, 214, 119, 0.1)'}
          ]
        }
      },
      label: {
        show: true,
        position: 'top',
        color: '#1f2d3d',
        fontWeight: 'bold',
        formatter: (params: any) => params.value.toLocaleString()
      }
    }
  ]
}));

// 定时器变量
let refreshTimer: number | null = null;

onMounted(async () => {
  // 初始加载
  await getCardData();

  // 每 5 秒自动刷新
  refreshTimer = window.setInterval(() => {
    getCardData();
  }, 5000);
});

// 组件卸载时务必清除定时器，防止内存泄漏
onUnmounted(() => {
  if (refreshTimer !== null) {
    clearInterval(refreshTimer);
    refreshTimer = null;
  }
});
</script>

<template>
  <div class="p-2 sm:p-0">
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 md:gap-5 mb-5">
      <div class="card">
        <div class="flex items-center justify-center mr-3.75 text-[#38D677]">
          <PhChartBar size="60" weight="duotone"/>
        </div>
        <div>
          <h3 class="text-[14px] text-[#606266]">累计调用</h3>
          <div class="font-bold text-[24px]">
            <CountTo :value="data.apiAllCount.count" :formatter="formatNumber"/>
          </div>
        </div>
      </div>

      <div class="card">
        <div class="flex items-center justify-center mr-3.75 text-[#38D677]">
          <PhCalendarDot size="60" weight="duotone"/>
        </div>
        <div>
          <h3 class="text-[14px] text-[#606266]">今日调用</h3>
          <div class="font-bold text-[24px]">
            <CountTo :value="data.apiTodayCount.count" :formatter="formatNumber"/>
          </div>
          <div v-if="data.apiTodayCount.change != 0"
               :class="{'down': data.apiTodayCount.change < 0, 'up':data.apiTodayCount.change > 0}"
               class="text-[12px] flex items-center">
            <PhArrowBendRightDown v-if="data.apiTodayCount.change < 0" class="inline"/>
            <PhArrowBendLeftUp v-else class="inline"/>
            较昨日
            <CountTo :value="data.apiTodayCount.change" :formatter="formatNumber"/>
          </div>
        </div>
      </div>

      <div class="card">
        <div class="flex items-center justify-center mr-3.75 text-[#38D677]">
          <PhCalendarDots size="60" weight="duotone"/>
        </div>
        <div>
          <h3 class="text-[14px] text-[#606266]">本周调用</h3>
          <div class="font-bold text-[24px]">
            <CountTo :value="data.apiWeekCount.count" :formatter="formatNumber"/>
          </div>
          <div v-if="data.apiWeekCount.change != 0"
               :class="{'down': data.apiWeekCount.change < 0, 'up':data.apiWeekCount.change > 0}"
               class="text-[12px] flex items-center">
            <PhArrowBendRightDown v-if="data.apiWeekCount.change < 0" class="inline"/>
            <PhArrowBendLeftUp v-else class="inline"/>
            较上周
            <CountTo :value="data.apiWeekCount.change" :formatter="formatNumber"/>
          </div>
        </div>
      </div>
      <div class="card">
        <div class="flex items-center justify-center mr-3.75 text-[#38D677]">
          <PhCalendar size="60" weight="duotone"/>
        </div>
        <div>
          <h3 class="text-[14px] text-[#606266]">本月调用</h3>
          <div class="font-bold text-[24px]">
            <CountTo :value="data.apiMonthCount.count" :formatter="formatNumber"/>
          </div>
          <div v-if="data.apiMonthCount.change != 0"
               :class="{'down': data.apiMonthCount.change < 0, 'up':data.apiMonthCount.change > 0}"
               class="text-[12px] flex items-center">
            <PhArrowBendRightDown v-if="data.apiMonthCount.change < 0" class="inline"/>
            <PhArrowBendLeftUp v-else class="inline"/>
            较上月
            <CountTo :value="data.apiMonthCount.change" :formatter="formatNumber"/>
          </div>
        </div>
      </div>
      <div class="card">
        <div class="flex items-center justify-center mr-3.75 text-[#38D677]">
          <PhUsers size="60" weight="duotone"/>
        </div>
        <div>
          <h3 class="text-[14px] text-[#606266]">用户总数</h3>
          <div class="font-bold text-[24px]">
            <CountTo :value="data.userCount.count" :formatter="formatNumber"/>
          </div>
        </div>
      </div>
      <div class="card">
        <div class="flex items-center justify-center mr-3.75 text-[#38D677]">
          <PhUserCirclePlus size="60" weight="duotone"/>
        </div>
        <div>
          <h3 class="text-[14px] text-[#606266]">本月注册数</h3>
          <div class="font-bold text-[24px]">
            <CountTo :value="data.userMonthRegisterCount.count" :formatter="formatNumber"/>
          </div>
        </div>
      </div>
      <div class="card">
        <div class="flex items-center justify-center mr-3.75 text-[#38D677]">
          <PhColumns size="60" weight="duotone"/>
        </div>
        <div>
          <h3 class="text-[14px] text-[#606266]">API总数量</h3>
          <div class="font-bold text-[24px]">
            <CountTo :value="data.apiAppCount.count" :formatter="formatNumber"/>
          </div>
        </div>
      </div>
      <div class="card">
        <div class="flex items-center justify-center mr-3.75 text-[#38D677]">
          <PhColumnsPlusRight size="60" weight="duotone"/>
        </div>
        <div>
          <h3 class="text-[14px] text-[#606266]">本月API发布数量</h3>
          <div class="font-bold text-[24px]">
            <CountTo :value="data.apiAppMonthCount.count" :formatter="formatNumber"/>
          </div>
        </div>
      </div>
    </div>


    <div class="grid grid-cols-1 lg:grid-cols-2 gap-4 md:gap-5">
      <div class="chart-card overflow-hidden">
        <div class="chart-header">近7天调用统计</div>
        <BaseChart :options="chart1Options" class="w-full h-75"/>
      </div>
      <div class="chart-card overflow-hidden">
        <div class="chart-header">今日热门接口</div>
        <BaseChart :options="chart2Options" class="w-full h-75"/>
      </div>
    </div>
  </div>
</template>

<style scoped lang="scss">
.card {
  background: #fff;
  border-radius: 8px;
  padding: 20px;
  display: flex;
  align-items: center;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.05);
  border: 1px solid #ebeef5;
  height: 120px;
}

.down {
  color: #fa6161;
}

.up {
  color: #39b16c;
}

.chart-card {
  background: #fff;
  border-radius: 8px;
  padding: 20px;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.05);
  border: 1px solid #ebeef5;

  .chart-header {
    font-size: 16px;
    font-weight: bold;
    margin-bottom: 20px;
    color: #303133;
  }
}
</style>
