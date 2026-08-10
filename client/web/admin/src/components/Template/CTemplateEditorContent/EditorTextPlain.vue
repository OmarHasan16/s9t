<template>
  <c-ace-editor
    v-model="template.template"
    data-test-id="template-plain-text-editor"
    name="editor/plain_text"
    lang="plain_text"
    min-height="400px"
    show-line-numbers
    resizable
  />
</template>

<script>
import { components } from '@cortezaproject/corteza-vue'

const { CAceEditor } = components

export default {
  components: {
    CAceEditor,
  },

  props: {
    template: {
      type: Object,
      required: true,
      default: () => ({}),
    },
  },
}
</script>
