<template>
  <div>
    <!-- <div class="my-2">
      <h6 class="text-primary mb-2">
        {{ $t('types.title') }}
      </h6>
      <b-form-checkbox-group
        :checked="storeResourceTypes"
        name="type"
        :disabled="storeProcessing"
        stacked
        class="mt-1"
        @change="updateResourceTypes"
      >
        <b-form-checkbox
          v-for="option in options"
          :key="option.value"
          :value="option.value"
          class="pointer ml-2"
        >
          <b-avatar
            :icon="option.icon"
            size="sm"
            class="align-center bg-light text-dark"
          />
          {{ option.text }}
        </b-form-checkbox>
      </b-form-checkbox-group>
    </div> -->

    <div
      v-for="agg in aggregationOptions"
      :key="agg.resource"
      class="my-2"
    >
      <div
        v-if="agg.items.length"
        class="d-flex justify-content-between align-items-center"
        style="min-height: 25px;"
      >
        <h6
          class="text-primary d-flex mb-0"
        >
          {{ agg.name }}
          <b-badge
            v-if="groups[agg.name].length"
            variant="light"
            pill
            class="ml-1 align-self-center"
          >
            {{ groups[agg.name].length }}
          </b-badge>
        </h6>
        <b-button
          v-if="groups[agg.name].length"
          variant="link"
          class="text-muted p-0 m-0"
          size="sm"
          @click="clearGroup(agg.name)"
        >
          {{ $t('reset') }}
        </b-button>
      </div>

      <b-form-checkbox-group
        v-model="groups[agg.name]"
        stacked
        class="mt-1 ml-2"
        :disabled="storeProcessing"
        @change="updateGroup(agg.name)"
      >
        <b-form-checkbox
          v-for="(resource, i) in agg.items"
          :key="i"
          :value="resource.name"
          class="mb-1"
        >
          <div class="d-flex align-items-center">
            <span class="d-inline-block text-truncate">
              {{ getResourceDisplayName(agg.resource, resource) }}
            </span>

            <!-- <span
              class="pl-3 ml-auto text-muted"
            >
              {{ resource.hits }}
            </span> -->
          </div>
        </b-form-checkbox>
      </b-form-checkbox-group>
    </div>
  </div>
</template>

<script>
import { mapGetters, mapActions } from 'vuex'

export default {
  i18nOptions: {
    namespaces: 'filters',
  },

  data () {
    return {
      groups: {
        Module: [],
        Namespace: [],
      },
    }
  },

  computed: {
    ...mapGetters({
      storeResourceTypes: 'discovery/resourceTypes',
      storeAggregations: 'discovery/aggregations',
      storeModules: 'discovery/modules',
      storeNamespaces: 'discovery/namespaces',
      storeProcessing: 'discovery/processing',
    }),

    options () {
      return [
        { text: this.$t('types.namespace'), value: 'compose:namespace', icon: 'code-square' },
        { text: this.$t('types.module'), value: 'compose:module', icon: 'list-ul' },
        { text: this.$t('types.record'), value: 'compose:record', icon: 'file-earmark-text' },
      ]
    },

    aggregationOptions () {
      let namespaceOptions = this.storeAggregations.find(({ resource }) => resource === 'compose:namespace') || {}
      let moduleOptions = this.storeAggregations.find(({ resource }) => resource === 'compose:module') || {}

      namespaceOptions = {
        resource: 'compose:namespace',
        name: 'Namespace',
        hits: namespaceOptions.hits || 0,
        items: namespaceOptions.resource_name || [],
      }

      // Get all modules that are missing from store aggregations but are in filter
      const missingModuleOptions = this.groups.Module.filter(name => !(moduleOptions.resource_name || []).some(o => o.name === name))
        .map(name => ({ name }))

      moduleOptions = {
        resource: 'compose:module',
        name: 'Module',
        hits: moduleOptions.hits || 0,
        items: [
          ...missingModuleOptions,
          ...(moduleOptions.resource_name || []),
        ],
      }

      return [namespaceOptions, moduleOptions]
    },
  },

  watch: {
    storeNamespaces: {
      immediate: true,
      handler (namespace) {
        this.groups.Namespace = namespace
      },
    },

    storeModules: {
      immediate: true,
      handler (module) {
        this.groups.Module = module
      },
    },
  },

  methods: {
    ...mapActions({
      updateResourceTypes: 'discovery/updateResourceTypes',
    }),

    getResourceDisplayName (type, { name, handle, slug }) {
      if (type === 'compose:namespace') {
        return name || slug || 'Unnamed namespace'
      } else if (type === 'compose:module') {
        return handle || name || 'Unnamed module'
      }
    },

    updateGroup (name) {
      this.$store.dispatch(`discovery/update${name}s`, this.groups[name])
    },

    clearGroup (name) {
      this.groups[name] = []
      this.updateGroup(name)
    },
  },
}
</script>
