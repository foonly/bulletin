<script setup>
import { ref, computed } from "vue";
import { useCircleStore } from "../stores/circles";
import { useToastStore } from "../stores/toast";

const props = defineProps({
	show: Boolean,
	id: String,
});

const emit = defineEmits(["close", "created"]);
const circleStore = useCircleStore();
const toast = useToastStore();

const inviteForm = ref({
	role_to_grant: "standard",
	max_uses: null,
	expires_in_hrs: null,
});

const canInviteAdmins = computed(() => circleStore.activeCircle?.role === "admin");

const generateInvite = async () => {
	try {
		await circleStore.createInvite(props.id, {
			role_to_grant: inviteForm.value.role_to_grant,
			max_uses: inviteForm.value.max_uses ? parseInt(inviteForm.value.max_uses) : null,
			expires_in_hrs: inviteForm.value.expires_in_hrs
				? parseInt(inviteForm.value.expires_in_hrs)
				: null,
		});
		toast.success("Invite link generated!");
		emit("created");
		emit("close");
		inviteForm.value = { role_to_grant: "standard", max_uses: null, expires_in_hrs: null };
	} catch (err) {
		toast.error("Failed to generate invite");
	}
};
</script>

<template>
	<div v-if="show" class="modal-backdrop" @click.self="$emit('close')">
		<div class="modal">
			<div class="modal__header">
				<h2>Generate New Invite</h2>
				<button class="btn-icon" @click="$emit('close')">✕</button>
			</div>

			<div class="field">
				<label class="label-uppercase">Role to Grant</label>
				<select v-model="inviteForm.role_to_grant">
					<option value="guest">Guest</option>
					<option value="standard">Standard</option>
					<option value="mod">Moderator</option>
					<option v-if="canInviteAdmins" value="admin">Admin</option>
				</select>
			</div>

			<div class="field-row">
				<div class="field">
					<label class="label-uppercase">Max Uses</label>
					<input v-model="inviteForm.max_uses" type="number" placeholder="∞" />
				</div>
				<div class="field">
					<label class="label-uppercase">Expires (hrs)</label>
					<input v-model="inviteForm.expires_in_hrs" type="number" placeholder="Never" />
				</div>
			</div>

			<div class="form-actions">
				<button class="btn btn-ghost" @click="$emit('close')">Cancel</button>
				<button class="btn btn-primary" @click="generateInvite">Create Invite</button>
			</div>
		</div>
	</div>
</template>
