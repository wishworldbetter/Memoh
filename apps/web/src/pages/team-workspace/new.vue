<template>
  <div class="flex h-full flex-col bg-background">
    <header class="border-b">
      <div class="mx-auto flex max-w-[1120px] items-start gap-3 px-4 py-3 lg:px-6">
        <Button
          variant="ghost"
          size="sm"
          class="mt-0.5 h-8 shrink-0 px-2"
          @click="goBack"
        >
          <ArrowLeft class="size-4" />
        </Button>
        <div class="min-w-0 flex-1">
          <h1 class="text-lg font-semibold leading-7">
            {{ t('teams.newIssue') }}
          </h1>
          <p class="text-xs text-muted-foreground">
            {{ team?.name ?? t('teams.loading') }}
          </p>
        </div>
      </div>
    </header>

    <div class="flex-1 overflow-auto">
      <form
        class="mx-auto grid max-w-[1120px] gap-4 px-4 py-4 lg:grid-cols-[minmax(0,1fr)_260px] lg:px-6"
        @submit.prevent="submitIssue"
      >
        <main class="min-w-0 space-y-3">
          <section class="rounded-md border bg-background">
            <div class="border-b bg-muted/30 px-2.5 py-1.5 text-xs font-medium">
              {{ t('teams.issueTitle') }}
            </div>
            <div class="p-2.5">
              <Input
                v-model="issueForm.title"
                :placeholder="t('teams.issueTitle')"
                class="h-9 text-sm font-medium"
                autofocus
                required
              />
            </div>
          </section>

          <IssueMarkdownComposer
            v-model="issueForm.description"
            :placeholder="t('teams.issueDescriptionPlaceholder')"
            :helper="t('teams.commentMentionHint')"
            :submit-label="t('teams.createIssue')"
            :write-label="t('teams.write')"
            :preview-label="t('common.preview')"
            :empty-preview-label="t('teams.previewEmpty')"
            :disabled="creating || !issueForm.title.trim()"
            :require-content="false"
            :rows="7"
            @submit="submitIssue"
          >
            <template #secondary-actions>
              <Button
                type="button"
                variant="ghost"
                size="sm"
                @click="goBack"
              >
                {{ t('common.cancel') }}
              </Button>
            </template>
          </IssueMarkdownComposer>
        </main>

        <aside class="min-w-0">
          <div class="sticky top-3 space-y-3">
            <IssueSidebarSection :title="t('teams.statusLabel')">
              <template #icon>
                <CircleDot class="size-3.5 text-muted-foreground" />
              </template>
              <Select v-model="issueForm.status">
                <SelectTrigger class="h-8">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem
                    v-for="status in ISSUE_COLUMNS"
                    :key="status.value"
                    :value="status.value"
                  >
                    {{ t(status.labelKey) }}
                  </SelectItem>
                </SelectContent>
              </Select>
            </IssueSidebarSection>

            <IssueSidebarSection :title="t('teams.assignee')">
              <template #icon>
                <UserCircle class="size-3.5 text-muted-foreground" />
              </template>
              <Select v-model="issueForm.assignee">
                <SelectTrigger class="h-8">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="none">
                    {{ t('teams.unassigned') }}
                  </SelectItem>
                  <SelectItem
                    v-for="member in assignableMembers"
                    :key="member.id"
                    :value="memberValue(member)"
                  >
                    {{ memberLabel(member) }}
                  </SelectItem>
                </SelectContent>
              </Select>
            </IssueSidebarSection>
          </div>
        </aside>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { HandlersMemberResponse } from '@memohai/sdk'
import { computed, reactive, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useMutation, useQuery, useQueryCache } from '@pinia/colada'
import { toast } from 'vue-sonner'
import { ArrowLeft, CircleDot, UserCircle } from 'lucide-vue-next'
import {
  Button,
  Input,
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@memohai/ui'
import {
  getTeamsByTeamId,
  getTeamsByTeamIdMembers,
  postTeamsByTeamIdIssues,
} from '@memohai/sdk'
import { resolveApiErrorMessage } from '@/utils/api-error'
import IssueMarkdownComposer from './components/IssueMarkdownComposer.vue'
import IssueSidebarSection from './components/IssueSidebarSection.vue'

type IssueStatus = 'backlog' | 'todo' | 'in_progress' | 'blocked' | 'review' | 'done' | 'cancelled'

const ISSUE_COLUMNS: Array<{ value: IssueStatus, labelKey: string }> = [
  { value: 'backlog', labelKey: 'teams.status.backlog' },
  { value: 'todo', labelKey: 'teams.status.todo' },
  { value: 'in_progress', labelKey: 'teams.status.inProgress' },
  { value: 'blocked', labelKey: 'teams.status.blocked' },
  { value: 'review', labelKey: 'teams.status.review' },
  { value: 'done', labelKey: 'teams.status.done' },
  { value: 'cancelled', labelKey: 'teams.status.cancelled' },
]

const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const queryCache = useQueryCache()

const teamId = computed(() => String(route.params.teamId ?? ''))
const issueForm = reactive({
  title: '',
  description: '',
  status: normalizeStatus(String(route.query.status ?? 'todo')),
  assignee: 'none',
})

const { data: teamData } = useQuery({
  key: () => ['team', teamId.value],
  query: async () => {
    const { data, error } = await getTeamsByTeamId({ path: { team_id: teamId.value } })
    if (error) throw error
    return data
  },
  enabled: () => !!teamId.value,
})

const { data: membersData } = useQuery({
  key: () => ['team', teamId.value, 'members'],
  query: async () => {
    const { data, error } = await getTeamsByTeamIdMembers({ path: { team_id: teamId.value } })
    if (error) throw error
    return data ?? []
  },
  enabled: () => !!teamId.value,
})

const team = computed(() => teamData.value)
const members = computed(() => membersData.value ?? [])
const assignableMembers = computed(() =>
  members.value.filter((member) => member.member_type === 'bot' ? !!member.bot_id : !!member.user_id),
)

watch(
  () => route.query.status,
  (status) => {
    issueForm.status = normalizeStatus(String(status ?? 'todo'))
  },
)

const { mutateAsync: createIssue, isLoading: creating } = useMutation({
  mutation: async () => {
    const assignment = parseAssigneeValue(issueForm.assignee)
    const { data, error } = await postTeamsByTeamIdIssues({
      path: { team_id: teamId.value },
      body: {
        title: issueForm.title.trim(),
        description: issueForm.description.trim(),
        status: issueForm.status,
        ...assignment,
      },
    })
    if (error) throw error
    return data
  },
})

async function submitIssue() {
  if (!issueForm.title.trim()) return
  try {
    const issue = await createIssue()
    toast.success(t('teams.issueCreated'))
    void queryCache.invalidateQueries({ key: ['team', teamId.value, 'issues'] })
    if (issue?.id) {
      router.push({ name: 'team-issue', params: { teamId: teamId.value, issueId: issue.id } })
    }
  }
  catch (err) {
    toast.error(resolveApiErrorMessage(err, t('teams.issueCreateFailed')))
  }
}

function goBack() {
  router.push({ name: 'team-workspace', params: { teamId: teamId.value } })
}

function normalizeStatus(status: string | undefined): IssueStatus {
  if (ISSUE_COLUMNS.some((column) => column.value === status)) {
    return status as IssueStatus
  }
  return 'todo'
}

function memberValue(member: HandlersMemberResponse) {
  if (member.member_type === 'bot') return `bot:${member.bot_id ?? ''}`
  return `user:${member.user_id ?? ''}`
}

function parseAssigneeValue(value: string) {
  if (value.startsWith('bot:')) {
    return { assignee_type: 'bot', assignee_bot_id: value.slice(4) }
  }
  if (value.startsWith('user:')) {
    return { assignee_type: 'user', assignee_user_id: value.slice(5) }
  }
  return {}
}

function memberLabel(member: HandlersMemberResponse) {
  return member.display_name || member.bot_id || member.user_id || t('teams.unknownMember')
}
</script>
