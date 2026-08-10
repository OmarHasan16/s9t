<template>
  <b-modal
    :visible="showModal"
    :title="$t('recordList.summaries.customSummaries.label')"
    body-class="p-0"
    footer-class="d-flex w-100 align-items-center"
    centered
    no-fade
    size="lg"
    @hide="onModalHide"
  >
    <b-card
      class="pt-0"
    >
      <b-form-group
        :label="$t('recordList.summaries.param.label.label')"
        label-class="text-primary"
      >
        <b-form-input
          v-model="summary.label"
          :placeholder="$t('recordList.summaries.param.label.placeholder')"
        />
      </b-form-group>

      <b-form-group
        :label="$t('recordList.summaries.param.field.label')"
        label-class="text-primary"
      >
        <c-input-select
          v-model="summary.field"
          :options="recordListModuleFields"
          :placeholder="$t('recordList.summaries.param.field.placeholder')"
          :reduce="field => field.name"
        />
      </b-form-group>

      <b-form-group
        :label="$t('recordList.summaries.param.metric.label')"
        label-class="text-primary"
      >
        <c-input-select
          v-model="summary.metric"
          :options="summaryMetrics"
          :placeholder="$t('recordList.summaries.param.metric.placeholder')"
          :reduce="m => m.value"
        />
      </b-form-group>
    </b-card>

    <template #modal-footer>
      <c-input-confirm
        v-if="summaryIndex !== -1"
        :text="$t('general:label.delete')"
        variant="danger"
        size="md"
        @confirmed="onDelete"
      />

      <b-button
        variant="light"
        class="ml-auto"
        @click="onModalHide"
      >
        {{ $t('general:label.cancel') }}
      </b-button>

      <b-button
        variant="primary"
        :disabled="isSaveDisabled"
        @click="onSave"
      >
        {{ $t('general:label.save') }}
      </b-button>
    </template>
  </b-modal>
</template>

<script>
export default {
  i18nOptions: {
    namespaces: 'block',
  },

  name: 'CustomSummary',

  props: {
    visible: {
      type: Boolean,
      default: false,
    },

    module: {
      type: Object,
      required: true,
    },

    summaryIndex: {
      type: Number,
      default: -1,
    },

    summary: {
      type: Object,
      default: () => ({}),
    },
  },

  data () {
    return {
      showModal: this.visible,
    }
  },

  computed: {
    recordListModuleFields () {
      if (this.module) {
        return [
          ...this.module.fields,
          ...this.module.systemFields().map(sf => {
            return {
              label: this.$t(`field:system.${sf.name}`),
              name: sf.name === 'recordID' ? 'ID' : sf.name,
            }
          }),
        ].map(({ name, label }) => ({ name, label }))
      }

      return []
    },

    summaryMetrics () {
      return [
        { value: 'sum', label: this.$t('recordList.summaries.metrics.sum.label') },
        { value: 'min', label: this.$t('recordList.summaries.metrics.min.label') },
        { value: 'max', label: this.$t('recordList.summaries.metrics.max.label') },
        { value: 'avg', label: this.$t('recordList.summaries.metrics.avg.label') },
        { value: 'emptyCount', label: this.$t('recordList.summaries.metrics.emptyCount.label') },
        { value: 'notEmptyCount', label: this.$t('recordList.summaries.metrics.notEmptyCount.label') },
        { value: 'uniqueCount', label: this.$t('recordList.summaries.metrics.uniqueCount.label') },
      ]
    },

    isSaveDisabled () {
      return !this.summary.label || !this.summary.field || !this.summary.metric
    },
  },

  watch: {
    visible: {
      immediate: true,
      handler (val) {
        this.showModal = val
      },
    },
  },

  methods: {
    onModalHide () {
      this.showModal = false
      this.$emit('close')
    },

    onSave () {
      this.$emit('save', this.summary)
    },

    onDelete () {
      this.$emit('delete')
    },
  },
}
</script>
