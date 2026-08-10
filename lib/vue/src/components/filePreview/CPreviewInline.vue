<template>
  <div class="inline h-100">
    <div class="h-100">
      <component
        :is="previewType"
        v-bind="$attrs"
        :max-pages="1"
        :initial-scale="1.5"
        inline
        v-on="$listeners"
      />
    </div>
  </div>
</template>

<script lang="js">
import { InlineIMG as IMG, PDF, NoPreview } from './common/types'
import base from './common/base.vue'

export default {
  components: {
    IMG,
    PDF,
    NoPreview,
  },

  extends: base,
}
</script>
