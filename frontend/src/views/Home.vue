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
					'w-12 h-12 rounded-3xl cursor-pointer flex items-center justify-center font-bold text-xl transition-all hover:rounded-xl',
					activeCircle?.id === circle.id
						? 'bg-purple-600 rounded-xl'
						: 'bg-gray-800 hover:bg-purple-500',
				]"
			>
				{{ circle.name[0].toUpperCase() }}
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

		<!-- Main Content -->
		<div class="flex-1 flex flex-col">
			<header
				class="h-12 border-b border-gray-800 flex items-center px-4 justify-between bg-gray-900"
			>
				<h2 class="font-bold text-lg">
					{{ activeCircle?.name || "Bulletin" }}
				</h2>
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

			<div class="flex-1 overflow-hidden">
				<router-view></router-view>
				<div
					v-if="route.path === '/'"
					class="h-full flex items-center justify-center text-gray-500"
				>
					Select a circle to start communicating
				</div>
			</div>
		</div>
	</div>
</template>

<script setup>
import { onMounted, computed, watch, ref } from "vue";
import { useAuthStore } from "../stores/auth";
import { useCircleStore } from "../stores/circles";
import { useRouter, useRoute } from "vue-router";

const auth = useAuthStore();
const circleStore = useCircleStore();
const router = useRouter();
const route = useRoute();

const showCreateModal = ref(false);
const newCircle = ref({
	name: "",
	description: "",
});

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

onMounted(async () => {
	await circleStore.fetchCircles();
	syncActiveCircle();
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
		showCreateModal.value = false;
		newCircle.value = { name: "", description: "" };
		router.push(`/circle/${res.id}`);
	} catch (err) {
		alert("Failed to create circle");
	}
};

const handleLogout = async () => {
	await auth.logout();
	router.push("/login");
};
</script>
