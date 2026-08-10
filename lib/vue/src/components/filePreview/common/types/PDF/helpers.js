/* eslint-disable @typescript-eslint/explicit-function-return-type */
/**
 * Represents a PDF's page + system metadata
 */
export class Page {
  constructor (params) {
    this.merge(params)
  }

  merge ({ index, page, loading, loaded, rendered, failed, node }) {
    this.index = index
    this.page = page
    this.loading = loading
    this.loaded = loaded
    this.rendered = rendered
    this.failed = failed
    this.node = node
  }
}

/**
 * Represents the document source
 */
export class Document {
  constructor (params) {
    this.merge(params)
  }

  merge ({ src, pdf, loading, loaded, failed, scale }) {
    this.src = src
    this.pdf = pdf
    this.loading = loading
    this.loaded = loaded
    this.failed = failed
    this.scale = scale
  }
}

/**
 * Helper to generate a placeholder page node
 * @param {Object} labels Available labels
 * @returns {Node}
 */
export function makePlaceholder (labels) {
  const node = document.createElement('div')
  node.classList.add('loader')
  if (labels.pageLoading) {
    node.innerHTML = labels.pageLoading
  }
  return node
}

/**
 * Helper to generate a failed page node
 * @param {Object} labels Available labels
 */
export function makeFailedPage (labels) {
  const node = document.createElement('div')
  node.classList.add('failed')
  if (labels.pageLoadFailed) {
    node.innerHTML = labels.pageLoadFailed
  }
  return node
}
