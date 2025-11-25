

const API_BASE_URL = window.__RUNTIME_CONFIG__.API_BASE_URL;

export const Endpoints = {
	GetAuthURL: `${API_BASE_URL}/authURL`,
	Authorize: `${API_BASE_URL}/authorize`,
	GetUserInfo: `${API_BASE_URL}/userInfo`,
	SetChosenColumn: `${API_BASE_URL}/chosenColumn`,
};


export type UserInfoResponse = {
	email: string;
	name: string;
	given_name: string;
	picture: string;
	is_active: boolean;
	chosen_column: number | null;
}


export const PK_WEBSITE_URL = "https://www.pk.edu.pl/index.php?lang=pl"
export const GOOGLE_CALENDAR_URL = "https://calendar.google.com/calendar?hl=pl"