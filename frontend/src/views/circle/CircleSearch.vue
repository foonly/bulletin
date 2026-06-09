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

const formatDate = (dateStr) => new Date(dateStr).toLocaleString();

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
	<div
		class="content-container"
		style="display: flex; flex-direction: column; gap: 1.5rem"
	>
		<div
			style="display: flex; align-items: center; justify-content: space-between"
		>
			<h1>Search Results for "{{ route.query.q }}"</h1>
			<span style="font-size: var(--text-sm); color: var(--fg-3)">
				{{ results.length }} results found
			</span>
		</div>

		<div v-if="loading" style="text-align: center; padding: 4rem 0">
			<div class="spinner"></div>
			<p style="margin-top: 1rem; color: var(--fg-2)">Searching…</p>
		</div>

		<div v-else-if="results.length > 0" class="thread-list">
			<div
				v-for="result in results"
				:key="result.id"
				class="thread-item"
				@click="goToResult(result)"
			>
				<div class="thread-item__header">
					<div class="thread-item__header-left">
						<span class="thread-item__author">{{ result.author_name }}</span>
						<span class="thread-item__date">{{
							formatDate(result.created_at)
						}}</span>
					</div>
					<div class="thread-item__header-right">
						<span v-if="result.parent_id" class="role-badge">Reply</span>
						<span v-else class="badge badge-accent">Thread</span>
					</div>
				</div>

				<div class="thread-item__title-row">
					<h3 v-if="result.title" class="thread-item__title">
						{{ result.title }}
					</h3>
					<span v-if="result.title" class="thread-item__divider">·</span>
					<span class="thread-item__preview-inline">{{
						stripMarkdown(result.content)
					}}</span>
				</div>
			</div>
		</div>

		<div v-else class="empty-state">
			<div class="empty-state__icon">🔍</div>
			<p>No results found for your search query.</p>
		</div>
	</div>
</template>
