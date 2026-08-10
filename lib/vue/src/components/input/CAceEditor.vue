<template>
  <div
    ref="container"
    class="position-relative ace-editor-wrapper"
    :class="{ 'resizable': resizable }"
    :style="containerStyle"
  >
    <ace-editor
      ref="aceeditor"
      v-model="editorValue"
      :lang="lang"
      :mode="lang"
      theme="chrome"
      width="100%"
      :height="effectiveHeight"
      :class="{ 'border-0 rounded-0': !border }"
      v-on="$listeners"
      @init="editorInit"
    />

    <b-button
      v-if="showPopout"
      variant="link"
      class="popout position-absolute px-2 py-1 mr-3"
      @click="$emit('open')"
    >
      <font-awesome-icon
        :icon="['fas', 'expand-alt']"
      />
    </b-button>
  </div>
</template>

<script>
import AceEditor from 'vue2-ace-editor'
import { library } from '@fortawesome/fontawesome-svg-core'
import { faExpandAlt } from '@fortawesome/free-solid-svg-icons'

library.add(faExpandAlt)

export default {
  components: {
    AceEditor,
  },

  props: {
    value: {
      type: String,
      default: '',
    },

    lang: {
      type: String,
      default: 'text',
    },

    minHeight: {
      type: String,
      default: '2.35rem',
    },

    showLineNumbers: {
      type: Boolean,
      default: false,
    },

    fontSize: {
      type: String,
      default: '14px',
    },

    border: {
      type: Boolean,
      default: true,
    },

    showPopout: {
      type: Boolean,
      default: false,
    },

    readOnly: {
      type: Boolean,
      default: false,
    },

    autoComplete: {
      type: Boolean,
      default: false,
    },

    highlightActiveLine: {
      type: Boolean,
      default: false,
    },

    showPrintMargin: {
      type: Boolean,
      default: false,
    },

    autoCompleteSuggestions: {
      type: [Array, Object],
      default: () => ([]),
    },

    initExpressions: {
      type: Boolean,
      required: false,
    },

    fontFamily: {
      type: String,
      default: '',
    },

    placeholder: {
      type: String,
      default: '',
    },

    resizable: {
      type: Boolean,
      default: false,
    },
  },

  data () {
    return {
      resizeObserver: null,
      manualHeight: null,
      contentHeight: null,
      programmaticUpdate: false,
      resizeTimeout: null,
    }
  },

  computed: {
    editorValue: {
      get () {
        return this.value
      },

      set (value = '') {
        this.$emit('update:value', value)
      },
    },

    containerStyle () {
      if (this.resizable) {
        return {
          minHeight: this.minHeight,
          height: this.effectiveHeight,
        }
      }
      return {}
    },

    effectiveHeight () {
      if (this.resizable) {
        // Use the larger of manual height, content height, or minimum height
        const manualHeightPx = this.manualHeight ? this.parseHeight(this.manualHeight) : 0
        const contentHeightPx = this.contentHeight ? this.parseHeight(this.contentHeight) : 0
        const minHeightPx = this.parseHeight(this.minHeight)
        
        const maxHeight = Math.max(manualHeightPx, contentHeightPx, minHeightPx)
        return `${maxHeight}px`
      }
      return this.minHeight
    },
  },

  watch: {
    value () {
      if (this.resizable && this.$refs.aceeditor && this.$refs.aceeditor.editor) {
        this.$nextTick(() => {
          this.calculateContentHeight(this.$refs.aceeditor.editor)
        })
      }
    },
  },

  mounted () {
    if (this.resizable) {
      this.setupResizeObserver()
    }
  },

  beforeDestroy () {
    if (this.resizeObserver) {
      this.resizeObserver.disconnect()
    }
    if (this.resizeTimeout) {
      clearTimeout(this.resizeTimeout)
    }
  },

  methods: {
    parseHeight (height) {
      if (typeof height === 'number') return height
      if (typeof height === 'string') {
        // Remove 'px' and parse to number
        return parseFloat(height.replace('px', '').replace('rem', '')) * (height.includes('rem') ? 16 : 1)
      }
      return 0
    },

    calculateContentHeight (editor) {
      if (!editor) return

      const session = editor.session
      const lineCount = session.getLength()
      const lineHeight = editor.renderer.lineHeight || 16
      const padding = 14 // Top and bottom padding
      
      const contentHeight = (lineCount * lineHeight) + padding
      
      // Parse minHeight to ensure we don't go below it
      const minHeightPx = this.parseHeight(this.minHeight)
      const finalHeight = Math.max(contentHeight, minHeightPx)
      
      // Set flag to prevent ResizeObserver from reacting to this change
      this.programmaticUpdate = true
      this.contentHeight = `${finalHeight}px`
      
      // Clear the flag after a short delay
      this.$nextTick(() => {
        setTimeout(() => {
          this.programmaticUpdate = false
        }, 50)
      })
    },

    setupResizeObserver () {
      let lastHeight = null
      let isInitialized = false
      
      this.resizeObserver = new ResizeObserver((entries) => {
        // Skip if we're in the middle of a programmatic update
        if (this.programmaticUpdate) {
          return
        }

        for (const entry of entries) {
          const newHeight = entry.contentRect.height
          
          // Skip the first observation (initialization)
          if (!isInitialized) {
            lastHeight = newHeight
            isInitialized = true
            return
          }
          
          // Only update if height actually changed significantly (more than 2px to avoid rounding issues)
          if (lastHeight !== null && Math.abs(lastHeight - newHeight) > 2) {
            // Clear any pending timeout
            if (this.resizeTimeout) {
              clearTimeout(this.resizeTimeout)
            }
            
            // Debounce the update slightly to prevent flickering
            this.resizeTimeout = setTimeout(() => {
              this.manualHeight = `${newHeight}px`
              this.updateEditorSize()
              lastHeight = newHeight
            }, 10)
          }
        }
      })

      if (this.$refs.container) {
        this.resizeObserver.observe(this.$refs.container)
      }
    },

    updateEditorSize () {
      // Force ace editor to resize
      this.$nextTick(() => {
        if (this.$refs.aceeditor && this.$refs.aceeditor.editor) {
          const editor = this.$refs.aceeditor.editor
          editor.resize()
        }
      })
    },

    editorInit (editor) {
      import('brace/mode/text')
      import('brace/mode/html')
      import('brace/mode/css')
      import('brace/mode/scss')
      import('brace/mode/json')
      import('brace/mode/javascript')

      import('brace/snippets/text')
      import('brace/snippets/html')
      import('brace/snippets/css')
      import('brace/snippets/scss')
      import('brace/snippets/json')
      import('brace/snippets/javascript')

      import('brace/theme/chrome')
      import('brace/ext/language_tools')
      import('brace/ext/emmet')

      editor.setOptions({
        tabSize: 2,
        fontSize: this.fontSize,
        wrap: true,
        indentedSoftWrap: false,
        showPrintMargin: this.showPrintMargin,
        showLineNumbers: this.showLineNumbers,
        showGutter: this.showLineNumbers,
        displayIndentGuides: this.lang !== 'text',
        useWorker: false,
        readOnly: this.readOnly,
        highlightActiveLine: this.highlightActiveLine,
        cursorStyle: 'smooth',

        ...(this.autoComplete && {
          enableBasicAutocompletion: true,
          enableLiveAutocompletion: true,
          enableSnippets: true,
          enableEmmet: true,
        }),

        ...(this.fontFamily && { fontFamily: this.fontFamily }),
        ...(this.fontSize && { fontSize: this.fontSize }),
      })

      editor.on('input', this.updatePlaceholder)
      this.updatePlaceholder(undefined, editor)

      // Calculate initial content height and listen for changes if resizable
      if (this.resizable) {
        // Calculate initial height
        this.$nextTick(() => {
          this.calculateContentHeight(editor)
        })

        // Recalculate on content change
        editor.session.on('change', () => {
          this.calculateContentHeight(editor)
        })
      }

      if (this.initExpressions) {
        this.processExpressionAutoComplete(editor)
      } else if (this.autoComplete) {
        const staticWordCompleter = {
          getCompletions: (editor, session, pos, prefix, callback) => {
            const autoCompleteSuggestions = this.autoCompleteSuggestions
            callback(
              null,
              autoCompleteSuggestions.map(({ caption, value, meta }) => ({
                caption,
                value,
                meta,
              })),
            )
          },
        }

        editor.completers.push(staticWordCompleter)
      }
    },

    updatePlaceholder (_, editor) {
      if (!this.placeholder) return

      const shouldShow = !editor.session.getValue().length
      let node = editor.renderer.emptyMessageNode

      if (!shouldShow && node) {
        editor.renderer.scroller.removeChild(editor.renderer.emptyMessageNode)
        editor.renderer.emptyMessageNode = null
      } else if (shouldShow && !node) {
        node = editor.renderer.emptyMessageNode = document.createElement('div')
        node.textContent = this.placeholder
        node.className = 'ace_placeholder'
        node.style.padding = '7px 10px'
        node.style.position = 'absolute'
        node.style.zIndex = 9
        node.style.opacity = 0.5
        editor.renderer.scroller.appendChild(node)
      }
    },

    processExpressionAutoComplete (editor) {
      // Extract text context from current cursor position
      const getTextContext = (pos) => {
        const session = editor.session
        const line = session.getLine(pos.row)
        const lastSpaceIndex = Math.max(0, line.lastIndexOf(' '))
        const textAfterSpace = line.slice(lastSpaceIndex, pos.column).trim()
        const lastDotIndex = textAfterSpace.lastIndexOf('.')
        const searchTextForCaption = lastDotIndex >= 0 ? textAfterSpace.slice(lastDotIndex + 1) : textAfterSpace
        
        return { line, textAfterSpace, searchTextForCaption }
      }

      // Check if suggestion matches search text
      const matchesSuggestion = (suggestion, textAfterSpace, searchTextForCaption) => {
        const suggestionValue = typeof suggestion === 'string' ? suggestion : suggestion.value
        const suggestionCaption = typeof suggestion === 'string' ? suggestion : suggestion.caption
        return suggestionValue.toLowerCase().startsWith(textAfterSpace.toLowerCase()) ||
               suggestionCaption.toLowerCase().startsWith(searchTextForCaption.toLowerCase())
      }

      // Check and trigger autocomplete if there are matching suggestions
      const checkAndTriggerAutocomplete = () => {
        setTimeout(() => {
          const pos = editor.getCursorPosition()
          const { line, textAfterSpace, searchTextForCaption } = getTextContext(pos)
          
          // Don't trigger after closing braces
          const charBeforeCursor = pos.column > 0 ? line[pos.column - 1] : ''
          if (charBeforeCursor === '}' || charBeforeCursor === ')' || textAfterSpace.length === 0) {
            return
          }
          
          const context = this.getContext(editor, editor.session, pos)
          const suggestions = this.getSuggestionsForContext(context)
          const hasMatches = suggestions.some(s => matchesSuggestion(s, textAfterSpace, searchTextForCaption))
          
          if (hasMatches) {
            editor.execCommand('startAutocomplete')
          }
        }, 10)
      }
      
      const staticWordCompleter = {
        identifierRegexps: [/[${\w]+/],
        getCompletions: (editor, session, pos, prefix, callback) => {
          const context = this.getContext(editor, session, pos)
          const suggestions = this.getSuggestionsForContext(context)
          const { textAfterSpace, searchTextForCaption } = getTextContext(pos)
          
          const filteredSuggestions = suggestions
            .filter(s => matchesSuggestion(s, textAfterSpace, searchTextForCaption))
            .map(suggestion => {
              const caption = typeof suggestion === 'string' ? suggestion : suggestion.caption
              const value = typeof suggestion === 'string' ? suggestion : suggestion.value
              const captionMatch = caption.toLowerCase().indexOf(searchTextForCaption.toLowerCase())
              
              return {
                caption,
                value,
                score: captionMatch === 0 ? 10000 : 1000,
                meta: 'variable',
                completer: {
                  insertMatch: (insertEditor, data) => {
                    insertEditor.jumpToMatching()
                    const line = session.getLine(pos.row)
                    const spaceIndex = line.lastIndexOf(' ')
                    const startCol = spaceIndex > 0 ? spaceIndex + 1 : 0

                    insertEditor.session.replace({
                      start: { row: pos.row, column: startCol },
                      end: { row: pos.row, column: pos.column },
                    }, data.value)
                    
                    checkAndTriggerAutocomplete()
                  },
                },
              }
            })

          callback(null, filteredSuggestions)
        },
      }

      editor.completers = [staticWordCompleter]

      // Only trigger autocomplete automatically when there's a partial match
      editor.commands.on('afterExec', (e) => {
        if (['insertstring', 'Return'].includes(e.command.name) || /^[\w.($]$/.test(e.args)) {
          checkAndTriggerAutocomplete()
        }
      })

      editor.renderer.setScrollMargin(7, 7)
      editor.renderer.setPadding(10)
    },

    getContext (editor, session, pos) {
      const line = session.getLine(pos.row)
      const lastSpaceIndex = Math.max(0, line.lastIndexOf(' '))
      const textBeforeCursor = line.slice(lastSpaceIndex, pos.column).trim()
      return textBeforeCursor.split('.').slice(0, -1).join('.')
    },

    getSuggestionsForContext (context) {
      return this.autoCompleteSuggestions[context] || []
    },
  },
}
</script>

<style lang="scss" scoped>
.popout {
  z-index: 7;
  bottom: 0;
  right: 0;
}

.ace-editor-wrapper {
  &.resizable {
    resize: vertical;
    overflow: auto;
    
    .ace_editor {
      height: 100% !important;
    }
  }
}
</style>

<style lang="scss">
.ace_editor {
  color: var(--black) !important;
  background-color: var(--white) !important;
  border-radius: 0.25rem;
  border: 2px solid var(--extra-light);
  font-size: initial !important;

  .ace_gutter {
    background-color: var(--light) !important;
    color: var(--black) !important;

    .ace_gutter-active-line {
      background-color: var(--extra-light) !important;
    }
  }
  .ace_hidden-cursors {
    .ace_cursor {
      color: var(--white) !important;
    }
  }

  .ace_cursor {
    color: var(--black) !important;
  }
}

.ace_autocomplete {
  border: none !important;
  .ace_active-line {
    background-color: var(--extra-light) !important;
  }

  .ace_line-hover {
    border: none !important;
    background-color: var(--extra-light) !important;
  }

  .ace_completion-highlight {
    color: var(--primary) !important;
  }
}
</style>
