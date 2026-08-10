<template>
  <b-container
    fluid
    class="h-100 mh-100 p-0 d-flex flex-column"
  >
    <split
      direction="horizontal"
      :gutter-size="12"
      class="h-100 overflow-hidden"
    >
      <split-area
        :size="map.show ? 70 : 100"
        :min-size="300"
        class="d-flex flex-column"
      >
        <div class="px-3 flex-shrink-0">
          <b-form-group class="mb-0">
            <c-input-search
              :value="query"
              :placeholder="$t('input-placeholder')"
              :autocomplete="'off'"
              :disabled="storeProcessing"
              submittable
              @search="onQuerySubmit"
            />
          </b-form-group>

          <div
            class="d-flex align-items-center px-1 mt-1 mb-2 text-muted"
          >
            <div class="d-flex align-items-center">
              <span
                :class="{ 'discovering': storeProcessing }"
                class="mt-1"
              >
                {{ searchDescription }}
              </span>
            </div>

            <div class="d-flex align-items-center ml-auto gap-2">
              <b-button
                variant="extra-light"
                size="sm"
                class="d-flex align-items-center gap-1 mt-2"
                @click="toggleMap"
              >
                <font-awesome-icon
                  :icon="['fas', 'map-marked-alt']"
                />
                {{ !map.show ? $t('search:show-map') : $t('search:hide-map') }}
              </b-button>

              <div class="d-flex align-items-center">
                <font-awesome-icon
                  :icon="['fas', 'grip-lines']"
                  class="mt-2 mr-1 pointer"
                  :class="{ 'text-primary': viewMode === 'list' }"
                  @click="viewMode = 'list'"
                />

                <b-form-checkbox
                  v-model="viewMode"
                  :value="'grid'"
                  :unchecked-value="'list'"
                  switch
                  class="pointer ml-2"
                />

                <font-awesome-icon
                  :icon="['fas', 'grip-horizontal']"
                  class="mt-2 ml-1 pointer"
                  :class="{ 'text-primary': viewMode === 'grid' }"
                  @click="viewMode = 'grid'"
                />
              </div>
            </div>
          </div>
        </div>

        <div class="d-flex flex-column flex-fill overflow-hidden">
          <div
            v-if="(storeProcessing || !total.actual) && !loadingMore"
            class="d-flex align-items-center justify-content-center w-100 my-5"
          >
            <h5 class="mb-0">
              <b-spinner
                v-if="storeProcessing"
                variant="primary"
                class="p-4"
              />
              <span
                v-else-if="!total.actual"
              >
                {{ $t('no-results') }}
              </span>
            </h5>
          </div>

          <div
            v-else
            class="results d-flex flex-wrap gap-3 px-4 py-3 overflow-auto"
            :class="{ 'list-view': viewMode === 'list' }"
          >
            <div
              v-for="(hit, i) in hits"
              :key="i"
              class="result-item w-100"
              :class="{ 'grid-view': viewMode === 'grid' }"
            >
              <result
                :id="`result-${i}`"
                :index="i"
                :hit="hit"
                :show-map="map.show"
                :class="{ 'border-primary border shadow': map.clickedMarker && map.clickedMarker === i }"
                class="border"
                @hover="onResultHover(i)"
              />
            </div>

            <div
              v-if="total.actual > 0 && total.actual < total.all"
              class="d-flex align-items-center justify-content-center py-3 w-100"
            >
              <b-button
                variant="primary"
                :disabled="loadingMore"
                class="d-flex align-items-center justify-content-center gap-1"
                @click="getSearchData({ append: true })"
              >
                <b-spinner
                  v-if="loadingMore"
                  small
                />
                {{ loadingMore ? $t('search:loading-more') : $t('search:show-more') }}
              </b-button>
            </div>
          </div>
        </div>
      </split-area>

      <split-area
        :size="map.show ? 30 : 0"
        :min-size="300"
      >
        <discovery-map
          v-if="map.show"
          :markers="map.markers"
          :hover-index="map.hoverIndex"
          class="pl-2"
          @marker-clicked="markerClicked"
        />
      </split-area>
    </split>
  </b-container>
</template>

<script>
import { mapGetters, mapActions } from 'vuex'
import Result from './Results'
import DiscoveryMap from './DiscoveryMap.vue'
import { components } from '@cortezaproject/corteza-vue'
import { Split, SplitArea } from 'vue-split-panel'
const { CInputSearch } = components

export default {
  i18nOptions: {
    namespaces: 'search',
  },

  components: {
    Result,
    DiscoveryMap,
    CInputSearch,
    Split,
    SplitArea,
  },

  data () {
    return {
      loadingMore: false,

      query: '',

      hits: [],

      pagination: {
        limit: 50,
        from: 0,
        size: 50,
      },

      total: {
        all: 0,
        actual: 0,
      },

      initial: false,

      map: {
        show: false,
        markers: [],
        clickedMarker: undefined,
        hoverIndex: undefined,
      },

      viewMode: 'list',
    }
  },

  computed: {
    ...mapGetters({
      storeProcessing: 'discovery/processing',
      storeResourceTypes: 'discovery/resourceTypes',
      storeModules: 'discovery/modules',
      storeNamespaces: 'discovery/namespaces',
    }),

    searchDescription () {
      if (this.storeProcessing) {
        return this.$t('discovering')
      }

      if (this.total.all > 0) {
        return this.$t('range', { actual: this.total.actual, all: this.total.all })
      }

      return ''
    },
  },

  watch: {
    storeResourceTypes: {
      handler () {
        if (this.initial) return
        this.pagination.size = this.pagination.limit
        this.getSearchData()
      },
    },

    storeModules: {
      handler () {
        if (this.initial) return
        this.pagination.size = this.pagination.limit
        this.getSearchData()
      },
    },

    storeNamespaces: {
      handler () {
        if (this.initial) return
        this.pagination.size = this.pagination.limit
        this.getSearchData()
      },
    },
  },

  mounted () {
    this.initial = true

    const { query = '', modules, namespaces, resourceTypes, size = 50 } = this.$route.query

    this.query = query
    this.pagination.size = size

    if (namespaces) {
      this.updateNamespaces(Array.isArray(namespaces) ? namespaces : [namespaces])
    }

    if (modules) {
      this.updateModules(Array.isArray(modules) ? modules : [modules])
    }

    if (resourceTypes) {
      this.updateResourceTypes(Array.isArray(resourceTypes) ? resourceTypes : [resourceTypes])
    }

    this.getSearchData()

    setTimeout(() => {
      this.initial = false
    }, 1000)
  },

  methods: {
    ...mapActions({
      fetchData: 'discovery/fetchData',
      updateModules: 'discovery/updateModules',
      updateNamespaces: 'discovery/updateNamespaces',
      updateResourceTypes: 'discovery/updateResourceTypes',
    }),

    getSearchData ({ query = this.query, append = false } = {}) {
      if (append) {
        this.pagination.size += this.pagination.limit
        this.loadingMore = true
      } else {
        this.map.markers = []
        this.hits = []
        this.map.hoverIndex = undefined
      }

      const modules = this.storeModules
      const namespaces = this.storeNamespaces

      const { size } = this.pagination

      this.updateRouteQuery({ query, modules, namespaces, size })

      this.fetchData({ query, modules, namespaces, size }).then((response = {}) => {
        if (response) {
          if (append) {
            this.hits = [...this.hits, ...(response.hits || [])]
          } else {
            this.hits = response.hits || []
          }
          this.total.all = response.total_results || 0
          this.total.actual = this.hits.length

          this.pagination = {
            ...this.pagination,
            from: response.from || 0,
            size: response.size || 0,
          }

          this.getMarkers()
        }
      }).catch(e => {
        this.toastErrorHandler(this.$t('notification:search.failed'))(e)
        this.hits = []
      }).finally(() => {
        this.loadingMore = false
      })
    },

    onQuerySubmit (query) {
      if (!this.storeProcessing) {
        this.query = query
        this.pagination.size = 50
        this.getSearchData()
      }
    },

    getMarkers () {
      const markers = []

      this.hits.forEach(({ type, value }) => {
        if (type === 'compose:record' && Array.isArray(value.values)) {
          const id = value.recordID
          value.values.forEach(({ value = [] }) => {
            const isGeometry = value && value.find(v => {
              return (v !== null ? v : '').toString().includes('{"coordinates":[')
            })

            if (isGeometry) {
              value.forEach(coordinates => {
                coordinates = JSON.parse(coordinates || '{}').coordinates
                if (coordinates && coordinates.length) {
                  markers.push({ id, coordinates })
                }
              })
            }
          })
        }
      })

      this.map.markers = markers
    },

    markerClicked (ID) {
      if (ID) {
        const result = document.getElementById(`result-${ID}`)

        if (result) {
          result.scrollIntoView({
            behavior: 'smooth',
            block: 'center',
          })
        }
      }

      this.map.clickedMarker = ID
      this.map.hoverIndex = undefined
    },

    onResultHover (index) {
      if (this.map.show) {
        this.map.hoverIndex = index
      }
    },

    toggleMap () {
      this.map.show = !this.map.show
    },

    updateRouteQuery ({ query = undefined, modules = [], namespaces = [], size = 0 }) {
      if (JSON.stringify(this.$route.query) !== JSON.stringify({ query, modules, namespaces, size })) {
        this.$router.push({ query: { query: query || undefined, modules, namespaces, size } })
      }
    },
  },
}
</script>

<style lang="scss">
.split .gutter {
  background-color: transparent;
}
</style>

<style lang="scss" scoped>
// https://stackoverflow.com/a/40991531/17926309
.discovering::after {
  display: inline-block;
  animation: discovering steps(1, end) 1s infinite;
  content: '';
}

@keyframes discovering {
  0% { content: ''; }
  25% { content: '.'; }
  50% { content: '..'; }
  75% { content: '...'; }
  100% { content: ''; }
}

.result-item {
  &.grid-view {
    min-width: 30rem;
    flex: 1;
  }
}

.list-view {
  .result-item {
    max-width: 100%;
  }
}

.results {
  flex: 1;
  min-height: 0;
  position: relative;
}
</style>
