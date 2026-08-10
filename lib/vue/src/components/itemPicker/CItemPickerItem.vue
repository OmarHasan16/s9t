<template>
  <div
    class="d-flex align-items-center"
  >
    <font-awesome-icon
      v-if="!disabled && !disabledSorting && !hideIcons"
      :icon="['fas', 'grip-vertical']"
      :class="{
        'text-muted': disabledDragging,
      }"
      class="align-baseline mr-3 text-primary"
    />
    <b
      class="text-truncate"
    >
      <slot
        v-bind="item"
      >
        {{ item[textField] }}
      </slot>
    </b>
    <b-button
      v-if="_hideIcons"
      v-b-tooltip.noninteractive.hover="{ title: selected ? 'Unselect' : 'Select', boundary: 'body' , delay: 1000 }"
      :data-test-id="`button-${selected ? 'unselect' : 'select'}`"
      variant="outline-light"
      class="d-flex align-items-center ml-auto p-2 border-0"
      @click.prevent.stop="$emit(selected ? 'unselect' : 'select')"
    >
      <font-awesome-icon
        :icon="[selected ? 'far' : 'fas', selected ? 'eye' : 'eye-slash']"
        class="text-muted"
      />
    </b-button>
  </div>
</template>

<script>
export default {
  name: 'CItemPickerItem',

  props: {
    item: {
      type: Object,
      required: true,
    },

    textField: {
      type: String,
      default: 'text',
    },

    selected: {
      type: Boolean,
    },

    disabled: {
      type: Boolean,
    },

    disabledDragging: {
      type: Boolean,
    },

    disabledSorting: {
      type: Boolean,
    },

    hideIcons: {
      type: Boolean,
    },
  },

  computed: {
    _hideIcons () {
      return !this.disabled && !this.hideIcons
    },
  },
}
</script>
