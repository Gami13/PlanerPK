

const API_BASE_URL = 'http://localhost:8080/api';

export const Endpoints = {
	GetAuthURL: `${API_BASE_URL}/authURL`,
	Authorize: `${API_BASE_URL}/authorize`,
	GetUserInfo: `${API_BASE_URL}/userInfo`,
	SetChosenColumn: `${API_BASE_URL}/chosenColumn`,
};