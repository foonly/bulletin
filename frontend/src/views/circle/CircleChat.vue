<script setup>
import { ref, onMounted, onUnmounted, nextTick, watch, computed } from "vue";
import { useCircleStore } from "../../stores/circles";
import { useAuthStore } from "../../stores/auth";
import { renderMarkdown } from "../../utils/markdown";

const props = defineProps(["id"]);
const circleStore = useCircleStore();
const auth = useAuthStore();
const chatInput = ref("");
const chatBox = ref(null);
const chatReadTimer = ref(null);

const unreadChatCount = computed(() => {
	return circleStore.chatMessages.filter(isUnread).length;
});

const isUnread = (msg) => {
	if (msg.user_id === auth.user?.id) return false;
	const lastRead = circleStore.activeCircle?.last_read_at;
	if (!lastRead) return true;
	return new Date(msg.created_at) > new Date(lastRead);
};

const formatDate = (dateStr) => {
	return new Date(dateStr).toLocaleString();
};

const scrollToBottom = () => {
	nextTick(() => {
		if (chatBox.value) {
			chatBox.value.scrollTop = chatBox.value.scrollHeight;
		}
	});
};

const startChatReadTracking = () => {
	if (chatReadTimer.value) clearTimeout(chatReadTimer.value);
	chatReadTimer.value = setTimeout(async () => {
		try {
			await circleStore.markRead(props.id, props.id);
			if (circleStore.activeCircle) {
				const chatUnread = unreadChatCount.value;
				circleStore.activeCircle.last_read_at = new Date().toISOString();
				circleStore.activeCircle.unread_count = Math.max(
					0,
					circleStore.activeCircle.unread_count - chatUnread,
				);
			}
		} catch (err) {
			console.error("Failed to mark chat as read", err);
		}
	}, 3000);
};

const sendChatMessage = () => {
	if (!chatInput.value.trim()) return;
	// We emit to the parent because the WebSocket is owned by the layout
	window.dispatchEvent(
		new CustomEvent("send-chat-message", { detail: chatInput.value }),
	);
	chatInput.value = "";
};

// Listen for new messages to trigger read tracking and scroll
const handleNewMessage = () => {
	scrollToBottom();
	startChatReadTracking();
};

onMounted(() => {
	circleStore.fetchChatHistory(props.id);
	scrollToBottom();
	startChatReadTracking();
	window.addEventListener("chat-message-received", handleNewMessage);
});

onUnmounted(() => {
	if (chatReadTimer.value) clearTimeout(chatReadTimer.value);
	window.removeEventListener("chat-message-received", handleNewMessage);
});

watch(() => circleStore.chatMessages.length, handleNewMessage);
</script>

<template>
	<div class="flex flex-col h-full">
		<div class="flex-1 overflow-y-auto space-y-1 mb-4 p-2" ref="chatBox">
			<div
				v-for="msg in circleStore.chatMessages"
				:key="msg.id"
				:class="[
					'text-sm px-2 py-1 rounded transition-all duration-500',
					msg.user_id === auth.user?.id ? 'bg-white/5' : '',
					isUnread(msg)
						? 'bg-purple-600/10 border-l-2 border-purple-500'
						: 'border-l-2 border-transparent',
				]"
			>
				<span
					:class="[
						'font-bold',
						msg.user_id === auth.user?.id
							? 'text-purple-300'
							: 'text-purple-500',
					]"
					>{{ msg.username }}:</span
				>
				<span
					class="ml-2 text-gray-300 markdown-content inline-block"
					v-html="renderMarkdown(msg.content)"
				></span>
				<span class="ml-2 text-[10px] text-gray-600">{{
					formatDate(msg.created_at)
				}}</span>
			</div>
		</div>
		<div class="flex space-x-2">
			<input
				v-model="chatInput"
				@keyup.enter="sendChatMessage"
				placeholder="Type a message..."
				class="flex-1 bg-gray-800 border border-gray-700 rounded px-4 py-2 focus:outline-none focus:border-purple-500"
			/>
			<button
				@click="sendChatMessage"
				class="bg-purple-600 px-6 rounded font-bold hover:bg-purple-700"
			>
				Send
			</button>
		</div>
	</div>
</template>
