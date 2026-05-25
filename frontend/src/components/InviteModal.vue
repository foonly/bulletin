<script setup>
import { ref, computed } from "vue";
import { useCircleStore } from "../stores/circles";
import { useToastStore } from "../stores/toast";

const props = defineProps({
	show: Boolean,
	id: String, // Circle ID
});

const emit = defineEmits(["close", "created"]);
const circleStore = useCircleStore();
const toast = useToastStore();

const inviteForm = ref({
	role_to_grant: "standard",
	max_uses: null,
	expires_in_hrs: null,
});

const canInviteAdmins = computed(() => {
	return circleStore.activeCircle?.role === "admin";
});

const generateInvite = async () => {
	try {
		await circleStore.createInvite(props.id, {
			role_to_grant: inviteForm.value.role_to_grant,
			max_uses: inviteForm.value.max_uses
				? parseInt(inviteForm.value.max_uses)
				: null,
			expires_in_hrs: inviteForm.value.expires_in_hrs
				? parseInt(inviteForm.value.expires_in_hrs)
				: null,
		});
		toast.success("Invite link generated!");
		emit("created");
		emit("close");
		// Reset form
		inviteForm.value = {
			role_to_grant: "standard",
			max_uses: null,
			expires_in_hrs: null,
		};
	} catch (err) {
		toast.error("Failed to generate invite");
	}
};
</script>

<template>
	<div
		v-if="show"
		class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-4"
	>
		<div
			class="bg-gray-800 rounded-lg shadow-2xl w-full max-w-md border border-gray-700"
		>
			<div class="p-6 space-y-4">
				<div class="flex items-center justify-between mb-2">
					<h2 class="text-xl font-bold">Generate New Invite</h2>
					<button @click="$emit('close')" class="text-gray-400 hover:text-white">
						✕
					</button>
				</div>

				<div class="space-y-4">
					<div class="flex flex-col">
						<label class="text-xs text-gray-400 mb-1 uppercase font-bold"
							>Role to Grant</label
						>
						<select
							v-model="inviteForm.role_to_grant"
							class="bg-gray-700 p-2 rounded focus:outline-none focus:ring-1 focus:ring-purple-500"
						>
							<option value="guest">Guest</option>
							<option value="standard">Standard</option>
							<option value="mod">Moderator</option>
							<option v-if="canInviteAdmins" value="admin">Admin</option>
						</select>
					</div>

					<div class="grid grid-cols-2 gap-4">
						<div class="flex flex-col">
							<label class="text-xs text-gray-400 mb-1 uppercase font-bold"
								>Max Uses</label
							>
							<input
								v-model="inviteForm.max_uses"
								type="number"
								placeholder="∞"
								class="bg-gray-700 p-2 rounded focus:outline-none focus:ring-1 focus:ring-purple-500"
							/>
						</div>
						<div class="flex flex-col">
							<label class="text-xs text-gray-400 mb-1 uppercase font-bold"
								>Expires (hrs)</label
							>
							<input
								v-model="inviteForm.expires_in_hrs"
								type="number"
								placeholder="Never"
								class="bg-gray-700 p-2 rounded focus:outline-none focus:ring-1 focus:ring-purple-500"
							/>
						</div>
					</div>
				</div>

				<div class="flex justify-end space-x-3 pt-4 border-t border-gray-700">
					<button
						@click="$emit('close')"
						class="px-4 py-2 text-sm text-gray-400 hover:text-white transition"
					>
						Cancel
					</button>
					<button
						@click="generateInvite"
						class="px-6 py-2 bg-purple-600 hover:bg-purple-700 rounded font-bold transition text-sm"
					>
						Create Invite
					</button>
				</div>
			</div>
		</div>
	</div>
</template>
