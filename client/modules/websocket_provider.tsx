import React, { useState, useEffect, createContext, useContext, useCallback } from "react";
import { AuthContext } from "@/modules/auth_provider";

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

    // Establish WebSocket connection
    const connectWebSocket = useCallback(() => {
        if (!currentChat || !user) return;

        const ws = new WebSocket(
            `${process.env.NEXT_PUBLIC_API_URL}/ws/joinChat/${currentChat}?userID=${user.id}&username=${user.username}`
        );

        ws.onopen = () => console.log(`WebSocket connected to chat ${currentChat}`);
        ws.onclose = () => {
            console.log("WebSocket connection closed");
            setConn(null);
        };
        ws.onerror = (error) => {
            console.error("WebSocket error:", error);
            setConn(null);
        };

        setConn(ws);

        // Clean up on unmount or chat switch
        return () => {
            ws.close();
            setConn(null);
        };
    }, [currentChat, user]);

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
