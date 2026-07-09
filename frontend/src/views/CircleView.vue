<template>
	<div class="circle-layout">
		<!-- Error state -->
		<div v-if="error" class="error-state">
			<div class="error-state__icon">🚫</div>
			<h2>{{ error.title }}</h2>
			<p>{{ error.message }}</p>
			<router-link to="/" class="btn btn-primary"
				>Return to Dashboard</router-link
			>
		</div>

		<template v-else>
			<!-- Circle sidebar -->
			<aside class="circle-sidebar">
				<div class="circle-sidebar__header">
					<h2 :title="circleStore.activeCircle?.name">
						{{ circleStore.activeCircle?.name }}
					</h2>
					<router-link
						v-if="canManage"
						:to="{ name: 'circle-settings', params: { id } }"
						class="btn-icon"
						:class="{ 'text-accent': $route.name === 'circle-settings' }"
						title="Settings"
					>
						<svg
							xmlns="http://www.w3.org/2000/svg"
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

				<div class="circle-sidebar__body">
					<!-- Search -->
					<div class="search-field">
						<span class="search-icon">🔍</span>
						<input
							v-model="searchQuery"
							@keyup.enter="handleSearch"
							placeholder="Search posts..."
							type="search"
						/>
					</div>

					<!-- Nav -->
					<nav class="circle-nav">
						<router-link
							:to="{ name: 'circle-chat', params: { id } }"
							:class="['nav-item', { active: $route.name === 'circle-chat' }]"
							active-class=""
							exact-active-class=""
						>
							<span>💬</span>
							<span>Chat</span>
							<span
								v-if="unreadChatCount > 0 && $route.name !== 'circle-chat'"
								class="unread-pill"
								>{{ unreadChatCount }}</span
							>
						</router-link>

						<router-link
							:to="{ name: 'circle-posts', params: { id } }"
							:class="[
								'nav-item',
								{
									active:
										($route.name === 'circle-posts' && !$route.query.tag) ||
										$route.name === 'circle-search',
								},
							]"
							active-class=""
							exact-active-class=""
						>
							<span>📊</span>
							<span>{{
								$route.name === "circle-search" ? "Search Results" : "Dashboard"
							}}</span>
						</router-link>
					</nav>

					<!-- Tags -->
					<div>
						<p class="tag-section-title" style="margin-bottom: 0.375rem">
							Browse Tags
						</p>
						<nav class="circle-nav">
							<div
								v-for="tag in circleStore.tags"
								:key="tag.id"
								class="tag-list-item"
							>
								<router-link
									:to="{
										name: 'circle-posts',
										params: { id },
										query: { tag: tag.name },
									}"
									:class="[
										'nav-item',
										{
											'router-link-tag-active':
												$route.name === 'circle-posts' &&
												$route.query.tag === tag.name,
										},
									]"
									active-class=""
									exact-active-class=""
									style="flex: 1"
								>
									<span class="truncate">
										<span v-if="tag.is_pinned">📌 </span>#{{ tag.name }}
									</span>
									<span v-if="tag.unread_count > 0" class="unread-pill">
										{{ tag.unread_count }}
									</span>
								</router-link>
								<button
									v-if="isAdmin && !tag.is_pinned"
									class="tag-pin-btn"
									@click="togglePin(tag.id, true)"
									title="Pin tag"
								>
									Pin
								</button>
								<button
									v-if="isAdmin && tag.is_pinned"
									class="tag-pin-btn pinned"
									@click="togglePin(tag.id, false)"
									title="Unpin tag"
								>
									Unpin
								</button>
							</div>
						</nav>
					</div>

					<!-- Members -->
					<div>
						<div
							class="member-section-header"
							@click="membersExpanded = !membersExpanded"
						>
							<h3>Members ({{ members.length }})</h3>
							<span class="chevron" :class="{ open: membersExpanded }">▼</span>
						</div>

						<div v-if="membersExpanded" class="circle-sidebar__members">
							<div class="member-list">
								<div
									v-for="member in members"
									:key="member.id"
									class="member-item"
									:class="{ online: socketStore.onlineUserIds.has(member.id) }"
									:title="member.username"
								>
									<span
										class="presence-dot"
										:class="
											socketStore.onlineUserIds.has(member.id)
												? 'online'
												: 'offline'
										"
									></span>
									<span class="member-name">{{ member.username }}</span>
								</div>
							</div>

							<router-link
								v-if="canInvite"
								:to="{ name: 'circle-invites', params: { id } }"
								class="invite-btn"
								@click="ui.closeSidebar"
							>
								<span>✉️</span>
								<span>Invite People</span>
							</router-link>
						</div>
					</div>
				</div>

				<div class="circle-sidebar__footer">
					<router-link to="/" @click="ui.closeSidebar">
						<span>←</span>
						<span>Dashboard</span>
					</router-link>
				</div>
			</aside>

			<!-- Circle content -->
			<div class="circle-content">
				<router-view v-slot="{ Component }">
					<component :is="Component" v-bind="route.params" :members="members" />
				</router-view>
			</div>
		</template>
	</div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, computed, watch } from "vue";
import { useCircleStore } from "../stores/circles";
import { useAuthStore } from "../stores/auth";
import { useUIStore } from "../stores/ui";
import { useToastStore } from "../stores/toast";
import axios from "../api";
import { useRouter, useRoute } from "vue-router";
import { showBrowserNotification } from "../utils/notifications";
import { useSocketStore } from "../stores/socket";

const props = defineProps(["id"]);
const circleStore = useCircleStore();
const auth = useAuthStore();
const ui = useUIStore();
const toast = useToastStore();
const socketStore = useSocketStore();
const router = useRouter();
const route = useRoute();

const members = ref([]);
const error = ref(null);
const searchQuery = ref("");
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

const canManage = computed(() => {
	const role = circleStore.activeCircle?.role;
	return role === "admin" || role === "mod";
});

const isAdmin = computed(() => circleStore.activeCircle?.role === "admin");

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
	return (circleStore.chatMessages || []).filter(isUnread).length;
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
		const res = await axios.get(`/api/circles/${props.id}/members`);
		members.value = res.data;

		if (members.value.length <= 8) {
			membersExpanded.value = true;
		}
	} catch (err) {
		console.error("Failed to load circle data:", err);
		error.value = {
			title: "Access Denied",
			message: "You are not a member of this circle or it has been deleted.",
		};
	}
};

const togglePin = async (tagId, isPinned) => {
	try {
		await circleStore.pinTag(props.id, tagId, isPinned);
	} catch (err) {
		toast.error("Failed to update tag pin");
	}
};

onMounted(() => {
	loadCircleData();
	window.addEventListener("refresh-members", loadCircleData);
});

onUnmounted(() => {
	window.removeEventListener("refresh-members", loadCircleData);
});

watch(() => props.id, loadCircleData);
watch(
	() => route.fullPath,
	() => {
		ui.closeSidebar();
	},
);
</script>
