<template>
  <b-card
    no-body
    class="result shadow-sm h-100"
    :class="{ 'shadow': hovered }"
    @mouseover="onHover"
    @mouseleave="onLeave"
  >
    <a
      v-if="hit.value.url"
      :href="hit.value.url"
      target="_blank"
      rel="noopener noreferrer"
      class="stretched-link"
    />

    <component
      :is="component"
      v-bind="$props"
    />
  </b-card>
</template>

<script>
import base from './base'
import * as Results from './loader'

export default {
  extends: base,

  props: {
    hit: {
      type: Object,
      required: true,
    },
  },

  data () {
    return {
      hovered: false,
    }
  },

  computed: {
    component () {
      const { type } = this.hit
      const resourceType = type.split(':')[1]

      const keys = Object.keys(Results)
      const i = keys.map(c => c.toLocaleLowerCase()).findIndex(c => c === resourceType)

      return Results[keys[i]]
    },
  },

  methods: {
    onHover () {
      this.hovered = true
      this.$emit('hover', this.hit.value.recordID)
    },

    onLeave () {
      this.hovered = false
      this.$emit('hover', undefined)
    },
  },
}
</script>

<style lang="scss" scoped>
.result {
  transition: all 0.3s ease-in;
}
</style>
