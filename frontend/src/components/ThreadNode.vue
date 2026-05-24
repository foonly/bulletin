<script setup>
import { ref, computed } from "vue";
import { useCircleStore } from "../stores/circles";
import { useAuthStore } from "../stores/auth";
import { useToastStore } from "../stores/toast";
import { renderMarkdown } from "../utils/markdown";

const props = defineProps({
	node: Object,
	circleId: String,
});

const emit = defineEmits(["reply-created"]);
const circleStore = useCircleStore();
const auth = useAuthStore();
const toast = useToastStore();

const canEdit = computed(() => {
	return (
		auth.user?.id === props.node.author_id ||
		circleStore.activeCircle?.role === "admin"
	);
});
const showReplyBox = ref(false);
const showReplyPreview = ref(false);
const replyContent = ref("");
const isEditing = ref(false);
const showEditPreview = ref(false);
const editContent = ref("");

const isUnread = computed(() => {
	if (props.node.author_id === auth.user?.id) return false;
	if (!props.node.last_read_at) return true;
	return new Date(props.node.created_at) > new Date(props.node.last_read_at);
});

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
		toast.success("Reply posted!");
		emit("reply-created");
	} catch (err) {
		toast.error("Failed to post reply");
	}
};

const startEdit = () => {
	editContent.value = props.node.content;
	isEditing.value = true;
};

const submitEdit = async () => {
	if (!editContent.value.trim()) return;
	try {
		await circleStore.updatePost(
			props.circleId,
			props.node.id,
			editContent.value,
		);
		isEditing.value = false;
		toast.success("Post updated!");
		emit("reply-created"); // Refresh to show new content
	} catch (err) {
		toast.error("Failed to update post");
	}
};

const handleDelete = async () => {
	if (
		!confirm(
			"Are you sure you want to delete this? This action cannot be undone.",
		)
	)
		return;

	try {
		await circleStore.deletePost(props.circleId, props.node.id);
		toast.success(props.node.parent_id ? "Reply deleted" : "Thread deleted");
		emit("reply-created"); // Refresh view
	} catch (err) {
		toast.error("Failed to delete post");
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
				'p-4 rounded-lg border transition-all duration-500',
				node.parent_id
					? 'bg-gray-800/50 border-gray-700'
					: 'bg-gray-800 border-purple-900/50 p-6',
				isUnread ? 'border-l-4 border-l-purple-500 bg-purple-500/5' : '',
				node.is_deleted ? 'opacity-50 grayscale' : '',
			]"
		>
			<div class="flex items-center justify-between mb-2">
				<div class="flex items-center space-x-2">
					<div
						class="w-6 h-6 bg-purple-900 rounded-full flex items-center justify-center text-[10px] font-bold"
					>
						{{
							node.is_deleted
								? "?"
								: node.author_name
									? node.author_name[0].toUpperCase()
									: "?"
						}}
					</div>
					<span class="font-bold text-purple-400 text-sm">
						{{ node.is_deleted ? "Deleted" : node.author_name }}
					</span>
					<span class="text-[10px] text-gray-500">{{
						formatDate(node.created_at)
					}}</span>
					<span
						v-if="node.updated_at && !node.is_deleted"
						class="text-[10px] text-gray-500 italic"
					>
						(edited {{ formatDate(node.updated_at) }})
					</span>
				</div>

				<div class="flex items-center space-x-2" v-if="!node.is_deleted">
					<button
						v-if="canEdit && !isEditing"
						@click="startEdit"
						class="text-xs text-gray-500 hover:text-white font-medium"
					>
						Edit
					</button>
					<button
						v-if="canEdit && !isEditing"
						@click="handleDelete"
						class="text-xs text-red-500 hover:text-red-400 font-medium"
					>
						Delete
					</button>
					<button
						@click="showReplyBox = !showReplyBox"
						class="text-xs text-purple-400 hover:text-purple-300 font-medium"
					>
						{{ showReplyBox ? "Cancel" : "Reply" }}
					</button>
				</div>
			</div>

			<h2 v-if="node.title" class="text-xl font-bold mb-3">{{ node.title }}</h2>

			<div v-if="isEditing" class="space-y-2">
				<div class="flex items-center justify-between">
					<label class="text-[10px] text-gray-500 font-bold uppercase"
						>Editing Post</label
					>
					<button
						@click="showEditPreview = !showEditPreview"
						class="text-[10px] uppercase font-bold text-purple-400 hover:text-purple-300"
					>
						{{ showEditPreview ? "Edit Text" : "Show Preview" }}
					</button>
				</div>
				<div
					v-if="showEditPreview"
					class="w-full bg-gray-900 p-3 rounded-lg border border-gray-700 markdown-content text-sm min-h-[100px]"
					v-html="renderMarkdown(editContent)"
				></div>
				<textarea
					v-else
					v-model="editContent"
					class="w-full bg-gray-700 p-2 rounded border border-gray-600 focus:outline-none text-sm"
					rows="4"
				></textarea>
				<div class="flex justify-end space-x-2">
					<button
						@click="isEditing = false"
						class="text-xs text-gray-400 hover:text-white"
					>
						Cancel
					</button>
					<button
						@click="submitEdit"
						class="bg-purple-600 px-3 py-1 rounded font-bold text-xs"
					>
						Save
					</button>
				</div>
			</div>
			<div
				v-else
				class="markdown-content text-gray-300 text-sm leading-relaxed"
				:class="{ 'italic text-gray-500': node.is_deleted }"
				v-html="renderMarkdown(node.content)"
			></div>

			<!-- Nested Reply Box -->
			<div v-if="showReplyBox" class="mt-4 pt-4 border-t border-gray-700">
				<div class="flex items-center justify-between mb-2">
					<label class="text-[10px] text-gray-500 font-bold uppercase"
						>Reply</label
					>
					<button
						@click="showReplyPreview = !showReplyPreview"
						class="text-[10px] uppercase font-bold text-purple-400 hover:text-purple-300"
					>
						{{ showReplyPreview ? "Edit Text" : "Show Preview" }}
					</button>
				</div>
				<div
					v-if="showReplyPreview"
					class="w-full bg-gray-900 p-3 rounded-lg border border-gray-700 markdown-content text-sm min-h-[120px]"
					v-html="renderMarkdown(replyContent)"
				></div>
				<textarea
					v-else
					v-model="replyContent"
					placeholder="Write a reply..."
					class="w-full bg-gray-700 p-2 rounded border border-gray-600 focus:outline-none text-sm"
					rows="5"
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
