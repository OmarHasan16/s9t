<template>
  <b-container
    class="pt-2 pb-3"
  >
    <c-content-header
      :title="$t('title')"
    />

    <c-ui-location-settings
      v-if="settings"
      :settings="settings"
      :processing="location.processing"
      :success="location.success"
      :can-manage="canManage"
      @submit="onSubmit($event, 'location')"
    />
  </b-container>
</template>

<script>
import editorHelpers from 'corteza-webapp-admin/src/mixins/editorHelpers'
import CUILocationSettings from 'corteza-webapp-admin/src/components/Settings/UI/CUILocationSettings'
import { mapGetters } from 'vuex'

export default {
  i18nOptions: {
    namespaces: 'ui.settings',
    keyPrefix: 'editor.location',
  },

  components: {
    'c-ui-location-settings': CUILocationSettings,
  },

  mixins: [
    editorHelpers,
  ],

  data () {
    return {
      settings: undefined,

      location: {
        processing: false,
        success: false,
      },
    }
  },

  computed: {
    ...mapGetters({
      can: 'rbac/can',
    }),

    canManage () {
      return this.can('system/', 'settings.manage')
    },
  },

  created () {
    this.fetchSettings()
  },

  methods: {
    fetchSettings () {
      this.incLoader()

      this.$Settings.fetch()
      return this.$SystemAPI.settingsList({ prefix: 'ui' })
        .then(settings => {
          this.settings = {}

          settings.forEach(({ name, value }) => {
            this.$set(this.settings, name, value)
          })
        })
        .catch(this.toastErrorHandler(this.$t('notification:settings.location.fetch.error')))
        .finally(() => {
          this.decLoader()
        })
    },

    onSubmit (settings, type) {
      this[type].processing = true

      const values = Object.entries(settings).map(([name, value]) => {
        return { name, value }
      })

      this.$SystemAPI.settingsUpdate({ values })
        .then(() => {
          return this.fetchSettings()
        })
        .then(() => {
          this.animateSuccess(type)
          this.toastSuccess(this.$t('notification:settings.location.update.success'))
        })
        .catch(this.toastErrorHandler(this.$t('notification:settings.location.update.error')))
        .finally(() => {
          this[type].processing = false
        })
    },
  },
}
</script>
