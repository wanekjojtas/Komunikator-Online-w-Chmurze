import { NextRouter } from "next/router";

const fetchWithAuth = async (
    url: string,
    options: RequestInit = {},
    router: NextRouter
): Promise<Response | null> => {
    const token = localStorage.getItem("jwt");

    if (!token) {
        console.warn("No token found. Redirecting to login.");
        localStorage.removeItem("user_info");
        router.push("/login");
        return null;
    }

    const headers = {
        ...options.headers,
        Authorization: `Bearer ${token}`,
    };

    const response = await fetch(url, { ...options, headers });

    if (response.status === 401) {
        console.error("Session expired or unauthorized. Redirecting to login...");
        localStorage.removeItem("jwt");
        localStorage.removeItem("user_info");
        router.push("/login");
        return null;
    }

    return response;
};

export default fetchWithAuth;
