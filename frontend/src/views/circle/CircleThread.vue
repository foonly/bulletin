<script setup>
import { onMounted, onUnmounted, computed, ref, watch } from "vue";
import { useCircleStore } from "../../stores/circles";
import { useRouter } from "vue-router";
import ThreadNode from "../../components/ThreadNode.vue";

const props = defineProps(["id", "threadId"]);
const circleStore = useCircleStore();
const router = useRouter();
const readTimer = ref(null);

const threadTree = computed(() => {
	const activeThread = circleStore.activeThread;
	if (!activeThread || !activeThread.length) return null;

	const posts = activeThread;
	const root = posts.find((p) => !p.parent_id);
	if (!root) return null;

	const buildNode = (node) => ({
		...node,
		replies: posts
			.filter((p) => p.parent_id === node.id)
			.map((p) => buildNode(p)),
	});

	return buildNode(root);
});

const startReadTracking = (entityId) => {
	if (readTimer.value) clearTimeout(readTimer.value);

	readTimer.value = setTimeout(async () => {
		try {
			await circleStore.markRead(props.id, entityId);

			if (circleStore.activeThread) {
				const now = new Date().toISOString();
				circleStore.activeThread.forEach((post) => {
					post.last_read_at = now;
				});
			}

			const thread = circleStore.threads.find((t) => t.id === entityId);
			if (thread && thread.unread_count > 0 && circleStore.activeCircle) {
				circleStore.activeCircle.unread_count = Math.max(
					0,
					circleStore.activeCircle.unread_count - thread.unread_count,
				);
				thread.unread_count = 0;
			}
			await circleStore.fetchThreads(props.id);
			await circleStore.fetchTags(props.id);
		} catch (err) {
			console.error("Failed to mark as read", err);
		}
	}, 3000);
};

const loadThread = async () => {
	try {
		await circleStore.fetchThread(props.id, props.threadId);
		startReadTracking(props.threadId);
	} catch (err) {
		router.push({ name: "circle-posts", params: { id: props.id } });
	}
};

onMounted(loadThread);
watch(() => props.threadId, loadThread);

onUnmounted(() => {
	if (readTimer.value) clearTimeout(readTimer.value);
});
</script>

<template>
	<div>
		<button
			class="btn btn-ghost"
			@click="router.push({ name: 'circle-posts', params: { id } })"
			style="margin-bottom: 1rem; color: var(--accent)"
		>
			← Back to all threads
		</button>

		<div style="display: flex; flex-direction: column; gap: 1rem; padding-bottom: 3rem">
			<ThreadNode
				v-if="threadTree"
				:node="threadTree"
				:circle-id="props.id"
				@reply-created="loadThread"
			/>
		</div>
	</div>
</template>
