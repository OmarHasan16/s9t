<template>
  <c-lightbox v-on="$listeners">
    <!-- Proxy slots -->
    <template
      v-for="(_, slot) of $scopedSlots"
      #[slot]="scope"
    >
      <slot
        :name="slot"
        v-bind="scope"
        v-on="$listeners"
      />
    </template>

    <component
      :is="previewType"
      v-if="previewType"
      v-bind="$attrs"
      v-on="$listeners"
    />
  </c-lightbox>
</template>

<script lang="js">
import { LightboxIMG as IMG, PDF, NoPreview } from './common/types'
import { CLightbox } from '../lightbox/index.ts'
import base from './common/base.vue'

export default {
  components: {
    CLightbox,
    IMG,
    PDF,
    NoPreview,
  },

  extends: base,
}
</script>
