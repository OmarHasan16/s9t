<template>
  <b-modal
    v-model="showModal"
    centered
    hide-header
    hide-footer
    no-close-on-backdrop
    no-close-on-esc
    body-class="d-flex flex-column justify-content-center align-items-center gap-2 p-4"
    @hide="stopCountdown"
  >
    <h5>{{ labels.warning(countdownTime) }}</h5>

    <b-button
      variant="primary"
      size="lg"
      @click="extendSession"
    >
      {{ labels.extend }}
    </b-button>
  </b-modal>
</template>

<script>
import { throttle } from 'lodash'

export default {
  name: 'CExtendSession',

  props: {
    timeout: {
      type: Number,
      default: 60,
    },

    labels: {
      type: Object,
      default: () => ({
        extend: 'Extend Session',
        warning: (countdownTime) => `You will be logged out in ${countdownTime} seconds`,
      }),
    },
  },

  data () {
    return {
      countdownTime: null,
      countdownTimer: null,
      lastActivityTime: 0,
      heartbeatInterval: null,
      debouncedUpdateActivity: null,
      showModal: false,
    }
  },

  mounted () {
    this.debouncedUpdateActivity = throttle(this.updateActivity, 5000, { leading: true, trailing: true })
    this.setupActivityListeners()
    this.setupHeartbeatMonitoring()
  },

  beforeDestroy () {
    clearInterval(this.heartbeatInterval)
    clearInterval(this.countdownTimer)
    this.removeActivityListeners()
    this.debouncedUpdateActivity.cancel()
  },

  methods: {
    extendSession () {
      this.showModal = false

      this.$auth.stopAutoLogout().then(() => {
        this.setupHeartbeatMonitoring()
      }).catch((err) => {
        console.error('Failed to stop auto logout', err)
        this.$auth.logout()
      })
    },

    startCountdown (expiresIn) {
      this.countdownTime = expiresIn

      this.countdownTimer = setInterval(() => {
        this.countdownTime--

        if (this.countdownTime <= 0) {
          clearInterval(this.countdownTimer)
          this.$auth.logout()
        }
      }, 1000)
    },

    stopCountdown () {
      if (!this.countdownTimer) {
        return
      }

      clearInterval(this.countdownTimer)
    },

    promptExtendSession () {
      clearInterval(this.heartbeatInterval)

      this.$auth.startAutoLogout().then((expiresIn) => {
        this.startCountdown(expiresIn)
        this.showModal = true
      }).catch((err) => {
        console.error('Failed to start auto logout', err)
        this.$auth.logout()
      })
    },

    setupActivityListeners () {
      // Desktop events
      window.addEventListener('mousemove', this.debouncedUpdateActivity, { passive: true })
      window.addEventListener('keypress', this.debouncedUpdateActivity, { passive: true })
      window.addEventListener('click', this.debouncedUpdateActivity, { passive: true })
      window.addEventListener('scroll', this.debouncedUpdateActivity, { passive: true })

      // Mobile events
      window.addEventListener('touchstart', this.debouncedUpdateActivity, { passive: true })
      window.addEventListener('touchmove', this.debouncedUpdateActivity, { passive: true })
      window.addEventListener('touchend', this.debouncedUpdateActivity, { passive: true })
    },

    removeActivityListeners () {
      // Desktop events
      window.removeEventListener('mousemove', this.debouncedUpdateActivity)
      window.removeEventListener('keypress', this.debouncedUpdateActivity)
      window.removeEventListener('click', this.debouncedUpdateActivity)
      window.removeEventListener('scroll', this.debouncedUpdateActivity)

      // Mobile events
      window.removeEventListener('touchstart', this.debouncedUpdateActivity)
      window.removeEventListener('touchmove', this.debouncedUpdateActivity)
      window.removeEventListener('touchend', this.debouncedUpdateActivity)
    },

    updateActivity () {
      this.lastActivityTime = Date.now()
    },

    setupHeartbeatMonitoring () {
      if (this.timeout === 0) {
        return
      }

      this.lastActivityTime = Date.now()

      this.heartbeatInterval = setInterval(() => {
        const secondsSinceLastActivity = Math.floor((Date.now() - this.lastActivityTime) / 1000)

        if (secondsSinceLastActivity > this.timeout) {
          this.promptExtendSession()
        }
      }, 5000)
    },
  },
}
</script>
