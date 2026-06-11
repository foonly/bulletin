<script setup>
import { ref, computed, watch, nextTick } from "vue";
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

const canEdit = computed(
	() =>
		auth.user?.id === props.node.author_id ||
		circleStore.activeCircle?.role === "admin",
);

const showReplyBox = ref(false);
const showReplyPreview = ref(false);
const replyContent = ref("");
const replyTextarea = ref(null);
const isEditing = ref(false);
const showEditPreview = ref(false);
const editContent = ref("");
const editTextarea = ref(null);

watch(showReplyBox, (val) => {
	if (val) nextTick(() => replyTextarea.value?.focus());
});

watch(isEditing, (val) => {
	if (val) nextTick(() => editTextarea.value?.focus());
});

watch(showReplyPreview, (val) => {
	if (!val && showReplyBox.value) nextTick(() => replyTextarea.value?.focus());
});

watch(showEditPreview, (val) => {
	if (!val && isEditing.value) nextTick(() => editTextarea.value?.focus());
});

const isUnread = computed(() => {
	if (props.node.author_id === auth.user?.id) return false;
	if (!props.node.last_read_at) return true;
	return new Date(props.node.created_at) > new Date(props.node.last_read_at);
});

const formatDate = (dateStr) => new Date(dateStr).toLocaleString();

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
		emit("reply-created");
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
		emit("reply-created");
	} catch (err) {
		toast.error("Failed to delete post");
	}
};
</script>

<template>
	<div class="thread-node" :class="{ 'is-reply': node.parent_id }">
		<div
			class="thread-post"
			:class="{
				'is-root': !node.parent_id,
				unread: isUnread,
				'is-deleted': node.is_deleted,
			}"
		>
			<div class="thread-post__header">
				<div class="thread-post__author-row">
					<div class="thread-post__avatar">
						{{
							node.is_deleted
								? "?"
								: (node.author_name?.[0]?.toUpperCase() ?? "?")
						}}
					</div>
					<span class="thread-post__author-name">
						{{ node.is_deleted ? "Deleted" : node.author_name }}
					</span>
					<span class="thread-post__date">{{
						formatDate(node.created_at)
					}}</span>
					<span
						v-if="node.updated_at && !node.is_deleted"
						class="thread-post__edited"
					>
						(edited {{ formatDate(node.updated_at) }})
					</span>
				</div>

				<div v-if="!node.is_deleted" class="thread-post__actions">
					<button
						v-if="canEdit && !isEditing"
						class="thread-post__action-btn"
						@click="startEdit"
					>
						Edit
					</button>
					<button
						v-if="canEdit && !isEditing"
						class="thread-post__action-btn delete"
						@click="handleDelete"
					>
						Delete
					</button>
					<button
						class="thread-post__action-btn reply"
						@click="showReplyBox = !showReplyBox"
					>
						{{ showReplyBox ? "Cancel" : "Reply" }}
					</button>
				</div>
			</div>

			<h2 v-if="node.title" class="thread-post__title">{{ node.title }}</h2>

			<!-- Edit mode -->
			<div v-if="isEditing" class="thread-post__edit">
				<div class="thread-post__edit-header">
					<span class="label-uppercase">Editing Post</span>
					<button
						class="btn btn-ghost btn-sm"
						style="color: var(--accent)"
						@click="showEditPreview = !showEditPreview"
					>
						{{ showEditPreview ? "Edit Text" : "Show Preview" }}
					</button>
				</div>
				<div
					v-if="showEditPreview"
					class="markdown-content"
					style="
						min-height: 100px;
						background: var(--bg-sunken);
						padding: 0.75rem;
						border-radius: var(--r-md);
						border: 1px solid var(--border);
						font-size: var(--text-sm);
					"
					v-html="renderMarkdown(editContent)"
				></div>
				<textarea
					v-else
					ref="editTextarea"
					v-model="editContent"
					rows="4"
					style="font-size: var(--text-sm)"
				></textarea>
				<div class="thread-post__edit-actions">
					<button class="btn btn-ghost btn-sm" @click="isEditing = false">
						Cancel
					</button>
					<button class="btn btn-primary btn-sm" @click="submitEdit">
						Save
					</button>
				</div>
			</div>

			<!-- Post body -->
			<div
				v-else
				class="thread-post__body markdown-content"
				:class="{ 'is-deleted': node.is_deleted }"
				v-html="renderMarkdown(node.content)"
			></div>

			<!-- Reply box -->
			<div v-if="showReplyBox" class="reply-box">
				<div class="reply-box__header">
					<span class="label-uppercase">Reply</span>
					<button
						class="btn btn-ghost btn-sm"
						style="color: var(--accent)"
						@click="showReplyPreview = !showReplyPreview"
					>
						{{ showReplyPreview ? "Edit Text" : "Show Preview" }}
					</button>
				</div>
				<div
					v-if="showReplyPreview"
					class="markdown-content"
					style="
						min-height: 120px;
						background: var(--bg-sunken);
						padding: 0.75rem;
						border-radius: var(--r-md);
						border: 1px solid var(--border);
						font-size: var(--text-sm);
					"
					v-html="renderMarkdown(replyContent)"
				></div>
				<textarea
					v-else
					ref="replyTextarea"
					v-model="replyContent"
					placeholder="Write a reply…"
					rows="5"
					style="font-size: var(--text-sm)"
				></textarea>
				<div class="reply-box__actions">
					<button
						class="btn btn-primary btn-sm"
						:disabled="!replyContent.trim()"
						@click="submitReply"
					>
						Post Reply
					</button>
				</div>
			</div>
		</div>

		<!-- Nested replies -->
		<div
			v-if="node.replies?.length"
			style="display: flex; flex-direction: column; gap: 1rem"
		>
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
