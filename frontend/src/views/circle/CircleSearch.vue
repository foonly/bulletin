<script setup>
import { ref, onMounted, watch } from "vue";
import { useCircleStore } from "../../stores/circles";
import { useRouter, useRoute } from "vue-router";
import { stripMarkdown } from "../../utils/markdown";

const props = defineProps(["id"]);
const circleStore = useCircleStore();
const router = useRouter();
const route = useRoute();

const results = ref([]);
const loading = ref(false);

const performSearch = async () => {
	const query = route.query.q;
	if (!query) {
		results.value = [];
		return;
	}

	loading.value = true;
	try {
		results.value = await circleStore.search(props.id, query);
	} catch (err) {
		console.error("Search failed", err);
	} finally {
		loading.value = false;
	}
};

const formatDate = (dateStr) => {
	return new Date(dateStr).toLocaleString();
};

const goToResult = (result) => {
	router.push({
		name: "circle-thread",
		params: { id: props.id, threadId: result.root_id },
	});
};

onMounted(performSearch);
watch(() => route.query.q, performSearch);
</script>

<template>
	<div class="space-y-6">
		<div class="flex items-center justify-between">
			<h1 class="text-2xl font-bold">
				Search Results for "{{ route.query.q }}"
			</h1>
			<span class="text-sm text-gray-500">{{ results.length }} results found</span>
		</div>

		<div v-if="loading" class="text-center py-20">
			<div
				class="animate-spin rounded-full h-12 w-12 border-b-2 border-purple-500 mx-auto"
			></div>
			<p class="mt-4 text-gray-400">Searching...</p>
		</div>

		<div v-else-if="results.length > 0" class="space-y-4">
			<div
				v-for="result in results"
				:key="result.id"
				@click="goToResult(result)"
				class="bg-gray-800 p-4 rounded-lg border border-gray-700 cursor-pointer hover:border-purple-500 transition-colors group"
			>
				<div class="flex items-center justify-between mb-2 text-xs">
					<div class="flex items-center space-x-2">
						<span class="font-bold text-purple-400">{{
							result.author_name
						}}</span>
						<span class="text-gray-500">{{ formatDate(result.created_at) }}</span>
					</div>
					<span v-if="result.parent_id" class="text-[10px] text-gray-600 uppercase font-bold">Reply</span>
                    <span v-else class="text-[10px] text-purple-600 uppercase font-bold">Thread</span>
				</div>

				<h3 v-if="result.title" class="font-bold text-lg mb-1 group-hover:text-purple-400 transition-colors">
					{{ result.title }}
				</h3>
				<p class="text-gray-300 text-sm line-clamp-3">
					{{ stripMarkdown(result.content) }}
				</p>
			</div>
		</div>

		<div v-else class="text-center py-20 text-gray-500">
			<div class="text-4xl mb-4">🔍</div>
			<p>No results found for your search query.</p>
		</div>
	</div>
</template>
