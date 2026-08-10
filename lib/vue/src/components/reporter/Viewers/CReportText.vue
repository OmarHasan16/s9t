<template>
  <div
    v-if="options"
    class="rt-content p-3"
    v-html="options.value"
  />
</template>

<script>
import base from './base.vue'

export default {
  extends: base,
}
</script>
