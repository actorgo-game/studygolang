import { ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

export function usePagination(defaultPerPage = 20) {
  const route = useRoute()
  const router = useRouter()

  const page = ref(Number(route.query.p) || 1)
  const total = ref(0)
  const perPage = ref(defaultPerPage)

  watch(
    () => route.query.p,
    (val) => {
      page.value = Number(val) || 1
    }
  )

  function onPageChange(newPage: number) {
    page.value = newPage
    router.push({ query: { ...route.query, p: String(newPage) } })
  }

  return { page, total, perPage, onPageChange }
}
