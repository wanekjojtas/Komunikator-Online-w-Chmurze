import { NextRouter } from "next/router";
import { API_URL } from "@/constants";

const fetchWithAuth = async (
    url: string,
    options: RequestInit = {},
    router: NextRouter
): Promise<Response | null> => {
    let token = localStorage.getItem("jwt");

    if (!token) {
        console.warn("No access token found. Redirecting to login.");
        clearUserSession(router);
        return null;
    }

    const headers = {
        ...options.headers,
        Authorization: `Bearer ${token}`,
    };

    // First attempt to fetch with the current access token
    let response = await fetch(url, { ...options, headers });

    if (response.status === 401) {
        console.warn("Access token expired. Attempting to refresh...");

        const refreshedToken = await refreshToken(router);

        if (!refreshedToken) {
            console.error("Token refresh failed. Redirecting to login...");
            clearUserSession(router);
            return null;
        }

        // Retry the original request with the refreshed token
        headers.Authorization = `Bearer ${refreshedToken}`;
        response = await fetch(url, { ...options, headers });
    }

    return response;
};

// Refresh the access token
const refreshToken = async (router: NextRouter): Promise<string | null> => {
    const storedRefreshToken = localStorage.getItem("refreshToken");

    if (!storedRefreshToken) {
        console.error("No refresh token found. Redirecting to login.");
        clearUserSession(router);
        return null;
    }

    try {
        const res = await fetch(`${API_URL}/auth/refresh-token`, {
            method: "POST",
            headers: {
                Authorization: `Bearer ${storedRefreshToken}`,
            },
        });

        if (res.ok) {
            const data = await res.json();
            localStorage.setItem("jwt", data.accessToken); // Store the new access token
            return data.accessToken;
        } else if (res.status === 429) {
            console.warn("Rate limit exceeded. Retrying after 1 minute...");
            await new Promise((resolve) => setTimeout(resolve, 60000)); // Wait 1 minute
            return refreshToken(router); // Retry after waiting
        } else if (res.status === 401) {
            console.error("Refresh token expired or invalid. Redirecting to login.");
            clearUserSession(router);
        } else {
            console.error("Failed to refresh token with status:", res.status);
        }
    } catch (error) {
        console.error("Error refreshing token:", error);
    }

    return null;
};

// Clear user session and redirect to login
const clearUserSession = (router: NextRouter) => {
    localStorage.removeItem("jwt");
    localStorage.removeItem("refreshToken");
    localStorage.removeItem("user_info");
    router.push("/login");
};

export default fetchWithAuth;
