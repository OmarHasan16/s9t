<template>
  <b-container
    class="pt-2 pb-3"
  >
    <c-content-header
      :title="$t('title')"
    />

    <c-ui-branding-editor
      v-if="settings"
      :settings="settings"
      :processing="branding.processing"
      :success="branding.success"
      :can-manage="canManage"
      @submit="onSubmit($event, 'branding')"
    />
  </b-container>
</template>

<script>
import editorHelpers from 'corteza-webapp-admin/src/mixins/editorHelpers'
import CUIBrandingEditor from 'corteza-webapp-admin/src/components/Settings/UI/CUIBrandingEditor.vue'
import { mapGetters } from 'vuex'

export default {
  i18nOptions: {
    namespaces: 'ui.settings',
    keyPrefix: 'editor.theming',
  },

  components: {
    'c-ui-branding-editor': CUIBrandingEditor,
  },

  mixins: [
    editorHelpers,
  ],

  data () {
    return {
      settings: undefined,

      branding: {
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
        .catch(this.toastErrorHandler(this.$t('notification:settings.theming.fetch.error')))
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
          return this.fetchSettings().then(() => {
            if ((type === 'branding' && this.settings['ui.studio.sass-installed'])) {
              return new Promise((resolve) => {
                setTimeout(() => {
                  const stylesheet = document.querySelector('link#corteza-custom-css')
                  stylesheet.href = 'custom.css?v=' + new Date().getTime()
                  resolve()
                }, 1000)
              })
            }
          })
        })
        .then(() => {
          this.animateSuccess(type)
          this.toastSuccess(this.$t('notification:settings.theming.update.success'))
        })
        .catch(this.toastErrorHandler(this.$t('notification:settings.theming.update.error')))
        .finally(() => {
          this[type].processing = false
        })
    },
  },
}
</script>
