<script setup>
import { ref } from "vue";
import { useCircleStore } from "../stores/circles";

const props = defineProps({
	node: Object,
	circleId: String,
});

const emit = defineEmits(["reply-created"]);
const circleStore = useCircleStore();
const showReplyBox = ref(false);
const replyContent = ref("");

const formatDate = (dateStr) => {
	return new Date(dateStr).toLocaleString();
};

const submitReply = async () => {
	if (!replyContent.value.trim()) return;

	try {
		await circleStore.createPost(props.circleId, {
			parent_id: props.node.id,
			content: replyContent.value,
			tags: [],
		});
		replyContent.value = "";
		showReplyBox.value = false;
		emit("reply-created");
	} catch (err) {
		alert("Failed to post reply");
	}
};
</script>

<template>
	<div
		:class="[
			'space-y-4',
			node.parent_id ? 'ml-2 md:ml-6 border-l-2 border-gray-800 pl-4 py-1' : '',
		]"
	>
		<!-- Post Content -->
		<div
			:class="[
				'p-4 rounded-lg border transition-colors',
				node.parent_id
					? 'bg-gray-800/50 border-gray-700'
					: 'bg-gray-800 border-purple-900/50 p-6',
			]"
		>
			<div class="flex items-center justify-between mb-2">
				<div class="flex items-center space-x-2">
					<div
						class="w-6 h-6 bg-purple-900 rounded-full flex items-center justify-center text-[10px] font-bold"
					>
						{{ node.author_name[0].toUpperCase() }}
					</div>
					<span class="font-bold text-purple-400 text-sm">{{
						node.author_name
					}}</span>
					<span class="text-[10px] text-gray-500">{{
						formatDate(node.created_at)
					}}</span>
				</div>

				<button
					@click="showReplyBox = !showReplyBox"
					class="text-xs text-purple-400 hover:text-purple-300 font-medium"
				>
					{{ showReplyBox ? "Cancel" : "Reply" }}
				</button>
			</div>

			<h2 v-if="node.title" class="text-xl font-bold mb-3">{{ node.title }}</h2>
			<div class="text-gray-300 text-sm whitespace-pre-wrap leading-relaxed">
				{{ node.content }}
			</div>

			<!-- Nested Reply Box -->
			<div v-if="showReplyBox" class="mt-4 pt-4 border-t border-gray-700">
				<textarea
					v-model="replyContent"
					placeholder="Write a reply..."
					class="w-full bg-gray-700 p-2 rounded border border-gray-600 focus:outline-none text-sm"
					rows="2"
				></textarea>
				<div class="flex justify-end mt-2">
					<button
						@click="submitReply"
						:disabled="!replyContent.trim()"
						class="bg-purple-600 px-3 py-1 rounded font-bold hover:bg-purple-700 text-xs disabled:opacity-50"
					>
						Post Reply
					</button>
				</div>
			</div>
		</div>

		<!-- Children Nodes -->
		<div v-if="node.replies && node.replies.length" class="space-y-4">
			<ThreadNode
				v-for="reply in node.replies"
				:key="reply.id"
				:node="reply"
				:circle-id="circleId"
				@reply-created="$emit('reply-created')"
			/>
		</div>
	</div>
</template>
