import { useState, useEffect, useContext } from "react";
import { useRouter } from "next/router";
import { API_URL } from "../constants";
import { AuthContext } from "../modules/auth_provider";
import { WebSocketContext } from "@/modules/websocket_provider";

const Index = () => {
    const [chats, setChats] = useState<{ id: string; name: string }[]>([]);
    const [searchQuery, setSearchQuery] = useState("");
    const [users, setUsers] = useState<{ id: string; username: string }[]>([]);
    const [selectedChat, setSelectedChat] = useState<{ id: string; name: string } | null>(null);
    const [messages, setMessages] = useState<{ id: string; username: string; content: string; createdAt: string }[]>([]);
    const [newMessage, setNewMessage] = useState("");
    const { user } = useContext(AuthContext);
    const { conn, setConn } = useContext(WebSocketContext);
    const router = useRouter();

    const redirectToLogin = () => {
        console.warn("No valid JWT token found. Redirecting to login...");
        localStorage.removeItem("jwt");
        localStorage.removeItem("user_info");
        router.push("/login");
    };

    // Fetch all users on component mount
    useEffect(() => {
        const fetchAllUsers = async () => {
            try {
                const token = localStorage.getItem("jwt");
                if (!token) {
                    redirectToLogin();
                    return;
                }

                const res = await fetch(`${API_URL}/users/all`, {
                    method: "GET",
                    headers: {
                        Authorization: `Bearer ${token}`,
                    },
                });

                if (!res.ok) {
                    if (res.status === 401) redirectToLogin();
                    else throw new Error("Failed to fetch users.");
                }

                const data = await res.json();
                const currentUser = { id: user?.id, username: user?.username };
                const uniqueUsers = [...data, currentUser].filter(
                    (value, index, self) =>
                        index === self.findIndex((u) => u.id === value.id)
                );

                setUsers(uniqueUsers || []);
            } catch (error) {
                console.error("Error fetching all users:", error);
                redirectToLogin();
            }
        };

        fetchAllUsers();
    }, [user?.id, user?.username]);

    // Fetch chats on component mount
    useEffect(() => {
        if (!user?.id) return;

        const fetchChats = async () => {
            try {
                const token = localStorage.getItem("jwt");
                if (!token) {
                    redirectToLogin();
                    return;
                }

                const res = await fetch(`${API_URL}/ws/getUserChats`, {
                    method: "GET",
                    headers: {
                        Authorization: `Bearer ${token}`,
                    },
                });

                if (!res.ok) {
                    if (res.status === 401) redirectToLogin();
                    else throw new Error("Failed to fetch chats.");
                }

                const data = await res.json();
                setChats(data || []);
            } catch (error) {
                console.error("Error fetching chats:", error);
                redirectToLogin();
            }
        };

        fetchChats();
    }, [user?.id]);

    // WebSocket connection
    useEffect(() => {
        if (!selectedChat?.id || !user) return;

        let ws: WebSocket | null = null;
        let reconnectTimeout: NodeJS.Timeout;

        const connectWebSocket = () => {
            const token = localStorage.getItem("jwt");
            if (!token) {
                redirectToLogin();
                return;
            }

            ws = new WebSocket(
                `${API_URL.replace("http", "ws")}/ws/joinChat/${selectedChat.id}?userID=${user.id}&username=${user.username}&token=${token}`
            );

            ws.onopen = () => {
                console.log(`Connected to chat: ${selectedChat.name}`);
            };

            ws.onmessage = (event) => {
                const newMessage = JSON.parse(event.data);
                setMessages((prev) => [...prev, newMessage]);
            };

            ws.onclose = (event) => {
                console.log("WebSocket connection closed:", event.code, event.reason);

                if (event.code === 4001 || event.reason.toLowerCase().includes("token")) {
                    redirectToLogin();
                } else if (event.code !== 1000) {
                    reconnectTimeout = setTimeout(connectWebSocket, 3000);
                }
            };

            ws.onerror = (error) => {
                console.error("WebSocket error:", error);
                redirectToLogin();
            };

            setConn(ws);
        };

        connectWebSocket();

        return () => {
            if (ws && ws.readyState === WebSocket.OPEN) {
                ws.close(1000, "Client disconnected");
            }
            clearTimeout(reconnectTimeout);
            setConn(null);
            setMessages([]);
        };
        }, [selectedChat?.id, user, setConn]);
    

    const handleSearch = async (query: string) => {
        setSearchQuery(query);

        if (!query.trim()) {
            setUsers([]);
            return;
        }

        try {
            const token = localStorage.getItem("jwt");
            if (!token) throw new Error("No JWT token found.");

            const res = await fetch(`${API_URL}/users/search?q=${encodeURIComponent(query)}`, {
                method: "GET",
                headers: {
                    Authorization: `Bearer ${token}`,
                },
            });

            if (!res.ok) throw new Error("Failed to search users.");
            const data = await res.json();
            setUsers(data || []);
        } catch (error) {
            console.error("Error searching users:", error);
        }
    };

    const handleSendMessage = async () => {
        if (!newMessage.trim() || !selectedChat?.id) return;

        try {
            const token = localStorage.getItem("jwt");
            if (!token) throw new Error("No JWT token found.");

            const res = await fetch(`${API_URL}/ws/sendMessage`, {
                method: "POST",
                headers: {
                    "Content-Type": "application/json",
                    Authorization: `Bearer ${token}`,
                },
                body: JSON.stringify({
                    chatID: selectedChat.id,
                    content: newMessage,
                }),
            });

            if (!res.ok) throw new Error("Failed to send message.");
            setNewMessage("");
        } catch (error) {
            console.error("Error sending message:", error);
        }
    };
    // Fetch chat details and messages when a chat is selected
    const fetchChatDetails = async (chatID: string) => {
        try {
            const token = localStorage.getItem("jwt");
            if (!token) {
                console.error("No JWT token found.");
                return;
            }
    
            const res = await fetch(`${API_URL}/ws/getChatDetails/${chatID}`, {
                method: "GET",
                headers: {
                    Authorization: `Bearer ${token}`,
                },
            });
    
            if (!res.ok) {
                console.error(`Failed to fetch chat details. Status: ${res.status}`);
                return;
            }
    
            const data = await res.json();
    
            if (!data?.id || !data?.name || !data?.members || !Array.isArray(data.members)) {
                console.error("Invalid chat data:", data);
                return;
            }
    
            // Handle self-chat
            if (data.name === "Chat with Yourself") {
                setSelectedChat({ id: chatID, name: "Chat with Yourself" });
            } else if (data.members.length === 2) {
                // One-on-one chat (backend should provide the correct name)
                setSelectedChat({ id: chatID, name: data.name });
            } else {
                // Group chat
                setSelectedChat({ id: chatID, name: data.name });
            }
    
            // Fetch chat messages after setting chat details
            await fetchChatMessages(chatID);
        } catch (error) {
            console.error("Error fetching chat details:", error);
        }
    };
    
    
    // Fetch chat messages when a chat is selected
    const fetchChatMessages = async (chatID: string) => {
        try {
            const token = localStorage.getItem("jwt");
            if (!token) throw new Error("No JWT token found.");

            const res = await fetch(`${API_URL}/ws/getChatMessages/${chatID}`, {
                method: "GET",
                headers: {
                    Authorization: `Bearer ${token}`,
                },
            });

            if (!res.ok) throw new Error("Failed to fetch chat messages.");
            const data = await res.json();
            setMessages(data || []);
        } catch (error) {
            console.error("Error fetching chat messages:", error);
        }
    };

    return (
        <div className="flex h-screen w-full">
            {/* Sidebar */}
            <div className="w-1/3 bg-gray-800 text-white flex flex-col">
                <div className="p-4 bg-gray-700">
                    <input
                        type="text"
                        className="w-full p-2 rounded bg-gray-600 placeholder-gray-400"
                        placeholder="Search for users"
                        value={searchQuery}
                        onChange={(e) => handleSearch(e.target.value)}
                    />
                </div>
                <div className="flex-1 overflow-y-auto p-4">
                    {!searchQuery.trim() ? (
                        <>
                            <p className="text-gray-400 mb-2">Started Chats:</p>
                            {chats.length > 0 ? (
                                chats.map((chat) => (
                                    <div
                                        key={chat.id}
                                        className={`p-4 rounded mb-2 cursor-pointer ${
                                            selectedChat?.id === chat.id
                                                ? "bg-blue-600 hover:bg-blue-700" // Slight hover effect for active chat
                                                : chat.name === "Chat with Yourself"
                                                ? "bg-purple-700 hover:bg-purple-800"
                                                : chat.name.includes(", ") // Group chat detection
                                                ? "bg-green-700 hover:bg-green-800"
                                                : "bg-gray-700 hover:bg-gray-600"
                                        }`}
                                        onClick={() => fetchChatDetails(chat.id)}
                                    >
                                        {chat.name || "Unnamed Chat"}
                                    </div>
                                ))
                            ) : (
                                <p className="text-gray-400">No chats started yet. Start chatting!</p>
                            )}
                        </>
                    ) : users.length > 0 ? (
                        <>
                            <p className="text-gray-400 mb-2">Search Results:</p>
                            {users.map((selectedUser) => (
                                <div
                                    key={selectedUser.id}
                                    className="p-4 bg-gray-700 rounded mb-2 cursor-pointer hover:bg-gray-600"
                                    onClick={async () => {
                                        try {
                                            const token = localStorage.getItem("jwt");
                                            if (!token) throw new Error("No JWT token found.");
    
                                            const members = [user.id, selectedUser.id].sort();
    
                                            const res = await fetch(`${API_URL}/ws/startChat`, {
                                                method: "POST",
                                                headers: {
                                                    "Content-Type": "application/json",
                                                    Authorization: `Bearer ${token}`,
                                                },
                                                body: JSON.stringify({ members }),
                                            });
    
                                            if (!res.ok) throw new Error("Failed to start chat.");
                                            const data = await res.json();
    
                                            setChats((prevChats) => {
                                                const exists = prevChats.some(
                                                    (chat) => chat.id === data.chatID
                                                );
                                                if (!exists) {
                                                    return [
                                                        ...prevChats,
                                                        { id: data.chatID, name: data.name }, // Use backend-provided name
                                                    ];
                                                }
                                                return prevChats;
                                            });
    
                                            setSelectedChat({ id: data.chatID, name: data.name }); // Set as active chat
                                            await fetchChatMessages(data.chatID);
    
                                            // Clear search bar and redirect to started chats
                                            setSearchQuery(""); // Clear the search query
                                        } catch (error) {
                                            console.error("Error starting chat:", error);
                                        }
                                    }}
                                >
                                    {selectedUser.username}
                                </div>
                            ))}
                        </>
                    ) : (
                        <p className="text-gray-400">No users found.</p>
                    )}
                </div>
            </div>
    
            {/* Main Chat Section */}
            <div className="w-2/3 bg-gray-900 text-white flex flex-col">
                {selectedChat ? (
                    <>
                        <div className="p-4 bg-gray-800 border-b border-gray-700">
                            <h2 className="text-lg font-bold">
                                {selectedChat.name === "Chat with Yourself"
                                    ? "Chat with Yourself"
                                    : `${selectedChat.name}`}
                            </h2>
                        </div>
                        <div
                            className="flex-1 overflow-y-auto p-4"
                            ref={(el) => {
                                if (el) el.scrollTop = el.scrollHeight; // Auto-scroll
                            }}
                        >
                            {messages.length > 0 ? (
                                messages.map((msg, index) => (
                                    <div key={index} className="flex mb-4">
                                        {msg.username !== user?.username && (
                                            <div className="mr-2 w-8 h-8 bg-gray-600 rounded-full flex items-center justify-center text-white font-bold uppercase">
                                                {msg.username.charAt(0)}
                                            </div>
                                        )}
                                        <div
                                            className={`p-2 rounded-lg max-w-xs ${
                                                msg.username === user?.username
                                                    ? "bg-blue-600 text-white self-end ml-auto"
                                                    : "bg-gray-700 text-white self-start"
                                            } whitespace-pre-wrap break-words`}
                                        >
                                            <p className="text-sm font-semibold">{msg.username}</p>
                                            <p className="mt-1">{msg.content}</p>
                                        </div>
                                    </div>
                                ))
                            ) : (
                                <p className="text-gray-400">No messages yet. Start the conversation!</p>
                            )}
                        </div>
                        <div className="p-4 border-t border-gray-700 flex items-center">
                            <input
                                type="text"
                                className="flex-1 p-2 rounded bg-gray-600 placeholder-gray-400 mr-2"
                                placeholder="Type a message..."
                                value={newMessage}
                                onChange={(e) => setNewMessage(e.target.value)}
                                onKeyDown={(e) => e.key === "Enter" && handleSendMessage()}
                            />
                            <button
                                onClick={handleSendMessage}
                                className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
                            >
                                Send
                            </button>
                        </div>
                    </>
                ) : (
                    <div className="flex-1 flex items-center justify-center">
                        <p className="text-gray-400">Select a chat to start chatting.</p>
                    </div>
                )}
            </div>
        </div>
    );    
};

export default Index;
