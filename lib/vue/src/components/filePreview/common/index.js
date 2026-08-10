import mime from 'mime'

const types = [
  { type: 'application/pdf', component: 'PDF' },
  { type: 'image/', component: 'IMG' },
]

/**
 * Tells what component (if any) can preview the given file
 * @param {String|undefined} type pre defined mime type of object
 * @param {String|undefined} src object's src
 * @param {String|undefined} name object's name
 * @returns {String|undefined} preview component or undefined
 */
export const getComponent = ({ type, src, name }) => {
  const srcType = type || mime.getType(src) || mime.getType(name)
  if (!srcType) {
    return
  }

  for (const { type, component } of types) {
    if (srcType.indexOf(type) >= 0) {
      return component
    }
  }

  return 'NoPreview'
}

/**
 * Tells if we support the given file type preview
 * @param {String|undefined} type pre defined mime type of object
 * @param {String|undefined} src object's src
 * @param {String|undefined} name object's name
 * @returns {Boolean} if file can be previewed
 */
export const canPreview = ({ type, src, name }) => {
  return !!getComponent({ type, src, name })
}

export const getExtensionIconType = (ext) => {
  switch (ext) {
    case 'odt':
    case 'doc':
    case 'docx':
      return 'word'
    case 'pdf':
      return 'pdf'
    case 'ppt':
    case 'pptx':
      return 'powerpoint'
    case 'zip':
    case 'rar':
      return 'archive'
    case 'xls':
    case 'xlsx':
    case 'csv':
      return 'excel'
    case 'mov':
    case 'mp3':
    case 'mp4':
      return 'video'
    case 'png':
    case 'jpg':
    case 'jpeg':
      return 'image'
    default: return 'alt'
  }
}
