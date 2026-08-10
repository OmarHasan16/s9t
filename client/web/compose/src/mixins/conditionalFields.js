// vue mixin for conditional fields
import { NoID } from '@cortezaproject/corteza-js'

export default {
  data () {
    return {
      conditions: [], // Array of fieldIDs that should be hidden
      evaluating: false,
    }
  },

  methods: {
    async evaluateExpressions () {
      if (!this.block.options.fieldConditions.length || this.$route.name === 'admin.pages.builder') return

      await new Promise(resolve => setTimeout(resolve, 300))

      const { expressions, variables } = this.prepareFieldConditionsData()

      return this.$SystemAPI
        .expressionEvaluate({ variables, expressions })
        .then(res => {
          // Store previous conditions before updating
          const previousConditions = [...this.conditions]

          // Update current conditions
          this.conditions = []

          Object.keys(res).forEach(v => {
            if (!res[v]) this.conditions.push(v)
          })

          this.clearValuesForHiddenFields(previousConditions)
        }).catch(this.toastErrorHandler(this.$t('notification:record.fieldConditions.failed')))
    },

    prepareFieldConditionsData () {
      const expressions = {}
      const record = this.record ? this.record.serialize() : {}
      const variables = {
        user: this.$auth.user,
        record,
      }

      this.block.options.fieldConditions.forEach(({ field, condition }) => {
        if (field && condition) {
          expressions[field] = condition
        }
      })

      return { expressions, variables }
    },

    canDisplay ({ fieldID, name }) {
      return !this.conditions.includes(fieldID !== NoID ? fieldID : name)
    },

    clearValuesForHiddenFields (previousConditions) {
      // Find fields that were visible before but are now hidden
      const newlyHiddenFields = this.conditions.filter(fieldId =>
        !previousConditions.includes(fieldId),
      )

      if (newlyHiddenFields.length === 0) return

      const clearAllOnHide = this.block.options.clearConditionalFieldsOnHide || false

      this.block.options.fieldConditions.forEach(({ field, clearOnHide }) => {
        const shouldClear = clearAllOnHide || clearOnHide

        if (!shouldClear) return
        if (!newlyHiddenFields.includes(field)) return

        const moduleField = this.module?.fields?.find(f =>
          f.fieldID === field || f.name === field,
        )

        if (!moduleField || !this.record || !this.record.values) return

        const fieldName = moduleField.name

        if (moduleField.isMulti) {
          this.record.values[fieldName] = []
        } else {
          this.record.values[fieldName] = undefined
        }
      })
    },
  },
}
