<template>
  <b-overlay>
    <b-card-header class="border-bottom">
      <div class="d-flex flex-wrap align-items-center justify-content-between gap-2">
        <h5
          class="text-primary text-capitalize text-truncate mb-0"
        >
          {{ hit.value.name || hit.value.slug }}
        </h5>

        <span class="text-nowrap">
          <b-avatar
            size="sm"
            icon="code-square"
            class="align-center bg-light text-dark"
          />
          {{ $t('types.namespace') }}
        </span>
      </div>
      <div class="d-flex justify-content-between small">
        <slot name="header" />
      </div>
    </b-card-header>

    <b-card-body>
      <div
        class="d-flex flex-wrap"
        style="gap: 2rem;"
      >
        <b-form-group
          v-for="(item, name, i) in limitData()"
          :key="i"
          :label="name"
          label-class="text-capitalize text-primary"
          class="mb-0"
          style="min-width: 20rem; max-width: 100%;"
        >
          {{ item }}
        </b-form-group>
      </div>
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
        'namespace',
      ]
    },
  },
}
</script>
