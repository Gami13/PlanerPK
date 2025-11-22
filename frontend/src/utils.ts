export function setAccessToken(token: string) {
	document.cookie = `access_token=${token}; path=/; max-age=604800; samesite=strict`;
}
export function getAccessToken(): string | null {
	const match = document.cookie.match(new RegExp('(^| )access_token=([^;]+)'));
	return match ? match[2] : null;
}