<script lang="ts" setup>
import {Plus, Delete, Edit, Search, Refresh} from "@element-plus/icons-vue";
import {
  GetUserApiKeyList,
  GetApiKeyInfo,
  UpdateApiKey,
  UserApiKeyListSearch,
  DeleteApiKey,
  CreateApiKey,
  ResetApiKey
} from "@/api";
import {onMounted, reactive, ref, watch} from "vue";
import type {Pagination, SelectFormApiKey, APIKey} from "@/types";
import {dayjs, type FormRules} from "element-plus";
import {HelloAPIConfig} from "@/config/config.ts";
import {useUserStore} from "@/stores";
import {copyText, formatDuration, formatNativeDate, isMobile} from "@/utils";

// --- Loading 状态 ---
const loading = ref(false); // 表格加载状态
const drawerLoading = ref(false); // 抽屉详情加载状态
const submitLoading = ref(false); // 表单提交加载状态

// 分页
const paging = ref<Pagination>({
  page: 1,
  pageSize: HelloAPIConfig.website.admin.pageSize,
  total: 0,
})

// 统一数据获取
const fetchData = async () => {
  loading.value = true; // 开启表格 Loading
  const isSearching = searchForm.value.keywords !== '' || searchForm.value.type !== -1 || searchForm.value.status !== -1;
  const apiCall = isSearching
      ? UserApiKeyListSearch(searchForm.value, userStore.user.id, paging.value.page, paging.value.pageSize)
      : GetUserApiKeyList(userStore.user.id, paging.value.page, paging.value.pageSize);

  try {
    const res = await apiCall;
    if (res.data.code === 200) {
      const {list, total} = res.data.data;
      paging.value.total = total;
      tableData.value = processTableData(list);
    }
  } catch (error) {
    console.error("Failed to fetch data:", error);
  } finally {
    loading.value = false; // 关闭表格 Loading
  }
}

// 处理表格数据
const processTableData = (rawData: any[]): APIKey[] => {
  const now = Date.now();
  return rawData.map((item: any) => ({
    ...item,
    isExpired: (Date.parse(item.expired) < now && item.type === 0) || (item.count === 0 && item.type === 1),
    remainingTime: formatDuration(now, item.expired),
    createdText: formatNativeDate(item.created)
  }));
}


// 分页改变时触发
const handlePageChange = () => {
  fetchData()
}

const userStore = useUserStore();
const tableData = ref<APIKey[]>();

// 搜索筛选表单
const searchForm = ref<SelectFormApiKey>({
  keywords: '',
  type: -1, // -1 代表不限，0 免费，1 收费
  status: -1, // -1 代表不限，0 正常，1 异常，2 维护
})

// 搜索
const handleSearch = () => {
  paging.value.page = 1;
  fetchData();
}

// 重置搜索
const resetSearch = () => {
  searchForm.value = {
    keywords: '',
    type: -1, // -1 代表不限，0 免费，1 收费
    status: -1, // -1 代表不限，0 正常，1 异常，2 维护
  };
  paging.value.page = 1
  fetchData()
}


// 抽屉
const showDrawer = ref(false)
const drawerTitle = ref()
const resizable = ref(true) // 是否可调整大小
const drawerSize = ref(isMobile ? '100%' : '30%') // 抽屉大小
const nowRow = ref()

// 日期表单
const dateArray = ref(['', ''])
watch(dateArray, (newVal, _) => {
  if (newVal && newVal[0] && newVal[1]) {
    if (showDrawer.value) {
      formData.value.started = dayjs(newVal[0]).format('YYYY-MM-DDTHH:mm:ss')
      formData.value.expired = dayjs(newVal[1]).format('YYYY-MM-DDTHH:mm:ss')
    }
  }
})

// 提交表单
const formRef = ref()
const formData = ref<APIKey>({
  api_id: null,
  key: '',
  type: 0,
  created: '',
  updated: '',
  started: '',
  expired: '',
  count: 0,
  desc: ''
});

// 重置formData
const resetFormData = () => {
  formData.value = {
    api_id: null,
    key: '',
    type: 0,
    created: '',
    updated: '',
    started: '',
    expired: '',
    count: 0,
    desc: ''
  }
  dateArray.value = ['', '']
}

// 打开抽屉
const openDrawer = (type: string, row?: object) => {
  resetFormData()
  if (type === 'edit') {
    nowRow.value = row
  }
  drawerTitle.value = type === 'create' ? '创建 APIKey' : '编辑 APIKey'
  showDrawer.value = true
}

// 创建APIKey
const submitFormCreate = async () => {
  // 校验表单
  const isValid = await formRef.value?.validate().catch(() => false);
  if (isValid) {
    submitLoading.value = true; // 开启提交 Loading
    try {
      // 提交表单
      if (formData.value.type == 0) {
        formData.value.count = 0
      } else {
        formData.value.started = ''
        formData.value.expired = ''
      }
      const res = await CreateApiKey(formData.value)
      if (res.data.code == 200) {
        await fetchData()
        ElMessage.success('创建成功')
        showDrawer.value = false
      }
    } finally {
      submitLoading.value = false; // 关闭提交 Loading
    }
  }
}

// 更新APIKey
const submitFormUpdate = async () => {
  // 校验表单
  const isValid = await formRef.value?.validate().catch(() => false);
  if (isValid) {
    submitLoading.value = true; // 开启提交 Loading
    try {
      // 提交表单
      if (formData.value.type == 0) {
        formData.value.count = 0
      } else {
        formData.value.started = ''
        formData.value.expired = ''
      }
      const res = await UpdateApiKey(nowRow.value.key, formData.value)
      if (res.data.code == 200) {
        await fetchData()
        ElMessage.success('编辑成功')
        showDrawer.value = false
      }
    } finally {
      submitLoading.value = false; // 关闭提交 Loading
    }
  }
}

// 表单校验
const rules: FormRules = reactive({
  type: [
    {required: true, message: '请选择接口类型', trigger: 'change'}
  ],
  api_id: [
    {required: true, message: '请输入接口 Api ID', trigger: 'blur'}
  ],
  // 备注长度校验
  desc: [
    {max: 255, message: '备注长度不得超过 255 个字符', trigger: 'blur'}
  ],
  // 自定义校验逻辑
  validateDynamicField: [
    {
      validator: (_, __, callback) => {
        if (formData.value.type == 0) {
          if (!dateArray.value || dateArray.value[0] === '' || dateArray.value[1] === '' || dateArray.value[0] == null || dateArray.value[1] == null) {
            return callback(new Error('请选择有效期限'));
          }
        }
        if (formData.value.type == 1) {
          if (formData.value.count === undefined || formData.value.count === null || formData.value.count < 0) {
            return callback(new Error('请输入可用次数'));
          }
        }
        callback();
      },
      trigger: ['blur', 'change']
    }
  ]
})

const onDrawerLoadOver = async () => {
  if (nowRow.value?.api_id) {
    drawerLoading.value = true; // 开启详情 Loading
    try {
      const res = await GetApiKeyInfo(nowRow.value.key)
      if (res.data.code == 200) {
        formData.value = res.data.data
        dateArray.value = [res.data.data.started, res.data.data.expired]
      }
    } finally {
      drawerLoading.value = false; // 关闭详情 Loading
    }
  }
}

// 重置APIKey
const resetAPIKey = async (key: string) => {
  ElMessageBox.confirm('你确定要重置此APIKey吗？之前的会立即失效，操作不可逆。', '提示',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning',
        dangerouslyUseHTMLString: true,
      }
  ).then(async () => {
    const res = await ResetApiKey(key)
    if (res.data.code == 200) {
      await fetchData()
      ElMessage.success('重置成功')
    }
  })
}

// 时间范围快捷选择
const shortcuts = [
  {
    text: '设置7天',
    value: () => {
      const end = new Date()
      const start = new Date()
      end.setTime(start.getTime() + 3600 * 1000 * 24 * 7)
      return [start, end]
    },
  },
  {
    text: '设置30天',
    value: () => {
      const end = new Date()
      const start = new Date()
      end.setTime(start.getTime() + 3600 * 1000 * 24 * 30)
      return [start, end]
    },
  },
  {
    text: '设置180天',
    value: () => {
      const end = new Date()
      const start = new Date()
      end.setTime(start.getTime() + 3600 * 1000 * 24 * 180)
      return [start, end]
    },
  },
  {
    text: '设置1年',
    value: () => {
      const end = new Date()
      const start = new Date()
      end.setTime(start.getTime() + 3600 * 1000 * 24 * 365)
      return [start, end]
    },
  },
]

const onDrawerClose = () => {
  resetFormData()
  nowRow.value = undefined;
}

const handleDelete = (key: string) => {
  ElMessageBox.confirm('确定要删除吗？删除后将无法恢复！', '提示',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning',
        dangerouslyUseHTMLString: true,
      }
  ).then(async () => {
    const res = await DeleteApiKey(key)
    if (res.data.code == 200) {
      // 如果当前页只有一条数据，并且不是第一页，则删除后请求前一页
      if (tableData.value && tableData.value.length === 1 && paging.value.page > 1) {
        paging.value.page--;
      }
      await fetchData()
      ElMessage.success('删除成功')
    }
  }).catch(() => {
  })
}


onMounted(() => {
  fetchData();
})
</script>
<template>
  <admin-main-body title="APIKey 管理">
    <template #header>
      <el-button :icon="Plus" type="primary" @click="openDrawer('create')">创建 APIKey</el-button>

    </template>
    <template #default>
      <div class="pl-2">
        <el-form :inline="true" :model="searchForm" class="searchForm">
          <el-form-item>
            <el-input
                v-model="searchForm.keywords"
                placeholder="Api ID / 关键字"
                @keyup.enter="handleSearch"
            />
          </el-form-item>

          <el-form-item label="类型">
            <el-select v-model="searchForm.type" placeholder="请选择">
              <el-option :value="-1" label="不限"/>
              <el-option :value="0" label="期限">
                <el-tag size="small" type="primary">期限</el-tag>
              </el-option>
              <el-option :value="1" label="计次">
                <el-tag size="small" type="warning">计次</el-tag>
              </el-option>
            </el-select>
          </el-form-item>

          <el-form-item label="状态">
            <el-select v-model="searchForm.status" placeholder="请选择">
              <el-option :value="-1" label="不限"/>
              <el-option :value="0" label="有效">
                <el-tag size="small" type="success">有效</el-tag>
              </el-option>

              <el-option :value="1" label="失效">
                <el-tag size="small" type="danger">失效</el-tag>
              </el-option>
            </el-select>
          </el-form-item>

          <el-form-item>
            <el-button :icon="Search" plain type="primary" @click="handleSearch">查询</el-button>
            <el-button :icon="Refresh" @click="resetSearch">重置</el-button>
          </el-form-item>
          <el-form-item>
            <el-tag size="large" type="info">当前条件共 {{ paging.total }} 条数据</el-tag>
          </el-form-item>
        </el-form>
      </div>

      <el-table :data="tableData" class="w-full" v-loading="loading">
        <el-table-column label="Api ID">
          <template #default="{row}">
            <el-link
                :href="'/info/' + row.api_id"
                target="_blank"
                type="primary"
                underline="never"
            >
            {{ row.api_id }}
            </el-link>
          </template>
        </el-table-column>
        <el-table-column label="APIKey" minWidth="160">
          <template #default="{row}">
            <el-link
                target="_blank"
                type="info"
                @click="copyText(row.key)"
            >
              {{ row.key }}
            </el-link>
          </template>
        </el-table-column>
        <el-table-column label="类型">
          <template #default="{ row }">
            <el-tag :type="row.type === 1 ? 'warning' : 'primary'">
              {{ row.type === 1 ? '计次' : '期限' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="创建日期" minWidth="150">
          <template #default="{ row }">
            {{ row.createdText }}
          </template>
        </el-table-column>
        <el-table-column label="有效期 / 调用次数" minWidth="150">
          <template #default="{ row }">
            <el-tag type="primary">
              <span v-if="row.started">{{ row.remainingTime }}</span>
              <span v-else>-</span>
            </el-tag>&nbsp;
            <el-tag type="warning">
              <span v-if="row.type==1">{{ row.count }}</span>
              <span v-else>-</span>
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态">
          <template #default="{ row }">
            <el-tag :type="row.isExpired ? 'danger' : 'success'">
              {{ row.isExpired ? '失效' : '有效' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="180">
          <template #default="{ row }">
            <el-tooltip content="重置APIKey" placement="left">
              <el-button
                  :icon="Refresh"
                  circle
                  plain
                  type="warning"
                  @click="resetAPIKey(row.key)"
              />
            </el-tooltip>
            <el-button
                :icon="Edit"
                circle
                plain
                type="primary"
                @click="openDrawer('edit', row)"
            />
            <el-button
                :icon="Delete"
                circle
                plain
                type="danger"
                @click="handleDelete(row.key)"
            />
          </template>
        </el-table-column>
        <template #empty>
          <el-empty description="空空的什么也没有＞﹏＜" style="user-select: none"/>
        </template>
      </el-table>
      <div class="flex items-center justify-center mt-6">
        <el-pagination
            v-model:current-page="paging.page"
            :page-size="paging.pageSize"
            :total="paging.total"
            layout="prev, pager, next"
            @current-change="handlePageChange"
        />
      </div>

      <el-drawer
          v-model="showDrawer"
          :destroy-on-close="true"
          :resizable="resizable"
          :size="drawerSize"
          :title="drawerTitle"
          class="rounded-tl-[20px] rounded-bl-[20px] m-auto custom-btt-drawer"
          direction="rtl"
          @closed="onDrawerClose"
          @open="onDrawerLoadOver"
          :close-on-press-escape="false"
      >
        <div v-loading="drawerLoading">
          <el-form ref="formRef" :model="formData" :rules="rules">
            <el-form-item label="接口类型" prop="type">
              <el-radio-group v-model="formData.type">
                <el-radio-button :value="0" label="期限"/>
                <el-radio-button :value="1" label="计次"/>
              </el-radio-group>
            </el-form-item>
            <el-form-item label="所属接口" prop="api_id">
              <el-input v-model="formData.api_id" placeholder="请输入接口Api ID，例如：1"/>
            </el-form-item>
            <el-form-item v-if="formData.type==0" label="有效期限" prop="validateDynamicField">
              <el-date-picker
                  v-model="dateArray"
                  :shortcuts="shortcuts"
                  end-placeholder="过期日期"
                  range-separator="至"
                  size="default"
                  start-placeholder="开始日期"
                  type="daterange"
                  unlink-panels
                  value-format="YYYY-MM-DD HH:mm:ss"
              />
            </el-form-item>
            <el-form-item v-if="formData.type==1" label="可用次数" prop="validateDynamicField">
              <el-input-number v-model="formData.count" :max="9999999999" :min="0" :step="20"/>
            </el-form-item>
            <el-form-item label="备&emsp;&emsp;注" prop="desc">
              <el-input v-model="formData.desc" :maxlength="255" autosize placeholder="可填写备注，例如：测试用"
                        type="textarea"/>
            </el-form-item>
          </el-form>
          <div class="flex justify-center">
            <el-button class="w-50" size="large" type="primary"
                       :loading="submitLoading"
                       @click="nowRow ? submitFormUpdate() : submitFormCreate()">
              {{ nowRow ? '编辑' : '创建' }} APIKey
            </el-button>
          </div>
        </div>

      </el-drawer>
    </template>
  </admin-main-body>
</template>
<style scoped>
/* 隐藏表单标签的伪元素 */
:deep(.el-form-item__label:before) {
  display: none;
}

:deep(.custom-btt-drawer) {
  width: 80%;
}

:deep(.searchForm .el-input) {
  --el-input-width: 150px;
}

:deep(.searchForm .el-select) {
  --el-select-width: 100px;
}

@media (max-width: 768px) {
  :deep(.custom-btt-drawer) {
    width: 100%;
  }
}
</style>