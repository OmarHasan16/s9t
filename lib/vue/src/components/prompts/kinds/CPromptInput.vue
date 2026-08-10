<template>
  <div class="d-flex flex-column gap-1">
    <p
      v-if="!!message"
      class="text-break"
      v-html="message"
    />

    <b-form-group
      :label="label"
      label-class="text-primary"
    >
      <c-input-date-time
        v-if="type === 'date' || type === 'time' || type === 'datetime'"
        v-model="value"
        :no-date="type === 'time'"
        :no-time="type === 'date'"
        :disabled="loading"
        :labels="{
          clear: $t('general:label.clear'),
          none: $t('general:label.none'),
          now: $t('general:label.now'),
          today: $t('general:label.today'),
        }"
      />
      <b-input
        v-else
        v-model="value"
        :type="type"
        :disabled="loading"
      />
    </b-form-group>

    <b-button
      :disabled="loading"
      variant="primary"
      class="ml-auto"
      @click="$emit('submit', { value: { '@value': value, '@type': 'String' }})"
    >
      {{ pVal('buttonLabel', 'Submit') }}
    </b-button>
  </div>
</template>

<script lang="js">
import base from './base.vue'
import { CInputDateTime } from '../../input'

const validTypes = [
  'text',
  'number',
  'email',
  'password',
  'search',
  'date',
  'time',
  'datetime',
]

export default {
  name: 'CPromptInput',

  components: {
    CInputDateTime,
  },

  extends: base,

  data () {
    return {
      value: undefined,
    }
  },

  computed: {
    type () {
      const t = this.pVal('type', 'text')
      if (validTypes.indexOf(t) === -1) {
        return 'text'
      }

      return t
    },
  },

  beforeMount () {
    this.value = this.pVal('inputValue')
  },

}
</script>
