<template>
  <div class="container">
    <div class="footer">
      <!-- Powered by {{ author }} @
      <el-link type="primary" href="https://github.com/yangjian102621/chatgpt-plus" target="_blank">
        {{ title }} -
        {{ version }}
      </el-link> -->
    </div>
  </div>
</template>
<script setup>

import {ref} from "vue";

const title = ref(process.env.VUE_APP_TITLE)
const version = ref(process.env.VUE_APP_VERSION)
const author = ref('欢豚')
</script>

<style scoped lang="stylus">
.container {
  position: fixed;
  left: 0;
  bottom: 0;
  width: 100%;
  display flex;
  justify-content center

  .footer {
    max-width 400px;
    text-align center;
    font-size 14px;
    padding 20px;
    width 100%
  }
}

</style>