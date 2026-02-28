import MarkdownIt from 'markdown-it'
import Prism from 'prismjs'

import 'prismjs/components/prism-go'
import 'prismjs/components/prism-javascript'
import 'prismjs/components/prism-typescript'
import 'prismjs/components/prism-python'
import 'prismjs/components/prism-bash'
import 'prismjs/components/prism-json'
import 'prismjs/components/prism-yaml'
import 'prismjs/components/prism-sql'

const md = new MarkdownIt({
  html: true,
  linkify: true,
  highlight(str: string, lang: string) {
    const grammar = lang && Prism.languages[lang]
    if (grammar) {
      try {
        return `<pre class="language-${lang}"><code>${Prism.highlight(str, grammar, lang)}</code></pre>`
      } catch {
        // fallthrough
      }
    }
    return `<pre class="language-text"><code>${md.utils.escapeHtml(str)}</code></pre>`
  },
})

export function renderMarkdown(text: string): string {
  if (!text) return ''
  return md.render(text)
}

export default md
