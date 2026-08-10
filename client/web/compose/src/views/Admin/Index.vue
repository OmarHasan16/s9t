<template>
  <div
    v-if="namespace.canManageNamespace"
    class="d-flex flex-column w-100 h-100"
  >
    <router-view
      class="flex-grow-1 overflow-auto"
      :namespace="namespace"
    />

    <portal-target
      name="admin-toolbar"
    />
  </div>
</template>

<script>
export default {
  name: 'AdminRoot',

  props: {
    namespace: {
      type: Object,
      required: false,
      default: undefined,
    },
  },

  mounted () {
    if (!this.namespace.canManageNamespace) {
      this.$router.push({ name: 'pages' })
    }
  },
}
</script>
