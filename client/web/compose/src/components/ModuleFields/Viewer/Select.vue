<template>
  <div>
    <span
      v-for="(v, index) of value"
      :key="index"
      :class="{ 'd-block': field.options.multiDelimiter === '\n' }"
    >
      <span
        :class="{ 'badge badge-pill': field.options.displayType === 'badge', 'mt-1': field.options.multiDelimiter === '\n' && index !== 0 }"
        :style="v.style"
      >
        {{ v.text }}
      </span>

      {{ index !== value.length - 1 ? field.options.multiDelimiter : '' }}
    </span>
  </div>
</template>

<script>
import base from './base'

export default {
  extends: base,

  computed: {
    /**
     * Overwrite default; allow values to resolve to their labels
     * @returns {String|Array<String>}
     */
    value () {
      let v
      if (this.field.isSystem) {
        v = this.record[this.field.name]
      }
      v = this.record ? this.record.values[this.field.name] : undefined

      if (this.field.isMulti) {
        if (!Array.isArray(v)) {
          v = []
        }

        return v.map(v => this.resolveValue(v) || v).filter(v => v && v.text)
      } else {
        return [this.resolveValue(v) || v].filter(v => v && v.text)
      }
    },
  },

  methods: {
    resolveValue (v) {
      const opt = this.field.options.options.find(({ value }) => value === v) || { text: v }

      return {
        text: opt.text,
        style: this.getOptionStyle(opt),
      }
    },

    getOptionStyle (opt) {
      const style = {}

      if (this.field.options.displayType === 'badge') {
        const optStyle = opt.style || {}
        style.fontSize = '0.9rem'
        style.color = optStyle.textColor || 'var(--dark)'
        style.backgroundColor = optStyle.backgroundColor || 'var(--extra-light)'
      }

      return style
    },
  },
}
</script>

<style lang="scss" scoped>
.badge {
  font-family: var(--font-medium);
}
</style>
