import { useState, createContext, useEffect } from "react";
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
}>({
    authenticated: false,
    setAuthenticated: () => {},
    user: { username: "", id: "" },
    setUser: () => {},
});

const AuthContextProvier = ({ children }: { children: React.ReactNode }) => {
    const [authenticated, setAuthenticated] = useState(false);
    const [user, setUser] = useState<UserInfo>({ username: "", id: "" });

    const router = useRouter();

    const validateToken = async () => {
      const token = localStorage.getItem("jwt");
      if (!token) {
          console.warn("No token found. Redirecting to login.");
          localStorage.removeItem("user_info");
          router.push("/login");
          return;
      }
  
      try {
          const res = await fetch(`${API_URL}/validate-token`, {
              method: "GET",
              headers: {
                  Authorization: `Bearer ${token}`,
              },
          });
  
          if (!res.ok) {
              console.error("Token validation failed. Redirecting to login.");
              localStorage.removeItem("jwt");
              localStorage.removeItem("user_info");
              router.push("/login");
          }
      } catch (error) {
          console.error("Error validating token:", error);
          localStorage.removeItem("jwt");
          localStorage.removeItem("user_info");
          router.push("/login");
      }
  };
  

    useEffect(() => {
        try {
            const userInfo = localStorage.getItem("user_info");
            console.log("Loaded user_info:", userInfo); // Debug log

            if (!userInfo) {
                router.push("/login");
                return;
            }

            const parsedUser = JSON.parse(userInfo);
            if (parsedUser && parsedUser.id && parsedUser.username) {
                setUser(parsedUser);
                setAuthenticated(true);
                validateToken(); // Validate the token after setting user info
            } else {
                throw new Error("Invalid user info");
            }
        } catch (error) {
            console.error("Error initializing AuthContextProvider:", error);
            router.push("/login");
        }
    }, []);

    return (
        <AuthContext.Provider
            value={{
                authenticated,
                setAuthenticated,
                user,
                setUser,
            }}
        >
            {children}
        </AuthContext.Provider>
    );
};

export default AuthContextProvier;
