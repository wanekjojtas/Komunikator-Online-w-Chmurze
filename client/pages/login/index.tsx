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
    setMessage(""); // Clear previous messages

    if (!email || !password) {
      setMessage("Please enter both email and password.");
      return;
    }

    try {
      const res = await fetch(`${API_URL}/login`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email, password }),
      });

      const data = await res.json();

      if (res.ok) {
        localStorage.setItem("user_info", JSON.stringify({ id: data.id, username: data.username }));
        localStorage.setItem("jwt", data.token);

        setUser({ id: data.id, username: data.username });
        setAuthenticated(true);

        router.push("/");
      } else {
        setMessage(data.error || "Invalid email or password.");
      }
    } catch (err) {
      console.error("Login error:", err);
      setMessage("An unexpected error occurred. Please try again.");
    }
  };

  const handleSignUp = async (e: React.SyntheticEvent) => {
    e.preventDefault();
    setMessage(""); // Clear previous messages

    if (!username || !email || !password) {
      setMessage("All fields are required.");
      return;
    }

    if (!validatePassword(password)) {
      setMessage(
        "Password must be at least 8 characters long, include one capital letter, one number, and one special character."
      );
      return;
    }

    try {
      const res = await fetch(`${API_URL}/signup`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ username, email, password }),
      });

      const data = await res.json();

      if (res.ok) {
        setMessage("User created successfully! You can now log in.");
        setIsSignUp(false);
        setUsername("");
        setEmail("");
        setPassword("");
      } else {
        switch (data.error) {
          case "email_already_exists":
            setMessage("The email address is already in use.");
            break;
          case "username_already_exists":
            setMessage("The username is already taken.");
            break;
          case "invalid_email_format":
            setMessage("The email format is invalid.");
            break;
          case "password_too_short":
            setMessage("The password does not meet the required criteria.");
            break;
          default:
            setMessage(data.error || "Failed to create user.");
        }
      }
    } catch (err) {
      console.error("Sign-up error:", err);
      setMessage("An unexpected error occurred during sign-up.");
    }
  };

  const toggleMode = () => {
    setIsSignUp(!isSignUp);
    setUsername(""); // Clear fields
    setEmail("");
    setPassword("");
    setMessage(""); // Clear messages
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
            onClick={toggleMode}
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