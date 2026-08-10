<template>
  <div>
    <c-uploader
      :labels="{
        placeholder: $t('general:label.dropFiles'),
        uploading: $t('general:label.uploading'),
        fileTypeNotAllowed: $t('general:label.fileTypeNotAllowed'),
      }"
      :endpoint="userImportEndpoint"
      :accepted-files="['application/zip']"
      @upload="onUploaded"
    />
  </div>
</template>

<script>
import { components } from '@cortezaproject/corteza-vue'
const { CUploader } = components

export default {
  i18nOptions: {
    namespaces: 'system.users',
  },

  name: 'CFileUpload',

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
    userImportEndpoint () {
      return this.$SystemAPI.baseURL + this.$SystemAPI.userImportEndpoint({})
    },
  },

  methods: {
    onUploaded () {
      this.$emit('imported', {})
    },
  },
}
</script>
