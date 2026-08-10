<template>
  <wrap
    v-bind="$props"
    body-class="pt-3 px-3"
    v-on="$listeners"
  >
    <div
      v-if="isProcessing"
      class="d-flex align-items-center justify-content-center h-100"
    >
      <b-spinner />
    </div>

    <div
      v-else-if="fieldModule"
      ref="fieldContainer"
      :class="fieldLayoutClass"
    >
      <template v-for="field in fields">
        <b-form-group
          v-if="canDisplay(field)"
          :key="`${field.fieldID}-${field.name}`"
          :data-test-id="getFieldCypressId(field.label || field.name)"
          :label-cols-md="options.horizontalFieldLayoutEnabled && '6'"
          :label-cols-xl="options.horizontalFieldLayoutEnabled && '5'"
          :content-cols-md="options.horizontalFieldLayoutEnabled && '6'"
          :content-cols-xl="options.horizontalFieldLayoutEnabled && '7'"
          :class="columnWrapClass"
          :style="fieldWidth"
          class="field-container"
        >
          <template #label>
            <div
              class="d-flex align-items-center text-primary mb-0"
            >
              <span
                class="d-flex"
                style="margin-top: 0.1rem;"
              >
                {{ field.label || field.name }}
              </span>

              <c-hint :tooltip="((field.options.hint || {}).view || '')" />

              <div
                v-if="!record.deletedAt && options.inlineRecordEditEnabled && isFieldEditable(field)"
                class="inline-actions ml-1"
              >
                <b-button
                  v-b-tooltip.noninteractive.hover="{ title: $t('record.inlineEdit.button.title'), boundary: 'body' }"
                  variant="outline-extra-light"
                  :disabled="editable"
                  size="sm"
                  class="text-secondary border-0"
                  @click="editInlineField(fieldRecord, field)"
                >
                  <font-awesome-icon
                    :icon="['fas', 'pen']"
                  />
                </b-button>
              </div>
            </div>

            <div
              class="small text-muted"
              :class="{ 'mb-1': !!(field.options.description || {}).view }"
            >
              {{ (field.options.description || {}).view }}
            </div>
          </template>

          <div
            v-if="field.canReadRecordValue"
            class="value align-self-center"
          >
            <field-viewer
              v-bind="{ ...$props, field }"
              :extra-options="options"
              :record="fieldRecord"
            />
          </div>
          <i
            v-else
            class="text-muted"
          >
            {{ $t('field.noPermission') }}
          </i>
        </b-form-group>
      </template>
    </div>

    <!-- Modal for inline editing -->
    <bulk-edit-modal
      v-if="options.inlineRecordEditEnabled && fieldModule"
      :modal-title="$t('record.inlineEdit.modal.title')"
      :namespace="namespace"
      :module="fieldModule"
      :selected-records="inlineEdit.recordIDs"
      :selected-fields="inlineEdit.fields"
      :initial-record="inlineEdit.record"
      :query="inlineEdit.query"
      :allow-add-field="options.inlineRecordEditAllowAddField"
      open-on-select
      @save="onInlineEdit()"
      @close="onInlineEditClose()"
    />
  </wrap>
</template>
<script>
import { compose, NoID } from '@cortezaproject/corteza-js'
import { mapActions } from 'vuex'
import axios from 'axios'
import base from './base'
import FieldViewer from 'corteza-webapp-compose/src/components/ModuleFields/Viewer'
import BulkEditModal from 'corteza-webapp-compose/src/components/Public/Record/BulkEdit'
import users from 'corteza-webapp-compose/src/mixins/users'
import records from 'corteza-webapp-compose/src/mixins/records'
import conditionalFields from 'corteza-webapp-compose/src/mixins/conditionalFields'
import recordLayout from 'corteza-webapp-compose/src/mixins/recordLayout'

export default {
  i18nOptions: {
    namespaces: 'block',
  },

  components: {
    FieldViewer,
    BulkEditModal,
  },

  extends: base,

  mixins: [
    users,
    records,
    conditionalFields,
    recordLayout,
  ],

  data () {
    return {
      referenceRecord: undefined,
      referenceModule: undefined,
      inlineEdit: {
        fields: [],
        recordIDs: [],
        record: {},
      },

      abortableRequests: [],
    }
  },

  computed: {
    fields () {
      if (!this.fieldModule) {
        // No module, no fields
        return []
      }

      if (!this.options.fields || this.options.fields.length === 0) {
        // No fields defined in the options, show all (buy system)
        return this.fieldModule.fields
      }

      // Show filtered & ordered list of fields
      return this.fieldModule.filterFields(this.options.fields).map(f => {
        f.label = f.isSystem ? this.$t(`field:system.${f.name}`) : f.label || f.name
        return f
      })
    },

    fieldLayoutClass () {
      const classes = {
        default: 'd-flex flex-column',
        noWrap: 'd-flex gap-2',
        wrap: 'row no-gutters',
      }

      return classes[this.options.recordFieldLayoutOption]
    },

    fieldModule () {
      return this.options.referenceField ? this.referenceModule : this.module
    },

    fieldRecord () {
      return this.options.referenceField ? this.referenceRecord : this.record
    },

    isProcessing () {
      return this.loadingRecord || !this.fieldRecord || this.evaluating
    },

    fieldWidth () {
      if (this.options.recordFieldLayoutOption !== 'noWrap') {
        return {}
      }

      return { 'min-width': '13rem' }
    },
  },

  watch: {
    loadingRecord: {
      immediate: true,
      handler (loadingRecord) {
        const { recordID } = this.record || {}

        if (!recordID || loadingRecord) return

        let resolutions = []

        if (recordID !== NoID) {
          resolutions = [
            this.fetchUsers(this.fields, [this.record]),
            this.fetchRecords(this.namespace.namespaceID, this.fields, [this.record]),
          ]
        }

        this.evaluating = true

        Promise.all([
          ...resolutions,
          this.evaluateExpressions(),
        ]).finally(() => {
          this.evaluating = false
        })

        if (this.options.referenceModuleID) {
          this.fetchReferenceModule(this.options.referenceModuleID)
        }
      },
    },

    options: {
      deep: true,
      handler (options) {
        if (options.referenceModuleID) {
          this.fetchReferenceModule(options.referenceModuleID)
        }
      },
    },

    'record.values': {
      deep: true,
      handler (newValues = {}) {
        if (this.options.referenceField) {
          const { recordID: oldValue } = this.referenceRecord || {}
          const newValue = newValues[this.options.referenceField]

          if (oldValue !== newValue) {
            this.loadRecord(this.referenceModule)
          }
        }
      },
    },

    isProcessing: {
      handler (newVal) {
        if (this.options.recordFieldLayoutOption !== 'wrap') return

        if (!newVal && this.fieldModule) {
          this.$nextTick(() => {
            this.initializeResizeObserver(this.$refs.fieldContainer, this.options.recordFieldLayoutOption)
          })
        }
      },
    },

    'options.recordFieldLayoutOption': {
      handler (newVal) {
        if (newVal === 'wrap' && this.fieldModule) {
          this.initializeResizeObserver(this.$refs.fieldContainer, this.options.recordFieldLayoutOption)
        } else if (this.resizeObserver) {
          this.resizeObserver.unobserve(this.$refs.fieldContainer)
          this.columnWrapClass = ''
        }
      },
    },

    referenceRecord: {
      handler () {
        if (!this.referenceRecord) return

        this.fetchUsers(this.fields, [this.referenceRecord])
        this.fetchRecords(this.namespace.namespaceID, this.fields, [this.referenceRecord])
      },
    },
  },

  beforeDestroy () {
    this.abortRequests()
    this.setDefaultValues()
  },

  methods: {
    ...mapActions({
      findModuleByID: 'module/findByID',
      updateRecordSet: 'record/updateRecords',
    }),

    fetchReferenceModule (moduleID) {
      if (!moduleID) {
        this.referenceModule = undefined
        return
      }

      this.findModuleByID({ namespace: this.namespace.namespaceID, moduleID: this.options.referenceModuleID })
        .then(module => {
          this.referenceModule = new compose.Module({ ...module })

          if (this.options.referenceField) {
            this.loadRecord(this.referenceModule)
          }
        })
    },

    loadRecord (module) {
      if (!module) return

      const { namespaceID, moduleID } = module
      const { referenceField } = this.options
      const field = this.module.fields.find(({ fieldID }) => fieldID === referenceField)

      const recordID = this.record.values[field.name]

      if (!recordID || !field) {
        this.referenceRecord = new compose.Record(this.fieldModule, {})
        return
      }

      if (field.isMulti) {
        this.referenceRecord = new compose.Record(this.fieldModule, {})
        return
      }

      const { response, cancel } = this.$ComposeAPI
        .recordReadCancellable({ namespaceID, moduleID, recordID })

      this.abortableRequests.push(cancel)

      response()
        .then(record => {
          this.referenceRecord = new compose.Record(this.fieldModule, { ...record })
          this.updateRecordSet(this.referenceRecord)
        })
        .catch(e => {
          if (!axios.isCancel(e)) {
            this.referenceRecord = new compose.Record(this.fieldModule, {})
            this.toastErrorHandler(this.$t('notification:record.loadFailed'))(e)
          }
        })
    },

    editInlineField (record, field) {
      this.inlineEdit.fields = [field.name]
      this.inlineEdit.record = record.clone()
      this.inlineEdit.recordIDs = [record.recordID]
      this.inlineEdit.query = `recordID = ${record.recordID}`
    },

    onInlineEdit () {
      this.inlineEdit.fields = []
      this.inlineEdit.recordIDs = []
      this.inlineEdit.record = {}
      this.inlineEdit.query = ''
    },

    onInlineEditClose () {
      this.inlineEdit.fields = []
      this.inlineEdit.record = {}
      this.inlineEdit.query = ''
    },

    setDefaultValues () {
      this.referenceRecord = undefined
      this.referenceModule = undefined
      this.inlineEdit = {}
      this.abortableRequests = []
    },

    abortRequests () {
      this.abortableRequests.forEach((cancel) => {
        cancel()
      })
    },

    getFieldCypressId (field) {
      return `field-${field.toLowerCase().split(' ').join('-')}`
    },
  },
}
</script>

<style scoped>
.field-col > * {
  margin-left: 1rem;
  margin-right: 1rem;
}
</style>
