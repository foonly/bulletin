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
				class="w-12 h-12 rounded-3xl bg-gray-800 cursor-pointer flex items-center justify-center text-green-500 hover:bg-green-600 hover:text-white transition-all hover:rounded-xl"
			>
				+
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
import { onMounted, computed, watch } from "vue";
import { useAuthStore } from "../stores/auth";
import { useCircleStore } from "../stores/circles";
import { useRouter, useRoute } from "vue-router";

const auth = useAuthStore();
const circleStore = useCircleStore();
const router = useRouter();
const route = useRoute();

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

const handleLogout = async () => {
	await auth.logout();
	router.push("/login");
};
</script>
