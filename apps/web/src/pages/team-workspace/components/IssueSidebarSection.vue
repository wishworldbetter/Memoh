<template>
  <section class="rounded-md border bg-background px-3 py-3">
    <div class="mb-2 flex items-center justify-between gap-2">
      <div class="flex min-w-0 items-center gap-1.5">
        <slot name="icon" />
        <h2 class="truncate text-xs font-semibold">
          {{ title }}
        </h2>
      </div>
      <slot name="action" />
    </div>
    <div class="space-y-2 text-xs">
      <slot />
    </div>
  </section>
</template>

<script setup lang="ts">
defineProps<{
  title: string
}>()
</script>
