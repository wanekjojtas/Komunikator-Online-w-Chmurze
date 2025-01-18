import React, { useState, useEffect, useContext, useRef } from "react";
import { WebSocketContext } from "@/modules/websocket_provider";
import { AuthContext } from "@/modules/auth_provider";
import { API_URL } from "@/constants";

type Message = {
    content: string;
    username: string;
    chatID: string;
    type: "recv" | "self";
};

type ChatBodyProps = {
    chatID: string;
};

const ChatBody: React.FC<ChatBodyProps> = ({ chatID }) => {
    const { conn, setConn } = useContext(WebSocketContext);
    const { user } = useContext(AuthContext);
    const [messages, setMessages] = useState<Message[]>([]);
    const textareaRef = useRef<HTMLTextAreaElement>(null);

    useEffect(() => {
        if (user && user.id) {
            const ws = new WebSocket(
                `${API_URL}/ws/joinChat/${chatID}?username=${user.username}&userID=${user.id}`
            );
    
            setConn(ws);
    
            ws.onmessage = (event) => {
                const newMessage: Message = JSON.parse(event.data);
                setMessages((prev) => [...prev, newMessage]); // Dynamically append new messages
            };
    
            ws.onclose = () => console.log("WebSocket closed");
            ws.onerror = (error) => console.error("WebSocket error:", error);
    
            return () => ws.close();
        }
    }, [chatID, user, setConn]);

    const sendMessage = () => {
        if (textareaRef.current?.value && conn) {
            const message = {
                content: textareaRef.current.value,
                chatID,
                username: user?.username,
            };

            conn.send(JSON.stringify(message));
            textareaRef.current.value = ""; // Clear textarea
            setMessages((prev) => [...prev, { ...message, type: "self" }]);
        }
    };

    return (
        <div>
            <h2>Chat ID: {chatID}</h2>
            <div>
                {messages.map((msg, index) => (
                    <div key={index}>
                        <strong>{msg.username}:</strong> {msg.content}
                    </div>
                ))}
            </div>
            <textarea ref={textareaRef} placeholder="Type your message" />
            <button onClick={sendMessage}>Send</button>
        </div>
    );
};

export default ChatBody;