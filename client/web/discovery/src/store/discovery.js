const storeTypes = {
  processing: 'processing',
  resourceTypes: 'resourceTypes',
  aggregations: 'aggregations',
  modules: 'modules',
  namespaces: 'namespaces',
}

export default function (DiscoveryAPI) {
  return {
    namespaced: true,

    state: {
      processing: false,
      resourceTypes: [
        'compose:record',
      ],
      aggregations: [],
      modules: [],
      namespaces: [],
    },

    getters: {
      processing: state => state.processing,
      resourceTypes: state => state.resourceTypes,
      aggregations: state => state.aggregations,
      modules: state => state.modules,
      namespaces: state => state.namespaces,
    },

    actions: {
      async fetchData ({ commit, state }, { query, modules = state.modules, namespaces = state.namespaces, size }) {
        commit(storeTypes.processing, true)

        return DiscoveryAPI.query({ query, modules, namespaces, size, resourceTypes: state.resourceTypes }).then((response = {}) => {
          if (response) {
            commit(storeTypes.aggregations, response.aggregations)
          }

          return response
        }).finally(() => {
          commit(storeTypes.processing, false)
        })
      },

      updateResourceTypes ({ commit }, resourceTypes) {
        commit(storeTypes.resourceTypes, resourceTypes)
      },

      updateModules ({ commit }, modules) {
        commit(storeTypes.modules, modules)
      },

      updateNamespaces ({ commit }, namespaces) {
        commit(storeTypes.namespaces, namespaces)
      },
    },

    mutations: {
      [storeTypes.processing] (state, value) {
        state.processing = value
      },

      [storeTypes.resourceTypes] (state, resourceTypes) {
        state.resourceTypes = resourceTypes
      },

      [storeTypes.aggregations] (state, aggregations) {
        state.aggregations = aggregations
      },

      [storeTypes.modules] (state, modules) {
        state.modules = modules
      },

      [storeTypes.namespaces] (state, namespaces) {
        state.namespaces = namespaces
      },
    },
  }
}
