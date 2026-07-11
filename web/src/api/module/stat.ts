import request from '@/utils/module/request'
import type {StatDashboardResponse} from '@/types'

const apiUrl = '/api/v1/stat'

// 一次获取统计页面所需的全部聚合数据
export const GetStatDashboard = () => request.get<StatDashboardResponse>(apiUrl + '/dashboard')
