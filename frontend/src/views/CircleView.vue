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
			<!-- Error State -->
			<div
				v-if="error"
				class="flex-1 flex flex-col items-center justify-center p-8 text-center space-y-4"
			>
				<div class="text-6xl text-gray-700">🚫</div>
				<h2 class="text-2xl font-bold">{{ error.title }}</h2>
				<p class="text-gray-500 max-w-md">{{ error.message }}</p>
				<router-link
					to="/"
					class="bg-purple-600 px-6 py-2 rounded-lg font-bold hover:bg-purple-700 transition"
				>
					Return to Dashboard
				</router-link>
			</div>

			<!-- Normal UI -->
			<template v-else>
				<!-- Tag Sidebar -->
				<div
					v-if="tab === 'posts' && !activeThread"
					class="w-48 bg-gray-950/50 border-r border-gray-800 p-4 flex flex-col space-y-2 overflow-y-auto"
				>
					<h3
						class="text-xs font-bold text-gray-500 uppercase tracking-wider mb-2"
					>
						Tags
					</h3>
					<button
						@click="filterByTag('')"
						:class="[
							'text-left px-2 py-1 rounded text-sm transition-colors',
							!selectedTag
								? 'bg-purple-600 text-white'
								: 'text-gray-400 hover:bg-gray-800',
						]"
					>
						All Tags
					</button>
					<div
						v-for="tag in circleStore.tags"
						:key="tag.id"
						class="group flex items-center justify-between"
					>
						<button
							@click="filterByTag(tag.name)"
							:class="[
								'flex-1 text-left px-2 py-1 rounded text-sm transition-colors truncate',
								selectedTag === tag.name
									? 'bg-purple-600 text-white'
									: 'text-gray-400 hover:bg-gray-800',
							]"
						>
							<span v-if="tag.is_pinned" class="mr-1">📌</span>{{ tag.name }}
						</button>
						<button
							v-if="isAdmin && !tag.is_pinned"
							@click="togglePin(tag.id, true)"
							class="hidden group-hover:block text-[10px] text-gray-600 hover:text-purple-400 p-1"
						>
							Pin
						</button>
						<button
							v-if="isAdmin && tag.is_pinned"
							@click="togglePin(tag.id, false)"
							class="text-[10px] text-purple-400 p-1"
						>
							Unpin
						</button>
					</div>
				</div>

				<div class="flex-1 overflow-y-auto p-4">
					<div v-if="tab === 'posts'" class="space-y-4">
						<!-- Thread List -->
						<div v-if="!activeThread" class="space-y-4">
							<div class="bg-gray-800 p-4 rounded-lg border border-gray-700">
								<h3 class="text-lg font-bold mb-4">Start a new thread</h3>
								<input
									v-model="newPost.title"
									placeholder="Thread Title"
									class="w-full bg-gray-700 p-2 rounded mb-2 border border-gray-600 focus:outline-none"
								/>
								<textarea
									v-model="newPost.content"
									placeholder="What's on your mind?"
									class="w-full bg-gray-700 p-2 rounded border border-gray-600 focus:outline-none"
									rows="3"
								></textarea>

								<!-- Tag Selector -->
								<div class="mt-3 space-y-2">
									<label class="text-xs text-gray-500 font-bold uppercase"
										>Tags (at least one required)</label
									>
									<div class="flex flex-wrap gap-2">
										<button
											v-for="tag in circleStore.tags"
											:key="tag.id"
											@click="toggleTagSelection(tag.name)"
											:class="[
												'px-2 py-1 rounded text-xs border transition-colors',
												newPost.tags.includes(tag.name)
													? 'bg-purple-600 border-purple-500 text-white'
													: 'bg-gray-900 border-gray-700 text-gray-400 hover:border-gray-500',
											]"
										>
											{{ tag.name }}
										</button>
										<div
											v-if="circleStore.activeCircle?.allow_freeform_tags"
											class="flex items-center space-x-1"
										>
											<input
												v-model="newTagName"
												@keyup.enter="addCustomTag"
												placeholder="Add tag..."
												class="bg-gray-900 border border-gray-700 text-xs px-2 py-1 rounded focus:outline-none focus:border-purple-500 w-24"
											/>
											<button
												@click="addCustomTag"
												class="text-xs text-purple-400 hover:text-purple-300"
											>
												+
											</button>
										</div>
									</div>
									<div
										v-if="newPost.tags.length > 0"
										class="flex flex-wrap gap-1 mt-1"
									>
										<span
											v-for="tag in newPost.tags"
											:key="tag"
											class="bg-purple-900/30 text-purple-400 text-[10px] px-1.5 py-0.5 rounded flex items-center"
										>
											{{ tag }}
											<button
												@click="toggleTagSelection(tag)"
												class="ml-1 text-purple-600 hover:text-purple-400"
											>
												×
											</button>
										</span>
									</div>
								</div>

								<div class="flex justify-end mt-4">
									<button
										@click="handleCreatePost"
										:disabled="
											!newPost.content.trim() ||
											!newPost.title.trim() ||
											newPost.tags.length === 0
										"
										class="bg-purple-600 px-4 py-1 rounded font-bold hover:bg-purple-700 disabled:opacity-50 disabled:cursor-not-allowed"
									>
										Post
									</button>
								</div>
							</div>

							<div
								v-for="thread in circleStore.threads"
								:key="thread.id"
								@click="openThread(thread.id)"
								class="bg-gray-800 p-4 rounded-lg border border-gray-700 cursor-pointer hover:border-purple-500 transition-colors"
							>
								<div class="flex items-center justify-between mb-2 text-xs">
									<div class="flex items-center space-x-2">
										<span class="font-bold text-purple-400">{{
											thread.author_name
										}}</span>
										<span class="text-gray-500">{{
											formatDate(thread.created_at)
										}}</span>
									</div>
									<div
										v-if="thread.unread_count > 0"
										class="bg-red-500 text-white px-2 py-0.5 rounded-full font-bold"
									>
										{{ thread.unread_count }} new
									</div>
								</div>
								<h3 class="font-bold text-lg mb-1">{{ thread.title }}</h3>
								<div class="flex flex-wrap gap-1 mb-2">
									<span
										v-for="tag in thread.tags"
										:key="tag"
										@click.stop="filterByTag(tag)"
										class="text-[10px] bg-gray-700 text-gray-300 px-1.5 py-0.5 rounded hover:bg-purple-600 hover:text-white transition-colors cursor-pointer"
									>
										#{{ tag }}
									</span>
								</div>
								<p
									class="text-gray-400 text-sm line-clamp-4 overflow-hidden"
									style="
										display: -webkit-box;
										-webkit-line-clamp: 4;
										-webkit-box-orient: vertical;
									"
								>
									{{ thread.content }}
								</p>

								<div
									class="mt-4 flex items-center justify-between text-xs text-gray-500 border-t border-gray-700 pt-2"
								>
									<div class="flex items-center space-x-4">
										<span>{{ thread.reply_count }} replies</span>
										<span v-if="thread.last_reply_at"
											>Last reply: {{ formatDate(thread.last_reply_at) }}</span
										>
									</div>
								</div>
							</div>
						</div>

						<!-- Single Thread View -->
						<div v-else class="space-y-4">
							<button
								@click="activeThread = null"
								class="text-purple-400 hover:underline mb-2 flex items-center"
							>
								← Back to all threads
							</button>

							<div class="space-y-4 pb-12">
								<ThreadNode
									v-if="threadTree"
									:node="threadTree"
									:circle-id="props.id"
									@reply-created="openThread(threadTree.id)"
								/>
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
								<span class="font-bold text-purple-400"
									>{{ msg.username }}:</span
								>
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

							<!-- Active Invites List -->
							<div v-if="circleStore.invites.length > 0" class="mt-6 space-y-3">
								<h4 class="text-sm font-bold text-gray-400">Active Invites</h4>
								<div
									v-for="invite in circleStore.invites"
									:key="invite.id"
									class="bg-gray-900/50 p-3 rounded border border-gray-700 flex items-center justify-between"
								>
									<div class="space-y-1">
										<div class="flex items-center space-x-2">
											<code
												class="bg-purple-900/30 text-purple-400 px-2 py-0.5 rounded font-mono font-bold"
												>{{ invite.code }}</code
											>
											<span class="text-xs text-gray-500"
												>grants {{ invite.role_to_grant }} • Issued by
												{{ invite.created_by }}</span
											>
										</div>
										<div class="text-[10px] text-gray-500">
											Uses: {{ invite.used_count }} /
											{{ invite.max_uses || "∞" }} • Expires:
											{{
												invite.expires_at
													? formatDate(invite.expires_at)
													: "Never"
											}}
										</div>
									</div>
									<button
										@click="revokeInvite(invite.id)"
										class="text-red-500 hover:text-red-400 text-xs"
									>
										Revoke
									</button>
								</div>
							</div>
						</section>

						<!-- Tag Management -->
						<section class="bg-gray-800 p-6 rounded-lg space-y-4">
							<h3 class="text-xl font-bold border-b border-gray-700 pb-2">
								Tags
							</h3>
							<div class="flex items-center space-x-2">
								<input
									v-model="adminNewTagName"
									@keyup.enter="adminAddTag"
									placeholder="New tag name..."
									class="flex-1 bg-gray-700 p-2 rounded focus:outline-none"
								/>
								<button
									@click="adminAddTag"
									class="bg-purple-600 px-4 py-2 rounded font-bold hover:bg-purple-700"
								>
									Add Tag
								</button>
							</div>
							<div class="flex flex-wrap gap-2 pt-2">
								<div
									v-for="tag in circleStore.tags"
									:key="tag.id"
									class="flex items-center bg-gray-700 rounded-full px-3 py-1 text-sm"
								>
									<span :class="tag.is_pinned ? 'text-purple-400' : ''">{{
										tag.name
									}}</span>
									<button
										@click="togglePin(tag.id, !tag.is_pinned)"
										class="ml-2 text-xs text-gray-500 hover:text-purple-400"
									>
										{{ tag.is_pinned ? "Unpin" : "Pin" }}
									</button>
								</div>
							</div>
						</section>

						<!-- Member Management (Admins only) -->
						<section
							v-if="isAdmin"
							class="bg-gray-800 p-6 rounded-lg space-y-4"
						>
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
										<select
											:value="m.role"
											@change="updateMemberRole(m.id, $event.target.value)"
											class="bg-gray-600 text-xs p-1 rounded"
										>
											<option value="guest">Guest</option>
											<option value="standard">Standard</option>
											<option value="mod">Moderator</option>
											<option value="admin">Admin</option>
										</select>
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
			</template>
		</div>
	</div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, computed, watch, nextTick } from "vue";
import { useCircleStore } from "../stores/circles";
import { useAuthStore } from "../stores/auth";
import { useToastStore } from "../stores/toast";
import axios from "axios";
import ThreadNode from "../components/ThreadNode.vue";

const props = defineProps(["id"]);
const circleStore = useCircleStore();
const auth = useAuthStore();
const toast = useToastStore();
const tab = ref("posts");
const chatInput = ref("");
const chatBox = ref(null);
const newPost = ref({ title: "", content: "", tags: [] });
const newTagName = ref("");
const adminNewTagName = ref("");
const selectedTag = ref("");
const replyContent = ref("");
const members = ref([]);
const onlineUserIds = ref(new Set());
const activeThread = ref(null);
const error = ref(null);
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

const threadTree = computed(() => {
	if (!activeThread.value || !activeThread.value.length) return null;

	const posts = activeThread.value;
	const root = posts.find((p) => !p.parent_id);
	if (!root) return null;

	const buildNode = (node) => {
		return {
			...node,
			replies: posts
				.filter((p) => p.parent_id === node.id)
				.map((p) => buildNode(p)),
		};
	};

	return buildNode(root);
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
	activeThread.value = null; // Reset when switching circles
	selectedTag.value = ""; // Reset tag filter
	error.value = null; // Reset error

	try {
		if (circleStore.circles.length === 0) {
			await circleStore.fetchCircles();
		}
		const circle = circleStore.circles.find((c) => c.id === props.id);

		if (!circle) {
			// Trigger a 404-style error if not found in store after fetch
			error.value = {
				title: "Circle Not Found",
				message:
					"The circle you are looking for doesn't exist or you don't have permission to view it.",
			};
			return;
		}

		circleStore.activeCircle = circle;

		await circleStore.fetchThreads(props.id);
		await circleStore.fetchTags(props.id);
		await circleStore.fetchChatHistory(props.id);
		if (canManage.value) {
			await circleStore.fetchInvites(props.id);
		}
		const res = await axios.get(`/api/circles/${props.id}/members`);
		members.value = res.data;

		settings.value = { ...circleStore.activeCircle };
		connectWS();
	} catch (err) {
		console.error("Failed to load circle data:", err);
		if (err.response?.status === 404 || err.response?.status === 403) {
			error.value = {
				title: "Access Denied",
				message: "You are not a member of this circle or it has been deleted.",
			};
		} else {
			error.value = {
				title: "Connection Error",
				message:
					"Something went wrong while loading the circle. Please try again later.",
			};
		}
	}
};

const openThread = async (postId) => {
	await circleStore.fetchThread(props.id, postId);
	activeThread.value = circleStore.activeThread;

	// Start read marker logic
	startReadTracking(postId);
};

let readTimer = null;
const startReadTracking = (entityId) => {
	if (readTimer) clearTimeout(readTimer);

	// Wait 3 seconds then mark as read
	readTimer = setTimeout(async () => {
		try {
			await circleStore.markRead(props.id, entityId);
			// Refresh threads to update unread counts
			await circleStore.fetchThreads(props.id);
		} catch (err) {
			console.error("Failed to mark as read", err);
		}
	}, 3000);
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
		tags: newPost.value.tags,
	});
	newPost.value = { title: "", content: "", tags: [] };
	await circleStore.fetchThreads(props.id, selectedTag.value);
	await circleStore.fetchTags(props.id);
};

const toggleTagSelection = (tagName) => {
	const index = newPost.value.tags.indexOf(tagName);
	if (index > -1) {
		newPost.value.tags.splice(index, 1);
	} else {
		newPost.value.tags.push(tagName);
	}
};

const addCustomTag = () => {
	const tag = newTagName.value.trim().toLowerCase();
	if (tag && !newPost.value.tags.includes(tag)) {
		newPost.value.tags.push(tag);
	}
	newTagName.value = "";
};

const filterByTag = async (tagName) => {
	selectedTag.value = tagName;
	await circleStore.fetchThreads(props.id, tagName);
};

const adminAddTag = async () => {
	const tag = adminNewTagName.value.trim().toLowerCase();
	if (!tag) return;

	try {
		await circleStore.createTag(props.id, tag);
		toast.success(`Tag #${tag} created!`);
	} catch (err) {
		toast.error(err.response?.data || "Failed to create tag");
	}
	adminNewTagName.value = "";
};

const togglePin = async (tagId, isPinned) => {
	await circleStore.pinTag(props.id, tagId, isPinned);
};

const saveSettings = async () => {
	try {
		await circleStore.updateCircle(props.id, settings.value);
		toast.success("Settings saved!");
	} catch (err) {
		toast.error(err.response?.data || "Failed to save settings");
	}
};

const generateInvite = async () => {
	try {
		const res = await circleStore.createInvite(props.id, inviteForm.value);
		await circleStore.fetchInvites(props.id);
		toast.success(`Invite created: ${res.code}`, 30000);
		inviteForm.value = {
			role_to_grant: "standard",
			max_uses: null,
			expires_in_hrs: null,
		};
	} catch (err) {
		toast.error(err.response?.data || "Failed to create invite");
	}
};

const revokeInvite = async (inviteId) => {
	if (!confirm("Are you sure you want to revoke this invite code?")) return;
	try {
		await circleStore.deleteInvite(props.id, inviteId);
		toast.success("Invite revoked");
	} catch (err) {
		toast.error("Failed to revoke invite");
	}
};

const updateMemberRole = async (userId, role) => {
	try {
		await circleStore.updateMember(props.id, userId, role);
		toast.success("Member role updated");
		await loadCircleData();
	} catch (err) {
		toast.error(err.response?.data || "Failed to update role");
	}
};

const kickMember = async (userId) => {
	if (!confirm("Are you sure you want to kick this member?")) return;
	try {
		await circleStore.deleteMember(props.id, userId);
		toast.success("Member kicked");
		await loadCircleData();
	} catch (err) {
		toast.error(err.response?.data || "Failed to kick member");
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
	if (newTab === "chat") {
		scrollToBottom();
		circleStore.markRead(props.id, props.id); // Mark circle chat as read
	}
});

onMounted(loadCircleData);
onUnmounted(() => {
	if (ws) ws.close();
});
</script>
