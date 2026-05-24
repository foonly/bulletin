import { marked } from "marked";
import DOMPurify from "dompurify";

// Configure marked to be as safe as possible
marked.setOptions({
	gfm: true,
	breaks: true,
});

/**
 * Renders markdown to sanitized HTML.
 * @param {string} text - The markdown text to render.
 * @returns {string} - The sanitized HTML.
 */
export function renderMarkdown(text) {
	if (!text) return "";

	const rawHtml = marked.parse(text);

	// Sanitize the HTML to prevent XSS
	return DOMPurify.sanitize(rawHtml, {
		ALLOWED_TAGS: [
			"h1",
			"h2",
			"h3",
			"h4",
			"h5",
			"h6",
			"blockquote",
			"p",
			"a",
			"ul",
			"ol",
			"nl",
			"li",
			"ins",
			"del",
			"strong",
			"em",
			"strike",
			"code",
			"hr",
			"br",
			"div",
			"table",
			"thead",
			"caption",
			"tbody",
			"tr",
			"th",
			"td",
			"pre",
			"img",
			"span",
		],
		ALLOWED_ATTR: ["href", "name", "target", "src", "alt", "title", "class"],
		// Ensure links and images don't use dangerous protocols
		ALLOWED_URI_REGEXP:
			/^(?:(?:(?:f|ht)tps?|mailto|tel|callto|cid|xmpp):|[^a-z]|[a-z+.-]+(?:[^a-z+.-:]|$))/i,
	});
}

/**
 * Strips markdown characters to get plain text.
 * @param {string} text - The markdown text.
 * @returns {string} - Plain text.
 */
export function stripMarkdown(text) {
	if (!text) return "";
	return (
		text
			// Remove headers
			.replace(/^#+\s+/gm, "")
			// Remove bold/italic
			.replace(/([*_]{1,3})(\S.*?\S?)\1/g, "$2")
			// Remove links [text](url)
			.replace(/\[(.*?)\]\(.*?\)/g, "$1")
			// Remove images ![alt](url)
			.replace(/!\[(.*?)\]\(.*?\)/g, "$1")
			// Remove code blocks
			.replace(/```[\s\S]*?```/g, "")
			// Remove inline code
			.replace(/`(.+?)`/g, "$1")
			// Remove blockquotes
			.replace(/^\s*>\s+/gm, "")
			// Remove extra newlines
			.replace(/\n+/g, " ")
			.trim()
	);
}
