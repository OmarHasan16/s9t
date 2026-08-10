<template>
  <b-overlay>
    <b-card-header class="border-bottom">
      <div class="d-flex flex-wrap align-items-center justify-content-between gap-2">
        <h5
          class="text-primary text-capitalize text-truncate mb-0"
        >
          <span
            v-if="hit.value.namespace.name || hit.value.namespace.handle"
          >
            {{ hit.value.namespace.name || hit.value.namespace.handle }}
          </span>
          <span
            v-if="hit.value.namespace.name || hit.value.namespace.handle"
            class="mx-2"
          >
            <b-icon
              icon="arrow-right"
              font-scale="1"
            />
          </span>
          <span>
            {{ hit.value.name || hit.value.handle }}
          </span>
        </h5>

        <span class="text-nowrap">
          <b-badge
            v-if="Object.keys(hit.value.labels || { }).includes('federation')"
            variant="light"
            class="mr-1 h5 p-2 mb-0"
          >
            {{ $t('general:federated') }}
          </b-badge>
          <b-avatar
            size="sm"
            icon="list-ul"
            class="align-center bg-light text-dark"
          />
          {{ $t('types.module') }}
        </span>
      </div>

      <div class="d-flex justify-content-between small">
        <slot name="header" />
      </div>
    </b-card-header>

    <b-card-body
      class="d-flex flex-wrap"
      style="gap: 2rem;"
    >
      <b-form-group
        v-for="(value, name, i) in limitData()"
        :key="i"
        :label="name"
        label-class="text-capitalize text-primary"
        class="mb-0"
        style="min-width: 20rem; max-width: 100%;"
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
        'module',
      ]
    },
  },
}
</script>
