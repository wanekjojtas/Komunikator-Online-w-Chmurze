import { useState, useContext, useEffect } from "react";
import { useRouter } from "next/router";
import { AuthContext } from "@/modules/auth_provider";
import { API_URL } from "../../constants";

const IndexPage = () => {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [username, setUsername] = useState("");
  const [isSignUp, setIsSignUp] = useState(false);
  const [message, setMessage] = useState("");

  const { authenticated, setAuthenticated, setUser } = useContext(AuthContext);
  const router = useRouter();

  useEffect(() => {
    if (authenticated) {
      router.push("/");
    }
  }, [authenticated]);

  const validatePassword = (password: string) => {
    const passwordRegex = /^(?=.*[A-Z])(?=.*\d)(?=.*[!@#$%^&*])[A-Za-z\d!@#$%^&*]{8,}$/;
    return passwordRegex.test(password);
  };
  

  const handleLogin = async (e: React.SyntheticEvent) => {
    e.preventDefault();

    try {
        const res = await fetch(`${API_URL}/login`, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ email, password }),
        });

        if (res.ok) {
            const data = await res.json();
            
            // Store user info and token in localStorage
            localStorage.setItem("user_info", JSON.stringify({ id: data.id, username: data.username }));
            localStorage.setItem("jwt", data.token);

            // Update AuthContext state
            setUser({ id: data.id, username: data.username });
            setAuthenticated(true);

            // Redirect to home page
            router.push("/");
        } else {
            const error = await res.json();
            setMessage(error.error || "Failed to log in");
        }
    } catch (err) {
        console.error("Login error:", err);
        setMessage("An error occurred during login.");
    }
  };


  const handleSignUp = async (e: React.SyntheticEvent) => {
    e.preventDefault();
  
    if (!validatePassword(password)) {
      setMessage("Password must be at least 8 characters long, include one capital letter, one number, and one special character.");
      return;
    }
  
    try {
      const res = await fetch(`${API_URL}/signup`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ username, email, password }),
      });
  
      if (res.ok) {
        setMessage("User created successfully! You can now log in.");
        setIsSignUp(false);
        setUsername("");
        setEmail("");
        setPassword("");
      } else {
        const error = await res.json();
        if (error.error === "Email already exists") {
          setMessage("The email address is already in use.");
        } else if (error.error === "Username already exists") {
          setMessage("The username is already taken.");
        } else {
          setMessage(error.error || "Failed to create user.");
        }
      }
    } catch (err) {
      console.error(err);
      setMessage("An error occurred during sign-up.");
    }
  };  

  return (
    <div className="flex items-center justify-center min-w-full min-h-screen">
      <form className="flex flex-col md:w-1/5">
        <div className="text-3xl font-bold text-center">
          <span className="text-blue-600">{isSignUp ? "Sign Up" : "Login"}</span>
        </div>
        {isSignUp && (
          <input
            type="text"
            placeholder="Username"
            className="p-3 mt-4 rounded-md border-2 border-grey-500 focus:outline-none focus:border-blue text-gray-500 bg-white"
            value={username}
            onChange={(e) => setUsername(e.target.value)}
          />
        )}
        <input
          type="email"
          placeholder="Email"
          className="p-3 mt-4 rounded-md border-2 border-grey-500 focus:outline-none focus:border-blue text-gray-500 bg-white"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
        />
        <input
          type="password"
          placeholder="Password"
          className="p-3 mt-4 rounded-md border-2 border-grey-500 focus:outline-none focus:border-blue text-gray-500 bg-white"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
        />
        <button
          className="p-3 mt-6 rounded-md bg-blue-500 font-bold text-white"
          type="submit"
          onClick={isSignUp ? handleSignUp : handleLogin}
        >
          {isSignUp ? "Sign Up" : "Login"}
        </button>
        <div className="mt-4 text-center">
          <span
            className="text-blue-500 cursor-pointer"
            onClick={() => setIsSignUp(!isSignUp)}
          >
            {isSignUp
              ? "Already have an account? Login"
              : "Don't have an account? Sign Up"}
          </span>
        </div>
        {message && <p className="mt-4 text-red-500 text-center">{message}</p>}
      </form>
    </div>
  );
};

export default IndexPage;
