import React, { createContext, useCallback, useContext, useEffect, useMemo, useState } from "react";

type UserRole = "admin" | "user";

export type AuthUser = {
  id: number;
  username: string;
  display_name?: string | null;
  role: UserRole;
};

type AuthContextValue = {
  user: AuthUser | null;
  token: string | null;
  loading: boolean;
  login: (username: string, password: string) => Promise<void>;
  logout: () => void;
};

const AuthContext = createContext<AuthContextValue | null>(null);

const TOKEN_KEY = "shushu_auth_token";

const extractErrorMessage = (data: unknown, fallback: string) => {
  if (!data || typeof data !== "object") {
    return fallback;
  }
  const message = (data as { error?: string }).error;
  if (typeof message === "string" && message.trim() !== "") {
    return message;
  }
  return fallback;
};

export const AuthProvider: React.FC<React.PropsWithChildren> = ({ children }) => {
  const [token, setToken] = useState<string | null>(() => localStorage.getItem(TOKEN_KEY));
  const [user, setUser] = useState<AuthUser | null>(null);
  const [loading, setLoading] = useState<boolean>(true);

  const fetchMe = useCallback(async (currentToken: string) => {
    const response = await fetch("/api/auth/me", {
      headers: {
        Authorization: `Bearer ${currentToken}`
      }
    });
    if (!response.ok) {
      throw new Error("unauthorized");
    }
    const data = await response.json();
    return data.user as AuthUser;
  }, []);

  useEffect(() => {
    const init = async () => {
      if (!token) {
        setUser(null);
        setLoading(false);
        return;
      }
      try {
        const me = await fetchMe(token);
        setUser(me);
      } catch {
        localStorage.removeItem(TOKEN_KEY);
        setToken(null);
        setUser(null);
      } finally {
        setLoading(false);
      }
    };
    void init();
  }, [fetchMe, token]);

  const login = useCallback(async (username: string, password: string) => {
    setLoading(true);
    try {
      const response = await fetch("/api/auth/login", {
        method: "POST",
        headers: {
          "Content-Type": "application/json"
        },
        body: JSON.stringify({ username, password })
      });
      const data = await response.json();
      if (!response.ok) {
        throw new Error(extractErrorMessage(data, "登录失败"));
      }
      const nextToken = data.token as string;
      const nextUser = data.user as AuthUser;
      localStorage.setItem(TOKEN_KEY, nextToken);
      setToken(nextToken);
      setUser(nextUser);
    } catch (error) {
      if (error instanceof Error) {
        throw error;
      }
      throw new Error("登录失败");
    } finally {
      setLoading(false);
    }
  }, []);

  const logout = useCallback(() => {
    localStorage.removeItem(TOKEN_KEY);
    setToken(null);
    setUser(null);
  }, []);

  const value = useMemo(
    () => ({
      user,
      token,
      loading,
      login,
      logout
    }),
    [user, token, loading, login, logout]
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
};

export const useAuth = () => {
  const ctx = useContext(AuthContext);
  if (!ctx) {
    throw new Error("AuthContext not found");
  }
  return ctx;
};
