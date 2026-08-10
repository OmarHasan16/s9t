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
      <c-input-select
        v-if="inputType === 'select'"
        v-model="value"
        :options="itemOptions"
        :disabled="loading"
        :multiple="multiple"
        append-to-body
        label="text"
        :get-option-key="r => r.value"
        :placeholder="placeholder"
        :reduce="r => r.value"
        class="w-100"
      />

      <b-form-radio-group
        v-else-if="inputType === 'radio'"
        v-model="value"
        :options="itemOptions"
        :disabled="loading"
      />
    </b-form-group>

    <b-button
      :disabled="loading"
      variant="primary"
      class="ml-auto"
      @click="$emit('submit', { value: encodeValue() })"
    >
      {{ pVal('buttonLabel', 'Submit') }}
    </b-button>
  </div>
</template>

<script lang="js">
import base from './base.vue'
import CInputSelect from '../../input/CInputSelect.vue'

const validTypes = [
  'select',
  'radio',
]

export default {
  name: 'CPromptOptions',

  components: {
    CInputSelect,
  },

  extends: base,

  data () {
    return {
      value: undefined,
    }
  },

  computed: {
    itemOptions () {
      const out = []
      const options = this.pVal('options', {})

      for (const value in options) {
        out.push({ value, text: options[value] })
      }

      return out
    },

    inputType () {
      const t = this.pVal('type', 'text')
      if (validTypes.indexOf(t) === -1) {
        return 'select'
      }

      return t
    },

    multiple () {
      return this.pVal('multiselect', false)
    },

    placeholder () {
      return this.pVal('placeholder', 'Select an option')
    },
  },

  beforeMount () {
    let value = this.pVal('value')

    if (this.pVal('multiselect', false)) {
      if (Array.isArray(value)) {
        value = value.map(v => v['@value'])
      } else {
        value = value ? [value] : []
      }
    }

    this.value = value
  },

  methods: {
    encodeValue () {
      if (Array.isArray(this.value)) {
        return {
          '@type': 'Array',
          '@value': this.value || [],
        }
      } else {
        return { '@type': 'String', '@value': this.value }
      }
    },
  },
}
</script>
