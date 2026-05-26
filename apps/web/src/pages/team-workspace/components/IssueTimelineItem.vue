<template>
  <article class="flex gap-2.5">
    <Avatar class="mt-1 size-7 shrink-0">
      <AvatarImage
        v-if="avatarUrl"
        :src="avatarUrl"
        :alt="authorLabel"
      />
      <AvatarFallback class="text-[9px]">
        {{ initials(authorLabel) }}
      </AvatarFallback>
    </Avatar>

    <div class="min-w-0 flex-1 rounded-md border bg-background">
      <header class="flex min-h-9 items-center justify-between gap-2.5 border-b bg-muted/30 px-2.5 py-1.5">
        <div class="flex min-w-0 flex-wrap items-center gap-x-2 gap-y-1 text-xs text-muted-foreground">
          <span class="font-medium text-foreground">{{ authorLabel }}</span>
          <Badge
            v-if="authorType"
            variant="outline"
            class="h-5 px-1.5 text-[10px]"
          >
            {{ authorType }}
          </Badge>
          <span v-if="authorMeta">{{ authorMeta }}</span>
        </div>
        <slot name="actions" />
      </header>

      <div class="px-2.5 py-2.5">
        <slot />
      </div>
    </div>
  </article>
</template>

<script setup lang="ts">
import {
  Avatar,
  AvatarFallback,
  AvatarImage,
  Badge,
} from '@memohai/ui'

defineProps<{
  authorLabel: string
  avatarUrl?: string
  authorMeta?: string
  authorType?: string
}>()

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
