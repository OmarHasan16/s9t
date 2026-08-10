<template>
  <b-overlay>
    <b-card-header class="border-bottom">
      <div class="d-flex align-items-center justify-content-between">
        <h5
          class="text-primary text-capitalize text-truncate mr-2 mb-0"
        >
          {{ $t('types.user') }}
        </h5>

        <b-avatar
          v-b-tooltip.noninteractive.hover="{ title: $t('types.user'), boundary: 'body' }"
          size="sm"
          class="align-center bg-light text-dark"
        />
      </div>

      <div class="d-flex justify-content-between small">
        <slot name="header" />
      </div>
    </b-card-header>

    <b-card-body class="pb-0">
      <b-form-group
        v-for="(value, name, i) in limitData()"
        :key="i"
        :label="name"
        label-class="text-capitalize text-primary"
        class="mb-0"
        style="min-width: 200px; max-width: 100%;"
      >
        {{ value }}
      </b-form-group>
    </b-card-body>
  </b-overlay>
</template>

<script>
import base from './base'

export default {
  i18nOptions: {
    namespaces: 'filters',
  },

  extends: base,

  computed: {
    blacklistedFields () {
      return [
        ...this.defaultBlacklistedFields,
        'meta',
        'fields',
        'namespace',
        'labels',
      ]
    },
  },
}
</script>
