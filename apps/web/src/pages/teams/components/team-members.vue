<template>
  <div class="max-w-2xl mx-auto pb-6 space-y-5">
    <div class="flex items-start justify-between pb-4 border-b border-border/50">
      <div class="space-y-1">
        <h3 class="text-sm font-semibold text-foreground">
          {{ t('teams.tabs.members') }}
        </h3>
        <p class="text-[11px] text-muted-foreground">
          {{ t('teams.membersHint') }}
        </p>
      </div>
      <div class="flex items-center gap-3 shrink-0">
        <Button
          size="sm"
          class="h-8 text-xs font-medium shadow-none"
          @click="openAdd"
        >
          <Plus class="mr-1.5 size-3.5" />
          {{ t('teams.addMember') }}
        </Button>
      </div>
    </div>

    <div class="space-y-4 rounded-md border p-4">
      <div class="space-y-1">
        <h4 class="text-xs font-medium">
          {{ t('teams.members') }}
        </h4>
        <p class="text-[11px] text-muted-foreground">
          {{ t('teams.tabs.membersSubtitle') }}
        </p>
      </div>

      <div
        v-if="members.length === 0"
        class="rounded-md border border-dashed px-4 py-8 text-center text-xs text-muted-foreground"
      >
        {{ t('teams.noMembers') }}
      </div>
      <ul
        v-else
        class="divide-y rounded-md border"
      >
        <li
          v-for="member in members"
          :key="member.id"
          class="flex flex-wrap items-center justify-between gap-3 px-3 py-2.5"
        >
          <div class="flex min-w-0 items-center gap-3">
            <Avatar class="size-8">
              <AvatarFallback class="text-[10px]">
                {{ initials(memberLabel(member)) }}
              </AvatarFallback>
            </Avatar>
            <div class="min-w-0">
              <div class="flex flex-wrap items-center gap-2">
                <span class="truncate text-xs font-medium">{{ memberLabel(member) }}</span>
                <Badge
                  variant="outline"
                  class="h-4 px-1.5 text-[10px] font-normal"
                >
                  {{ member.member_type }}
                </Badge>
                <Badge
                  v-if="member.role"
                  variant="secondary"
                  class="h-4 px-1.5 text-[10px] font-normal"
                >
                  {{ member.role }}
                </Badge>
              </div>
              <p
                v-if="member.instructions"
                class="mt-1 line-clamp-1 text-[11px] text-muted-foreground"
              >
                {{ member.instructions }}
              </p>
            </div>
          </div>
          <div class="flex items-center gap-1">
            <Button
              variant="ghost"
              size="sm"
              class="h-7 px-2 text-xs"
              @click="openEdit(member)"
            >
              <Edit3 class="mr-1.5 size-3" />
              {{ t('common.edit') }}
            </Button>
            <Button
              variant="ghost"
              size="sm"
              class="h-7 px-2 text-xs text-destructive hover:text-destructive"
              @click="handleRemove(member.id)"
            >
              {{ t('common.remove') }}
            </Button>
          </div>
        </li>
      </ul>
    </div>

    <Dialog v-model:open="dialogOpen">
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{{ editingMember ? t('teams.editMember') : t('teams.addMember') }}</DialogTitle>
          <DialogDescription>{{ t('teams.memberDialogHint') }}</DialogDescription>
        </DialogHeader>
        <form
          class="space-y-4"
          @submit.prevent="onSubmit"
        >
          <div
            v-if="!editingMember"
            class="space-y-1.5"
          >
            <Label class="text-xs">{{ t('teams.memberPrincipal') }}</Label>
            <PrincipalSelect
              :model-value="principalId"
              :principal-type="form.member_type"
              :kinds="['bot', 'user']"
              trigger-class="w-full"
              :placeholder="t('teams.memberPrincipalPlaceholder')"
              @update:model-value="setPrincipalId"
              @update:principal-type="setPrincipalType"
            />
          </div>
          <div
            v-if="editingMember"
            class="rounded-md border bg-muted/40 p-3 text-xs"
          >
            <span class="text-muted-foreground">{{ t('teams.member') }}: </span>
            <span class="font-medium">{{ memberLabel(editingMember) }}</span>
          </div>
          <div class="space-y-1.5">
            <Label class="text-xs">{{ t('teams.role') }}</Label>
            <Input
              v-model="form.role"
              :placeholder="t('teams.rolePlaceholder')"
            />
          </div>
          <div class="space-y-1.5">
            <Label class="text-xs">{{ t('teams.memberInstructions') }}</Label>
            <Textarea
              v-model="form.instructions"
              rows="4"
              class="text-xs"
            />
          </div>
          <DialogFooter>
            <Button
              type="button"
              variant="ghost"
              @click="dialogOpen = false"
            >
              {{ t('common.cancel') }}
            </Button>
            <Button
              type="submit"
              :disabled="saving"
            >
              <Spinner
                v-if="saving"
                class="mr-1.5 size-3"
              />
              {{ t('common.save') }}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useMutation, useQuery, useQueryCache } from '@pinia/colada'
import { toast } from 'vue-sonner'
import {
  Avatar,
  AvatarFallback,
  Badge,
  Button,
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  Input,
  Label,
  Spinner,
  Textarea,
} from '@memohai/ui'
import { Edit3, Plus } from 'lucide-vue-next'
import {
  deleteTeamsByTeamIdMembersByMemberId,
  getTeamsByTeamIdMembers,
  postTeamsByTeamIdMembers,
  putTeamsByTeamIdMembersByMemberId,
} from '@memohai/sdk'
import type { HandlersMemberResponse } from '@memohai/sdk'
import PrincipalSelect, { type PrincipalKind } from '@/components/principal-select/index.vue'
import { resolveApiErrorMessage } from '@/utils/api-error'

const props = defineProps<{
  teamId: string
}>()

const { t } = useI18n()
const queryCache = useQueryCache()

const membersKey = computed(() => ['team', props.teamId, 'members'])

const { data: membersData } = useQuery({
  key: () => membersKey.value,
  query: async () => {
    const { data, error } = await getTeamsByTeamIdMembers({ path: { team_id: props.teamId } })
    if (error) throw error
    return data ?? []
  },
  enabled: () => !!props.teamId,
})

const members = computed(() => membersData.value ?? [])

const dialogOpen = ref(false)
const editingMember = ref<HandlersMemberResponse | null>(null)

const form = reactive({
  member_type: 'bot' as PrincipalKind,
  bot_id: '',
  user_id: '',
  role: '',
  instructions: '',
})

const principalId = computed(() =>
  form.member_type === 'user' ? form.user_id : form.bot_id,
)

function setPrincipalId(value: string) {
  if (form.member_type === 'user') {
    form.user_id = value
    form.bot_id = ''
  }
  else {
    form.bot_id = value
    form.user_id = ''
  }
}

function setPrincipalType(value: PrincipalKind | undefined) {
  if (!value) return
  form.member_type = value
  if (value === 'user') form.bot_id = ''
  else form.user_id = ''
}

function resetForm() {
  form.member_type = 'bot'
  form.bot_id = ''
  form.user_id = ''
  form.role = ''
  form.instructions = ''
}

function openAdd() {
  editingMember.value = null
  resetForm()
  dialogOpen.value = true
}

function openEdit(member: HandlersMemberResponse) {
  editingMember.value = member
  form.member_type = member.member_type === 'user' ? 'user' : 'bot'
  form.bot_id = member.bot_id ?? ''
  form.user_id = member.user_id ?? ''
  form.role = member.role ?? ''
  form.instructions = member.instructions ?? ''
  dialogOpen.value = true
}

const { mutateAsync: addMember, isLoading: adding } = useMutation({
  mutation: async () => {
    const { data, error } = await postTeamsByTeamIdMembers({
      path: { team_id: props.teamId },
      body: {
        member_type: form.member_type,
        bot_id: form.bot_id,
        user_id: form.user_id,
        role: form.role.trim(),
        instructions: form.instructions.trim(),
      },
    })
    if (error) throw error
    return data
  },
})

const { mutateAsync: updateMember, isLoading: updating } = useMutation({
  mutation: async (memberId: string) => {
    const { data, error } = await putTeamsByTeamIdMembersByMemberId({
      path: { team_id: props.teamId, member_id: memberId },
      body: {
        role: form.role.trim(),
        instructions: form.instructions.trim(),
      },
    })
    if (error) throw error
    return data
  },
})

const { mutateAsync: removeMember } = useMutation({
  mutation: async (memberId: string) => {
    const { error } = await deleteTeamsByTeamIdMembersByMemberId({
      path: { team_id: props.teamId, member_id: memberId },
    })
    if (error) throw error
  },
})

const saving = computed(() => adding.value || updating.value)

async function onSubmit() {
  try {
    if (editingMember.value?.id) {
      await updateMember(editingMember.value.id)
      toast.success(t('teams.memberUpdated'))
    }
    else {
      await addMember()
      toast.success(t('teams.memberAdded'))
    }
    dialogOpen.value = false
    void queryCache.invalidateQueries({ key: membersKey.value })
  }
  catch (err) {
    toast.error(
      resolveApiErrorMessage(
        err,
        editingMember.value ? t('teams.memberUpdateFailed') : t('teams.memberAddFailed'),
      ),
    )
  }
}

async function handleRemove(id: string | undefined) {
  if (!id) return
  try {
    await removeMember(id)
    toast.success(t('teams.memberRemoved'))
    void queryCache.invalidateQueries({ key: membersKey.value })
  }
  catch (err) {
    toast.error(resolveApiErrorMessage(err, t('teams.memberRemoveFailed')))
  }
}

function memberLabel(member: HandlersMemberResponse) {
  return member.display_name || member.bot_id || member.user_id || t('teams.unknownMember')
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
