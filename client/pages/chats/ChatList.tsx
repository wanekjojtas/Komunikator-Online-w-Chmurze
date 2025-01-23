import React, { useEffect, useState, useContext } from "react";
import { API_URL } from "@/constants/constants";
import { AuthContext } from "@/modules/auth_provider";

type ChatListProps = {
    onSelect: (chatID: string) => void;
};

type Chat = {
    id: string;
    name: string;
};

const ChatList: React.FC<ChatListProps> = ({ onSelect }) => {
    const [chats, setChats] = useState<Chat[]>([]);
    const { user } = useContext(AuthContext);

    useEffect(() => {
        if (user && user.id) {
            // Fetch user-specific chats from the backend
            fetch(`${API_URL}/ws/getUserChats`, {
                method: "GET",
                headers: {
                    Authorization: `Bearer ${localStorage.getItem("jwt")}`,
                },
            })
                .then((res) => res.json())
                .then((data) => setChats(data))
                .catch((err) => console.error("Failed to fetch chats:", err));
        }
    }, [user]);

    return (
        <div>
            {chats.map((chat) => (
                <div
                    key={chat.id}
                    onClick={() => onSelect(chat.id)}
                    style={{
                        padding: "10px",
                        border: "1px solid gray",
                        borderRadius: "5px",
                        marginBottom: "5px",
                        cursor: "pointer",
                    }}
                >
                    {chat.name || "Unnamed Chat"}
                </div>
            ))}
        </div>
    );
};

export default ChatList;
