<script setup>
import { onMounted, watch } from "vue";
import { useAuthStore } from "./stores/auth";
import { useSocketStore } from "./stores/socket";
import ToastContainer from "./components/ToastContainer.vue";

const auth = useAuthStore();
const socket = useSocketStore();

onMounted(() => {
	if (auth.user) {
		socket.connect();
	}
});

watch(
	() => auth.user,
	(user) => {
		if (user) {
			socket.connect();
		} else {
			socket.disconnect();
		}
	},
);
</script>

<template>
	<router-view></router-view>
	<ToastContainer />
</template>
