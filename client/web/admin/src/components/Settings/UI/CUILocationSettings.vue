<template>
  <b-card
    header-class="border-bottom"
    footer-class="border-top d-flex flex-wrap flex-fill-child gap-1"
    class="shadow-sm"
  >
    <template #header>
      <h4 class="m-0">
        {{ $t('geosearch.title') }}
      </h4>
    </template>

    <b-form
      @submit.prevent="$emit('submit', settings)"
    >
      <b-form-group
        :label="$t('geosearch.provider.label')"
        :description="$t('geosearch.provider.description')"
        label-class="text-primary"
      >
        <c-input-select
          v-model="locationSettings.geoSearchProvider"
          :options="providerOptions"
          :reduce="o => o.value"
          label="text"
          :clearable="false"
        />
      </b-form-group>

      <b-form-group
        v-if="requiresApiKey"
        :label="$t('geosearch.apiKey.label')"
        label-class="text-primary"
      >
        <b-form-input
          v-model="locationSettings.geoSearchApiKey"
          type="text"
          :placeholder="$t('geosearch.apiKey.placeholder')"
        />
      </b-form-group>
    </b-form>

    <template #footer>
      <c-button-submit
        v-if="canManage"
        :processing="processing"
        :success="success"
        :disabled="isSubmitDisabled"
        :text="$t('admin:general.label.submit')"
        class="ml-auto"
        @submit="onSubmit"
      />
    </template>
  </b-card>
</template>

<script>
import { components } from '@cortezaproject/corteza-vue'

const { CInputSelect } = components

export default {
  name: 'CUILocationSettings',

  components: {
    CInputSelect,
  },

  i18nOptions: {
    namespaces: 'ui.settings',
    keyPrefix: 'editor.location',
  },

  props: {
    settings: {
      type: Object,
      required: true,
    },

    processing: {
      type: Boolean,
      value: false,
    },

    success: {
      type: Boolean,
      value: false,
    },

    canManage: {
      type: Boolean,
      required: true,
    },
  },

  data () {
    return {
      locationSettings: {},

      providerOptions: [
        { value: 'openstreetmap', text: 'OpenStreetMap', requiresApiKey: false },
        { value: 'opencage', text: 'OpenCage', requiresApiKey: true },
        { value: 'esri', text: 'Esri', requiresApiKey: false },
        { value: 'geoapify', text: 'Geoapify', requiresApiKey: true },
        { value: 'geocodeearth', text: 'Geocode Earth', requiresApiKey: true },
        { value: 'google', text: 'Google Maps', requiresApiKey: true },
        { value: 'locationiq', text: 'LocationIQ', requiresApiKey: true },
        { value: 'mapbox', text: 'Mapbox', requiresApiKey: true },
        { value: 'pelias', text: 'Pelias', requiresApiKey: true },
      ],
    }
  },

  computed: {
    selectedProvider () {
      return this.providerOptions.find(p => p.value === this.locationSettings.geoSearchProvider)
    },

    requiresApiKey () {
      return this.selectedProvider?.requiresApiKey || false
    },

    isSubmitDisabled () {
      return this.requiresApiKey && !this.locationSettings.geoSearchApiKey
    },
  },

  watch: {
    settings: {
      immediate: true,
      handler (settings) {
        this.locationSettings = settings['ui.location'] || {}

        if (!this.locationSettings.geoSearchProvider) {
          this.$set(this.locationSettings, 'geoSearchProvider', 'openstreetmap')
        }

        if (!this.locationSettings.geoSearchApiKey) {
          this.$set(this.locationSettings, 'geoSearchApiKey', '')
        }
      },
    },

    'locationSettings.geoSearchProvider' (newProvider) {
      this.locationSettings.geoSearchApiKey = ''
    },
  },

  methods: {
    onSubmit () {
      this.$emit('submit', { 'ui.location': this.locationSettings })
    },
  },
}
</script>
