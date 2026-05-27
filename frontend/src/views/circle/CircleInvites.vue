<script setup>
import { ref, onMounted, computed } from "vue";
import { useCircleStore } from "../../stores/circles";
import { useToastStore } from "../../stores/toast";
import InviteModal from "../../components/InviteModal.vue";

const props = defineProps(["id"]);
const circleStore = useCircleStore();
const toast = useToastStore();

const showInviteModal = ref(false);

const activeInvites = computed(() =>
	circleStore.invites.filter((i) => {
		const notExpired = !i.expires_at || new Date(i.expires_at) > new Date();
		const notDepleted = !i.max_uses || i.used_count < i.max_uses;
		return notExpired && notDepleted;
	}),
);

const inactiveInvites = computed(() =>
	circleStore.invites.filter((i) => {
		const expired = i.expires_at && new Date(i.expires_at) <= new Date();
		const depleted = i.max_uses && i.used_count >= i.max_uses;
		return expired || depleted;
	}),
);

const formatDate = (dateStr) => new Date(dateStr).toLocaleString();

const copyInviteLink = (code) => {
	const url = `${window.location.origin}/join/${code}`;
	navigator.clipboard.writeText(url);
	toast.success("Invite link copied to clipboard!");
};

const revokeInvite = async (inviteId) => {
	try {
		await circleStore.deleteInvite(props.id, inviteId);
		toast.success("Invite removed");
	} catch (err) {
		toast.error("Failed to remove invite");
	}
};

const clearAllInactive = async () => {
	if (!confirm("Clear all inactive and depleted invites?")) return;
	try {
		for (const invite of inactiveInvites.value) {
			await circleStore.deleteInvite(props.id, invite.id);
		}
		toast.success("Inactive invites cleared");
	} catch (err) {
		toast.error("Failed to clear some invites");
	}
};

onMounted(() => {
	circleStore.fetchInvites(props.id);
});
</script>

<template>
	<div class="circle-settings">
		<section class="section-card">
			<div class="section-card__header" style="padding-bottom: 0.75rem; border-bottom: 1px solid var(--border)">
				<h3>Manage Invites</h3>
				<button class="btn btn-primary btn-sm" @click="showInviteModal = true">+ New Invite</button>
			</div>

			<div v-if="activeInvites.length > 0" style="display: flex; flex-direction: column; gap: 0.75rem">
				<h4 style="font-size: var(--text-sm); font-weight: 700; color: var(--fg-2)">Active Invites</h4>
				<div v-for="invite in activeInvites" :key="invite.id" class="invite-row">
					<div style="display: flex; flex-direction: column; gap: 0.25rem">
						<div style="display: flex; align-items: center; gap: 0.5rem">
							<code class="invite-code">{{ invite.code }}</code>
							<button
								class="btn-icon"
								style="font-size: var(--text-base)"
								title="Copy invite link"
								@click="copyInviteLink(invite.code)"
							>🔗</button>
							<span style="font-size: var(--text-xs); color: var(--fg-3)">
								grants {{ invite.role_to_grant }} · Issued by {{ invite.created_by }}
							</span>
						</div>
						<div style="font-size: 10px; color: var(--fg-3)">
							Uses: {{ invite.used_count }} / {{ invite.max_uses || "∞" }} ·
							Expires: {{ invite.expires_at ? formatDate(invite.expires_at) : "Never" }}
						</div>
					</div>
					<button class="btn btn-ghost btn-sm" style="color: var(--danger)" @click="revokeInvite(invite.id)">Revoke</button>
				</div>
			</div>

			<div v-if="inactiveInvites.length > 0" style="display: flex; flex-direction: column; gap: 0.75rem; opacity: 0.6">
				<div style="display: flex; align-items: center; justify-content: space-between">
					<h4 style="font-size: var(--text-sm); font-weight: 700; color: var(--fg-3)">Inactive / Depleted</h4>
					<button class="btn btn-ghost btn-sm" style="color: var(--danger)" @click="clearAllInactive">Clear All</button>
				</div>
				<div v-for="invite in inactiveInvites" :key="invite.id" class="invite-row inactive">
					<div style="display: flex; flex-direction: column; gap: 0.25rem">
						<div style="display: flex; align-items: center; gap: 0.5rem">
							<code class="invite-code inactive">{{ invite.code }}</code>
							<span style="font-size: var(--text-xs); color: var(--fg-3)">
								grants {{ invite.role_to_grant }} · Issued by {{ invite.created_by }}
							</span>
						</div>
						<div style="font-size: 10px; color: var(--fg-3)">
							Uses: {{ invite.used_count }} / {{ invite.max_uses || "∞" }} ·
							<span style="font-weight: 700; text-transform: uppercase">
								{{ invite.max_uses && invite.used_count >= invite.max_uses ? "Depleted" : "Expired" }}
							</span>
						</div>
					</div>
					<button class="btn btn-ghost btn-sm" style="color: var(--fg-3)" @click="revokeInvite(invite.id)">Clear</button>
				</div>
			</div>

			<div v-if="circleStore.invites.length === 0" class="empty-state">
				<div class="empty-state__icon">✉️</div>
				<p>You haven't created any invite links yet.</p>
				<button class="btn btn-primary btn-sm" @click="showInviteModal = true">Create your first invite</button>
			</div>
		</section>
	</div>

	<InviteModal
		:show="showInviteModal"
		:id="id"
		@close="showInviteModal = false"
		@created="circleStore.fetchInvites(id)"
	/>
</template>
