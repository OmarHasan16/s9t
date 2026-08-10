<template>
  <div class="d-flex flex-column gap-1">
    <p
      v-if="!!message"
      class="text-break"
      v-html="message"
    />

    <b-button
      :variant="pVal('buttonVariant', 'primary')"
      :disabled="loading"
      class="ml-auto"
      @click="$emit('submit', { confirmed: pRaw(undefined, true, 'Boolean') })"
    >
      {{ pVal('buttonLabel', 'Ok') }}
    </b-button>
  </div>
</template>

<script lang="js">
import base from './base.vue'

export default {
  name: 'CPromptAlert',
  extends: base,
}
</script>
