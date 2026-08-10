<template>
  <div>
    <b-button
      :id="popoverTarget"
      v-b-tooltip.noninteractive.hover="{ title: $t('recordList.filter.title'), boundary: 'body' }"
      :variant="variant"
      class="d-flex align-items-center d-print-none border-0 px-1 h-100"
      :class="buttonClass"
      :style="buttonStyle"
      @click.stop
    >
      <font-awesome-icon
        :icon="['fas', 'filter']"
        :class="[inFilter ? 'text-primary' : inactiveIconClass]"
      />
    </b-button>

    <b-popover
      ref="popover"
      custom-class="record-list-filter shadow-sm"
      triggers="click blur"
      placement="bottom"
      delay="0"
      boundary="window"
      boundary-padding="2"
      :target="popoverTarget"
      @hide="onHide"
      @show="onOpen"
    >
      <b-card
        no-body
        class="position-static w-100"
      >
        <b-card-body class="px-3 pb-0 overflow-auto">
          <filter-toolbox
            v-model="componentFilter"
            :module="module"
            :namespace="namespace"
            :selected-field="selectedField"
            @value-change="preventClose"
          />
        </b-card-body>

        <b-card-footer class="d-flex justify-content-between shadow-sm rounded">
          <b-button
            variant="light"
            @click="resetFilter"
          >
            {{ $t('general:label.reset') }}
          </b-button>

          <div class="d-flex">
            <b-button
              v-if="allowFilterPresetSave"
              variant="outline-primary"
              class="mr-2"
              @click="onSave(true, 'filter-preset')"
            >
              {{ $t('recordList.filter.addFilterToPreset') }}
            </b-button>
            <b-button
              ref="btnSave"
              variant="primary"
              @click="onSave"
            >
              {{ $t('general.label.save') }}
            </b-button>
          </div>
        </b-card-footer>
      </b-card>

      <a
        ref="focusMe"
        href=""
        disabled
      />
    </b-popover>
  </div>
</template>

<script>
import FilterToolbox from 'corteza-webapp-compose/src/components/Common/FilterToolbox.vue'

export default {
  i18nOptions: {
    namespaces: 'block',
  },

  components: {
    FilterToolbox,
  },

  props: {
    target: {
      type: String,
      default: '',
    },

    selectedField: {
      type: Object,
      default: undefined,
    },

    namespace: {
      type: Object,
      required: true,
    },

    module: {
      type: Object,
      required: true,
    },

    recordListFilter: {
      type: Array,
      required: true,
    },

    variant: {
      type: String,
      default: 'outline-light',
    },

    inactiveIconClass: {
      type: String,
      default: 'text-secondary',
    },

    buttonClass: {
      type: String,
      default: '',
    },

    buttonStyle: {
      type: String,
      default: '',
    },

    allowFilterPresetSave: {
      type: Boolean,
      default: false,
    },
  },

  data () {
    return {
      componentFilter: [],

      // Used to prevent unwanted closure of popover
      preventPopoverClose: false,
    }
  },

  computed: {
    inFilter () {
      return this.recordListFilter.some(({ filter }) => {
        return filter.some(({ name }) => name === (this.selectedField || {}).name)
      })
    },

    popoverTarget () {
      return `${this.target || '0'}-${(this.selectedField || {}).name}`
    },
  },

  watch: {
    recordListFilter: {
      immediate: true,
      deep: true,
      handler (recordListFilter) {
        this.componentFilter = [...recordListFilter]
      },
    },
  },

  beforeDestroy () {
    this.setDefaultValues()
  },

  methods: {
    onHide (e) {
      if (this.preventPopoverClose) {
        e.preventDefault()
        // Focuses invisible element to refocus popover (problems with closing it otherwise)
        this.$nextTick(() => {
          this.$refs.focusMe.focus()
        })
      }
    },

    onOpen () {
      this.componentFilter = [...this.recordListFilter]
    },

    preventClose () {
      this.preventPopoverClose = true

      setTimeout(() => {
        this.preventPopoverClose = false
      }, 100)
    },

    resetFilter () {
      this.componentFilter = undefined
      this.$emit('reset')
    },

    onSave (close = true, type = 'filter') {
      if (close) {
        this.$refs.popover.$emit('close')
      }

      setTimeout(() => {
        this.$emit(type, this.componentFilter.filter(({ filter = [] }) => filter.filter((f = {}) => !!f.name).length > 0))
      }, 100)
    },

    setDefaultValues () {
      this.componentFilter = []
      this.preventPopoverClose = false
    },
  },
}
</script>

<style lang="scss">
.record-list-filter {
  z-index: 1040;
  max-width: 800px !important;
  opacity: 1 !important;
  border-color: transparent;

  .popover-body {
    display: flex;
    width: 800px;
    min-width: min(99vw, 350px);
    max-width: 99vw;
    max-height: 25rem;
    padding: 0;
    color: var(--black);
    background: var(--white);
    border-radius: 0.25rem;
    opacity: 1 !important;
    box-shadow: 0 3px 48px #00000026;
    font-size: 0.9rem;
  }

  .v-select,
  .field-operator,
  .field-editor {
    min-width: 120px;
  }

  .arrow {
    &::before {
      border-bottom-color: var(--white);
      border-top-color: var(--white);
    }

    &::after {
      border-top-color: var(--white);
    }
  }
}
</style>
