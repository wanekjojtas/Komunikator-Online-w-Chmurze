import React, { useContext, useState, useEffect } from "react";
import ChatList from "../chats/ChatList";
import ChatBody from "@/components/chat_body";
import { WebSocketContext } from "@/modules/websocket_provider";

const App = () => {
    const { currentChat, setCurrentChat, conn } = useContext(WebSocketContext);
    const [loading, setLoading] = useState(false);

    useEffect(() => {
        if (conn) {
            setLoading(false);
        }
    }, [conn]);

    const handleSelectChat = (chatID: string) => {
        setLoading(true);
        setCurrentChat(chatID);
    };

    return (
        <div style={{ display: "flex", flexDirection: "column", alignItems: "center", padding: "20px" }}>
            {loading ? (
                <div>Loading...</div>
            ) : !currentChat ? (
                <ChatList onSelect={handleSelectChat} />
            ) : (
                <div style={{ width: "100%", maxWidth: "600px" }}>
                    <button
                        onClick={() => setCurrentChat(null)}
                        style={{
                            marginBottom: "10px",
                            padding: "10px 20px",
                            backgroundColor: "#007BFF",
                            color: "#fff",
                            border: "none",
                            borderRadius: "5px",
                            cursor: "pointer",
                        }}
                    >
                        Back to Chat List
                    </button>
                    <ChatBody chatID={currentChat} />
                </div>
            )}
        </div>
    );
};

export default App;
