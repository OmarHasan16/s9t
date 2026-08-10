<script>
import { compose, NoID, validator } from '@cortezaproject/corteza-js'
import Wrap from './Wrap'

export default {
  i18nOptions: {
    namespaces: 'block',
  },

  components: {
    Wrap,
  },

  props: {
    blockIndex: {
      type: Number,
      default: -1,
    },

    namespace: {
      type: compose.Namespace,
      required: true,
    },

    page: {
      type: compose.Page,
      required: true,
    },

    blocks: {
      type: Array,
      default: () => [],
    },

    block: {
      type: compose.PageBlock,
      required: true,
    },

    module: {
      type: compose.Module,
      required: false,
      default: undefined,
    },

    record: {
      type: compose.Record,
      required: false,
      default: undefined,
    },

    mode: {
      type: String,
      required: false,
      default: '',
    },

    editable: {
      type: Boolean,
      required: false,
      default: false,
    },

    resizing: {
      type: Boolean,
      required: false,
      default: false,
    },

    magnified: {
      type: Boolean,
      required: false,
      default: false,
    },

    unsavedBlocks: {
      type: Set,
      default: () => new Set(),
    },

    loadingRecord: {
      type: Boolean,
      required: false,
      default: false,
    },

    errors: {
      type: validator.Validated,
      required: false,
      default: () => new validator.Validated(),
    },
  },

  data () {
    return {
      processing: false,
      refreshInterval: null,
      key: 0,
    }
  },

  computed: {
    options: {
      get () {
        return this.block.options
      },
      set (options) {
        this.block.options = options
      },
    },

    isProcessing () {
      return this.processing || this.loadingRecord
    },

    autoRefreshEnabled () {
      return this.options.refreshRate >= 5 && ['page', 'page.record'].includes(this.$route.name)
    },

    // detect when a page block is opened in a modal through magnification or record open type
    inModal () {
      const { recordPageID, magnifiedBlockID } = this.$route.query

      return !!recordPageID || !!magnifiedBlockID
    },

    isRecordPage () {
      return this.page && this.page.moduleID !== NoID
    },

    errorID () {
      const { recordID = NoID } = this.record || {}
      return recordID === NoID ? 'parent:0' : recordID
    },
  },

  beforeDestroy () {
    this.setBaseDefaultValues()
  },

  methods: {
    setBaseDefaultValues () {
      if (this.refreshInterval) {
        clearInterval(this.refreshInterval)
        this.refreshInterval = null
      }

      this.processing = false
      this.key = 0
    },

    /**
     *
     * @param {*} refreshFunction
     * If reloading data does not refresh the page block
     * You should attach :key="key" to it and increment it in the refreshFunction
     */
    refreshBlock (refreshFunction, ...params) {
      if (!this.autoRefreshEnabled || this.refreshInterval) return

      const interval = setInterval(() => {
        refreshFunction(...params)
      }, this.options.refreshRate * 1000)

      this.refreshInterval = interval
    },

    /**
     * Returns errors, filtered for a specific field
     * @param name
     * @returns {validator.Validated} filtered validation results
     */
    fieldErrors (name) {
      if (!this.errors) {
        this.$emit('errors', { errors: undefined, id: `${this.errorID}:${name}` })
        return new validator.Validated()
      }

      const errors = this.errors.filterByMeta('field', name).filterByMeta('id', this.errorID)

      if (errors.set.length > 0) {
        this.$emit('errors', { errors, id: `${this.errorID}:${name}` })
      } else {
        this.$emit('errors', { errors: undefined, id: `${this.errorID}:${name}` })
      }

      return errors
    },
  },
}
</script>
