<template>
  <b-form
    @submit.prevent="$emit('upload')"
  >
    <c-uploader
      v-if="!disabled"
      :endpoint="uploadEndpoint"
      :labels="{
        placeholder: $t('general:label.dropFiles'),
        uploading: $t('general:label.uploading'),
        fileTypeNotAllowed: $t('general:label.fileTypeNotAllowed'),
      }"
      :accepted-files="['image/*']"
      @upload="$emit('upload', $event)"
    />

    <div
      v-if="value"
      class="d-flex justify-content-center w-100 mt-2"
    >
      <b-img
        :src="value"
        class="mw-100 h-auto"
      />
    </div>
  </b-form>
</template>

<script>
import { components } from '@cortezaproject/corteza-vue'
const { CUploader } = components

export default {
  name: 'CUploaderWithPreview',

  components: {
    CUploader,
  },

  props: {
    value: {
      type: String,
      default: () => undefined,
    },

    disabled: {
      type: Boolean,
      default: () => false,
    },

    labels: {
      type: Object,
      default: () => ({}),
    },

    endpoint: {
      type: String,
      required: true,
    },

    acceptedFiles: {
      type: Array,
      default: () => [],
    },

    maxFilesize: {
      type: Number,
      default: 100,
    },
  },

  data () {
    return {
      panels: [],
    }
  },

  computed: {
    uploadEndpoint () {
      return this.$SystemAPI.baseURL + this.endpoint
    },
  },
}
</script>
