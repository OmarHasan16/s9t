<template>
  <b-form-group class="p-0 m-0">
    <div
      v-for="(expr, ei) in value"
      :key="ei"
      class="d-flex align-items-center gap-1 mb-2"
    >
      <b-input-group>
        <b-input-group-prepend>
          <b-input-group-text
            v-b-tooltip.noninteractive.hover="{ title: $t('validators.expression.tooltip'), boundary: 'body' }"
            variant="extra-light"
          >
            ƒ
          </b-input-group-text>
        </b-input-group-prepend>
        <slot :value="value[ei]">
          <b-form-input
            v-model="value[ei]"
            :placeholder="placeholder"
          />
        </slot>
      </b-input-group>

      <c-input-confirm
        :no-prompt="noPrompt(value[ei])"
        show-icon
        @confirmed="$emit('remove', ei)"
      />
    </div>
  </b-form-group>
</template>
<script>

export default {
  i18nOptions: {
    namespaces: 'field',
  },

  props: {
    value: {
      type: Array,
      default: () => ([]),
    },

    placeholder: {
      type: String,
      default: () => {},
    },

    noPrompt: {
      type: Function,
      default: v => v.length === 0,
    },
  },
}
</script>
