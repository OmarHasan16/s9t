<template>
  <div>
    <b-modal
      :visible="visible"
      size="lg"
      :title="(filter || {}).label"
      :ok-title="$t('filters.modal.ok')"
      body-class="p-0"
      cancel-variant="light"
      @ok="onSave"
      @hidden="onHidden"
    >
      <div
        v-if="internalFilter"
        class="card-body"
      >
        <c-filter-params
          :filter="internalFilter"
        />
        <b-form-checkbox
          v-model="internalFilter.enabled"
          data-test-id="checkbox-filter-enable"
        >
          {{ $t('filters.enabled') }}
        </b-form-checkbox>
      </div>
    </b-modal>
  </div>
</template>

<script>
import CFilterParams from 'corteza-webapp-admin/src/components/Apigw/CFilterParams'

export default {
  components: {
    CFilterParams,
  },

  props: {
    filter: {
      type: Object,
      default: () => ({}),
    },

    visible: {
      type: Boolean,
      required: false,
      default: false,
    },
  },

  data () {
    return {
      internalFilter: undefined,
    }
  },

  watch: {
    visible: {
      handler (visible) {
        if (visible) {
          // Convert to FE structure
          this.internalFilter = {
            ...this.filter,
            params: this.filter.params.map(p => {
              let value = p.value || {}

              if (this.filter.ref === 'response') {
                if (p.type === 'header') {
                  value = Object.entries(value).map(([name, v = []]) => ({ name, expr: v.join('') }))
                } else if (p.type === 'input') {
                  value = { type: 'Any', expr: '', ...value }
                }
              }

              return { ...p, value }
            }),
          }
        } else {
          this.internalFilter = undefined
        }
      },
    },
  },

  methods: {
    onSave () {
      // Convert to BE structure
      const filter = {
        ...this.internalFilter,
        params: this.internalFilter.params.map(p => {
          if (this.filter.ref === 'response') {
            if (p.type === 'header') {
              p.value = p.value.reduce((obj, { name, expr = '' }) => {
                return { ...obj, [name]: [expr] }
              }, {})
            }
          }

          return p
        }),
      }

      this.$emit('submit', { ...filter, updated: true })
    },

    onHidden () {
      this.internalFilter = undefined
      this.$emit('reset')
    },
  },
}
</script>
