<script lang="js">
import { getComponent } from './index.js'

export default {
  inheritAttrs: false,

  props: {
    inline: {
      type: Boolean,
      default: false,
    },
  },

  computed: {
    previewType () {
      return getComponent({ type: this.$attrs.mime, src: this.$attrs.src, name: this.$attrs.name })
    },
  },
}
</script>
