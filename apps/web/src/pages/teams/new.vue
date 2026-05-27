<template>
  <section class="relative mx-auto px-4 pt-2 pb-10 lg:px-6 md:pt-4 md:pb-12 max-w-2xl">
    <form
      :aria-busy="isCreateFlowBlocked"
      :class="{ 'pointer-events-none select-none opacity-60': isCreateFlowBlocked }"
      @submit.prevent="handleSubmit"
    >
      <div>
        <h3 class="text-sm font-medium mb-4">
          {{ t('teams.basicInfo') }}
        </h3>
        <div class="flex items-start gap-4">
          <div class="group/avatar relative size-16 shrink-0 rounded-full overflow-hidden cursor-pointer">
            <Avatar class="size-16 rounded-full">
              <AvatarImage
                v-if="form.avatar_url.trim()"
                :src="form.avatar_url.trim()"
                :alt="form.name"
              />
              <AvatarFallback class="text-xl">
                <Users
                  v-if="!avatarFallback"
                  class="size-5"
                />
                <template v-else>
                  {{ avatarFallback }}
                </template>
              </AvatarFallback>
            </Avatar>
            <button
              type="button"
              class="absolute inset-0 flex items-center justify-center rounded-full bg-black/40 opacity-0 transition-opacity group-hover/avatar:opacity-100"
              :title="t('common.edit')"
              :aria-label="t('common.edit')"
              @click="avatarDialogOpen = true"
            >
              <SquarePen class="size-6 text-white" />
            </button>
          </div>
          <div class="flex-1 min-w-0">
            <Label class="mb-2">
              {{ t('teams.name') }}
              <span class="text-destructive">*</span>
            </Label>
            <Input
              v-model="form.name"
              type="text"
              :placeholder="t('teams.namePlaceholder')"
            />
          </div>
        </div>
      </div>

      <Separator class="my-6" />

      <div class="space-y-4">
        <div>
          <Label class="mb-2">{{ t('teams.description') }}</Label>
          <Textarea
            v-model="form.description"
            :placeholder="t('teams.descriptionPlaceholder')"
            rows="3"
            class="text-xs"
          />
        </div>
        <div>
          <Label class="mb-2">{{ t('teams.sharedDir') }}</Label>
          <Input
            v-model="form.shared_dir_name"
            :placeholder="t('teams.sharedDirPlaceholder')"
          />
        </div>
      </div>

      <div class="mt-8 flex items-center justify-end gap-2">
        <Button
          type="button"
          variant="ghost"
          @click="router.back()"
        >
          {{ t('common.cancel') }}
        </Button>
        <Button
          type="submit"
          :disabled="!canSubmit || isCreateFlowBlocked"
        >
          <Spinner
            v-if="submitLoading"
            class="mr-1.5 size-3"
          />
          {{ t('common.save') }}
        </Button>
      </div>
    </form>

    <AvatarEditDialog
      v-model:open="avatarDialogOpen"
      v-model:avatar-url="form.avatar_url"
      :fallback-text="avatarFallback"
      :title="t('teams.editAvatar')"
      :description="t('teams.editAvatarDescription')"
      :placeholder="t('teams.avatarUrlPlaceholder')"
    />
  </section>
</template>

<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useMutation, useQueryCache } from '@pinia/colada'
import { toast } from 'vue-sonner'
import { SquarePen, Users } from 'lucide-vue-next'
import {
  Avatar,
  AvatarFallback,
  AvatarImage,
  Button,
  Input,
  Label,
  Separator,
  Spinner,
  Textarea,
} from '@memohai/ui'
import { postTeams } from '@memohai/sdk'
import AvatarEditDialog from '@/components/avatar-edit-dialog/index.vue'
import { useAvatarInitials } from '@/composables/useAvatarInitials'
import { resolveApiErrorMessage } from '@/utils/api-error'

const router = useRouter()
const { t } = useI18n()
const queryCache = useQueryCache()

const form = reactive({
  name: '',
  description: '',
  shared_dir_name: '',
  avatar_url: '',
})

const avatarDialogOpen = ref(false)
const avatarFallback = useAvatarInitials(() => form.name)

const canSubmit = computed(() => Boolean(form.name.trim()))

const { mutateAsync: createTeam, isLoading: submitLoading } = useMutation({
  mutation: async () => {
    const { data, error } = await postTeams({
      body: {
        name: form.name.trim(),
        description: form.description.trim(),
        avatar_url: form.avatar_url.trim() || undefined,
        shared_dir_name: form.shared_dir_name.trim(),
      },
    })
    if (error) throw error
    return data
  },
  onSettled: () => {
    void queryCache.invalidateQueries({ key: ['teams'] })
  },
})

const isCreateFlowBlocked = computed(() => submitLoading.value)

async function handleSubmit() {
  if (!canSubmit.value || isCreateFlowBlocked.value) return
  try {
    const team = await createTeam()
    toast.success(t('teams.createSuccess'))
    const teamId = team?.id
    if (teamId) {
      router.push({ name: 'team-detail', params: { teamId } })
    }
    else {
      router.push({ name: 'teams' })
    }
  }
  catch (err) {
    toast.error(resolveApiErrorMessage(err, t('teams.createFailed')))
  }
}
</script>
