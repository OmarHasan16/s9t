<template>
  <div class="d-flex flex-column h-100 w-100">
    <div
      v-if="showSteps"
      class="d-flex flex-column m-5 vh-75"
    >
      <h1 class="display-3">
        {{ $t('label.welcome') }}
      </h1>

      <p class="lead">
        {{ $t('message.noPages') }}
        <span v-if="namespace.canManageNamespace">
          {{ $t('message.startBuilding') }}
        </span>
        <span v-else>
          {{ $t('message.notifyAdministrator') }}
        </span>
      </p>

      <b-container
        v-if="namespace.canManageNamespace"
        fluid="xl"
        class="align-items-center border-top steps"
      >
        <b-row
          align-v="center"
          class="text-center justify-content-between"
        >
          <b-col>
            <circle-step
              step-number="1"
              :done="hasModules"
            >
              <b-button
                v-if="!hasModules"
                data-test-id="button-module-create"
                :disabled="!namespace.canCreateModule"
                variant="outline-primary"
                size="lg"
                :to="{ name: 'admin.modules.create' }"
              >
                {{ $t('step.module.create') }}
              </b-button>
              <router-link
                v-else
                :to="{ name: 'admin.modules' }"
              >
                <b-button
                  data-test-id="button-module-view"
                  :disabled="!namespace.canManageNamespace"
                  variant="primary"
                  size="lg"
                >
                  {{ $t('step.module.view') }}
                </b-button>
              </router-link>
            </circle-step>
          </b-col>
          <b-col>
            <hr>
          </b-col>
          <b-col>
            <circle-step
              :done="hasCharts"
              :disabled="!hasModules"
              optional
            >
              <b-button
                v-if="!hasCharts"
                :disabled="!hasModules || !namespace.canCreateChart"
                variant="outline-primary"
                size="lg"
                :to="{ name: 'admin.charts.create', params: { category: 'generic' } }"
              >
                {{ $t('step.chart.create') }}
              </b-button>
              <router-link
                v-else
                :to="{ name: 'admin.charts' }"
              >
                <b-button
                  :disabled="!namespace.canManageNamespace"
                  variant="primary"
                  size="lg"
                >
                  {{ $t('step.chart.view') }}
                </b-button>
              </router-link>
            </circle-step>
          </b-col>
          <b-col>
            <hr>
          </b-col>
          <b-col>
            <circle-step
              step-number="2"
              :done="hasPages"
            >
              <b-button
                v-if="!hasPages"
                data-test-id="button-page-build"
                :disabled="!namespace.canCreatePage"
                variant="outline-primary"
                size="lg"
                @click="createNewPage"
              >
                {{ $t('step.page.create') }}
              </b-button>
              <router-link
                v-else
                :to="{ name: 'admin.pages' }"
              >
                <b-button
                  :disabled="!namespace.canManageNamespace"
                  data-test-id="button-page-view"
                  variant="primary"
                  size="lg"
                >
                  {{ $t('step.page.view') }}
                </b-button>
              </router-link>
            </circle-step>
          </b-col>
        </b-row>
      </b-container>
    </div>

    <router-view
      v-else
      class="flex-grow-1 overflow-auto"
      :namespace="namespace"
      :page="page"
    />

    <portal-target
      name="toolbar"
    />
  </div>
</template>

<script>
import { mapGetters, mapActions } from 'vuex'
import { compose } from '@cortezaproject/corteza-js'
import CircleStep from 'corteza-webapp-compose/src/components/Common/CircleStep'

const demoPageHandle = 'demo_page'

export default {
  i18nOptions: {
    namespaces: 'onboarding',
  },

  name: 'PublicRoot',

  components: {
    CircleStep,
  },

  props: {
    pageID: {
      type: String,
      required: false,
      default: '',
    },

    namespace: { // via router-view
      type: compose.Namespace,
      required: true,
    },
  },

  data () {
    return {
      navVisible: false,
      documentWidth: 0,
      loaded: false,
    }
  },

  computed: {
    ...mapGetters({
      modules: 'module/set',
      pages: 'page/set',
      charts: 'chart/set',
    }),

    page () {
      return this.$store.getters['page/getByID'](this.pageID) || new compose.Page()
    },

    showSteps () {
      return !this.pageID && this.loaded
    },

    hasModules () {
      return !!this.modules.length
    },

    hasCharts () {
      return !!this.charts.length
    },

    hasPages () {
      return this.pages.filter(p => p.visible || p.handle === demoPageHandle).length > 0
    },
  },

  watch: {
    pageID: {
      immediate: true,
      handler (pageID) {
        // If we redirect to page index, try to find & redirect to a first
        // available public page.
        if (!this.pageID) {
          const { pageID } = this.$store.getters['page/homePage'] || {}
          if (pageID) {
            // Use replace so we don't push to history stack
            this.$router.replace({ name: 'page', params: { pageID } })
          } else {
            document.title = this.namespace.name || this.namespace.slug || this.$t('general:label.app-name.namespace.view')
            this.loaded = true
          }
        }
      },
    },
  },

  created () {
    this.documentWidth = document.body.offsetWidth
    window.onresize = () => {
      this.documentWidth = document.body.offsetWidth
    }
  },

  beforeDestroy () {
    this.setDefaultValues()
  },

  methods: {
    ...mapActions({
      createPage: 'page/create',
      createPageLayout: 'pageLayout/create',
      createChart: 'chart/create',
    }),

    createNewPage () {
      const { namespaceID } = this.namespace
      const newPage = new compose.Page({
        namespaceID,
        title: 'Demo Page',
        handle: demoPageHandle,
        visible: true,
        blocks: [],
      })

      this.createPage(newPage).then(({ pageID, title }) => {
        const pageLayout = new compose.PageLayout({ namespaceID, pageID, handle: 'primary', meta: { title } })
        return this.createPageLayout(pageLayout).then(() => {
          this.$router.push({ name: 'admin.pages.builder', params: { pageID } })
        })
      }).catch(this.toastErrorHandler(this.$t('notification:page.saveFailed')))
    },

    setDefaultValues () {
      this.navVisible = false
      this.documentWidth = 0
      this.loaded = false
    },
  },
}
</script>
<style lang="scss" scoped>
.steps {
  padding: 0;
  padding-top: 20vh;
}
</style>
