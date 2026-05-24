<script setup>
import { useToastStore } from "../stores/toast";

const toastStore = useToastStore();
</script>

<template>
	<div
		class="fixed bottom-4 right-4 z-[100] flex flex-col gap-2 pointer-events-none"
	>
		<transition-group name="toast">
			<div
				v-for="toast in toastStore.toasts"
				:key="toast.id"
				:class="[
					'pointer-events-auto min-w-[250px] max-w-md p-4 rounded-lg shadow-2xl border flex items-start justify-between gap-4 transition-all duration-300',
					toast.type === 'success'
						? 'bg-green-900 border-green-700 text-green-100'
						: '',
					toast.type === 'error' ? 'bg-red-900 border-red-700 text-red-100' : '',
					toast.type === 'info'
						? 'bg-blue-900 border-blue-700 text-blue-100'
						: '',
					toast.type === 'warning'
						? 'bg-yellow-900 border-yellow-700 text-yellow-100'
						: '',
				]"
			>
				<p class="text-sm font-medium">{{ toast.message }}</p>
				<button
					@click="toastStore.removeToast(toast.id)"
					class="text-white/50 hover:text-white transition-colors"
				>
					<svg
						xmlns="http://www.w3.org/2000/svg"
						class="h-4 w-4"
						viewBox="0 0 20 20"
						fill="currentColor"
					>
						<path
							fill-rule="evenodd"
							d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z"
							clip-rule="evenodd"
						/>
					</svg>
				</button>
			</div>
		</transition-group>
	</div>
</template>

<style scoped>
.toast-enter-from {
	opacity: 0;
	transform: translateX(100%);
}
.toast-enter-to {
	opacity: 1;
	transform: translateX(0);
}
.toast-leave-from {
	opacity: 1;
	transform: translateX(0);
}
.toast-leave-to {
	opacity: 0;
	transform: translateX(100%);
}
.toast-move {
	transition: transform 0.3s ease;
}
</style>
