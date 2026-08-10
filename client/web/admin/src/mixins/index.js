import Vue from 'vue'

import resourceTranslations from './resource-translations'

import { mixins } from '@cortezaproject/corteza-vue'

import './eventbus'

Vue.mixin(mixins.toast)
Vue.mixin(resourceTranslations)
