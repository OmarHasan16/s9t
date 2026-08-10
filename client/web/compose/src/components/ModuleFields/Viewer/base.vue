<template>
  <div>
    <!-- Extra empty line is added thanks to white-space: pre-line (multivalue) if we write div in multiple lines  -->
    <!-- eslint-disable-next-line -->
    <div :class="classes">{{ formatted }}</div>
  </div>
</template>
<script>
import { compose } from '@cortezaproject/corteza-js'

export default {
  props: {
    namespace: {
      type: compose.Namespace,
      required: true,
    },

    field: {
      type: compose.ModuleField,
      required: true,
    },

    record: {
      type: compose.Record,
      required: true,
    },

    valueOnly: {
      type: Boolean,
      required: false,
    },

    extraOptions: {
      type: Object,
      default: () => ({}),
    },

    includeStyles: {
      type: Boolean,
      default: false,
    },

    disableClick: {
      type: Boolean,
      default: false,
    },
  },

  computed: {
    value () {
      if (this.field.isSystem) {
        return this.record[this.field.name]
      }

      return this.record ? this.record.values[this.field.name] : undefined
    },

    formatted () {
      if (this.field.isMulti) {
        return this.value.join(this.field.options.multiDelimiter)
      }
      return this.value
    },

    classes () {
      const classes = []
      const { fieldID } = this.field
      const { textStyles = {} } = this.extraOptions

      if (this.field.isMulti) {
        classes.push('multiline')
      } else if (this.includeStyles) {
        if (!textStyles.wrappedFields || !textStyles.wrappedFields.includes(fieldID)) {
          classes.push('text-nowrap')
        }
      }

      return classes
    },

    options () {
      return this.field.options
    },

    // detect when a page block is opened in a modal through magnification or record open type
    inModal () {
      const { recordPageID, magnifiedBlockID } = this.$route.query

      return !!recordPageID || !!magnifiedBlockID
    },
  },
}
</script>

<style>
.multiline {
  white-space: pre-line;
}
</style>
