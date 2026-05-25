import { defineStore } from "pinia";
import axios from "axios";
import { useAuthStore } from "./auth";
import { showBrowserNotification } from "../utils/notifications";

export const useCircleStore = defineStore("circles", {
	state: () => ({
		circles: [],
		activeCircle: null,
		threads: [],
		activeThread: null,
		posts: [],
		tags: [],
		invites: [],
		chatMessages: [],
	}),
	actions: {
		async fetchCircles() {
			const res = await axios.get("/api/circles");
			this.circles = res.data;
		},
		async fetchInvites(circleId) {
			const res = await axios.get(`/api/circles/${circleId}/invites`);
			this.invites = res.data;
		},
		async deleteInvite(circleId, inviteId) {
			await axios.delete(`/api/circles/${circleId}/invites/${inviteId}`);
			await this.fetchInvites(circleId);
		},
		async createCircle(circleData) {
			const res = await axios.post("/api/circles", circleData);
			await this.fetchCircles();
			return res.data;
		},
		async fetchTags(circleId) {
			const res = await axios.get(`/api/circles/${circleId}/tags`);
			this.tags = res.data;
		},
		async createTag(circleId, name) {
			await axios.post(`/api/circles/${circleId}/tags`, { name });
			await this.fetchTags(circleId);
		},
		async pinTag(circleId, tagId, isPinned) {
			await axios.post(`/api/circles/${circleId}/tags/${tagId}/pin`, {
				is_pinned: isPinned,
			});
			await this.fetchTags(circleId);
		},
		async fetchThreads(circleId, tag = "") {
			let url = `/api/circles/${circleId}/threads`;
			if (tag) {
				url += `?tag=${encodeURIComponent(tag)}`;
			}
			const res = await axios.get(url);
			this.threads = res.data;
		},
		async fetchThread(circleId, postId) {
			const res = await axios.get(`/api/circles/${circleId}/threads/${postId}`);
			this.activeThread = res.data;
		},
		async updatePost(circleId, postId, content) {
			await axios.put(`/api/circles/${circleId}/threads/${postId}`, {
				content,
			});
		},
		async deletePost(circleId, postId) {
			await axios.delete(`/api/circles/${circleId}/threads/${postId}`);
		},
		async markRead(circleId, entityId) {
			await axios.post(`/api/circles/${circleId}/read/${entityId}`);
		},
		async fetchPosts(circleId) {
			const res = await axios.get(`/api/circles/${circleId}/posts`);
			this.posts = res.data;
		},
		async fetchChatHistory(circleId) {
			const res = await axios.get(`/api/circles/${circleId}/chat/history`);
			this.chatMessages = res.data;
		},
		async search(circleId, query) {
			const res = await axios.get(
				`/api/circles/${circleId}/search?q=${encodeURIComponent(query)}`,
			);
			return res.data;
		},
		async createPost(circleId, postData) {
			await axios.post(`/api/circles/${circleId}/posts`, postData);
			await this.fetchPosts(circleId);
		},
		async updateCircle(circleId, circleData) {
			await axios.put(`/api/circles/${circleId}`, circleData);
			await this.fetchCircles();
		},
		async createInvite(circleId, inviteData) {
			await axios.post(`/api/circles/${circleId}/invites`, inviteData);
		},
		async updateMember(circleId, userId, role) {
			await axios.put(`/api/circles/${circleId}/members/${userId}`, { role });
		},
		async deleteMember(circleId, userId) {
			await axios.delete(`/api/circles/${circleId}/members/${userId}`);
		},
		addChatMessage(msg) {
			this.chatMessages.push(msg);
		},
		incrementUnreadChat(circleId) {
			const circle = this.circles.find((c) => c.id === circleId);
			if (circle) {
				circle.unread_chat_count++;
				circle.unread_count++;
			}
		},
		async refreshUnreadCounts() {
			const res = await axios.get("/api/circles");
			const newCircles = res.data;
			const auth = useAuthStore();

			// Intelligently update existing objects to maintain reactivity
			newCircles.forEach((newCircle) => {
				const existing = this.circles.find((c) => c.id === newCircle.id);
				if (existing) {
					// Check for new notifications before updating
					if (
						newCircle.unread_count > existing.unread_count &&
						auth.notificationsEnabled &&
						document.visibilityState !== "visible"
					) {
						const diff = newCircle.unread_count - existing.unread_count;
						showBrowserNotification(`Activity in ${newCircle.name}`, {
							body: `You have ${diff} new notification${diff > 1 ? "s" : ""}.`,
						});
					}

					existing.unread_count = newCircle.unread_count;
					existing.unread_chat_count = newCircle.unread_chat_count;
					existing.unread_post_count = newCircle.unread_post_count;
					existing.member_count = newCircle.member_count;
					existing.last_post_title = newCircle.last_post_title;
					existing.last_post_at = newCircle.last_post_at;
					existing.last_read_at = newCircle.last_read_at;

					// Also update activeCircle if it matches
					if (this.activeCircle && this.activeCircle.id === existing.id) {
						this.activeCircle.unread_count = existing.unread_count;
						this.activeCircle.unread_chat_count = existing.unread_chat_count;
						this.activeCircle.unread_post_count = existing.unread_post_count;
						this.activeCircle.last_read_at = existing.last_read_at;
					}
				}
			});

			// If lengths differ, just replace the whole array
			if (newCircles.length !== this.circles.length) {
				this.circles = newCircles;
			}
		},
		async refreshActiveTags() {
			if (!this.activeCircle) return;
			const res = await axios.get(`/api/circles/${this.activeCircle.id}/tags`);
			const newTags = res.data;

			newTags.forEach((newTag) => {
				const existing = this.tags.find((t) => t.id === newTag.id);
				if (existing) {
					existing.unread_count = newTag.unread_count;
					existing.use_count = newTag.use_count;
					existing.is_pinned = newTag.is_pinned;
				}
			});

			if (newTags.length !== this.tags.length) {
				this.tags = newTags;
			}
		},
	},
});
