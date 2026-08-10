<script>
import multi from './multi'
import errors from '../errors'
import { compose, validator } from '@cortezaproject/corteza-js'
import { mapGetters } from 'vuex'

export default {
  components: {
    // multi is used in the components that extends base
    // eslint-disable-next-line vue/no-unused-components
    multi,

    // errors is used in the components that extends base
    // eslint-disable-next-line vue/no-unused-components
    errors,
  },

  props: {
    namespace: {
      type: compose.Namespace,
      required: true,
    },

    field: {
      type: compose.ModuleField,
      required: true,
    },

    record: {
      type: compose.Record,
      required: true,
    },

    errors: {
      type: validator.Validated,
      required: true,
    },

    valueOnly: {
      type: Boolean,
      default: false,
    },

    horizontal: {
      type: Boolean,
      default: false,
    },

    extraOptions: {
      type: Object,
      default: () => ({}),
    },
  },

  computed: {
    ...mapGetters({
      isFieldRequiredByLayout: 'ui/isFieldRequiredByLayout',
    }),

    formGroupStyleClasses () {
      return {
        required: this.isRequired,

        // CSS class "small" is set on form group
        // because of font-size: inherit prop on .col-form-label on
        // wrapping element
        small: false,
        'value-only': this.valueOnly,
      }
    },

    foo () {
      return this.isFieldRequiredByLayout(this.field.name || this.field.fieldID)
    },

    isRequired () {
      return this.field.isRequired || this.isFieldRequiredByLayout(this.field.name || this.field.fieldID)
    },

    state () {
      if (!this.errors.valid()) {
        return null
      }

      return this.errors.valid() === true ? null : false
    },

    value: {
      get () {
        if (this.field.isSystem) {
          return this.record[this.field.name]
        }

        return this.record.values[this.field.name]
      },

      set (value) {
        if (this.field.isSystem) {
          this.record[this.field.name] = value
        } else {
          this.record.values[this.field.name] = value
        }

        this.$emit('change', value)
      },
    },

    showPopover: {
      get () {
        return this.preventPopoverClose
      },

      set (v) {
        this.$emit('update:preventPopoverClose', v)
      },
    },

    label () {
      return this.field.label || this.field.name
    },

    description () {
      if (this.valueOnly) {
        return ''
      }

      const { view, edit } = this.field.options.description

      return edit || view
    },

    hint () {
      if (this.valueOnly) {
        return ''
      }
      const { view, edit } = this.field.options.hint

      return edit || view
    },

    // detect when a page block is opened in a modal through magnification or record open type
    inModal () {
      const { recordPageID, magnifiedBlockID } = this.$route.query

      return !!recordPageID || !!magnifiedBlockID
    },
  },

  methods: {
    getFieldCypressId (field) {
      return `field-${field.toLowerCase().split(' ').join('-')}`
    },

    setMultiValue (value, index) {
      this.$set(this.value, index, value)

      this.$emit('change', this.value)
    },
  },
}
</script>
