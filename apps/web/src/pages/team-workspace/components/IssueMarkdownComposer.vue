<template>
  <div class="rounded-md border bg-background">
    <Tabs
      v-model="activeTab"
      class="w-full"
    >
      <div class="flex items-center justify-between gap-2 border-b px-2 py-1.5">
        <TabsList class="h-7 gap-0.5 p-0.5">
          <TabsTrigger
            value="write"
            class="h-6 px-2.5 text-xs"
          >
            {{ writeLabel }}
          </TabsTrigger>
          <TabsTrigger
            value="preview"
            class="h-6 px-2.5 text-xs"
          >
            {{ previewLabel }}
          </TabsTrigger>
        </TabsList>
      </div>

      <TabsContent
        value="write"
        class="m-0 p-2.5"
      >
        <Textarea
          v-model="value"
          :placeholder="placeholder"
          :rows="rows"
          class="min-h-24 border-0 p-0 !text-sm leading-6 shadow-none placeholder:text-sm focus-visible:ring-0"
        />
      </TabsContent>

      <TabsContent
        value="preview"
        class="m-0 min-h-24 p-2.5"
      >
        <MarkdownPreview
          v-if="value.trim()"
          :content="value"
          class="!h-auto !overflow-visible !bg-transparent [&>.prose]:px-0 [&>.prose]:py-0 [&_.markdown-renderer]:text-sm [&_.markdown-renderer]:leading-6 [&_.markdown-renderer_li]:text-sm [&_.markdown-renderer_li]:leading-6 [&_.markdown-renderer_p]:text-sm [&_.markdown-renderer_p]:leading-6 [&_.prose]:text-sm [&_.prose]:leading-6 [&_.prose_li]:text-sm [&_.prose_li]:leading-6 [&_.prose_p]:text-sm [&_.prose_p]:leading-6"
        />
        <div
          v-else
          class="rounded-md border border-dashed px-3 py-6 text-center text-sm text-muted-foreground"
        >
          {{ emptyPreviewLabel }}
        </div>
      </TabsContent>
    </Tabs>

    <div class="flex items-center justify-between gap-2.5 border-t px-2.5 py-1.5">
      <p class="min-w-0 text-xs text-muted-foreground">
        {{ helper }}
      </p>
      <div class="flex shrink-0 items-center gap-2">
        <slot name="secondary-actions" />
        <Button
          size="sm"
          :disabled="disabled || (requireContent && !value.trim())"
          @click="$emit('submit')"
        >
          <Send class="mr-1.5 size-3.5" />
          {{ submitLabel }}
        </Button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { Send } from 'lucide-vue-next'
import {
  Button,
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
  Textarea,
} from '@memohai/ui'
import MarkdownPreview from '@/components/markdown-preview/index.vue'

const props = withDefaults(defineProps<{
  modelValue: string
  placeholder?: string
  helper?: string
  submitLabel: string
  writeLabel: string
  previewLabel: string
  emptyPreviewLabel: string
  disabled?: boolean
  requireContent?: boolean
  rows?: number
}>(), {
  placeholder: '',
  helper: '',
  disabled: false,
  requireContent: true,
  rows: 4,
})

const emit = defineEmits<{
  'update:modelValue': [value: string]
  submit: []
}>()

const activeTab = ref('write')
const value = computed({
  get: () => props.modelValue,
  set: (next: string) => emit('update:modelValue', next),
})
</script>
