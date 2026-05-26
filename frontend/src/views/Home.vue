<template>
	<div class="flex h-screen bg-gray-900 text-gray-100 overflow-hidden">
		<!-- Sidebar -->
		<div
			class="w-20 bg-gray-950 flex flex-col items-center py-4 space-y-4 border-r border-gray-800"
		>
			<div
				v-for="circle in circleStore.circles"
				:key="circle.id"
				@click="selectCircle(circle)"
				:class="[
					'w-12 h-12 rounded-3xl cursor-pointer flex items-center justify-center font-bold text-xl transition-all hover:rounded-xl relative group',
					activeCircle?.id === circle.id
						? 'bg-purple-600 rounded-xl'
						: 'bg-gray-800 hover:bg-purple-500',
				]"
				:title="circle.name"
			>
				{{
					circle.name && circle.name.length > 0
						? circle.name[0].toUpperCase()
						: "?"
				}}
				<!-- Unread Badge -->
				<div
					v-if="circle.unread_count > 0"
					class="absolute -top-1 -right-1 bg-red-500 text-white text-[10px] min-w-[18px] h-[18px] flex items-center justify-center rounded-full border-2 border-gray-950 font-bold px-1"
				>
					{{ circle.unread_count > 99 ? "99+" : circle.unread_count }}
				</div>

				<!-- Tooltip -->
				<div
					class="absolute left-16 bg-gray-800 text-white text-xs py-1 px-2 rounded opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none whitespace-nowrap z-50 border border-gray-700 shadow-xl"
				>
					{{ circle.name }}
				</div>
			</div>
			<div
				@click="showJoinModal = true"
				class="w-12 h-12 rounded-3xl bg-gray-800 cursor-pointer flex items-center justify-center text-blue-500 hover:bg-blue-600 hover:text-white transition-all hover:rounded-xl group relative"
				title="Join a Circle"
			>
				<span class="text-xl">#</span>
				<!-- Tooltip -->
				<div
					class="absolute left-16 bg-gray-800 text-white text-xs py-1 px-2 rounded opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none whitespace-nowrap z-50 border border-gray-700 shadow-xl"
				>
					Join a Circle
				</div>
			</div>
			<div
				@click="showCreateModal = true"
				class="w-12 h-12 rounded-3xl bg-gray-800 cursor-pointer flex items-center justify-center text-green-500 hover:bg-green-600 hover:text-white transition-all hover:rounded-xl"
			>
				+
			</div>
		</div>

		<!-- Create Circle Modal -->
		<div
			v-if="showCreateModal"
			class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-4"
		>
			<div
				class="bg-gray-800 rounded-lg shadow-2xl w-full max-w-md border border-gray-700"
			>
				<div class="p-6 space-y-4">
					<h2 class="text-xl font-bold">Create a New Circle</h2>
					<div class="space-y-1">
						<label class="text-xs text-gray-400 font-bold uppercase"
							>Name</label
						>
						<input
							v-model="newCircle.name"
							placeholder="Circle name"
							class="w-full bg-gray-900 border border-gray-700 p-2 rounded focus:outline-none focus:border-purple-500"
						/>
					</div>
					<div class="space-y-1">
						<label class="text-xs text-gray-400 font-bold uppercase"
							>Description</label
						>
						<textarea
							v-model="newCircle.description"
							placeholder="What is this circle about?"
							class="w-full bg-gray-900 border border-gray-700 p-2 rounded focus:outline-none focus:border-purple-500 h-24"
						></textarea>
					</div>
					<div class="flex justify-end space-x-3 pt-2">
						<button
							@click="showCreateModal = false"
							class="px-4 py-2 text-sm text-gray-400 hover:text-white transition"
						>
							Cancel
						</button>
						<button
							@click="handleCreateCircle"
							:disabled="!newCircle.name.trim()"
							class="px-4 py-2 bg-purple-600 hover:bg-purple-700 rounded font-bold transition disabled:opacity-50 disabled:cursor-not-allowed"
						>
							Create Circle
						</button>
					</div>
				</div>
			</div>
		</div>

		<!-- Join Circle Modal -->
		<div
			v-if="showJoinModal"
			class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-4"
		>
			<div
				class="bg-gray-800 rounded-lg shadow-2xl w-full max-w-md border border-gray-700"
			>
				<div class="p-6 space-y-4">
					<h2 class="text-xl font-bold">Join a Circle</h2>
					<p class="text-sm text-gray-400">
						Enter an invite code to join an existing circle.
					</p>
					<div class="space-y-1">
						<label class="text-xs text-gray-400 font-bold uppercase"
							>Invite Code</label
						>
						<input
							v-model="joinInviteCode"
							placeholder="e.g. welcome"
							class="w-full bg-gray-900 border border-gray-700 p-2 rounded focus:outline-none focus:border-purple-500"
							@keyup.enter="handleJoinCircle"
						/>
					</div>
					<div class="flex justify-end space-x-3 pt-2">
						<button
							@click="showJoinModal = false"
							class="px-4 py-2 text-sm text-gray-400 hover:text-white transition"
						>
							Cancel
						</button>
						<button
							@click="handleJoinCircle"
							:disabled="!joinInviteCode.trim()"
							class="px-4 py-2 bg-blue-600 hover:bg-blue-700 rounded font-bold transition disabled:opacity-50 disabled:cursor-not-allowed"
						>
							Join Circle
						</button>
					</div>
				</div>
			</div>
		</div>

		<!-- Main Content -->
		<div class="flex-1 flex flex-col">
			<header
				class="h-12 border-b border-gray-800 flex items-center px-4 justify-between bg-gray-900"
			>
				<h2 class="font-bold text-lg text-gray-400">{{ siteName }}</h2>
				<div class="flex items-center space-y-0 space-x-4">
					<router-link
						to="/settings"
						class="text-sm text-gray-400 hover:text-purple-400"
						>{{ auth.user?.username }}</router-link
					>
					<button
						@click="handleLogout"
						class="text-sm text-red-400 hover:underline"
					>
						Logout
					</button>
				</div>
			</header>

			<div class="flex-1 overflow-y-auto bg-gray-950/20">
				<router-view></router-view>
				<div v-if="route.path === '/'" class="h-full p-8">
					<div class="max-w-6xl mx-auto">
						<div class="flex items-center justify-between mb-8">
							<h1 class="text-3xl font-bold">Your Circles</h1>
							<div class="flex space-x-4">
								<button
									@click="showJoinModal = true"
									class="bg-gray-700 hover:bg-gray-600 px-6 py-2 rounded-lg font-bold transition flex items-center space-x-2 border border-gray-600"
								>
									<span>#</span>
									<span>Join Circle</span>
								</button>
								<button
									@click="showCreateModal = true"
									class="bg-purple-600 hover:bg-purple-700 px-6 py-2 rounded-lg font-bold transition flex items-center space-x-2"
								>
									<span>+</span>
									<span>Create Circle</span>
								</button>
							</div>
						</div>

						<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
							<div
								v-for="circle in circleStore.circles"
								:key="circle.id"
								@click="selectCircle(circle)"
								class="bg-gray-800 border border-gray-700 rounded-xl p-6 cursor-pointer hover:border-purple-500 transition-all hover:shadow-xl group"
							>
								<div class="flex items-start justify-between mb-4">
									<div
										class="w-12 h-12 bg-purple-600 rounded-lg flex items-center justify-center font-bold text-2xl shadow-lg group-hover:scale-110 transition-transform"
									>
										{{ circle.name ? circle.name[0].toUpperCase() : "?" }}
									</div>
									<div class="flex flex-col items-end space-y-1">
										<span
											class="text-[10px] uppercase font-bold px-2 py-0.5 rounded bg-gray-900 text-gray-400 border border-gray-700"
										>
											{{ circle.role }}
										</span>
										<span class="text-xs text-gray-500">
											{{ circle.member_count }} members
										</span>
									</div>
								</div>

								<h2
									class="text-xl font-bold mb-2 group-hover:text-purple-400 transition-colors"
								>
									{{ circle.name }}
								</h2>
								<p class="text-gray-400 text-sm line-clamp-2 mb-6 h-10">
									{{ circle.description || "No description provided." }}
								</p>

								<div class="space-y-3 pt-4 border-t border-gray-700/50">
									<!-- Unread Stats -->
									<div class="flex items-center space-x-4">
										<div
											class="flex items-center space-x-1.5"
											:title="circle.unread_post_count + ' unread posts'"
										>
											<span class="text-lg">📰</span>
											<span
												:class="[
													'text-xs font-bold',
													circle.unread_post_count > 0
														? 'text-purple-400'
														: 'text-gray-600',
												]"
											>
												{{ circle.unread_post_count }}
											</span>
										</div>
										<div
											class="flex items-center space-x-1.5"
											:title="circle.unread_chat_count + ' unread messages'"
										>
											<span class="text-lg">💬</span>
											<span
												:class="[
													'text-xs font-bold',
													circle.unread_chat_count > 0
														? 'text-purple-400'
														: 'text-gray-600',
												]"
											>
												{{ circle.unread_chat_count }}
											</span>
										</div>
									</div>

									<!-- Last Activity -->
									<div
										v-if="circle.last_post_at"
										class="flex flex-col space-y-1"
									>
										<span class="text-[10px] uppercase font-bold text-gray-600"
											>Last Activity</span
										>
										<div class="flex items-center justify-between">
											<span
												class="text-xs text-gray-300 truncate max-w-[150px]"
												:title="circle.last_post_title"
											>
												{{ circle.last_post_title || "New post" }}
											</span>
											<span class="text-[10px] text-gray-500 italic">
												{{ formatDate(circle.last_post_at) }}
											</span>
										</div>
									</div>
									<div v-else class="text-[10px] italic text-gray-600">
										No recent activity
									</div>
								</div>
							</div>
						</div>
					</div>
				</div>
			</div>
		</div>
	</div>
</template>

<script setup>
import { onMounted, onUnmounted, computed, watch, ref } from "vue";
import { useAuthStore } from "../stores/auth";
import { useCircleStore } from "../stores/circles";
import { useToastStore } from "../stores/toast";
import { useRouter, useRoute } from "vue-router";

const auth = useAuthStore();
const circleStore = useCircleStore();
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
const newCircle = ref({
	name: "",
	description: "",
});

const showJoinModal = ref(false);
const joinInviteCode = ref("");

const activeCircle = computed(() => circleStore.activeCircle);

const syncActiveCircle = () => {
	const circleId = route.params.id;
	if (circleId && circleStore.circles.length > 0) {
		const circle = circleStore.circles.find((c) => c.id === circleId);
		if (circle) {
			circleStore.activeCircle = circle;
		}
	} else if (!circleId) {
		circleStore.activeCircle = null;
	}
};

const loadingCircles = ref(false);
let pollInterval = null;

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

	// Background polling for unread counts (every 15 seconds)
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
	router.push("/login");
};
</script>
