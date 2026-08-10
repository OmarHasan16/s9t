import { NoID } from '@cortezaproject/corteza-js'

export default {
  computed: {
    isRecordPage () {
      return this.page && this.page.moduleID !== NoID
    },

    recordAutoCompleteParams () {
      return this.processRecordAutoCompleteParams({})
    },

    visibilityAutoCompleteParams () {
      return this.processVisibilityAutoCompleteParams({})
    },
  },

  methods: {
    processRecordAutoCompleteParams ({ module = this.module, operators = false } = {}) {
      const { fields = [] } = module || {}
      const moduleFields = fields.map(({ name }) => name)
      const userProperties = this.$auth.user.properties() || []

      const recordSuggestions = this.isRecordPage && this.record
        ? [
            ...(['ownerID', 'recordID'].map(value => ({ interpolate: true, value }))),
            {
              interpolate: true,
              value: 'record',
              properties: [
                { value: 'values', properties: Object.keys(this.record.values) || [] },
                ...(this.record.properties || []),
              ],
            },
          ]
        : []

      return [
        ...recordSuggestions,
        ...(operators ? ['AND', 'OR'] : []),
        { interpolate: true, value: 'userID' },
        { interpolate: true, value: 'user', properties: userProperties },
        ...moduleFields,
      ]
    },

    processVisibilityAutoCompleteParams ({ module = this.module } = {}) {
      const { fields = [] } = module || {}
      const moduleFields = fields.map(({ name }) => name)
      const userProperties = this.$auth.user.properties() || []

      const recordSuggestions = this.isRecordPage && this.record
        ? [
            {
              value: 'record',
              properties: [
                { value: 'values', properties: Object.keys(this.record.values) || [] },
                ...(this.record.properties || []),
              ],
            },
          ]
        : []

      return [
        ...recordSuggestions,
        { value: 'user', properties: userProperties },
        { value: 'screen', properties: ['width', 'height', 'userAgent', 'breakpoint'] },
        ...moduleFields,
      ]
    },
  },
}
