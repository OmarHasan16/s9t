<template>
  <b-container
    class="pt-2 pb-3"
  >
    <c-content-header
      :title="$t('title')"
    />

    <c-ui-topbar-settings
      v-if="settings"
      :settings="settings"
      :processing="topbar.processing"
      :success="topbar.success"
      :can-manage="canManage"
      @submit="onSubmit($event, 'topbar')"
    />
  </b-container>
</template>

<script>
import editorHelpers from 'corteza-webapp-admin/src/mixins/editorHelpers'
import CUITopbarSettings from 'corteza-webapp-admin/src/components/Settings/UI/CUITopbarSettings'
import { mapGetters } from 'vuex'

export default {
  i18nOptions: {
    namespaces: 'ui.settings',
    keyPrefix: 'editor.navigation',
  },

  components: {
    'c-ui-topbar-settings': CUITopbarSettings,
  },

  mixins: [
    editorHelpers,
  ],

  data () {
    return {
      settings: undefined,

      topbar: {
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
        .catch(this.toastErrorHandler(this.$t('notification:settings.navigation.fetch.error')))
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
          this.toastSuccess(this.$t('notification:settings.navigation.update.success'))
        })
        .catch(this.toastErrorHandler(this.$t('notification:settings.navigation.update.error')))
        .finally(() => {
          this[type].processing = false
        })
    },
  },
}
</script>
