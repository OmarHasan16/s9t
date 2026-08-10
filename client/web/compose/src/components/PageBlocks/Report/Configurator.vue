<template>
  <b-tab :title="$t('report.label')">
    <b-row>
      <b-col>
        <b-form-group
          :label="$t('report.label')"
          label-class="text-primary"
        >
          <c-input-select
            v-model="options.reportID"
            :options="reports"
            :get-option-label="getReportLabel"
            default-value="0"
            :reduce="o => o.reportID"
            @input="handleReportChange"
          />
        </b-form-group>
      </b-col>

      <b-col
        v-if="selectedReport && selectedReport.scenarios && selectedReport.scenarios.length > 1"
      >
        <b-form-group
          :label="$t('report.scenario.label')"
          label-class="text-primary"
        >
          <c-input-select
            v-model="options.scenarioID"
            :options="selectedReport.scenarios"
            default-value="0"
            :reduce="o => o.scenarioID"
          />
        </b-form-group>
      </b-col>
    </b-row>

    <b-form-group
      v-if="selectedReport"
      :label="$t('report.element.label')"
      label-class="text-primary"
    >
      <c-input-select
        v-model="options.elementID"
        :options="allElements"
        :reduce="o => o.elementID"
        default-value="0"
      >
        <template #selected-option="option">
          {{ option.label }}
          <small class="text-muted"> ({{ option.blockLabel }})</small>
        </template>

        <template #option="option">
          {{ option.label }}
          <small class="text-muted"> ({{ option.blockLabel }})</small>
        </template>
      </c-input-select>
    </b-form-group>
  </b-tab>
</template>
<script>
import base from '../base'
import { NoID } from '@cortezaproject/corteza-js'

export default {
  i18nOptions: {
    namespaces: 'block',
  },

  extends: base,

  data () {
    return {
      reports: [],
    }
  },

  computed: {
    selectedReport () {
      const { reportID = NoID } = this.options

      if (reportID !== NoID) {
        return this.reports.find(r => r.reportID === reportID)
      }

      return undefined
    },

    allElements () {
      if (!this.selectedReport) {
        return []
      }

      // Flatten all elements from all blocks
      const elements = []
      this.selectedReport.blocks.forEach(block => {
        const blockLabel = block.title || `${this.$t('general:label.block')} ${block.key}`

        if (block.elements && Array.isArray(block.elements)) {
          block.elements.forEach(element => {
            elements.push({
              elementID: element.elementID,
              blockLabel,
              label: element.name || element.kind,
            })
          })
        }
      })

      return elements
    },
  },

  watch: {
    'options.blockID' () {
      this.options.elementID = NoID
    },
  },

  beforeDestroy () {
    this.setDefaultValues()
  },

  created () {
    this.fetchReports()
  },

  methods: {
    fetchReports () {
      this.$SystemAPI.reportList()
        .then(({ set = [] }) => {
          this.reports = set
        })
        .catch(this.toastErrorHandler(this.$t('notification:report.listFetchFailed')))
    },

    setDefaultValues () {
      this.reports = []
    },

    getReportLabel ({ kind, meta = {} } = {}) {
      return meta.name || kind
    },

    handleReportChange () {
      if (this.options.elementID) {
        this.options.elementID = NoID
        this.options.scenarioID = NoID
      }
    },
  },
}
</script>
