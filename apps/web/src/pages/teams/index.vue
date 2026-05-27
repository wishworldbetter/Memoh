<template>
  <section class="px-4 pt-2 pb-10 lg:px-6 md:pt-4 md:pb-12">
    <div class="flex items-center justify-end mb-6 flex-wrap">
      <div class="flex items-center gap-3">
        <div class="relative">
          <Search
            class="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground size-3.5"
          />
          <Input
            v-model="searchText"
            :placeholder="t('teams.searchPlaceholder')"
            class="pl-9 w-64"
          />
        </div>
        <Button
          variant="default"
          @click="router.push({ name: 'team-new' })"
        >
          <Plus class="mr-1.5" />
          {{ t('teams.create') }}
        </Button>
      </div>
    </div>

    <div
      v-if="filteredTeams.length > 0"
      class="grid gap-4 grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4"
    >
      <TeamCard
        v-for="team in filteredTeams"
        :key="team.id"
        :team="team"
      />
    </div>

    <Empty
      v-else-if="!isLoading"
      class="mt-20 flex flex-col items-center justify-center"
    >
      <EmptyHeader>
        <EmptyMedia variant="icon">
          <Users />
        </EmptyMedia>
      </EmptyHeader>
      <EmptyTitle>{{ t('teams.emptyTitle') }}</EmptyTitle>
      <EmptyDescription>{{ t('teams.emptyDescription') }}</EmptyDescription>
      <EmptyContent />
    </Empty>
  </section>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useQuery } from '@pinia/colada'
import { Plus, Search, Users } from 'lucide-vue-next'
import {
  Button,
  Empty,
  EmptyContent,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
  Input,
} from '@memohai/ui'
import { getTeams } from '@memohai/sdk'
import TeamCard from './components/team-card.vue'

const router = useRouter()
const { t } = useI18n()
const searchText = ref('')

const { data, status } = useQuery({
  key: () => ['teams'],
  query: async () => {
    const { data, error } = await getTeams()
    if (error) throw error
    return data ?? []
  },
})

const teams = computed(() => data.value ?? [])
const isLoading = computed(() => status.value === 'loading')

const filteredTeams = computed(() => {
  const keyword = searchText.value.trim().toLowerCase()
  if (!keyword) return teams.value
  return teams.value.filter((team) =>
    team.name?.toLowerCase().includes(keyword)
    || team.description?.toLowerCase().includes(keyword)
    || team.id?.toLowerCase().includes(keyword),
  )
})
</script>
