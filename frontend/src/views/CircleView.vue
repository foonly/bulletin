<template>
	<div class="flex flex-col h-full">
		<!-- Tabs -->
		<div class="flex bg-gray-800 px-4">
			<button
				@click="tab = 'posts'"
				:class="[
					'px-4 py-2 border-b-2 transition',
					tab === 'posts'
						? 'border-purple-500 text-purple-500'
						: 'border-transparent text-gray-400 hover:text-gray-200',
				]"
			>
				Posts
			</button>
			<button
				@click="tab = 'chat'"
				:class="[
					'px-4 py-2 border-b-2 transition',
					tab === 'chat'
						? 'border-purple-500 text-purple-500'
						: 'border-transparent text-gray-400 hover:text-gray-200',
				]"
			>
				Chat
			</button>
			<button
				v-if="canManage"
				@click="tab = 'settings'"
				:class="[
					'px-4 py-2 border-b-2 transition',
					tab === 'settings'
						? 'border-purple-500 text-purple-500'
						: 'border-transparent text-gray-400 hover:text-gray-200',
				]"
			>
				Settings
			</button>
		</div>

		<!-- Content -->
		<div class="flex-1 flex overflow-hidden">
			<div class="flex-1 overflow-y-auto p-4">
				<div v-if="tab === 'posts'" class="space-y-4">
					<div class="bg-gray-800 p-4 rounded-lg">
						<input
							v-model="newPost.title"
							placeholder="Post Title"
							class="w-full bg-gray-700 p-2 rounded mb-2 border border-gray-600 focus:outline-none"
						/>
						<textarea
							v-model="newPost.content"
							placeholder="What's on your mind?"
							class="w-full bg-gray-700 p-2 rounded border border-gray-600 focus:outline-none"
							rows="3"
						></textarea>
						<div class="flex justify-end mt-2">
							<button
								@click="handleCreatePost"
								class="bg-purple-600 px-4 py-1 rounded font-bold hover:bg-purple-700"
							>
								Post
							</button>
						</div>
					</div>

					<div
						v-for="post in threadedPosts"
						:key="post.id"
						class="bg-gray-800 p-4 rounded-lg border border-gray-700"
					>
						<div class="flex items-center justify-between mb-2">
							<span class="font-bold text-purple-400">{{
								post.author_name
							}}</span>
							<span class="text-xs text-gray-500">{{
								formatDate(post.created_at)
							}}</span>
						</div>
						<h3 v-if="post.title" class="font-bold text-lg mb-1">
							{{ post.title }}
						</h3>
						<p class="text-gray-300">{{ post.content }}</p>

						<!-- Replies (simplified) -->
						<div
							v-if="post.replies && post.replies.length"
							class="mt-4 pl-4 border-l-2 border-gray-700 space-y-3"
						>
							<div
								v-for="reply in post.replies"
								:key="reply.id"
								class="text-sm"
							>
								<span class="font-bold text-gray-400"
									>{{ reply.author_name }}:</span
								>
								{{ reply.content }}
							</div>
						</div>
					</div>
				</div>

				<div v-else-if="tab === 'chat'" class="flex flex-col h-full">
					<div class="flex-1 overflow-y-auto space-y-2 mb-4" ref="chatBox">
						<div
							v-for="msg in circleStore.chatMessages"
							:key="msg.id"
							class="text-sm"
						>
							<span class="font-bold text-purple-400">{{ msg.username }}:</span>
							<span class="ml-2 text-gray-300">{{ msg.content }}</span>
							<span class="ml-2 text-[10px] text-gray-600">{{
								formatDate(msg.created_at)
							}}</span>
						</div>
					</div>
					<div class="flex space-x-2">
						<input
							v-model="chatInput"
							@keyup.enter="sendChatMessage"
							placeholder="Type a message..."
							class="flex-1 bg-gray-800 border border-gray-700 rounded px-4 py-2 focus:outline-none focus:border-purple-500"
						/>
						<button
							@click="sendChatMessage"
							class="bg-purple-600 px-6 rounded font-bold hover:bg-purple-700"
						>
							Send
						</button>
					</div>
				</div>

				<!-- Settings Tab -->
				<div
					v-else-if="tab === 'settings'"
					class="space-y-8 w-full max-w-2xl mx-auto"
				>
					<!-- Circle Config -->
					<section class="bg-gray-800 p-6 rounded-lg space-y-4">
						<h3 class="text-xl font-bold border-b border-gray-700 pb-2">
							Circle Settings
						</h3>
						<div class="grid grid-cols-2 gap-4">
							<div class="flex flex-col">
								<label class="text-xs text-gray-400 mb-1">Name</label>
								<input
									v-model="settings.name"
									class="bg-gray-700 p-2 rounded focus:outline-none focus:ring-1 focus:ring-purple-500"
								/>
							</div>
							<div class="flex flex-col">
								<label class="text-xs text-gray-400 mb-1"
									>Invite Min Role</label
								>
								<select
									v-model="settings.invite_min_role"
									class="bg-gray-700 p-2 rounded focus:outline-none"
								>
									<option value="guest">Guest</option>
									<option value="standard">Standard</option>
									<option value="mod">Moderator</option>
									<option value="admin">Admin</option>
								</select>
							</div>
							<div class="flex flex-col col-span-2">
								<label class="text-xs text-gray-400 mb-1">Description</label>
								<textarea
									v-model="settings.description"
									class="bg-gray-700 p-2 rounded focus:outline-none h-20"
								></textarea>
							</div>
							<div class="flex items-center space-x-2">
								<input
									type="checkbox"
									v-model="settings.allow_freeform_tags"
									id="freeform"
								/>
								<label for="freeform" class="text-sm"
									>Allow freeform tags</label
								>
							</div>
						</div>
						<div class="flex justify-end">
							<button
								@click="saveSettings"
								class="bg-green-600 px-4 py-2 rounded font-bold hover:bg-green-700"
							>
								Save Changes
							</button>
						</div>
					</section>

					<!-- Invite Generation -->
					<section class="bg-gray-800 p-6 rounded-lg space-y-4">
						<h3 class="text-xl font-bold border-b border-gray-700 pb-2">
							Generate Invite
						</h3>
						<div class="grid grid-cols-2 gap-4 text-sm">
							<div class="flex flex-col">
								<label class="text-xs text-gray-400 mb-1">Code</label>
								<input
									v-model="inviteForm.code"
									class="bg-gray-700 p-2 rounded focus:outline-none"
									placeholder="e.g. awesome-friends"
								/>
							</div>
							<div class="flex flex-col">
								<label class="text-xs text-gray-400 mb-1">Grant Role</label>
								<select
									v-model="inviteForm.role_to_grant"
									class="bg-gray-700 p-2 rounded"
								>
									<option value="guest">Guest</option>
									<option value="standard">Standard</option>
									<option value="mod">Moderator</option>
									<option value="admin">Admin</option>
								</select>
							</div>
							<div class="flex flex-col">
								<label class="text-xs text-gray-400 mb-1"
									>Max Uses (optional)</label
								>
								<input
									v-model.number="inviteForm.max_uses"
									type="number"
									class="bg-gray-700 p-2 rounded"
								/>
							</div>
							<div class="flex flex-col">
								<label class="text-xs text-gray-400 mb-1"
									>Expires in (hours)</label
								>
								<input
									v-model.number="inviteForm.expires_in_hrs"
									type="number"
									class="bg-gray-700 p-2 rounded"
								/>
							</div>
						</div>
						<div class="flex justify-end">
							<button
								@click="generateInvite"
								class="bg-purple-600 px-4 py-2 rounded font-bold hover:bg-purple-700"
							>
								Create Invite
							</button>
						</div>
					</section>

					<!-- Member Management (Admins only) -->
					<section v-if="isAdmin" class="bg-gray-800 p-6 rounded-lg space-y-4">
						<h3 class="text-xl font-bold border-b border-gray-700 pb-2">
							Manage Members
						</h3>
						<div class="space-y-2">
							<div
								v-for="m in members"
								:key="m.id"
								class="flex items-center justify-between bg-gray-700 p-3 rounded"
							>
								<div>
									<span class="font-bold">{{ m.username }}</span>
									<span class="text-xs text-gray-400 ml-2"
										>Invited by {{ m.invited_by }}</span
									>
								</div>
								<div class="flex items-center space-x-2">
									<button
										v-if="m.id !== auth.user.id"
										@click="kickMember(m.id)"
										class="text-red-400 hover:text-red-300 text-sm"
									>
										Kick
									</button>
								</div>
							</div>
						</div>
					</section>
				</div>
			</div>

			<!-- Right Sidebar (Members) -->
			<div
				class="w-64 bg-gray-900 border-l border-gray-800 p-4 overflow-y-auto hidden lg:block"
			>
				<h3
					class="text-xs font-bold text-gray-500 uppercase tracking-wider mb-4"
				>
					Members
				</h3>
				<div class="space-y-4">
					<div
						v-for="member in members"
						:key="member.id"
						class="flex items-center space-x-2"
					>
						<div
							:class="[
								'w-2 h-2 rounded-full',
								onlineUserIds.has(member.id) ? 'bg-green-500' : 'bg-gray-600',
							]"
						></div>
						<div>
							<div class="font-bold text-sm leading-tight">
								{{ member.username }}
							</div>
							<div class="text-[10px] text-gray-500">
								Invited by: {{ member.invited_by }}
							</div>
						</div>
					</div>
				</div>
			</div>
		</div>
	</div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, computed, watch, nextTick } from "vue";
import { useCircleStore } from "../stores/circles";
import { useAuthStore } from "../stores/auth";
import axios from "axios";

const props = defineProps(["id"]);
const circleStore = useCircleStore();
const auth = useAuthStore();
const tab = ref("posts");
const chatInput = ref("");
const chatBox = ref(null);
const newPost = ref({ title: "", content: "" });
const members = ref([]);
const onlineUserIds = ref(new Set());
const settings = ref({
	name: "",
	description: "",
	allow_freeform_tags: true,
	invite_min_role: "standard",
	chat_retention_days: 14,
	chat_retention_count: 50,
});
const inviteForm = ref({
	code: "",
	role_to_grant: "standard",
	max_uses: null,
	expires_in_hrs: null,
});

let ws = null;

const canManage = computed(() => {
	const role = circleStore.activeCircle?.role;
	return role === "admin" || role === "mod";
});

const isAdmin = computed(() => {
	return circleStore.activeCircle?.role === "admin";
});

const threadedPosts = computed(() => {
	const topLevel = circleStore.posts.filter((p) => !p.parent_id);
	return topLevel.map((p) => ({
		...p,
		replies: circleStore.posts.filter((r) => r.parent_id === p.id),
	}));
});

const formatDate = (dateStr) => {
	return new Date(dateStr).toLocaleString();
};

const loadCircleData = async () => {
	if (!circleStore.circles.length) {
		await circleStore.fetchCircles();
	}
	const circle = circleStore.circles.find((c) => c.id === props.id);
	if (circle) {
		circleStore.activeCircle = circle;
	}

	await circleStore.fetchPosts(props.id);
	await circleStore.fetchChatHistory(props.id);
	const res = await axios.get(`/api/circles/${props.id}/members`);
	members.value = res.data;

	if (circleStore.activeCircle) {
		settings.value = { ...circleStore.activeCircle };
	}

	connectWS();
};

const connectWS = () => {
	if (ws) ws.close();
	const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
	ws = new WebSocket(
		`${protocol}//${window.location.host}/api/circles/${props.id}/chat/ws`,
	);

	ws.onmessage = (event) => {
		console.log("WebSocket message received:", event.data);
		const msg = JSON.parse(event.data);

		if (msg.type === "presence") {
			onlineUserIds.value = new Set(msg.online_ids);
		} else if (msg.type === "join") {
			onlineUserIds.value.add(msg.user_id);
			onlineUserIds.value = new Set(onlineUserIds.value);
		} else if (msg.type === "leave") {
			onlineUserIds.value.delete(msg.user_id);
			onlineUserIds.value = new Set(onlineUserIds.value);
		} else {
			circleStore.addChatMessage(msg);
			scrollToBottom();
		}
	};

	ws.onopen = () => {
		console.log("WebSocket connected");
	};

	ws.onerror = (error) => {
		console.error("WebSocket error:", error);
	};

	ws.onclose = () => {
		console.log("WebSocket disconnected");
	};
};

const sendChatMessage = () => {
	if (!chatInput.value.trim() || !ws) return;
	ws.send(JSON.stringify({ content: chatInput.value }));
	chatInput.value = "";
};

const handleCreatePost = async () => {
	if (!newPost.value.content.trim()) return;
	await circleStore.createPost(props.id, {
		title: newPost.value.title,
		content: newPost.value.content,
		tags: [],
	});
	newPost.value = { title: "", content: "" };
};

const saveSettings = async () => {
	try {
		await circleStore.updateCircle(props.id, settings.value);
		alert("Settings saved!");
	} catch (err) {
		alert(err.response?.data || "Failed to save settings");
	}
};

const generateInvite = async () => {
	try {
		await circleStore.createInvite(props.id, inviteForm.value);
		alert("Invite created!");
		inviteForm.value = {
			code: "",
			role_to_grant: "standard",
			max_uses: null,
			expires_in_hrs: null,
		};
	} catch (err) {
		alert(err.response?.data || "Failed to create invite");
	}
};

const kickMember = async (userId) => {
	if (!confirm("Are you sure you want to kick this member?")) return;
	try {
		await circleStore.deleteMember(props.id, userId);
		await loadCircleData();
	} catch (err) {
		alert(err.response?.data || "Failed to kick member");
	}
};

const scrollToBottom = () => {
	nextTick(() => {
		if (chatBox.value) {
			chatBox.value.scrollTop = chatBox.value.scrollHeight;
		}
	});
};

watch(() => props.id, loadCircleData);
watch(tab, (newTab) => {
	if (newTab === "chat") scrollToBottom();
});

onMounted(loadCircleData);
onUnmounted(() => {
	if (ws) ws.close();
});
</script>
