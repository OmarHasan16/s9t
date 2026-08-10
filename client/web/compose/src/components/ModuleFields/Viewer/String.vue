<template>
  <div class="rt-content">
    <p
      v-if="formatted"
      :style="{ 'white-space': field.options.useRichTextEditor && 'pre-line' }"
      :class="[ 'multiline' && field.isMulti || field.options.multiLine, ...classes ]"
      v-html="formatted"
    />
  </div>
</template>
<script>
import base from './base'

export default {
  extends: base,
}
</script>
