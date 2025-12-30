import { ENV } from "./env";

export const Endpoints = {
	GetAuthURL: `${ENV.apiBaseUrl}/authURL`,
	Authorize: `${ENV.apiBaseUrl}/authorize`,
	GetUserInfo: `${ENV.apiBaseUrl}/userInfo`,
	SetChosenColumn: `${ENV.apiBaseUrl}/chosenColumn`,
};

export type UserInfoResponse = {
	email: string;
	name: string;
	given_name: string;
	picture: string;
	is_active: boolean;
	chosen_column: number | null;
};

export const PK_WEBSITE_URL = "https://www.pk.edu.pl/index.php?lang=pl";
export const GOOGLE_CALENDAR_URL = "https://calendar.google.com/calendar?hl=pl";
