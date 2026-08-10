<script>
export default {
  props: {
    index: {
      type: Number,
      required: true,
    },
    hit: {
      type: Object,
      required: true,
    },
    showMap: {
      type: Boolean,
      required: true,
    },
  },

  data () {
    return {
      defaultBlacklistedFields: ['deleted', 'created', 'updated', 'security', 'catch_all'],
    }
  },

  computed: {
    blacklistedFields () {
      return this.defaultBlacklistedFields
    },

    query () {
      return [this.$route.query.query || '']
    },

    createdBy () {
      const { by } = this.hit.value.created || {}
      return by
    },

    createdAt () {
      const { at } = this.hit.value.created || {}
      return at ? this.$options.filters.locFullDateTime(at) : at
    },

    updatedAt () {
      const { at } = this.hit.value.updated || {}
      return at ? this.$options.filters.locFullDateTime(at) : at
    },
  },

  methods: {
    limitData () {
      const out = {}

      if (this.hit.value) {
        for (const key in this.hit.value) {
          const value = this.hit.value[key]

          if (!!value && this.blacklistedFields.indexOf(key) < 0) {
            out[key] = value
          }
        }
      }

      if (this.createdBy) {
        out.createdBy = this.createdBy
      }

      if (this.createdAt) {
        out.createdAt = this.createdAt
      }

      if (this.updatedAt) {
        out.updatedAt = this.updatedAt
      }

      return out
    },
  },
}
</script>
