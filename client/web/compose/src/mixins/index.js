import Vue from 'vue'

import { mixins } from '@cortezaproject/corteza-vue'

import resourceTranslations from './resource-translations'
import uiHelpers from './uiHelpers'

Vue.mixin(mixins.toast)
Vue.mixin(resourceTranslations)
Vue.mixin(uiHelpers)
