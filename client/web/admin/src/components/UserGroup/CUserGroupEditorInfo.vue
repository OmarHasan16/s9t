<template>
  <b-card
    data-test-id="card-user-group-info"
    header-class="border-bottom"
    footer-class="border-top d-flex flex-wrap flex-fill-child gap-1"
    class="shadow-sm"
  >
    <b-form
      @submit.prevent="$emit('submit', userGroup)"
    >
      <b-row>
        <b-col
          cols="12"
          lg="6"
        >
          <b-form-group
            :label="$t('meta.short')"
            label-class="text-primary"
          >
            <b-form-input
              v-model="userGroup.meta.short"
              data-test-id="input-name"
              required
            />
          </b-form-group>
        </b-col>

        <b-col
          cols="12"
          lg="6"
        >
          <b-form-group
            :label="$t('handle')"
            :class="{ 'mb-0': !userGroup.userGroupID }"
            label-class="text-primary"
          >
            <b-form-input
              v-model="userGroup.handle"
              data-test-id="input-handle"
              :placeholder="$t('placeholder-handle')"
              :state="handleState"
            />
            <b-form-invalid-feedback :state="handleState">
              {{ $t('invalid-handle-characters') }}
            </b-form-invalid-feedback>
          </b-form-group>
        </b-col>
      </b-row>

      <b-row>
        <b-col
          cols="12"
        >
          <b-form-group
            :label="$t('meta.description')"
            label-class="text-primary"
          >
            <b-form-textarea
              v-model="userGroup.meta.description"
              data-test-id="textarea-description"
            />
          </b-form-group>
        </b-col>
      </b-row>

      <c-system-fields
        :id="userGroup.userGroupID"
        :resource="userGroup"
      />

      <template v-if="!isRoot">
        <hr>

        <div v-if="!isRoot">
          <h5 class="mb-3">
            {{ $t('parents.title') }}
          </h5>

          <c-form-table-wrapper
            :labels="{
              addButton: $t('general:label.add')
            }"
            class="my-3"
            @add-item="addParent"
          >
            <b-table-simple
              v-if="userGroup.config.path"
              borderless
              small
              responsive
            >
              <b-thead>
                <b-tr>
                  <b-th class="text-primary">
                    {{ $t('parents.parent.label') }}
                  </b-th>
                  <b-th class="text-primary">
                    {{ $t('parents.name.label') }}
                  </b-th>
                  <b-th v-if="userGroup.config.path.length > 1" />
                </b-tr>
              </b-thead>
              <b-tbody>
                <b-tr
                  v-for="(parent, i) in userGroup.config.path"
                  :key="i"
                >
                  <b-td
                    class="align-middle"
                    style="min-width: 250px;"
                  >
                    <c-input-user-group
                      v-model="parent.selfID"
                      :placeholder="$t('parents.parent.placeholder')"
                    />
                  </b-td>
                  <b-td
                    class="align-middle"
                    style="min-width: 200px;"
                  >
                    <b-form-input
                      v-model="parent.name"
                      :placeholder="$t('parents.name.placeholder')"
                    />
                  </b-td>
                  <b-td
                    v-if="userGroup.config.path.length > 1"
                    class="text-right align-middle"
                    style="width: 1%"
                  >
                    <c-input-confirm
                      show-icon
                      @confirmed="deleteParent(i)"
                    />
                  </b-td>
                </b-tr>
              </b-tbody>
            </b-table-simple>
          </c-form-table-wrapper>
        </div>
      </template>

      <!--
        include hidden input to enable
        trigger submit event w/ ENTER
      -->
      <input
        type="submit"
        class="d-none"
        :disabled="saveDisabled"
      >
    </b-form>

    <template #header>
      <h4 class="m-0">
        {{ $t('title') }}
      </h4>
    </template>

    <template #footer>
      <c-input-confirm
        v-if="!fresh && userGroup.canDeleteUserGroup"
        :data-test-id="deletedButtonStatusCypressId"
        :text="getDeleteStatus"
        variant="danger"
        size="md"
        @confirmed="$emit('delete')"
      />

      <c-corredor-manual-buttons
        ui-page="user-group/editor"
        ui-slot="infoFooter"
        resource-type="system:user-group"
        default-variant="light"
        @click="dispatchCortezaSystemUserGroupEvent($event, { userGroup })"
      />

      <c-button-submit
        :disabled="saveDisabled"
        :processing="processing"
        :success="success"
        :text="$t('admin:general.label.submit')"
        class="ml-auto"
        @submit="$emit('submit', userGroup)"
      />
    </template>
  </b-card>
</template>

<script>
import { NoID } from '@cortezaproject/corteza-js'
import { handle, components } from '@cortezaproject/corteza-vue'
const { CInputUserGroup } = components

export default {
  name: 'CUserGroupEditorInfo',

  i18nOptions: {
    namespaces: 'system.user-groups',
    keyPrefix: 'editor.info',
  },

  components: {
    CInputUserGroup,
  },

  props: {
    userGroup: {
      type: Object,
      required: true,
    },

    processing: {
      type: Boolean,
      value: false,
    },

    success: {
      type: Boolean,
      value: false,
    },

    canCreate: {
      type: Boolean,
      required: true,
    },
  },

  computed: {
    isRoot () {
      return this.userGroup.isRoot
    },

    getDeleteStatus () {
      return this.userGroup.deletedAt ? this.$t('undelete') : this.$t('delete')
    },

    userGroupID () {
      if (this.userGroup) {
        return this.userGroup.userGroupID
      }

      return undefined
    },

    fresh () {
      return !this.userGroupID || this.userGroupID === NoID
    },

    editable () {
      return this.fresh ? this.canCreate : this.userGroup.canUpdateUserGroup
    },

    nameState () {
      return this.userGroup.meta.short ? null : false
    },

    handleState () {
      return handle.handleState(this.userGroup.handle)
    },

    saveDisabled () {
      return !this.editable || [this.nameState, this.handleState, this.parentState].includes(false)
    },

    parentState () {
      if (this.isRoot) {
        return null
      }

      if (!this.userGroup.config.path || this.userGroup.config.path.length === 0) {
        return false
      }

      return this.userGroup.config.path.every(parent => parent.selfID) ? null : false
    },

    deletedButtonStatusCypressId () {
      return `button-${this.getDeleteStatus.toLowerCase()}`
    },

    suspendButtonStatusCypressId () {
      return `button-${this.getSuspendStatus.toLowerCase()}`
    },
  },

  watch: {
    userGroupID: {
      immediate: true,
      handler () {
        const { config = {} } = this.userGroup || {}
        const { path = [] } = config || {}

        if (!this.isRoot && path.length === 0) {
          this.addParent()
        }
      },
    },
  },

  methods: {
    addParent () {
      this.userGroup.config.path.push({
        selfID: '',
        name: '',
      })
    },

    deleteParent (i) {
      this.userGroup.config.path.splice(i, 1)
    },
  },
}
</script>
