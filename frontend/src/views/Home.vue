<template>
	<div class="app-layout" :class="{ 'sidebar-open': ui.sidebarOpen }">
		<!-- Sidebar overlay (mobile) -->
		<div class="sidebar-overlay" @click="ui.closeSidebar"></div>

		<!-- Circle rail -->
		<nav class="app-sidebar">
			<div
				v-for="circle in circleStore.circles"
				:key="circle.id"
				class="circle-icon"
				:class="{ active: activeCircle?.id === circle.id }"
				:title="circle.name"
				@click="selectCircle(circle)"
			>
				{{ circle.name?.length > 0 ? circle.name[0].toUpperCase() : "?" }}
				<div v-if="circle.unread_count > 0" class="badge">
					{{ circle.unread_count > 99 ? "99+" : circle.unread_count }}
				</div>
				<span class="tooltip">{{ circle.name }}</span>
			</div>
		</nav>

		<!-- Create Circle Modal -->
		<div
			v-if="showCreateModal"
			class="modal-backdrop"
			@click.self="showCreateModal = false"
		>
			<div class="modal">
				<div class="modal__header">
					<h2>Create a New Circle</h2>
					<button class="btn-icon" @click="showCreateModal = false">✕</button>
				</div>
				<div class="field">
					<label class="label-uppercase">Name</label>
					<input v-model="newCircle.name" placeholder="Circle name" />
				</div>
				<div class="field">
					<label class="label-uppercase">Description</label>
					<textarea
						v-model="newCircle.description"
						placeholder="What is this circle about?"
						style="height: 96px"
					></textarea>
				</div>
				<div class="form-actions">
					<button class="btn btn-ghost" @click="showCreateModal = false">
						Cancel
					</button>
					<button
						class="btn btn-primary"
						@click="handleCreateCircle"
						:disabled="!newCircle.name.trim()"
					>
						Create Circle
					</button>
				</div>
			</div>
		</div>

		<!-- Join Circle Modal -->
		<div
			v-if="showJoinModal"
			class="modal-backdrop"
			@click.self="showJoinModal = false"
		>
			<div class="modal">
				<div class="modal__header">
					<h2>Join a Circle</h2>
					<button class="btn-icon" @click="showJoinModal = false">✕</button>
				</div>
				<p style="font-size: var(--text-sm); color: var(--fg-2)">
					Enter an invite code to join an existing circle.
				</p>
				<div class="field">
					<label class="label-uppercase">Invite Code</label>
					<input
						v-model="joinInviteCode"
						placeholder="e.g. welcome"
						@keyup.enter="handleJoinCircle"
					/>
				</div>
				<div class="form-actions">
					<button class="btn btn-ghost" @click="showJoinModal = false">
						Cancel
					</button>
					<button
						class="btn btn-primary"
						@click="handleJoinCircle"
						:disabled="!joinInviteCode.trim()"
					>
						Join Circle
					</button>
				</div>
			</div>
		</div>

		<!-- Main content area -->
		<div class="app-main">
			<header class="app-header">
				<div class="app-header__left">
					<button class="btn-icon mobile-only" @click="ui.toggleSidebar">
						<svg
							v-if="!ui.sidebarOpen"
							xmlns="http://www.w3.org/2000/svg"
							fill="none"
							viewBox="0 0 24 24"
							stroke="currentColor"
						>
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M4 6h16M4 12h16M4 18h16"
							/>
						</svg>
						<svg
							v-else
							xmlns="http://www.w3.org/2000/svg"
							fill="none"
							viewBox="0 0 24 24"
							stroke="currentColor"
						>
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M6 18L18 6M6 6l12 12"
							/>
						</svg>
					</button>
					<router-link to="/" class="app-header__title">
						<LogoIcon class="app-header__logo" />
						<span>{{ siteName }}</span>
					</router-link>
				</div>
				<nav class="app-header__nav">
					<router-link to="/settings">{{ auth.user?.username }}</router-link>
					<button class="app-header__logout" @click="handleLogout">
						Logout
					</button>
				</nav>
			</header>

			<main class="app-content">
				<router-view v-slot="{ Component }">
					<component :is="Component" />
				</router-view>

				<!-- Dashboard (shown when at root path) -->
				<div v-if="route.path === '/'" class="home-dashboard">
					<div class="home-dashboard__header">
						<h1>Your Circles</h1>
						<div style="display: flex; gap: 0.75rem">
							<button class="btn btn-secondary" @click="showJoinModal = true">
								# Join Circle
							</button>
							<button class="btn btn-primary" @click="showCreateModal = true">
								+ Create Circle
							</button>
						</div>
					</div>

					<div class="circles-grid">
						<div
							v-for="circle in circleStore.circles"
							:key="circle.id"
							class="circle-card"
							@click="selectCircle(circle)"
						>
							<div class="circle-card__header">
								<div class="circle-card__avatar">
									{{ circle.name ? circle.name[0].toUpperCase() : "?" }}
								</div>
								<div class="circle-card__meta">
									<span class="role-badge">{{ circle.role }}</span>
									<span style="font-size: var(--text-xs); color: var(--fg-3)">
										{{ circle.member_count }} members
									</span>
								</div>
							</div>

							<h2>{{ circle.name }}</h2>
							<p class="circle-card__desc">
								{{ circle.description || "No description provided." }}
							</p>

							<div class="circle-card__stats">
								<div class="circle-card__stat-row">
									<div
										class="circle-card__stat"
										:title="circle.unread_post_count + ' unread posts'"
									>
										<span>📰</span>
										<span
											class="circle-card__stat-count"
											:class="{ 'has-unread': circle.unread_post_count > 0 }"
											>{{ circle.unread_post_count }}</span
										>
									</div>
									<div
										class="circle-card__stat"
										:title="circle.unread_chat_count + ' unread messages'"
									>
										<span>💬</span>
										<span
											class="circle-card__stat-count"
											:class="{ 'has-unread': circle.unread_chat_count > 0 }"
											>{{ circle.unread_chat_count }}</span
										>
									</div>
								</div>

								<div v-if="circle.last_post_at">
									<div class="circle-card__activity-label">Last Activity</div>
									<div class="circle-card__activity-row">
										<span
											class="circle-card__activity-title"
											:title="circle.last_post_title"
											>{{ circle.last_post_title || "New post" }}</span
										>
										<span class="circle-card__activity-time">
											{{ formatDate(circle.last_post_at) }}
										</span>
									</div>
								</div>
								<div
									v-else
									style="
										font-size: 10px;
										font-style: italic;
										color: var(--fg-3);
									"
								>
									No recent activity
								</div>
							</div>
						</div>
					</div>
				</div>
			</main>
		</div>
	</div>
</template>

<script setup>
import { onMounted, onUnmounted, computed, watch, ref } from "vue";
import { useAuthStore } from "../stores/auth";
import { useCircleStore } from "../stores/circles";
import { useUIStore } from "../stores/ui";
import { useToastStore } from "../stores/toast";
import { useRouter, useRoute } from "vue-router";
import LogoIcon from "../components/LogoIcon.vue";

const auth = useAuthStore();
const circleStore = useCircleStore();
const ui = useUIStore();
const toast = useToastStore();
const router = useRouter();
const route = useRoute();

const siteName = import.meta.env.VITE_SITE_NAME || "Bulletin";

const formatDate = (dateStr) => {
	if (!dateStr) return "";
	const date = new Date(dateStr);
	const now = new Date();
	const diff = now - date;

	if (diff < 24 * 60 * 60 * 1000) {
		return date.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" });
	}
	return date.toLocaleDateString([], { month: "short", day: "numeric" });
};

const showCreateModal = ref(false);
const newCircle = ref({ name: "", description: "" });

const showJoinModal = ref(false);
const joinInviteCode = ref("");

const activeCircle = computed(() => circleStore.activeCircle);

const applyPalette = (circle) => {
	const palette = circle?.palette || "violet";
	const palettes = ["violet", "ocean", "ember", "forest", "rose", "slate"];
	palettes.forEach((p) => document.body.classList.remove(`palette-${p}`));
	if (palette !== "violet") {
		document.body.classList.add(`palette-${palette}`);
	}
};

const syncActiveCircle = () => {
	const circleId = route.params.id;
	if (circleId && circleStore.circles.length > 0) {
		const circle = circleStore.circles.find((c) => c.id === circleId);
		if (circle) {
			circleStore.activeCircle = circle;
			applyPalette(circle);
		}
	} else if (!circleId) {
		circleStore.activeCircle = null;
		applyPalette(null);
	}
};

const loadingCircles = ref(false);
let pollInterval = null;

watch(
	() => ui.sidebarOpen,
	(open) => {
		if (open) {
			document.body.classList.add("lock-scroll");
		} else {
			document.body.classList.remove("lock-scroll");
		}
	},
);

onMounted(async () => {
	if (circleStore.circles.length === 0 && !loadingCircles.value) {
		loadingCircles.value = true;
		try {
			await circleStore.fetchCircles();
		} finally {
			loadingCircles.value = false;
		}
	}
	syncActiveCircle();

	pollInterval = setInterval(() => {
		if (auth.user) {
			circleStore.refreshUnreadCounts();
			if (circleStore.activeCircle) {
				circleStore.refreshActiveTags();
			}
		}
	}, 15000);
});

onUnmounted(() => {
	if (pollInterval) clearInterval(pollInterval);
});

watch(() => route.params.id, syncActiveCircle);

const selectCircle = (circle) => {
	circleStore.activeCircle = circle;
	applyPalette(circle);
	ui.closeSidebar();
	router.push(`/circle/${circle.id}`);
};

const handleCreateCircle = async () => {
	if (!newCircle.value.name.trim()) return;
	try {
		const res = await circleStore.createCircle(newCircle.value);
		toast.success(`Circle "${newCircle.value.name}" created!`);
		showCreateModal.value = false;
		newCircle.value = { name: "", description: "" };
		router.push(`/circle/${res.id}`);
	} catch (err) {
		toast.error("Failed to create circle");
	}
};

const handleJoinCircle = async () => {
	if (!joinInviteCode.value.trim()) return;
	try {
		const res = await circleStore.joinCircle(joinInviteCode.value);
		toast.success("Joined circle!");
		showJoinModal.value = false;
		joinInviteCode.value = "";
		router.push(`/circle/${res.id}`);
	} catch (err) {
		toast.error(err.response?.data || "Failed to join circle");
	}
};

const handleLogout = async () => {
	await auth.logout();
	applyPalette(null);
	router.push("/login");
};
</script>
