<script setup>
import { onMounted, watch, computed } from "vue";
import { useCircleStore } from "../../stores/circles";
import { useRouter, useRoute } from "vue-router";
import { stripMarkdown } from "../../utils/markdown";

const props = defineProps(["id"]);
const circleStore = useCircleStore();
const router = useRouter();
const route = useRoute();

const formatDate = (dateStr) => {
	if (!dateStr) return "";
	return new Date(dateStr).toLocaleString();
};

const openThread = (postId) => {
	router.push({
		name: "circle-thread",
		params: { id: props.id, threadId: postId },
	});
};

const filterByTag = (tagName) => {
	router.push({
		name: "circle-posts",
		params: { id: props.id },
		query: tagName ? { tag: tagName } : {},
	});
};

const filteredThreads = computed(() => {
	if (route.query.tag) return circleStore.threads;
	return circleStore.threads.filter((t) => t.unread_count > 0);
});

const loadThreads = () => {
	circleStore.fetchThreads(props.id, route.query.tag || "");
};

onMounted(loadThreads);
watch(() => route.query.tag, loadThreads);
watch(() => props.id, loadThreads);
</script>

<template>
	<div class="thread-list">
		<div
			style="
				margin-bottom: 1.25rem;
				display: flex;
				justify-content: space-between;
				align-items: flex-start;
			"
		>
			<div>
				<h1 v-if="!route.query.tag">Circle Dashboard</h1>
				<h1 v-else>#{{ route.query.tag }}</h1>
				<p
					style="
						font-size: var(--text-sm);
						color: var(--fg-3);
						margin-top: 0.25rem;
					"
				>
					<template v-if="!route.query.tag"
						>Showing threads with unread activity.</template
					>
					<template v-else
						>Showing all threads tagged with #{{ route.query.tag }}.</template
					>
				</p>
			</div>
			<router-link
				:to="{
					name: 'circle-new-thread',
					params: { id },
					query: route.query.tag ? { tag: route.query.tag } : {},
				}"
				class="btn btn-primary"
				>Start new thread</router-link
			>
		</div>

		<div
			v-for="thread in filteredThreads"
			:key="thread.id"
			class="thread-item"
			:class="{ unread: thread.unread_count > 0 }"
			@click="openThread(thread.id)"
		>
			<div class="thread-item__header">
				<div>
					<span class="thread-item__author">{{ thread.author_name }}</span>
					<span class="thread-item__date">{{
						formatDate(thread.created_at)
					}}</span>
				</div>
				<span v-if="thread.unread_count > 0" class="new-pill">
					{{ thread.unread_count }} new
				</span>
			</div>

			<h3>{{ thread.title }}</h3>

			<div class="thread-item__tags">
				<span
					v-for="tag in thread.tags"
					:key="tag"
					class="tag-chip"
					:class="{ active: route.query.tag === tag }"
					@click.stop="filterByTag(tag)"
					>#{{ tag }}</span
				>
			</div>

			<p class="thread-item__preview">{{ stripMarkdown(thread.content) }}</p>

			<div class="thread-item__footer">
				<div style="display: flex; gap: 1rem">
					<span>{{ thread.reply_count }} replies</span>
					<span v-if="thread.last_reply_at"
						>Last reply: {{ formatDate(thread.last_reply_at) }}</span
					>
				</div>
			</div>
		</div>

		<div v-if="filteredThreads.length === 0" class="empty-state">
			<template v-if="!route.query.tag">
				<div class="empty-state__icon">✅</div>
				<h3>You're all caught up!</h3>
				<p>No unread threads in this circle.</p>
			</template>
			<template v-else>
				<div class="empty-state__icon">📭</div>
				<p>No threads found with this tag.</p>
			</template>
		</div>
	</div>
</template>
