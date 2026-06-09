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

const isUnread = (msg) => {
	if (msg.user_id === auth.user?.id) return false;
	const lastRead = circleStore.activeCircle?.last_read_at;
	if (!lastRead) return true;
	return new Date(msg.created_at) > new Date(lastRead);
};

const unreadChatCount = computed(() => {
	return (circleStore.chatMessages || []).filter(isUnread).length;
});

const formatDate = (dateStr) => new Date(dateStr).toLocaleString();

const scrollToBottom = () => {
	nextTick(() => {
		if (chatBox.value) chatBox.value.scrollTop = chatBox.value.scrollHeight;
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
	window.dispatchEvent(
		new CustomEvent("send-chat-message", { detail: chatInput.value }),
	);
	chatInput.value = "";
};

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

watch(() => circleStore.chatMessages?.length, handleNewMessage);
</script>

<template>
	<div class="chat-view content-container">
		<div class="chat-messages" ref="chatBox">
			<div
				v-for="msg in circleStore.chatMessages"
				:key="msg.id"
				class="chat-message"
				:class="{
					'is-own': msg.user_id === auth.user?.id,
					'is-unread': isUnread(msg),
				}"
			>
				<span
					class="chat-message__author"
					:class="{ 'is-own': msg.user_id === auth.user?.id }"
					>{{ msg.username }}:</span
				>
				<span
					class="chat-message__content markdown-content"
					v-html="renderMarkdown(msg.content)"
				></span>
				<span class="chat-message__time">{{ formatDate(msg.created_at) }}</span>
			</div>
		</div>

		<div class="chat-input-row">
			<input
				v-model="chatInput"
				@keyup.enter="sendChatMessage"
				placeholder="Type a message..."
			/>
			<button class="btn btn-primary" @click="sendChatMessage">Send</button>
		</div>
	</div>
</template>
