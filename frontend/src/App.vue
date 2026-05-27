<script setup>
import ToastContainer from "./components/ToastContainer.vue";
</script>

<template>
	<div class="app-wrapper">
		<router-view></router-view>
		<ToastContainer />
	</div>
</template>
