<template>
  <b-button
    data-test-id="button-submit"
    type="submit"
    :variant="variant"
    :disabled="disabled || processing || success"
    :size="size"
    :block="block"
    :title="title"
    :class="buttonClass"
    @click.prevent="$emit('submit')"
  >
    <b-spinner
      v-if="processing"
      data-test-id="spinner"
      small
      class="align-middle"
    />

    <font-awesome-icon
      v-else-if="success"
      data-test-id="icon-success"
      :icon="['fas', 'check']"
    />

    <span
      v-else-if="text"
      data-test-id="button-text"
    >
      {{ text }}
    </span>
  </b-button>
</template>

<script>
export default {
  name: 'CButtonSubmit',

  props: {
    processing: {
      type: Boolean,
    },

    success: {
      type: Boolean,
    },

    disabled: {
      type: Boolean,
    },

    title: {
      type: String,
      default: '',
    },

    buttonClass: {
      type: String,
      default: '',
    },

    text: {
      type: String,
      default: '',
    },

    loadingText: {
      type: String,
      default: '',
    },

    size: {
      type: String,
      default: 'md',
    },

    block: {
      type: Boolean,
    },

    variant: {
      type: String,
      default: 'primary',
    },
  },
}
</script>
