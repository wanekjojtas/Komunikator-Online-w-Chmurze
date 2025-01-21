import React, { useState, useEffect, createContext, useContext, useCallback } from "react";
import { AuthContext } from "@/modules/auth_provider";
import { useRouter } from "next/router";

type Conn = WebSocket | null;

export const WebSocketContext = createContext<{
    conn: Conn;
    setConn: (c: Conn) => void;
    currentChat: string | null;
    setCurrentChat: (chatID: string | null) => void;
    user: { id: string; username: string } | null;
}>({
    conn: null,
    setConn: () => {},
    currentChat: null,
    setCurrentChat: () => {},
    user: null,
});

const WebSocketProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
    const [conn, setConn] = useState<Conn>(null);
    const [currentChat, setCurrentChat] = useState<string | null>(null);
    const { user } = useContext(AuthContext);
    const router = useRouter();

    const redirectToLogin = useCallback(() => {
        console.warn("Redirecting to login page due to WebSocket authentication error.");
        localStorage.removeItem("jwt");
        localStorage.removeItem("user_info");
        router.push("/login");
    }, [router]);

    // Establish WebSocket connection
    const connectWebSocket = useCallback(() => {
        if (!currentChat || !user) return;

        const ws = new WebSocket(
            `${process.env.NEXT_PUBLIC_API_URL}/ws/joinChat/${currentChat}?userID=${user.id}&username=${user.username}&token=${localStorage.getItem("jwt")}`
        );

        ws.onopen = () => {
            console.log(`WebSocket connected to chat ${currentChat}`);
        };

        ws.onmessage = (event) => {
            console.log("Message received:", event.data);
        };

        ws.onclose = (event) => {
            console.log("WebSocket connection closed:", event.code, event.reason);

            // Redirect to login on invalid or expired token
            if (event.code === 4001 || event.reason.toLowerCase().includes("invalid") || event.reason.toLowerCase().includes("expired")) {
                redirectToLogin();
            } else {
                setConn(null);
            }
        };

        ws.onerror = (error) => {
            console.error("WebSocket error:", error);
            redirectToLogin();
        };

        setConn(ws);

        // Clean up on unmount or chat switch
        return () => {
            ws.close();
            setConn(null);
        };
    }, [currentChat, user, redirectToLogin]);

    // Automatically reconnect when currentChat changes
    useEffect(() => {
        const cleanup = connectWebSocket();
        return cleanup; // Cleanup function
    }, [connectWebSocket]);

    return (
        <WebSocketContext.Provider
            value={{
                conn,
                setConn,
                currentChat,
                setCurrentChat,
                user,
            }}
        >
            {children}
        </WebSocketContext.Provider>
    );
};

export default WebSocketProvider;
