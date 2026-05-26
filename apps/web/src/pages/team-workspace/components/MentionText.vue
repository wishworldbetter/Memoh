<template>
  <span class="whitespace-pre-wrap break-words text-sm leading-6">
    <template
      v-for="(segment, index) in segments"
      :key="index"
    >
      <span v-if="segment.type === 'text'">{{ segment.text }}</span>
      <button
        v-else-if="segment.member"
        type="button"
        class="mx-0.5 inline-flex translate-y-0.5 items-center gap-1 rounded-full border bg-muted px-1.5 py-0.5 text-xs font-medium text-foreground hover:bg-accent"
        @click="$emit('open', segment.member)"
      >
        <Avatar class="size-4">
          <AvatarImage
            v-if="memberAvatar(segment.member)"
            :src="memberAvatar(segment.member)"
            :alt="memberLabel(segment.member)"
          />
          <AvatarFallback class="text-[8px]">
            {{ initials(memberLabel(segment.member)) }}
          </AvatarFallback>
        </Avatar>
        <span>@{{ memberLabel(segment.member) }}</span>
      </button>
      <span
        v-else
        class="mx-0.5 rounded-full bg-primary/10 px-1.5 py-0.5 text-xs font-medium text-primary"
      >
        @{{ segment.name }}
      </span>
    </template>
  </span>
</template>

<script setup lang="ts">
import type {
  BotsBot,
  HandlersMemberResponse,
} from '@memohai/sdk'
import { computed } from 'vue'
import {
  Avatar,
  AvatarFallback,
  AvatarImage,
} from '@memohai/ui'

type MentionSegment =
  | { type: 'text', text: string }
  | { type: 'mention', name: string, member?: HandlersMemberResponse }

const props = defineProps<{
  content?: string
  members: HandlersMemberResponse[]
  bots: BotsBot[]
}>()

defineEmits<{
  open: [member: HandlersMemberResponse]
}>()

const segments = computed(() => mentionSegments(props.content))

function mentionSegments(content: string | undefined): MentionSegment[] {
  const text = content ?? ''
  const segments: MentionSegment[] = []
  const pattern = /@"([^"]+)"|@([\p{L}\p{N}_\-.]+)/gu
  let cursor = 0
  for (const match of text.matchAll(pattern)) {
    const index = match.index ?? 0
    if (index > cursor) {
      segments.push({ type: 'text', text: text.slice(cursor, index) })
    }
    const name = match[1] || match[2] || ''
    segments.push({ type: 'mention', name, member: findMentionMember(name) })
    cursor = index + match[0].length
  }
  if (cursor < text.length) {
    segments.push({ type: 'text', text: text.slice(cursor) })
  }
  return segments.length > 0 ? segments : [{ type: 'text', text }]
}

function findMentionMember(name: string) {
  const key = normalizeMentionName(name)
  return props.members.find((member) => {
    const candidates = [
      member.display_name,
      member.bot_id,
      member.user_id,
    ].filter(Boolean).map((value) => normalizeMentionName(value ?? ''))
    return candidates.includes(key)
  })
}

function normalizeMentionName(value: string) {
  return value.trim().toLowerCase()
}

function memberLabel(member: HandlersMemberResponse) {
  return member.display_name || member.bot_id || member.user_id || 'Unknown'
}

function memberAvatar(member: HandlersMemberResponse) {
  if (!member.bot_id) return ''
  return props.bots.find((bot) => bot.id === member.bot_id)?.avatar_url ?? ''
}

function initials(name: string): string {
  return name
    .split(/[\s_-]+/)
    .filter(Boolean)
    .slice(0, 2)
    .map((word) => word[0])
    .join('')
    .toUpperCase() || '?'
}
</script>
