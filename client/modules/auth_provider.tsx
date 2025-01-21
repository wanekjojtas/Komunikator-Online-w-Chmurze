import { useState, createContext, useEffect, useMemo } from "react";
import { useRouter } from "next/router";
import { API_URL } from "@/constants";

export type UserInfo = {
    username: string;
    id: string;
};

export const AuthContext = createContext<{
    authenticated: boolean;
    setAuthenticated: (auth: boolean) => void;
    user: UserInfo;
    setUser: (user: UserInfo) => void;
    logout: () => void;
}>({
    authenticated: false,
    setAuthenticated: () => {},
    user: { username: "", id: "" },
    setUser: () => {},
    logout: () => {},
});

const AuthContextProvider = ({ children }: { children: React.ReactNode }) => {
    const [authenticated, setAuthenticated] = useState(false);
    const [user, setUser] = useState<UserInfo>({ username: "", id: "" });
    const router = useRouter();

    // Refresh access token
    const refreshToken = async (retryCount = 0): Promise<string | null> => {
        if (retryCount > 3) {
            console.error("Max retry limit reached for token refresh.");
            handleTokenError("Token refresh retry limit reached.");
            return null;
        }

        try {
            const storedRefreshToken = localStorage.getItem("refreshToken");
            if (!storedRefreshToken) {
                console.error("No refresh token found.");
                return null;
            }

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
                console.warn("Rate limit exceeded. Retrying...");
                await new Promise((resolve) => setTimeout(resolve, 60000));
                return refreshToken(retryCount + 1);
            } else {
                console.error("Failed to refresh token with status:", res.status);
                return null;
            }
        } catch (error) {
            console.error("Error refreshing token:", error);
            return null;
        }
    };

    // Validate JWT token
    const validateToken = async () => {
        const token = localStorage.getItem("jwt");
        if (!token) {
            console.warn("No access token found.");
            return;
        }

        try {
            const res = await fetch(`${API_URL}/auth/validate-token`, {
                method: "GET",
                headers: {
                    Authorization: `Bearer ${token}`,
                },
            });

            if (res.ok) {
                console.info("Token validation successful.");
                return;
            }

            if (res.status === 401) {
                console.warn("Token expired. Attempting to refresh...");
                const newToken = await refreshToken();
                if (!newToken) {
                    handleTokenError("Token refresh failed.");
                }
            } else {
                console.error(`Unexpected validation error: ${res.statusText}`);
                handleTokenError("Unexpected validation error.");
            }
        } catch (error) {
            console.error("Error validating token:", error);
            handleTokenError("Network or server error during token validation.");
        }
    };

    // Handle token errors by resetting state and redirecting to login
    const handleTokenError = (message: string) => {
        console.warn(message, "Redirecting to login...");
        alert("Your session has expired. Please log in again.");
        localStorage.clear();
        setAuthenticated(false);
        setUser({ username: "", id: "" });
        router.push("/login");
    };

    // Logout function for user-initiated logouts
    const logout = () => {
        console.info("Logging out...");
        handleTokenError("User logged out.");
    };

    // Initialize user info and validate token
    useEffect(() => {
        const initializeAuth = async () => {
            try {
                const userInfo = localStorage.getItem("user_info");
                const storedRefreshToken = localStorage.getItem("refreshToken");

                if (!userInfo || !storedRefreshToken) {
                    console.warn("User info or refresh token missing.");
                    handleTokenError("User info or refresh token not found.");
                    return;
                }

                const parsedUser = JSON.parse(userInfo);
                if (parsedUser?.id && parsedUser?.username) {
                    setUser(parsedUser);
                    setAuthenticated(true);
                    await validateToken();
                } else {
                    throw new Error("Invalid user info format.");
                }
            } catch (error) {
                console.error("Error initializing AuthContextProvider:", error);
                handleTokenError("Initialization failed.");
            }
        };

        initializeAuth();
    }, []);

    // Periodic token validation
    useEffect(() => {
        const interval = setInterval(() => validateToken(), 15 * 60 * 1000); // Validate every 15 minutes
        return () => clearInterval(interval);
    }, []);

    return (
        <AuthContext.Provider
            value={{
                authenticated,
                setAuthenticated,
                user,
                setUser,
                logout,
            }}
        >
            {children}
        </AuthContext.Provider>
    );
};

export default AuthContextProvider;
