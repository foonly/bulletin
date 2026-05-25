<script setup>
import { onMounted, watch } from "vue";
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
	router.push({ name: "circle-thread", params: { id: props.id, threadId: postId } });
};

const filterByTag = (tagName) => {
	router.push({
		name: "circle-posts",
		params: { id: props.id },
		query: tagName ? { tag: tagName } : {},
	});
};

const loadThreads = () => {
	circleStore.fetchThreads(props.id, route.query.tag || "");
};

onMounted(loadThreads);
watch(() => route.query.tag, loadThreads);
watch(() => props.id, loadThreads);
</script>

<template>
	<div class="space-y-4">
		<div
			v-for="thread in circleStore.threads"
			:key="thread.id"
			@click="openThread(thread.id)"
			:class="[
				'bg-gray-800 p-4 rounded-lg border cursor-pointer hover:border-purple-500 transition-all duration-500',
				thread.unread_count > 0
					? 'border-l-4 border-l-purple-500 bg-purple-500/5 border-gray-700'
					: 'border-gray-700',
			]"
		>
			<div class="flex items-center justify-between mb-2 text-xs">
				<div class="flex items-center space-x-2">
					<span class="font-bold text-purple-400">{{
						thread.author_name
					}}</span>
					<span class="text-gray-500">{{
						formatDate(thread.created_at)
					}}</span>
				</div>
				<div
					v-if="thread.unread_count > 0"
					class="bg-red-500 text-white px-2 py-0.5 rounded-full font-bold"
				>
					{{ thread.unread_count }} new
				</div>
			</div>
			<h3 class="font-bold text-lg mb-1">{{ thread.title }}</h3>
			<div class="flex flex-wrap gap-1 mb-2">
				<span
					v-for="tag in thread.tags"
					:key="tag"
					@click.stop="filterByTag(tag)"
					:class="[
                        'text-[10px] px-1.5 py-0.5 rounded transition-colors cursor-pointer',
                        route.query.tag === tag ? 'bg-purple-600 text-white' : 'bg-gray-700 text-gray-300 hover:bg-purple-600 hover:text-white'
                    ]"
				>
					#{{ tag }}
				</span>
			</div>
			<p
				class="text-gray-400 text-sm line-clamp-4 overflow-hidden"
				style="
					display: -webkit-box;
					-webkit-line-clamp: 4;
					-webkit-box-orient: vertical;
				"
			>
				{{ stripMarkdown(thread.content) }}
			</p>

			<div
				class="mt-4 flex items-center justify-between text-xs text-gray-500 border-t border-gray-700 pt-2"
			>
				<div class="flex items-center space-x-4">
					<span>{{ thread.reply_count }} replies</span>
					<span v-if="thread.last_reply_at"
						>Last reply: {{ formatDate(thread.last_reply_at) }}</span
					>
				</div>
			</div>
		</div>
        <div v-if="circleStore.threads.length === 0" class="text-center py-20 text-gray-500">
            <div class="text-4xl mb-4">📭</div>
            <p>No threads found in this circle.</p>
        </div>
	</div>
</template>
