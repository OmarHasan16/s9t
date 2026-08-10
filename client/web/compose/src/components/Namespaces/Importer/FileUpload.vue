<template>
  <b-card>
    <c-uploader
      :endpoint="fileUploadEndpoint"
      :accepted-files="['application/zip']"
      :labels="{
        uploading: $t('general:label.uploading'),
        placeholder: $t('import.uploadFilePlaceholder'),
        fileTypeNotAllowed: $t('general:label.fileTypeNotAllowed'),
      }"
      @upload="onUploaded"
    />
  </b-card>
</template>

<script>
import { components } from '@cortezaproject/corteza-vue'
const { CUploader } = components

export default {
  i18nOptions: {
    namespaces: 'namespace',
  },

  components: {
    CUploader,
  },

  data () {
    return {
      session: null,
      sessionFile: null,
    }
  },

  computed: {
    fileUploadEndpoint () {
      return this.$ComposeAPI.baseURL + this.$ComposeAPI.namespaceImportInitEndpoint({})
    },

    canContinue () {
      return !!this.session
    },
  },

  methods: {
    onUploaded (e, f) {
      this.session = e
      this.sessionFile = f
      this.fileUploaded()
    },

    fileUploaded () {
      this.$emit('fileUploaded', {
        ...this.session || {},
      })
    },
  },
}
</script>
