<template>
  <div class="d-flex flex-column gap-1">
    <p
      v-if="!!message"
      class="text-break"
      v-html="message"
    />

    <div class="d-flex flex-wrap align-items-center gap-2">
      <b-button
        :variant="pVal('confirmButtonVariant', 'primary')"
        :disabled="loading"
        class="flex-grow-1"
        @click="$emit('submit', { value: pRaw('confirmButtonValue', true, 'Boolean') })"
      >
        {{ pVal('confirmButtonLabel', 'Yes') }}
      </b-button>

      <b-button
        :disabled="loading"
        :variant="pVal('rejectButtonVariant', 'light')"
        class="flex-grow-1"
        @click="$emit('submit', { value: pRaw('rejectButtonValue', false, 'Boolean') })"
      >
        {{ pVal('rejectButtonLabel', 'No') }}
      </b-button>
    </div>
  </div>
</template>

<script lang="js">
import base from './base.vue'

export default {
  name: 'CPromptChoice',
  extends: base,
}
</script>
