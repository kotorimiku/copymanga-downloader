<template>
    <div class="config">
        <div class="form-item">
            <label>保存路径</label>
            <input type="text" v-model="outputPath" placeholder="请输入保存路径" class="input-box" />
        </div>
        <div class="form-item">
            <label>打包方式</label>
            <div class="select-container">
                <select v-model="packageType" class="styled-select">
                    <option value="cbz" title="会添加元数据ComicInfo.xml">cbz</option>
                    <option value="zip">zip</option>
                    <option value="image">图片</option>
                </select>
            </div>
        </div>
        <!-- 新增保存配置按钮 -->
        <div class="form-item">
            <button @click="saveConfig" class="btn save-btn">保存配置</button>
        </div>
    </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useToast } from "vue-toastification";
import { SaveConfig, GetConfig } from '../../wailsjs/go/main/Config'
import { main } from '../../wailsjs/go/models';

const outputPath = ref<string>("")
const packageType = ref<string>("")
const urlBase = ref<string>("")
const toast = useToast();

const saveConfig = () => {
    SaveConfig({
        urlBase: urlBase.value, outputPath: outputPath.value, packageType: packageType.value, userList: [],
        convertValues: function (a: any, classs: any, asMap?: boolean) {
            throw new Error('Function not implemented.');
        }
    }).then((res: any) => {
        console.log("配置已保存", res)
        toast.success('配置已保存', {
            timeout: 2000,
        });
    }).catch((err: any) => {
        console.error("保存配置失败", err)
        toast.error('保存配置失败', {
            timeout: 2000,
        });
    })
}

onMounted(() => {
    GetConfig().then((res: main.Config) => {
        if (res) {
            urlBase.value = res.urlBase
            outputPath.value = res.outputPath
            packageType.value = res.packageType
        }
    }).catch(() => {
        console.log("获取配置失败")
        toast.error('获取配置失败', {
            timeout: 2000,
        });
    })
})
</script>

<style scoped>
/* 配置项容器 */
.config {
    width: 300px;
    padding: 20px;
    background-color: #f9f9f9;
    border-radius: 8px;
    box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
    margin: 20px auto;
}

/* 表单项样式 */
.form-item {
    margin-bottom: 15px;
    display: flex;
    align-items: center;
    justify-content: space-between;
}

label {
    font-weight: bold;
    font-size: 14px;
    width: 80px;
    text-align: left;
}

/* 输入框样式 */
.input-box {
    width: 170px;
    padding: 8px;
    font-size: 14px;
    border-radius: 5px;
    border: 1px solid #ddd;
    margin-left: 10px;
}

/* 按钮样式 */
.btn {
    padding: 8px 16px;
    font-size: 14px;
    background-color: #ffffff;
    border: 1px solid #ddd;
    border-radius: 5px;
    cursor: pointer;
    transition: background-color 0.3s ease;
    margin-left: 10px;
}

.btn:hover {
    background-color: #f1f1f1;
}

.save-btn {
    width: 100%;
    background-color: #4CAF50;
    color: white;
    font-weight: bold;
}

.save-btn:hover {
    background-color: #45a049;
}

/* 美化 select 元素 */
.select-container {
    position: relative;
    width: 170px;
}

.styled-select {
    width: 100%;
    padding: 8px;
    font-size: 14px;
    border-radius: 5px;
    border: 1px solid #ddd;
    background-color: #fff;
    appearance: none;
    /* 去掉默认的箭头 */
    -webkit-appearance: none;
    /* 在 Safari 中去掉默认的箭头 */
    -moz-appearance: none;
    /* 在 Firefox 中去掉默认的箭头 */
    cursor: pointer;
}

.styled-select:focus {
    outline: none;
    border-color: #4CAF50;
}

.styled-select option {
    padding: 8px;
}

/* 自定义下拉箭头 */
.styled-select::after {
    content: '\2193';
    /* 向下箭头符号 */
    position: absolute;
    right: 10px;
    top: 50%;
    transform: translateY(-50%);
    font-size: 18px;
    color: #888;
    pointer-events: none;
}

.disabled {
    background-color: #e0e0e0;
    cursor: not-allowed;
}
</style>
