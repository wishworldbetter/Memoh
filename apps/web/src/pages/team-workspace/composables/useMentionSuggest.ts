import type { BotsBot, HandlersMemberResponse } from '@memohai/sdk'
import type { Ref } from 'vue'
import {
  computed,
  nextTick,
  onBeforeUnmount,
  onMounted,
  reactive,
  ref,
  watch,
} from 'vue'

// 与 MentionMarkdown 的渲染规则保持一致：
//  - `@unquoted`：Unicode letter/digit/_/-/.
//  - `@"…"`：引号包裹支持空格
// `@` 前必须是开头或非字母数字，避免误吞 email 等。
const MENTION_TRIGGER_RE = /(?:^|[^\p{L}\p{N}])(@(?:"([^"\n]*)|([\p{L}\p{N}_\-.]*)))$/u

export interface UseMentionSuggestOptions {
  getTextarea: () => HTMLTextAreaElement | null
  emitUpdate: (next: string) => void
  members: Ref<HandlersMemberResponse[]>
  bots: Ref<BotsBot[]>
  limit?: number
  /** Defaults to true. Set false when the textarea is rendered conditionally
   * and the caller wants to bind/unbind manually around v-if mount cycles. */
  autoBind?: boolean
}

export interface CaretCoords {
  top: number
  left: number
  lineHeight: number
}

export function useMentionSuggest(options: UseMentionSuggestOptions) {
  const limit = options.limit ?? 8
  const autoBind = options.autoBind ?? true
  const suggestion = reactive({
    open: false,
    triggerStart: 0,
    query: '',
    activeIndex: 0,
  })

  const caret = ref<CaretCoords>({ top: 0, left: 0, lineHeight: 20 })

  const candidates = computed<HandlersMemberResponse[]>(() => {
    if (!suggestion.open) return []
    const q = suggestion.query.trim().toLowerCase()
    const pool = options.members.value.filter((m) => {
      if (m.member_type === 'bot') return !!m.bot_id
      if (m.member_type === 'user') return !!m.user_id
      return true
    })
    if (!q) return pool.slice(0, limit)
    return pool
      .filter((m) => [m.display_name, m.bot_id, m.user_id]
        .filter((v): v is string => Boolean(v))
        .map((v) => v.toLowerCase())
        .some((token) => token.includes(q)))
      .slice(0, limit)
  })

  watch(candidates, (next) => {
    if (suggestion.activeIndex >= next.length) {
      suggestion.activeIndex = 0
    }
  })

  function detect() {
    const el = options.getTextarea()
    if (!el) return
    if (options.members.value.length === 0) {
      suggestion.open = false
      return
    }
    const pos = el.selectionStart ?? 0
    const before = el.value.slice(0, pos)
    const match = before.match(MENTION_TRIGGER_RE)
    if (!match) {
      suggestion.open = false
      return
    }
    const fullMention = match[1] ?? ''
    if (fullMention.startsWith('@"') && fullMention.length > 2 && fullMention.endsWith('"')) {
      suggestion.open = false
      return
    }
    const triggerStart = (match.index ?? 0) + match[0].length - fullMention.length
    suggestion.triggerStart = triggerStart
    suggestion.query = match[2] ?? match[3] ?? ''
    suggestion.activeIndex = 0
    suggestion.open = true
    updateCaret(el, triggerStart)
  }

  // 计算 caret 在视口坐标系下的位置（用于 fixed/Teleport 弹层）。按光标
  // 前的换行数 × 行高估算行号；不处理软折行，对 mention 输入场景足够。
  function updateCaret(el: HTMLTextAreaElement, pos: number) {
    const style = window.getComputedStyle(el)
    const lineHeightPx = parseFloat(style.lineHeight)
    const fontSizePx = parseFloat(style.fontSize) || 14
    const lineHeight = Number.isFinite(lineHeightPx) && lineHeightPx > 0
      ? lineHeightPx
      : fontSizePx * 1.5
    const paddingTop = parseFloat(style.paddingTop) || 0
    const paddingLeft = parseFloat(style.paddingLeft) || 0
    const before = el.value.substring(0, pos)
    const newlineCount = (before.match(/\n/g) || []).length
    const rect = el.getBoundingClientRect()
    const localTop = paddingTop + newlineCount * lineHeight - el.scrollTop
    // 不超出 textarea 可视范围。
    const clampedLocalTop = Math.min(
      Math.max(0, localTop),
      Math.max(0, el.clientHeight - lineHeight),
    )
    const top = rect.top + clampedLocalTop
    const left = rect.left + Math.max(0, paddingLeft - el.scrollLeft)
    caret.value = { top, left, lineHeight }
  }

  function onInput() {
    detect()
  }

  function onKeydown(ev: KeyboardEvent) {
    if (!suggestion.open || candidates.value.length === 0) return
    if (ev.key === 'ArrowDown') {
      ev.preventDefault()
      suggestion.activeIndex = (suggestion.activeIndex + 1) % candidates.value.length
    }
    else if (ev.key === 'ArrowUp') {
      ev.preventDefault()
      suggestion.activeIndex = (suggestion.activeIndex - 1 + candidates.value.length) % candidates.value.length
    }
    else if (ev.key === 'Enter' || ev.key === 'Tab') {
      const member = candidates.value[suggestion.activeIndex]
      if (member) {
        ev.preventDefault()
        applySelection(member)
      }
    }
    else if (ev.key === 'Escape') {
      ev.preventDefault()
      suggestion.open = false
    }
  }

  function onBlur() {
    // 延迟关闭，确保候选项 click（mousedown.prevent 后）仍能触发。
    setTimeout(() => {
      suggestion.open = false
    }, 120)
  }

  function applySelection(member: HandlersMemberResponse) {
    const el = options.getTextarea()
    if (!el) return
    const name = memberLabel(member)
    const replacement = (/\s/.test(name) ? `@"${name}"` : `@${name}`) + ' '
    const text = el.value
    const before = text.slice(0, suggestion.triggerStart)
    const cursor = el.selectionStart ?? text.length
    const after = text.slice(cursor)
    const next = before + replacement + after
    options.emitUpdate(next)
    suggestion.open = false
    nextTick(() => {
      const updated = options.getTextarea()
      if (!updated) return
      const caret = before.length + replacement.length
      updated.focus()
      updated.setSelectionRange(caret, caret)
    })
  }

  function memberLabel(member: HandlersMemberResponse): string {
    return member.display_name || member.bot_id || member.user_id || 'member'
  }

  function memberAvatar(member: HandlersMemberResponse): string {
    if (member.bot_id) {
      return options.bots.value.find((bot) => bot.id === member.bot_id)?.avatar_url ?? ''
    }
    return ''
  }

  function memberInitials(member: HandlersMemberResponse): string {
    const label = memberLabel(member)
    return label
      .split(/[\s_-]+/)
      .filter(Boolean)
      .slice(0, 2)
      .map((word) => word[0])
      .join('')
      .toUpperCase() || '?'
  }

  function bindTextarea() {
    const el = options.getTextarea()
    if (!el) return
    el.addEventListener('input', onInput)
    el.addEventListener('keydown', onKeydown)
    el.addEventListener('blur', onBlur)
    el.addEventListener('click', onInput)
  }

  function unbindTextarea() {
    const el = options.getTextarea()
    if (!el) return
    el.removeEventListener('input', onInput)
    el.removeEventListener('keydown', onKeydown)
    el.removeEventListener('blur', onBlur)
    el.removeEventListener('click', onInput)
  }

  if (autoBind) {
    onMounted(() => {
      bindTextarea()
    })

    onBeforeUnmount(() => {
      unbindTextarea()
    })
  }
  else {
    onBeforeUnmount(() => {
      unbindTextarea()
    })
  }

  return {
    suggestion,
    candidates,
    caret,
    applySelection,
    memberLabel,
    memberAvatar,
    memberInitials,
    bindTextarea,
    unbindTextarea,
  }
}
