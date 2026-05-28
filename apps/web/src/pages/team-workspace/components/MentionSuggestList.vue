<template>
  <Teleport to="body">
    <div
      v-if="open && candidates.length > 0"
      class="fixed z-50 max-h-64 w-72 overflow-y-auto rounded-md border bg-popover text-popover-foreground shadow-lg"
      :style="popoverStyle"
      role="listbox"
      @mousedown.prevent
    >
      <button
        v-for="(member, idx) in candidates"
        :key="member.id || `${member.member_type}-${idx}`"
        type="button"
        role="option"
        :aria-selected="idx === activeIndex"
        class="flex w-full items-center gap-2 px-2 py-1.5 text-left text-xs transition-colors"
        :class="idx === activeIndex ? 'bg-accent text-accent-foreground' : 'hover:bg-accent/60'"
        @mouseenter="$emit('hover', idx)"
        @click="$emit('select', member)"
      >
        <span class="inline-flex size-5 shrink-0 items-center justify-center overflow-hidden rounded-full bg-accent text-[9px] font-medium text-muted-foreground">
          <img
            v-if="memberAvatar(member)"
            :src="memberAvatar(member)"
            :alt="memberLabel(member)"
            class="size-full rounded-full object-cover"
          >
          <template v-else>
            {{ memberInitials(member) }}
          </template>
        </span>
        <span
          class="min-w-0 flex-1 truncate"
          :title="memberLabel(member)"
        >{{ memberLabel(member) }}</span>
        <Badge
          variant="outline"
          class="ml-auto h-4 shrink-0 px-1.5 text-[10px] font-normal"
        >
          {{ member.member_type === 'user' ? $t('principalSelect.kindUser') : $t('principalSelect.kindBot') }}
        </Badge>
      </button>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import type { HandlersMemberResponse } from '@memohai/sdk'
import type { CSSProperties } from 'vue'
import { computed } from 'vue'
import { Badge } from '@memohai/ui'

const props = defineProps<{
  open: boolean
  candidates: HandlersMemberResponse[]
  activeIndex: number
  memberLabel: (member: HandlersMemberResponse) => string
  memberAvatar: (member: HandlersMemberResponse) => string
  memberInitials: (member: HandlersMemberResponse) => string
  /** Caret coords in viewport (page) coordinates; the popover will appear
   * just below this caret using position: fixed. */
  caret?: { top: number, left: number, lineHeight: number } | null
}>()

defineEmits<{
  select: [member: HandlersMemberResponse]
  hover: [index: number]
}>()

const popoverStyle = computed<CSSProperties>(() => {
  if (!props.caret) {
    return { top: '0px', left: '0px', visibility: 'hidden' }
  }
  return {
    top: `${props.caret.top + props.caret.lineHeight + 4}px`,
    left: `${props.caret.left}px`,
  }
})
</script>
