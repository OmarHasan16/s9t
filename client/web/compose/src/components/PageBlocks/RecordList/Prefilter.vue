<template>
  <c-form-table-wrapper hide-add-button>
    <b-form-group
      :label="$t('recordList.record.prefilterCommand')"
      label-class="text-primary"
      class="m-0"
    >
      <template v-if="textInput">
        <c-input-expression
          v-model="options.prefilter"
          min-height="3.688rem"
          :suggestion-params="recordAutoCompleteParams"
        />

        <i18next
          path="recordList.record.prefilterFootnote"
          tag="small"
          class="text-muted"
        >
          <code>${record.values.fieldName}</code>
          <code>${recordID}</code>
          <code>${ownerID}</code>
          <span><code>${userID}</code>, <code>${user.name}</code></span>
        </i18next>

        <div class="d-flex align-items-center justify-content-end mt-1">
          <b-button
            variant="link"
            size="sm"
            class="text-decoration-none"
            @click="toggleFilterInputType"
          >
            {{ $t('recordList.prefilter.toggleInputType') }}
          </b-button>
        </div>
      </template>

      <template v-else>
        <filter-toolbox
          v-model="filterGroup"
          :module="module"
          :namespace="namespace"
          reset-filter-on-created
          start-empty
        />

        <div class="d-flex align-items-center justify-content-end mt-1 gap-1">
          <b-button
            variant="light"
            size="sm"
            @click="toggleFilterInputType"
          >
            {{ $t('general:label.cancel') }}
          </b-button>

          <b-button
            variant="primary"
            size="sm"
            class="ml-1"
            @click="saveFilter"
          >
            {{ $t('general:label.save') }}
          </b-button>
        </div>
      </template>
    </b-form-group>
  </c-form-table-wrapper>
</template>

<script>
import { components } from '@cortezaproject/corteza-vue'
import { compose } from '@cortezaproject/corteza-js'
import { convertRecordListFilter, queryToFilter } from 'corteza-webapp-compose/src/lib/record-filter.js'
import FilterToolbox from 'corteza-webapp-compose/src/components/Common/FilterToolbox.vue'
import autocomplete from 'corteza-webapp-compose/src/mixins/autocomplete.js'

const { CInputExpression } = components

export default {
  i18nOptions: {
    namespaces: 'block',
  },

  name: 'RecordListConfiguratorPrefilter',

  components: {
    FilterToolbox,
    CInputExpression,
  },

  mixins: [
    autocomplete,
  ],

  props: {
    options: {
      type: Object,
      required: true,
    },

    namespace: {
      type: compose.Namespace,
      required: true,
    },

    module: {
      type: compose.Module,
      required: true,
    },

    record: {
      type: [Object, null],
      required: false,
      default: null,
    },
  },

  data () {
    return {
      textInput: true,
      filterGroup: [],
    }
  },

  computed: {
    recordAutoCompleteParams () {
      return this.processRecordAutoCompleteParams({ operators: true })
    },
  },

  methods: {
    saveFilter (filter) {
      if (filter && filter[0] && !filter[0].filter[0].name) {
        return
      }

      this.options.prefilter = this.parseFilter()
      this.toggleFilterInputType()
    },

    toggleFilterInputType () {
      this.textInput = !this.textInput
      this.filterGroup = []
    },

    getOptionKey ({ name }) {
      return name
    },

    parseFilter () {
      return queryToFilter('', '', [], this.filterGroup.map(group => {
        group.filter = convertRecordListFilter(group.filter)

        return group
      }).filter(({ filter }) => filter.length))
    },
  },
}
</script>
