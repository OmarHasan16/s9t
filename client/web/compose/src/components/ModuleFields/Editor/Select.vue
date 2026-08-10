<template>
  <b-form-group
    :label-cols-md="horizontal && '5'"
    :label-cols-xl="horizontal && '4'"
    :content-cols-md="horizontal && '7'"
    :content-cols-xl="horizontal && '8'"
    :class="formGroupStyleClasses"
  >
    <template
      #label
    >
      <div
        v-if="!valueOnly"
        class="d-flex align-items-center text-primary p-0"
      >
        <span
          :title="label"
          class="d-inline-block mw-100"
        >
          {{ label }}
        </span>

        <c-hint :tooltip="hint" />

        <slot name="tools" />
      </div>
      <div
        class="small text-muted"
        :class="{ 'mb-1': description }"
      >
        {{ description }}
      </div>
    </template>

    <template v-if="field.isMulti">
      <div v-if="field.options.selectType === 'list'">
        <b-form-checkbox
          v-for="option in selectOptions"
          :key="option.value"
          v-model="value"
          :value="option.value"
          class="d-block mb-1"
        >
          <span
            class="pointer"
            :class="{ 'badge badge-pill': field.options.displayType === 'badge' }"
            :style="getOptionStyle(option.value)"
          >
            {{ option.text }}
          </span>
        </b-form-checkbox>

        <errors :errors="errors" />
      </div>

      <multi
        v-else
        :value.sync="value"
        :errors="errors"
        :single-input="field.options.selectType !== 'each'"
      >
        <template #single>
          <c-input-select
            v-if="field.options.selectType === 'default'"
            ref="singleSelect"
            :options="selectOptions"
            :placeholder="$t('kind.select.placeholder')"
            :reduce="o => o.value"
            :selectable="isSelectable"
            label="text"
            :badge="field.options.displayType === 'badge'"
            @input="selectChange"
          />

          <c-input-select
            v-if="field.options.selectType === 'multiple'"
            v-model="value"
            :options="selectOptions"
            :placeholder="$t('kind.select.placeholder')"
            :reduce="o => o.value"
            :selectable="isSelectable"
            label="text"
            multiple
            :badge="field.options.displayType === 'badge'"
          />
        </template>

        <template #default="ctx">
          <c-input-select
            v-if="field.options.selectType === 'each'"
            :value="value[ctx.index]"
            :options="selectOptions"
            :reduce="o => o.value"
            :placeholder="$t('kind.select.placeholder')"
            :selectable="isSelectable"
            label="text"
            :badge="field.options.displayType === 'badge'"
            @input="setMultiValue($event, ctx.index)"
          />

          <span
            v-else
            :class="{ 'badge badge-pill': field.options.displayType === 'badge' }"
            :style="getOptionStyle(value[ctx.index])"
          >
            {{ findLabel(value[ctx.index]) }}
          </span>
        </template>
      </multi>
    </template>

    <template
      v-else
    >
      <b-form-radio-group
        v-if="field.options.selectType === 'list'"
        v-model="value"
        :options="selectOptions"
        stacked
      />

      <c-input-select
        v-else
        v-model="value"
        :placeholder="$t('kind.select.optionNotSelected')"
        :options="selectOptions"
        :reduce="o => o.value"
        :selectable="isSelectable"
        label="text"
        :badge="field.options.displayType === 'badge'"
      />

      <errors :errors="errors" />
    </template>
  </b-form-group>
</template>

<script>
import base from './base'

export default {
  i18nOptions: {
    namespaces: 'field',
  },

  extends: base,

  computed: {
    selectOptions () {
      return this.field.options.options.filter(({ value = '', text = '' }) => value && text)
    },
  },

  methods: {
    selectChange (value) {
      this.value.push(value)
      // Reset select
      this.$refs.singleSelect.localValue = undefined
    },

    /**
     * Helper to resolve a label for a given value
     * @param {String} v Value in question
     * @returns {String}
     */
    findLabel (v) {
      return (this.selectOptions.find(({ value }) => value === v) || {}).text || v
    },

    isSelectable ({ value } = {}) {
      if (this.field.options.selectType === 'list') return true

      if (this.field.isMulti) {
        return !this.field.options.isUniqueMultiValue || !(this.value || []).includes(value)
      } else {
        return this.value !== value
      }
    },

    getOptionStyle (v) {
      const style = {}

      if (this.field.options.displayType === 'badge') {
        const opt = this.selectOptions.find(({ value }) => value === v) || { style: {} }

        style.fontSize = '0.9rem'
        style.color = opt.style.textColor || 'var(--dark)'
        style.backgroundColor = opt.style.backgroundColor || 'var(--extra-light)'
      }

      return style
    },
  },
}
</script>
