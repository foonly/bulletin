<template>
	<div class="flex flex-col h-full">
		<!-- Content -->
		<div class="flex-1 flex overflow-hidden">
			<!-- Error State -->
			<div
				v-if="error"
				class="flex-1 flex flex-col items-center justify-center p-8 text-center space-y-4"
			>
				<div class="text-6xl text-gray-700">🚫</div>
				<h2 class="text-2xl font-bold">{{ error.title }}</h2>
				<p class="text-gray-500 max-w-md">{{ error.message }}</p>
				<router-link
					to="/"
					class="bg-purple-600 px-6 py-2 rounded-lg font-bold hover:bg-purple-700 transition"
				>
					Return to Dashboard
				</router-link>
			</div>

			<!-- Normal UI -->
			<template v-else>
				<!-- Sidebar -->
				<div
					class="w-64 bg-gray-950/50 border-r border-gray-800 flex flex-col overflow-hidden"
				>
					<!-- Circle Header -->
					<div
						class="p-4 border-b border-gray-800 flex items-center justify-between"
					>
						<h2
							class="font-bold truncate text-gray-100"
							:title="circleStore.activeCircle?.name"
						>
							{{ circleStore.activeCircle?.name }}
						</h2>
						<router-link
							v-if="canManage"
							:to="{ name: 'circle-settings', params: { id: id } }"
							:class="[
								'p-1 rounded transition-colors',
								$route.name === 'circle-settings'
									? 'text-purple-400'
									: 'text-gray-500 hover:text-gray-300',
							]"
							title="Settings"
						>
							<svg
								xmlns="http://www.w3.org/2000/svg"
								class="h-5 w-5"
								viewBox="0 0 20 20"
								fill="currentColor"
							>
								<path
									fill-rule="evenodd"
									d="M11.49 3.17c-.38-1.56-2.6-1.56-2.98 0a1.532 1.532 0 01-2.286.948c-1.372-.836-2.942.734-2.106 2.106.54.886.061 2.042-.947 2.287-1.561.379-1.561 2.6 0 2.978a1.532 1.532 0 01.947 2.287c-.836 1.372.734 2.942 2.106 2.106a1.532 1.532 0 012.287.947c.379 1.561 2.6 1.561 2.978 0a1.533 1.533 0 012.287-.947c1.372.836 2.942-.734 2.106-2.106a1.533 1.533 0 01.947-2.287c1.561-.379 1.561-2.6 0-2.978a1.532 1.532 0 01-.947-2.287c.836-1.372-.734-2.942-2.106-2.106a1.532 1.532 0 01-2.287-.947zM10 13a3 3 0 100-6 3 3 0 000 6z"
									clip-rule="evenodd"
								/>
							</svg>
						</router-link>
					</div>

					<div class="flex-1 overflow-y-auto p-4 space-y-6">
						<!-- Search -->
						<div class="px-2">
							<div class="relative">
								<span class="absolute left-3 top-2.5 text-gray-500">🔍</span>
								<input
									v-model="searchQuery"
									@keyup.enter="handleSearch"
									placeholder="Search posts..."
									class="w-full bg-gray-900 border border-gray-800 rounded-lg py-2 pl-9 pr-4 text-xs focus:outline-none focus:border-purple-500 transition-colors"
								/>
							</div>
						</div>

						<!-- Main Nav -->
						<div class="space-y-1">
							<router-link
								:to="{ name: 'circle-new-thread', params: { id: id } }"
								class="w-full flex items-center justify-center space-x-2 px-3 py-2 bg-purple-600 hover:bg-purple-700 text-white rounded-lg transition-colors font-bold mb-4 shadow-lg"
							>
								<span>+</span>
								<span>Start New Thread</span>
							</router-link>

							<router-link
								:to="{ name: 'circle-chat', params: { id: id } }"
								:class="[
									'w-full flex items-center space-x-3 px-3 py-2 rounded-lg transition-colors',
									$route.name === 'circle-chat'
										? 'bg-purple-600 text-white'
										: 'text-gray-400 hover:bg-gray-800 hover:text-gray-200',
								]"
							>
								<span class="text-lg">💬</span>
								<span class="font-medium">Chat</span>
								<span
									v-if="unreadChatCount > 0 && $route.name !== 'circle-chat'"
									class="ml-auto bg-purple-500 text-white text-[10px] px-1.5 py-0.5 rounded-full font-bold shadow-sm"
								>
									{{ unreadChatCount }}
								</span>
							</router-link>
							<router-link
								:to="{ name: 'circle-posts', params: { id: id } }"
								:class="[
									'w-full flex items-center space-x-3 px-3 py-2 rounded-lg transition-colors',
									($route.name === 'circle-posts' && !$route.query.tag) ||
									$route.name === 'circle-search'
										? 'bg-purple-600 text-white'
										: 'text-gray-400 hover:bg-gray-800 hover:text-gray-200',
								]"
							>
								<span class="text-lg">📊</span>
								<span class="font-medium">{{
									$route.name === "circle-search"
										? "Search Results"
										: "Dashboard"
								}}</span>
							</router-link>
						</div>

						<!-- Tags -->
						<div class="space-y-2">
							<h3
								class="px-3 text-xs font-bold text-gray-500 uppercase tracking-wider"
							>
								Browse Tags
							</h3>
							<div class="space-y-1">
								<div
									v-for="tag in circleStore.tags"
									:key="tag.id"
									class="group flex items-center justify-between"
								>
									<router-link
										:to="{
											name: 'circle-posts',
											params: { id: id },
											query: { tag: tag.name },
										}"
										:class="[
											'flex-1 flex items-center justify-between px-3 py-2 rounded-lg text-sm transition-colors truncate font-medium',
											$route.name === 'circle-posts' &&
											$route.query.tag === tag.name
												? 'bg-purple-600/20 text-purple-300 border border-purple-500/50'
												: 'text-gray-400 hover:bg-gray-800 hover:text-gray-200 border border-transparent',
										]"
									>
										<span class="truncate">
											<span v-if="tag.is_pinned" class="mr-2">📌</span>#{{
												tag.name
											}}
										</span>
										<span
											v-if="tag.unread_count > 0"
											class="ml-2 bg-purple-500 text-white text-[10px] px-1.5 py-0.5 rounded-full font-bold shrink-0 shadow-sm"
										>
											{{ tag.unread_count }}
										</span>
									</router-link>
									<div class="flex items-center">
										<button
											v-if="isAdmin && !tag.is_pinned"
											@click="togglePin(tag.id, true)"
											class="hidden group-hover:block text-[10px] text-gray-600 hover:text-purple-400 p-1 ml-1"
											title="Pin tag"
										>
											Pin
										</button>
										<button
											v-if="isAdmin && tag.is_pinned"
											@click="togglePin(tag.id, false)"
											class="text-[10px] text-purple-400 p-1 ml-1"
											title="Unpin tag"
										>
											Unpin
										</button>
									</div>
								</div>
							</div>
						</div>

						<!-- Members Section -->
						<div class="space-y-2">
							<div
								@click="membersExpanded = !membersExpanded"
								class="flex items-center justify-between px-3 cursor-pointer group"
							>
								<h3
									class="text-xs font-bold text-gray-500 uppercase tracking-wider group-hover:text-gray-300 transition-colors"
								>
									Members ({{ members.length }})
								</h3>
								<span
									:class="[
										'text-[10px] text-gray-500 transition-transform duration-200',
										membersExpanded ? 'rotate-180' : '',
									]"
									>▼</span
								>
							</div>

							<div v-if="membersExpanded" class="space-y-1 mt-1">
								<button
									v-if="canInvite"
									@click.stop="showInviteModal = true"
									class="w-full flex items-center space-x-3 px-3 py-2 rounded-lg transition-colors text-purple-400 hover:bg-purple-900/20"
								>
									<span class="text-sm">✉️</span>
									<span class="text-xs font-bold uppercase tracking-tight"
										>Invite People</span
									>
								</button>

								<div class="max-h-48 overflow-y-auto px-1 space-y-0.5">
									<div
										v-for="member in members"
										:key="member.id"
										class="flex items-center space-x-2 px-2 py-1.5 rounded-lg hover:bg-gray-800/50 transition-colors group"
										:title="member.username"
									>
										<div
											:class="[
												'w-2 h-2 rounded-full shrink-0',
												onlineUserIds.has(member.id)
													? 'bg-green-500'
													: 'bg-gray-700',
											]"
										></div>
										<span
											:class="[
												'text-xs truncate',
												onlineUserIds.has(member.id)
													? 'text-gray-200 font-medium'
													: 'text-gray-500',
											]"
											>{{ member.username }}</span
										>
									</div>
								</div>
							</div>
						</div>
					</div>

					<!-- Bottom Link -->
					<div class="p-4 border-t border-gray-800">
						<router-link
							to="/"
							class="flex items-center space-x-2 text-xs text-gray-500 hover:text-gray-300 transition-colors"
						>
							<span>←</span>
							<span>Dashboard</span>
						</router-link>
					</div>
				</div>

				<div class="flex-1 overflow-y-auto p-4">
					<router-view :members="members"></router-view>
				</div>
			</template>
		</div>
	</div>

	<InviteModal
		:show="showInviteModal"
		:id="id"
		@close="showInviteModal = false"
		@created="circleStore.fetchInvites(id)"
	/>
</template>

<script setup>
import { ref, onMounted, onUnmounted, computed, watch } from "vue";
import { useCircleStore } from "../stores/circles";
import { useAuthStore } from "../stores/auth";
import { useToastStore } from "../stores/toast";
import axios from "axios";
import { useRouter, useRoute } from "vue-router";
import InviteModal from "../components/InviteModal.vue";
import { showBrowserNotification } from "../utils/notifications";

const props = defineProps(["id"]);
const circleStore = useCircleStore();
const auth = useAuthStore();
const toast = useToastStore();
const router = useRouter();
const route = useRoute();

const members = ref([]);
const onlineUserIds = ref(new Set());
const error = ref(null);
const searchQuery = ref("");
const showInviteModal = ref(false);
const membersExpanded = ref(false);

const handleSearch = () => {
	if (!searchQuery.value.trim()) return;
	router.push({
		name: "circle-search",
		params: { id: props.id },
		query: { q: searchQuery.value },
	});
	searchQuery.value = "";
};

let ws = null;

const canManage = computed(() => {
	const role = circleStore.activeCircle?.role;
	return role === "admin" || role === "mod";
});

const isAdmin = computed(() => {
	return circleStore.activeCircle?.role === "admin";
});

const canInvite = computed(() => {
	const role = circleStore.activeCircle?.role;
	const minRole = circleStore.activeCircle?.invite_min_role || "standard";

	const roles = { guest: 0, standard: 1, mod: 2, admin: 3 };
	return roles[role] >= roles[minRole];
});

const isUnread = (msg) => {
	if (msg.user_id === auth.user?.id) return false;
	const lastRead = circleStore.activeCircle?.last_read_at;
	if (!lastRead) return true;
	return new Date(msg.created_at) > new Date(lastRead);
};

const unreadChatCount = computed(() => {
	return circleStore.chatMessages.filter(isUnread).length;
});

const loadCircleData = async () => {
	error.value = null;

	try {
		if (circleStore.circles.length === 0) {
			await circleStore.fetchCircles();
		}
		const circle = circleStore.circles.find((c) => c.id === props.id);

		if (!circle) {
			error.value = {
				title: "Circle Not Found",
				message: "The circle you are looking for doesn't exist.",
			};
			return;
		}

		circleStore.activeCircle = circle;

		await circleStore.fetchChatHistory(props.id);
		await circleStore.fetchTags(props.id);
		if (canManage.value) {
			await circleStore.fetchInvites(props.id);
		}
		const res = await axios.get(`/api/circles/${props.id}/members`);
		members.value = res.data;

		// Expand members list by default if there are 8 or fewer members
		if (members.value.length <= 8) {
			membersExpanded.value = true;
		}

		connectWS();
	} catch (err) {
		console.error("Failed to load circle data:", err);
		error.value = {
			title: "Access Denied",
			message: "You are not a member of this circle or it has been deleted.",
		};
	}
};

const connectWS = () => {
	if (ws) ws.close();
	const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
	ws = new WebSocket(
		`${protocol}//${window.location.host}/api/circles/${props.id}/chat/ws`,
	);

	ws.onmessage = (event) => {
		const msg = JSON.parse(event.data);

		if (msg.type === "presence") {
			onlineUserIds.value = new Set(msg.online_ids);
		} else if (msg.type === "join") {
			onlineUserIds.value.add(msg.user_id);
			onlineUserIds.value = new Set(onlineUserIds.value);
		} else if (msg.type === "leave") {
			onlineUserIds.value.delete(msg.user_id);
			onlineUserIds.value = new Set(onlineUserIds.value);
		} else {
			circleStore.addChatMessage(msg);
			window.dispatchEvent(
				new CustomEvent("chat-message-received", { detail: msg }),
			);

			// Increment unread count if we are not looking at the chat
			if (route.name !== "circle-chat" && msg.user_id !== auth.user?.id) {
				circleStore.incrementUnreadChat(props.id);

				if (auth.notificationsEnabled) {
					showBrowserNotification(
						`New message in ${circleStore.activeCircle?.name}`,
						{
							body: `${msg.username}: ${msg.content}`,
						},
					);
				}
			}
		}
	};
};

const togglePin = async (tagId, isPinned) => {
	try {
		await circleStore.pinTag(props.id, tagId, isPinned);
	} catch (err) {
		toast.error("Failed to update tag pin");
	}
};

const handleSendChatMessage = (e) => {
	if (ws && ws.readyState === WebSocket.OPEN) {
		ws.send(JSON.stringify({ content: e.detail }));
	}
};

onMounted(() => {
	loadCircleData();
	window.addEventListener("send-chat-message", handleSendChatMessage);
	window.addEventListener("refresh-members", loadCircleData);
});

onUnmounted(() => {
	if (ws) ws.close();
	window.removeEventListener("send-chat-message", handleSendChatMessage);
	window.removeEventListener("refresh-members", loadCircleData);
});

watch(() => props.id, loadCircleData);
</script>
